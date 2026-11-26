package codex

import "github.com/relux-works/curator-runtimes/pkg/harnessapi"

// CodexCLIHarnessID is the canonical Curator harness identifier for Codex CLI.
const CodexCLIHarnessID harnessapi.ID = "codex_cli"

// HarnessAdapter exposes the staged harness surface for Codex CLI.
type HarnessAdapter struct{ harnessapi.UnportedOperations }

// NewHarnessAdapter constructs the staged harness adapter for Codex CLI.
func NewHarnessAdapter() harnessapi.Adapter { return HarnessAdapter{} }

// ID returns the canonical harness identifier.
func (HarnessAdapter) ID() harnessapi.ID { return CodexCLIHarnessID }

// APIVersion returns the implemented internal contract revision.
func (HarnessAdapter) APIVersion() harnessapi.APIVersion { return harnessapi.AdapterAPIVersionV1 }

// Capabilities returns only actually ported operations.
func (HarnessAdapter) Capabilities() harnessapi.CapabilitySet {
	set, _ := harnessapi.NewCapabilitySet(harnessapi.ReleaseProbing)
	return set
}

// ReleaseProbe returns the baseline release probe tuple for codex.
func (HarnessAdapter) ReleaseProbe(harnessapi.Context) (harnessapi.ProbeSpec, error) {
	return harnessapi.NewProbeSpec("codex", []string{"--version"}, harnessapi.ProbeTimeoutV1, nil), nil
}
