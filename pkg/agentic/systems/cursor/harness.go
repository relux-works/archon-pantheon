// Package cursor owns the registered skill and MCP inspection facts for Cursor.
package cursor

import (
	"github.com/relux-works/archon-pantheon/pkg/harnessapi"
	"github.com/relux-works/archon-pantheon/pkg/harnessvalue"
)

const (
	skillPath = ".cursor/rules"
	mcpPath   = ".cursor/mcp.json"
)

// CursorHarnessID is this module's stable canonical identifier.
const CursorHarnessID harnessapi.ID = "cursor"

// HarnessAdapter exposes the staged harness surface with explicit operational refusals.
type HarnessAdapter struct{ harnessapi.UnportedOperations }

// NewHarnessAdapter constructs the module's staged adapter.
func NewHarnessAdapter() harnessapi.Adapter { return HarnessAdapter{} }

// ID returns the module's canonical identifier.
func (HarnessAdapter) ID() harnessapi.ID { return CursorHarnessID }

// APIVersion returns the implemented internal contract revision.
func (HarnessAdapter) APIVersion() harnessapi.APIVersion { return harnessapi.AdapterAPIVersionV1 }

// Capabilities returns only the operations implemented by this partial adapter.
func (HarnessAdapter) Capabilities() harnessapi.CapabilitySet {
	capabilities, _ := harnessapi.NewCapabilitySet(harnessapi.SkillProjection, harnessapi.MCPInspection)
	return capabilities
}

// Skills describes the native discovery and projection paths for a scope.
func (HarnessAdapter) Skills(_ harnessapi.Context, scope harnessapi.Scope) (harnessapi.SkillDiscovery, error) {
	if scope != harnessapi.UserScope && scope != harnessapi.ProjectScope {
		return harnessapi.SkillDiscovery{}, harnessapi.NewRefusal("harness_scope_unsupported")
	}
	paths, err := namedPaths(skillPath)
	if err != nil {
		return harnessapi.SkillDiscovery{}, err
	}
	ignores := harnessapi.NamedPaths{}
	if scope == harnessapi.ProjectScope {
		ignores = paths
	}
	return harnessapi.NewSkillDiscovery(paths, ignores, scope, false, true)
}

// MCP describes read-only inspection of the native JSON configuration.
func (HarnessAdapter) MCP(_ harnessapi.Context, scope harnessapi.Scope) (harnessapi.MCPChannel, error) {
	if scope != harnessapi.UserScope && scope != harnessapi.ProjectScope {
		return harnessapi.MCPChannel{}, harnessapi.NewRefusal("harness_scope_unsupported")
	}
	paths, err := namedPaths(mcpPath)
	if err != nil {
		return harnessapi.MCPChannel{}, err
	}
	return harnessapi.NewMCPChannel(paths, harnessapi.JSON, true, false, scope == harnessapi.ProjectScope)
}

// ReleaseProbe returns the real baseline tuple or an unsupported-capability refusal.
func (HarnessAdapter) ReleaseProbe(harnessapi.Context) (harnessapi.ProbeSpec, error) {
	return harnessapi.ProbeSpec{}, harnessapi.NewRefusal(harnessapi.CapabilityUnsupported)
}

func namedPaths(values ...string) (harnessapi.NamedPaths, error) {
	paths := make([]harnessvalue.RelativePath, 0, len(values))
	for _, value := range values {
		path, err := harnessvalue.NewRelativePath(value)
		if err != nil {
			return harnessapi.NamedPaths{}, err
		}
		paths = append(paths, path)
	}
	return harnessapi.NewNamedPaths(paths)
}
