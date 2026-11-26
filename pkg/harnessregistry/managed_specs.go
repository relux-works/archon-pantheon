package harnessregistry

import (
	"github.com/relux-works/curator-runtimes/pkg/agentic/systems/agy"
	"github.com/relux-works/curator-runtimes/pkg/agentic/systems/kimi"
	"github.com/relux-works/curator-runtimes/pkg/harnessapi"
)

// ManagedHomeSpecs returns the pure adapter-owned managed-home specifications
// admitted beyond the frozen legacy parity set.
func ManagedHomeSpecs() []harnessapi.ManagedHomeSpec {
	return []harnessapi.ManagedHomeSpec{
		agy.ManagedHomeSpec(),
		kimi.ManagedHomeSpec(),
	}
}
