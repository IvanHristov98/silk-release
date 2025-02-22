// Code generated by counterfeiter. DO NOT EDIT.
package fakes

import (
	"sync"

	"code.cloudfoundry.org/vxlan-policy-agent/converger"
	"code.cloudfoundry.org/vxlan-policy-agent/enforcer"
)

type Planner struct {
	GetRulesAndChainStub        func() (enforcer.RulesWithChain, error)
	getRulesAndChainMutex       sync.RWMutex
	getRulesAndChainArgsForCall []struct {
	}
	getRulesAndChainReturns struct {
		result1 enforcer.RulesWithChain
		result2 error
	}
	getRulesAndChainReturnsOnCall map[int]struct {
		result1 enforcer.RulesWithChain
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *Planner) GetRulesAndChain() (enforcer.RulesWithChain, error) {
	fake.getRulesAndChainMutex.Lock()
	ret, specificReturn := fake.getRulesAndChainReturnsOnCall[len(fake.getRulesAndChainArgsForCall)]
	fake.getRulesAndChainArgsForCall = append(fake.getRulesAndChainArgsForCall, struct {
	}{})
	stub := fake.GetRulesAndChainStub
	fakeReturns := fake.getRulesAndChainReturns
	fake.recordInvocation("GetRulesAndChain", []interface{}{})
	fake.getRulesAndChainMutex.Unlock()
	if stub != nil {
		return stub()
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *Planner) GetRulesAndChainCallCount() int {
	fake.getRulesAndChainMutex.RLock()
	defer fake.getRulesAndChainMutex.RUnlock()
	return len(fake.getRulesAndChainArgsForCall)
}

func (fake *Planner) GetRulesAndChainCalls(stub func() (enforcer.RulesWithChain, error)) {
	fake.getRulesAndChainMutex.Lock()
	defer fake.getRulesAndChainMutex.Unlock()
	fake.GetRulesAndChainStub = stub
}

func (fake *Planner) GetRulesAndChainReturns(result1 enforcer.RulesWithChain, result2 error) {
	fake.getRulesAndChainMutex.Lock()
	defer fake.getRulesAndChainMutex.Unlock()
	fake.GetRulesAndChainStub = nil
	fake.getRulesAndChainReturns = struct {
		result1 enforcer.RulesWithChain
		result2 error
	}{result1, result2}
}

func (fake *Planner) GetRulesAndChainReturnsOnCall(i int, result1 enforcer.RulesWithChain, result2 error) {
	fake.getRulesAndChainMutex.Lock()
	defer fake.getRulesAndChainMutex.Unlock()
	fake.GetRulesAndChainStub = nil
	if fake.getRulesAndChainReturnsOnCall == nil {
		fake.getRulesAndChainReturnsOnCall = make(map[int]struct {
			result1 enforcer.RulesWithChain
			result2 error
		})
	}
	fake.getRulesAndChainReturnsOnCall[i] = struct {
		result1 enforcer.RulesWithChain
		result2 error
	}{result1, result2}
}

func (fake *Planner) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.getRulesAndChainMutex.RLock()
	defer fake.getRulesAndChainMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *Planner) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ converger.Planner = new(Planner)
