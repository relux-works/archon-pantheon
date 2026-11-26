package muse

import "github.com/relux-works/curator-runtimes/pkg/harnessapi"

// MuseHarnessID is the canonical Curator harness identifier for Muse CLI.
const MuseHarnessID harnessapi.ID = "muse"

// HarnessAdapter exposes the staged harness surface for Muse CLI.
type HarnessAdapter struct{ harnessapi.UnportedOperations }

// NewHarnessAdapter constructs the staged harness adapter for Muse CLI.
func NewHarnessAdapter() harnessapi.Adapter { return HarnessAdapter{} }

// ID returns the canonical harness identifier.
func (HarnessAdapter) ID() harnessapi.ID { return MuseHarnessID }

// APIVersion returns the implemented internal contract revision.
func (HarnessAdapter) APIVersion() harnessapi.APIVersion { return harnessapi.AdapterAPIVersionV1 }

// Capabilities returns only actually ported operations.
func (HarnessAdapter) Capabilities() harnessapi.CapabilitySet {
	set, _ := harnessapi.NewCapabilitySet(harnessapi.ReleaseProbing)
	return set
}

// ReleaseProbe returns the baseline release probe tuple for muse.
func (HarnessAdapter) ReleaseProbe(harnessapi.Context) (harnessapi.ProbeSpec, error) {
	return harnessapi.NewProbeSpec("muse", []string{"--version"}, harnessapi.ProbeTimeoutV1, nil), nil
}
