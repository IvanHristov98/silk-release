package converger_test

import (
	"errors"

	"code.cloudfoundry.org/lib/rules"
	"code.cloudfoundry.org/vxlan-policy-agent/converger"
	"code.cloudfoundry.org/vxlan-policy-agent/converger/fakes"
	"code.cloudfoundry.org/vxlan-policy-agent/enforcer"

	"code.cloudfoundry.org/lager/lagertest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("Single Poll Cycle", func() {
	Describe("Run", func() {
		var (
			p                    *converger.SinglePollCycle
			fakePolicyPlanner    *fakes.Planner
			fakeLocalPlanner     *fakes.Planner
			fakeRemotePlanner    *fakes.Planner
			fakeEnforcer         *fakes.RuleEnforcer
			metricsSender        *fakes.MetricsSender
			localRulesWithChain  enforcer.RulesWithChain
			remoteRulesWithChain enforcer.RulesWithChain
			policyRulesWithChain enforcer.RulesWithChain
			logger               *lagertest.TestLogger
			locker               *Locker
		)

		BeforeEach(func() {
			fakePolicyPlanner = &fakes.Planner{}
			fakeLocalPlanner = &fakes.Planner{}
			fakeRemotePlanner = &fakes.Planner{}
			fakeEnforcer = &fakes.RuleEnforcer{}
			metricsSender = &fakes.MetricsSender{}
			locker = &Locker{}
			logger = lagertest.NewTestLogger("test")

			p = &converger.SinglePollCycle{
				Planners:      []converger.Planner{fakeLocalPlanner, fakeRemotePlanner, fakePolicyPlanner},
				Enforcer:      fakeEnforcer,
				MetricsSender: metricsSender,
				Logger:        logger,
				Mutex:         locker,
			}

			localRulesWithChain = enforcer.RulesWithChain{
				Rules: []rules.IPTablesRule{[]string{"local-rule"}},
				Chain: enforcer.Chain{
					Table:       "local-table",
					ParentChain: "INPUT",
					Prefix:      "some-prefix",
				},
			}
			remoteRulesWithChain = enforcer.RulesWithChain{
				Rules: []rules.IPTablesRule{[]string{"remote-rule"}},
				Chain: enforcer.Chain{
					Table:       "remote-table",
					ParentChain: "INPUT",
					Prefix:      "some-prefix",
				},
			}
			policyRulesWithChain = enforcer.RulesWithChain{
				Rules: []rules.IPTablesRule{[]string{"policy-rule"}},
				Chain: enforcer.Chain{
					Table:       "policy-table",
					ParentChain: "INPUT",
					Prefix:      "some-prefix",
				},
			}

			fakeLocalPlanner.GetRulesAndChainReturns(localRulesWithChain, nil)
			fakeRemotePlanner.GetRulesAndChainReturns(remoteRulesWithChain, nil)
			fakePolicyPlanner.GetRulesAndChainReturns(policyRulesWithChain, nil)
		})

		It("enforces local,remote and policy rules on configured interval", func() {
			err := p.DoCycle()
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeLocalPlanner.GetRulesAndChainCallCount()).To(Equal(1))
			Expect(fakeRemotePlanner.GetRulesAndChainCallCount()).To(Equal(1))
			Expect(fakePolicyPlanner.GetRulesAndChainCallCount()).To(Equal(1))
			Expect(fakeEnforcer.EnforceRulesAndChainCallCount()).To(Equal(3))
			Expect(locker.LockCallCount).To(Equal(1))
			Expect(locker.UnlockCallCount).To(Equal(1))

			rws := fakeEnforcer.EnforceRulesAndChainArgsForCall(0)
			Expect(rws).To(Equal(localRulesWithChain))
			rws = fakeEnforcer.EnforceRulesAndChainArgsForCall(1)
			Expect(rws).To(Equal(remoteRulesWithChain))
			rws = fakeEnforcer.EnforceRulesAndChainArgsForCall(2)
			Expect(rws).To(Equal(policyRulesWithChain))
		})

		It("emits time metrics", func() {
			err := p.DoCycle()
			Expect(err).NotTo(HaveOccurred())
			Expect(metricsSender.SendDurationCallCount()).To(Equal(2))
			name, _ := metricsSender.SendDurationArgsForCall(0)
			Expect(name).To(Equal("iptablesEnforceTime"))
			name, _ = metricsSender.SendDurationArgsForCall(1)
			Expect(name).To(Equal("totalPollTime"))
		})

		Context("when a ruleset has not changed since the last poll cycle", func() {
			BeforeEach(func() {
				err := p.DoCycle()
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeEnforcer.EnforceRulesAndChainCallCount()).To(Equal(3))
			})

			It("does not re-write the ip tables rules", func() {
				err := p.DoCycle()
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeLocalPlanner.GetRulesAndChainCallCount()).To(Equal(2))
				Expect(fakeRemotePlanner.GetRulesAndChainCallCount()).To(Equal(2))
				Expect(fakePolicyPlanner.GetRulesAndChainCallCount()).To(Equal(2))

				Expect(fakeEnforcer.EnforceRulesAndChainCallCount()).To(Equal(3))
			})
		})

		Context("when a ruleset has changed since the last poll cycle", func() {
			BeforeEach(func() {
				err := p.DoCycle()
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeEnforcer.EnforceRulesAndChainCallCount()).To(Equal(3))
				localRulesWithChain.Rules = []rules.IPTablesRule{[]string{"new-rule"}}
				fakeLocalPlanner.GetRulesAndChainReturns(localRulesWithChain, nil)
			})

			It("re-writes the ip tables rules for that chain", func() {
				err := p.DoCycle()
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeLocalPlanner.GetRulesAndChainCallCount()).To(Equal(2))
				Expect(fakeRemotePlanner.GetRulesAndChainCallCount()).To(Equal(2))
				Expect(fakePolicyPlanner.GetRulesAndChainCallCount()).To(Equal(2))

				Expect(fakeEnforcer.EnforceRulesAndChainCallCount()).To(Equal(4))
			})

			It("logs a message about writing ip tables rules", func() {
				err := p.DoCycle()
				Expect(err).NotTo(HaveOccurred())
				Expect(logger).To(gbytes.Say("poll-cycle.*updating iptables rules.*new rules.*new-rule.*num new rules.*1.*num old rules.*1.*old rules.*local-rule"))
			})
		})

		Context("when a ruleset has all rules removed since the last poll cycle", func() {
			BeforeEach(func() {
				err := p.DoCycle()
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeEnforcer.EnforceRulesAndChainCallCount()).To(Equal(3))
				localRulesWithChain.Rules = []rules.IPTablesRule{}
				fakeLocalPlanner.GetRulesAndChainReturns(localRulesWithChain, nil)
			})

			It("re-writes the ip tables rules for that chain", func() {
				err := p.DoCycle()
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeLocalPlanner.GetRulesAndChainCallCount()).To(Equal(2))
				Expect(fakeRemotePlanner.GetRulesAndChainCallCount()).To(Equal(2))
				Expect(fakePolicyPlanner.GetRulesAndChainCallCount()).To(Equal(2))

				Expect(fakeEnforcer.EnforceRulesAndChainCallCount()).To(Equal(4))
			})
		})

		Context("when a new empty chain is created", func() {
			BeforeEach(func() {
				localRulesWithChain.Rules = []rules.IPTablesRule{}
				fakeLocalPlanner.GetRulesAndChainReturns(localRulesWithChain, nil)
			})

			It("enforces the rules for that chain", func() {
				err := p.DoCycle()
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeEnforcer.EnforceRulesAndChainCallCount()).To(Equal(3))
				Expect(metricsSender.SendDurationCallCount()).To(Equal(2))
			})
		})

		Context("when the local planner errors", func() {
			BeforeEach(func() {
				fakeLocalPlanner.GetRulesAndChainReturns(policyRulesWithChain, errors.New("eggplant"))
			})

			It("logs the error and returns", func() {
				err := p.DoCycle()
				Expect(err).To(MatchError("get-rules: eggplant"))

				Expect(fakeEnforcer.EnforceRulesAndChainCallCount()).To(Equal(0))
				Expect(metricsSender.SendDurationCallCount()).To(Equal(0))

				Expect(locker.LockCallCount).To(Equal(1))
				Expect(locker.UnlockCallCount).To(Equal(1))
			})
		})

		Context("when the remote planner errors", func() {
			BeforeEach(func() {
				fakeRemotePlanner.GetRulesAndChainReturns(policyRulesWithChain, errors.New("eggplant"))
			})

			It("logs the error and returns", func() {
				err := p.DoCycle()
				Expect(err).To(MatchError("get-rules: eggplant"))

				Expect(fakeEnforcer.EnforceRulesAndChainCallCount()).To(Equal(1))
				Expect(metricsSender.SendDurationCallCount()).To(Equal(0))

				Expect(locker.LockCallCount).To(Equal(1))
				Expect(locker.UnlockCallCount).To(Equal(1))
			})
		})

		Context("when the policy planner errors", func() {
			BeforeEach(func() {
				fakePolicyPlanner.GetRulesAndChainReturns(policyRulesWithChain, errors.New("eggplant"))
			})

			It("logs the error and returns", func() {
				err := p.DoCycle()
				Expect(err).To(MatchError("get-rules: eggplant"))

				Expect(fakeEnforcer.EnforceRulesAndChainCallCount()).To(Equal(2))
				Expect(metricsSender.SendDurationCallCount()).To(Equal(0))

				Expect(locker.LockCallCount).To(Equal(1))
				Expect(locker.UnlockCallCount).To(Equal(1))
			})
		})

		Context("when policy enforcer errors", func() {
			BeforeEach(func() {
				fakeEnforcer.EnforceRulesAndChainReturns(errors.New("eggplant"))
			})

			It("logs the error and returns", func() {
				err := p.DoCycle()
				Expect(err).To(MatchError("enforce: eggplant"))

				Expect(metricsSender.SendDurationCallCount()).To(Equal(0))

				Expect(locker.LockCallCount).To(Equal(1))
				Expect(locker.UnlockCallCount).To(Equal(1))
			})
		})
	})
})

type Locker struct {
	LockCallCount   int
	UnlockCallCount int
}

func (l *Locker) Lock() {
	l.LockCallCount++
}

func (l *Locker) Unlock() {
	l.UnlockCallCount++
}
