// +build tools

package tools

import (
	_ "github.com/containernetworking/cni/plugins/test/noop"
)

// This file imports packages that are used when running go generate, or used
// during the development process but not otherwise depended on by built code.
