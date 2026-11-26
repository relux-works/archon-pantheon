package claude

import (
	"github.com/relux-works/curator-runtimes/pkg/harnessapi"
	"github.com/relux-works/curator-runtimes/pkg/harnessvalue"
)

// ClaudeCodeHarnessID is the canonical Curator harness identifier for Claude Code.
const ClaudeCodeHarnessID harnessapi.ID = "claude_code"

// HarnessAdapter implements the Claude Code surfaces expressible by harnessapi v1.
type HarnessAdapter struct{ harnessapi.UnportedOperations }

// NewHarnessAdapter constructs the adapter registered for Claude Code.
func NewHarnessAdapter() harnessapi.Adapter { return HarnessAdapter{} }

// ID returns the canonical harness identifier.
func (HarnessAdapter) ID() harnessapi.ID { return ClaudeCodeHarnessID }

// APIVersion returns the supported adapter contract version.
func (HarnessAdapter) APIVersion() harnessapi.APIVersion { return harnessapi.AdapterAPIVersionV1 }

// Capabilities lists only the v1 operations this adapter can describe.
func (HarnessAdapter) Capabilities() harnessapi.CapabilitySet {
	set, _ := harnessapi.NewCapabilitySet(
		harnessapi.InstructionOutput,
		harnessapi.SkillProjection,
		harnessapi.MCPInspection,
		harnessapi.NativeMCPWrite,
		harnessapi.CredentialMetadata,
		harnessapi.TrustObservation,
		harnessapi.ReleaseProbing,
	)
	return set
}

// ReleaseProbe returns Claude Code's pinned read-only version command.
func (HarnessAdapter) ReleaseProbe(harnessapi.Context) (harnessapi.ProbeSpec, error) {
	return harnessapi.NewProbeSpec("claude", []string{"--version"}, harnessapi.ProbeTimeoutV1, nil), nil
}

// Home returns the direct Claude configuration home and its default location.
func (HarnessAdapter) Home(harnessapi.Context) (harnessapi.HomeLayout, error) {
	roots, err := claudeNamedPaths(".claude")
	if err != nil {
		return harnessapi.HomeLayout{}, err
	}
	return harnessapi.NewHomeLayout("CLAUDE_CONFIG_DIR", roots, false)
}

// Instructions describes the shared Claude root-instruction file.
func (HarnessAdapter) Instructions(_ harnessapi.Context, form harnessapi.Form) (harnessapi.InstructionSpec, error) {
	paths, err := claudeNamedPaths("CLAUDE.md")
	if err != nil {
		return harnessapi.InstructionSpec{}, err
	}
	return harnessapi.NewInstructionSpec(paths, form, nil)
}

// Skills describes Claude's native user and project skill locations.
func (HarnessAdapter) Skills(_ harnessapi.Context, scope harnessapi.Scope) (harnessapi.SkillDiscovery, error) {
	var path string
	var ignores []string
	switch scope {
	case harnessapi.UserScope:
		path = "skills"
	case harnessapi.ProjectScope:
		path = ".claude/skills"
		ignores = []string{path}
	default:
		return harnessapi.SkillDiscovery{}, harnessapi.NewRefusal("harness_scope_unsupported")
	}
	paths, err := claudeNamedPaths(path)
	if err != nil {
		return harnessapi.SkillDiscovery{}, err
	}
	ignorePaths, err := claudeNamedPaths(ignores...)
	if err != nil {
		return harnessapi.SkillDiscovery{}, err
	}
	return harnessapi.NewSkillDiscovery(paths, ignorePaths, scope, false, true)
}

// Settings describes the native settings file for the requested scope.
func (HarnessAdapter) Settings(context harnessapi.Context) (harnessapi.SettingsSpec, error) {
	paths := []string{"settings.json"}
	if context.Scope() == harnessapi.ProjectScope {
		paths = []string{".claude/settings.json", ".claude/settings.local.json"}
	} else if context.Scope() != harnessapi.UserScope {
		return harnessapi.SettingsSpec{}, harnessapi.NewRefusal("harness_scope_unsupported")
	}
	named, err := claudeNamedPaths(paths...)
	if err != nil {
		return harnessapi.SettingsSpec{}, err
	}
	return harnessapi.NewSettingsSpec(named, harnessapi.JSON, nil)
}

// MCP describes the Claude native MCP files.
func (HarnessAdapter) MCP(_ harnessapi.Context, scope harnessapi.Scope) (harnessapi.MCPChannel, error) {
	var path string
	var write, trust bool
	switch scope {
	case harnessapi.UserScope:
		path = ".claude.json"
	case harnessapi.ProjectScope:
		path, write, trust = ".mcp.json", true, true
	default:
		return harnessapi.MCPChannel{}, harnessapi.NewRefusal("harness_scope_unsupported")
	}
	paths, err := claudeNamedPaths(path)
	if err != nil {
		return harnessapi.MCPChannel{}, err
	}
	return harnessapi.NewMCPChannel(paths, harnessapi.JSON, true, write, trust)
}

// Credentials describes the Linux file-link route.
func (HarnessAdapter) Credentials(context harnessapi.Context) (harnessapi.CredentialRoute, error) {
	switch context.Platform() {
	case "linux":
		paths, err := claudeNamedPaths(".credentials.json")
		if err != nil {
			return harnessapi.CredentialRoute{}, err
		}
		return harnessapi.NewCredentialRoute(paths, "file-link", true)
	case "darwin":
		return harnessapi.CredentialRoute{}, harnessapi.NewRefusal(harnessapi.OperationNotPorted)
	default:
		paths, err := claudeNamedPaths()
		if err != nil {
			return harnessapi.CredentialRoute{}, err
		}
		return harnessapi.NewCredentialRoute(paths, "unsupported", false)
	}
}

// Trust reports project admission as observation-only policy.
func (HarnessAdapter) Trust(context harnessapi.Context) (harnessapi.TrustPolicy, error) {
	return harnessapi.NewTrustPolicy(context.Scope() == harnessapi.ProjectScope), nil
}

// Preflight leaves generic trust, ownership, and transaction checks to shared owners.
func (HarnessAdapter) Preflight(harnessapi.Context, harnessapi.Snapshot, harnessapi.Request) ([]harnessapi.Refusal, error) {
	return []harnessapi.Refusal{}, nil
}

func claudeNamedPaths(values ...string) (harnessapi.NamedPaths, error) {
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
