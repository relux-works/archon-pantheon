package gemini

import (
	"github.com/relux-works/archon-pantheon/pkg/harnessapi"
	"github.com/relux-works/archon-pantheon/pkg/harnessvalue"
)

// GeminiHarnessID is the canonical Curator harness identifier for Gemini CLI.
const GeminiHarnessID harnessapi.ID = "gemini"

// HarnessAdapter exposes the staged harness surface for Gemini CLI.
type HarnessAdapter struct{ harnessapi.UnportedOperations }

// NewHarnessAdapter constructs the staged harness adapter for Gemini CLI.
func NewHarnessAdapter() harnessapi.Adapter { return HarnessAdapter{} }

// ID returns the canonical harness identifier.
func (HarnessAdapter) ID() harnessapi.ID { return GeminiHarnessID }

// APIVersion returns the implemented internal contract revision.
func (HarnessAdapter) APIVersion() harnessapi.APIVersion { return harnessapi.AdapterAPIVersionV1 }

// Capabilities returns only the skill projection and MCP inspection operations.
func (HarnessAdapter) Capabilities() harnessapi.CapabilitySet { return HarnessCapabilities() }

// HarnessCapabilities returns the operations registered for this harness adapter.
func HarnessCapabilities() harnessapi.CapabilitySet {
	set, _ := harnessapi.NewCapabilitySet(harnessapi.SkillProjection, harnessapi.MCPInspection)
	return set
}

// ReleaseProbe returns an unsupported-capability refusal for Gemini CLI.
func (HarnessAdapter) ReleaseProbe(harnessapi.Context) (harnessapi.ProbeSpec, error) {
	return harnessapi.ProbeSpec{}, harnessapi.NewRefusal(harnessapi.CapabilityUnsupported)
}

// Skills describes the native destination and project ignore requirement.
func (HarnessAdapter) Skills(context harnessapi.Context, scope harnessapi.Scope) (harnessapi.SkillDiscovery, error) {
	if !geminiValidScope(context, scope) {
		return harnessapi.SkillDiscovery{}, harnessapi.NewRefusal("harness_scope_unsupported")
	}
	paths, err := geminiNamedPath(".gemini/skills")
	if err != nil {
		return harnessapi.SkillDiscovery{}, err
	}
	ignores := harnessapi.NamedPaths{}
	if scope == harnessapi.ProjectScope {
		ignores = paths
	}
	return harnessapi.NewSkillDiscovery(paths, ignores, scope, false, true)
}

// MCP describes the read-only JSON inspection surfaces and project trust gate.
func (HarnessAdapter) MCP(context harnessapi.Context, scope harnessapi.Scope) (harnessapi.MCPChannel, error) {
	if !geminiValidScope(context, scope) {
		return harnessapi.MCPChannel{}, harnessapi.NewRefusal("harness_scope_unsupported")
	}
	paths, err := geminiNamedPath(".gemini/settings.json")
	if err != nil {
		return harnessapi.MCPChannel{}, err
	}
	return harnessapi.NewMCPChannel(paths, harnessapi.JSON, true, false, scope == harnessapi.ProjectScope)
}

func geminiValidScope(context harnessapi.Context, scope harnessapi.Scope) bool {
	return (scope == harnessapi.UserScope || scope == harnessapi.ProjectScope) && context.Scope() == scope
}

func geminiNamedPath(value string) (harnessapi.NamedPaths, error) {
	path, err := harnessvalue.NewRelativePath(value)
	if err != nil {
		return harnessapi.NamedPaths{}, err
	}
	return harnessapi.NewNamedPaths([]harnessvalue.RelativePath{path})
}
