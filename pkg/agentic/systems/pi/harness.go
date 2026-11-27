package pi

import "github.com/relux-works/archon-pantheon/pkg/harnessapi"

// PiHarnessID is the canonical Curator harness identifier for Pi CLI.
const PiHarnessID harnessapi.ID = "pi"

// HarnessAdapter exposes the staged harness surface for Pi CLI.
type HarnessAdapter struct{ harnessapi.UnportedOperations }

// NewHarnessAdapter constructs the staged harness adapter for Pi CLI.
func NewHarnessAdapter() harnessapi.Adapter { return HarnessAdapter{} }

// ID returns the canonical harness identifier.
func (HarnessAdapter) ID() harnessapi.ID { return PiHarnessID }

// APIVersion returns the implemented internal contract revision.
func (HarnessAdapter) APIVersion() harnessapi.APIVersion { return harnessapi.AdapterAPIVersionV1 }

// Capabilities returns only actually ported operations.
func (HarnessAdapter) Capabilities() harnessapi.CapabilitySet {
	set, _ := harnessapi.NewCapabilitySet(harnessapi.ReleaseProbing)
	return set
}

// ReleaseProbe returns the baseline release probe tuple for pi.
func (HarnessAdapter) ReleaseProbe(harnessapi.Context) (harnessapi.ProbeSpec, error) {
	return harnessapi.NewProbeSpec("pi", []string{"--version"}, harnessapi.ProbeTimeoutV1, nil), nil
}
