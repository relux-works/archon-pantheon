// Package opencode declares the staged harness surface for OpenCode CLI.
package opencode

import "github.com/relux-works/curator-runtimes/pkg/harnessapi"

// OpenCodeHarnessID is this module's stable canonical identifier.
const OpenCodeHarnessID harnessapi.ID = "opencode"

// HarnessAdapter exposes the staged harness surface with explicit operational refusals.
type HarnessAdapter struct{ harnessapi.UnportedOperations }

// NewHarnessAdapter constructs the module's staged adapter.
func NewHarnessAdapter() harnessapi.Adapter { return HarnessAdapter{} }

// ID returns the module's canonical identifier.
func (HarnessAdapter) ID() harnessapi.ID { return OpenCodeHarnessID }

// APIVersion returns the implemented internal contract revision.
func (HarnessAdapter) APIVersion() harnessapi.APIVersion { return harnessapi.AdapterAPIVersionV1 }

// Capabilities returns only actually ported operations.
func (HarnessAdapter) Capabilities() harnessapi.CapabilitySet {
	set, _ := harnessapi.NewCapabilitySet(harnessapi.ReleaseProbing)
	return set
}

// ReleaseProbe returns the real baseline tuple or an unsupported-capability refusal.
func (HarnessAdapter) ReleaseProbe(harnessapi.Context) (harnessapi.ProbeSpec, error) {
	return harnessapi.NewProbeSpec("opencode", []string{"--version"}, harnessapi.ProbeTimeoutV1, nil), nil
}
