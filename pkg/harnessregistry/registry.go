// Package harnessregistry owns the closed, compile-time harness catalog.
// Native operations are staged until their individual migrations are qualified.
package harnessregistry

import (
	"github.com/relux-works/archon-pantheon/pkg/agentic"
	"github.com/relux-works/archon-pantheon/pkg/agentic/systems/claude"
	"github.com/relux-works/archon-pantheon/pkg/agentic/systems/codex"
	"github.com/relux-works/archon-pantheon/pkg/agentic/systems/cursor"
	"github.com/relux-works/archon-pantheon/pkg/agentic/systems/gemini"
	"github.com/relux-works/archon-pantheon/pkg/agentic/systems/muse"
	"github.com/relux-works/archon-pantheon/pkg/agentic/systems/opencode"
	"github.com/relux-works/archon-pantheon/pkg/agentic/systems/pi"
	"github.com/relux-works/archon-pantheon/pkg/agentic/systems/windsurf"
	"github.com/relux-works/archon-pantheon/pkg/harnessapi"
)

// DefaultID retains the existing default canonical identifier.
const DefaultID harnessapi.ID = codex.CodexCLIHarnessID

type registration struct {
	id           harnessapi.ID
	version      harnessapi.APIVersion
	capabilities harnessapi.CapabilitySet
	constructor  func() harnessapi.Adapter
	probe        harnessapi.ProbeSpec
	hasProbe     bool
}

func probing() harnessapi.CapabilitySet {
	set, _ := harnessapi.NewCapabilitySet(harnessapi.ReleaseProbing)
	return set
}

func skillAndMCPInspection() harnessapi.CapabilitySet {
	set, _ := harnessapi.NewCapabilitySet(harnessapi.SkillProjection, harnessapi.MCPInspection)
	return set
}

func nativeSkillsAndMCPInspection() harnessapi.CapabilitySet {
	set, _ := harnessapi.NewCapabilitySet(harnessapi.NativeSkillDiscovery, harnessapi.MCPInspection)
	return set
}

var entries = [...]registration{
	{id: claude.ClaudeCodeHarnessID, version: 1, capabilities: claude.NewHarnessAdapter().Capabilities(), constructor: claude.NewHarnessAdapter, probe: harnessapi.NewProbeSpec("claude", []string{"--version"}, harnessapi.ProbeTimeoutV1, nil), hasProbe: true},
	{id: codex.CodexCLIHarnessID, version: 1, capabilities: probing(), constructor: codex.NewHarnessAdapter, probe: harnessapi.NewProbeSpec("codex", []string{"--version"}, harnessapi.ProbeTimeoutV1, nil), hasProbe: true},
	{id: cursor.CursorHarnessID, version: 1, capabilities: skillAndMCPInspection(), constructor: cursor.NewHarnessAdapter, probe: harnessapi.ProbeSpec{}, hasProbe: false},
	{id: gemini.GeminiHarnessID, version: 1, capabilities: gemini.HarnessCapabilities(), constructor: gemini.NewHarnessAdapter, probe: harnessapi.ProbeSpec{}, hasProbe: false},
	{id: muse.MuseHarnessID, version: 1, capabilities: probing(), constructor: muse.NewHarnessAdapter, probe: harnessapi.NewProbeSpec("muse", []string{"--version"}, harnessapi.ProbeTimeoutV1, nil), hasProbe: true},
	{id: opencode.OpenCodeHarnessID, version: 1, capabilities: probing(), constructor: opencode.NewHarnessAdapter, probe: harnessapi.NewProbeSpec("opencode", []string{"--version"}, harnessapi.ProbeTimeoutV1, nil), hasProbe: true},
	{id: pi.PiHarnessID, version: 1, capabilities: probing(), constructor: pi.NewHarnessAdapter, probe: harnessapi.NewProbeSpec("pi", []string{"--version"}, harnessapi.ProbeTimeoutV1, nil), hasProbe: true},
	{id: windsurf.WindsurfHarnessID, version: 1, capabilities: nativeSkillsAndMCPInspection(), constructor: windsurf.NewHarnessAdapter, probe: harnessapi.ProbeSpec{}, hasProbe: false},
}

// CanonicalID normalizes the existing CLI aliases without choosing a fallback.
func CanonicalID(id harnessapi.ID) harnessapi.ID {
	return harnessapi.ID(agentic.CanonicalCuratorHarnessID(string(id)))
}

// ResolveSystemID maps a Curator harness or runtime alias ("agy", "antigravity",
// "claude", "claude_code", "codex", "codex_cli", "kimi", "kimi_code", "gemini",
// "muse", etc.) to its canonical agentic.SystemID.
func ResolveSystemID(id harnessapi.ID) (agentic.SystemID, error) {
	return agentic.NormalizeSystemID(agentic.CanonicalSystemID(string(id)))
}

// IDs returns an independent list of canonical registrations in static order.
func IDs() []harnessapi.ID {
	ids := make([]harnessapi.ID, len(entries))
	for i, entry := range entries {
		ids[i] = entry.id
	}
	return ids
}

// IDsWithCapability returns the static registrations that declare a capability.
func IDsWithCapability(capability harnessapi.Capability) []harnessapi.ID {
	var ids []harnessapi.ID
	for _, entry := range entries {
		if entry.capabilities.Has(capability) {
			ids = append(ids, entry.id)
		}
	}
	return ids
}

// Lookup returns only a checked static registration or a typed refusal.
func Lookup(id harnessapi.ID) (harnessapi.Adapter, error) {
	id = CanonicalID(id)
	for _, entry := range entries {
		if entry.id == id {
			return Verify(entry.version, entry.id, entry.capabilities, entry.constructor)
		}
	}
	return nil, harnessapi.NewRefusal(harnessapi.UnknownID)
}

// Verify checks registration metadata in refusal order without adding an entry
// to the catalog.
func Verify(version harnessapi.APIVersion, id harnessapi.ID, capabilities harnessapi.CapabilitySet, constructor func() harnessapi.Adapter) (harnessapi.Adapter, error) {
	if version != harnessapi.AdapterAPIVersionV1 {
		return nil, harnessapi.NewRefusal(harnessapi.APIVersionUnsupported)
	}
	if constructor == nil {
		return nil, harnessapi.NewRefusal(harnessapi.RegistryMismatch)
	}
	adapter := constructor()
	if adapter == nil {
		return nil, harnessapi.NewRefusal(harnessapi.RegistryMismatch)
	}
	if adapter.APIVersion() != version {
		return nil, harnessapi.NewRefusal(harnessapi.APIVersionUnsupported)
	}
	if adapter.ID() != id || adapter.Capabilities() != capabilities {
		return nil, harnessapi.NewRefusal(harnessapi.RegistryMismatch)
	}
	return adapter, nil
}

// ProbePolicy returns a defensive copy of the canonical ID's closed probe tuple.
func ProbePolicy(id harnessapi.ID) (harnessapi.ProbeSpec, error) {
	id = CanonicalID(id)
	for _, entry := range entries {
		if entry.id == id {
			if !entry.hasProbe {
				return harnessapi.ProbeSpec{}, harnessapi.NewRefusal(harnessapi.CapabilityUnsupported)
			}
			return harnessapi.NewProbeSpec(entry.probe.ExecutableAlias(), entry.probe.Arguments(), entry.probe.Timeout(), entry.probe.Environment()), nil
		}
	}
	return harnessapi.ProbeSpec{}, harnessapi.NewRefusal(harnessapi.UnknownID)
}
