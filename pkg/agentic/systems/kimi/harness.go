package kimi

import "github.com/relux-works/archon-pantheon/pkg/harnessapi"

// KimiCodeHarnessID is the canonical Curator harness identifier for Kimi Code CLI.
const KimiCodeHarnessID harnessapi.ID = "kimi_code"

// HarnessAdapter exposes the staged harness surface for Kimi Code CLI.
type HarnessAdapter struct{ harnessapi.UnportedOperations }

// NewHarnessAdapter constructs the staged harness adapter for Kimi Code CLI.
func NewHarnessAdapter() harnessapi.Adapter { return HarnessAdapter{} }

// ID returns the canonical harness identifier.
func (HarnessAdapter) ID() harnessapi.ID { return KimiCodeHarnessID }

// APIVersion returns the implemented internal contract revision.
func (HarnessAdapter) APIVersion() harnessapi.APIVersion { return harnessapi.AdapterAPIVersionV1 }

// Capabilities returns only actually ported operations.
func (HarnessAdapter) Capabilities() harnessapi.CapabilitySet {
	set, _ := harnessapi.NewCapabilitySet(harnessapi.ReleaseProbing)
	return set
}

// ReleaseProbe returns the read-only release probe tuple for kimi.
func (HarnessAdapter) ReleaseProbe(harnessapi.Context) (harnessapi.ProbeSpec, error) {
	return harnessapi.NewProbeSpec("kimi", []string{"--version"}, harnessapi.ProbeTimeoutV1, nil), nil
}

// ManagedHomeSpec returns the pure managed-home specification for Kimi Code CLI
// (verified against kimi 2.8.1 on macOS ARM64).
func ManagedHomeSpec() harnessapi.ManagedHomeSpec {
	// #nosec G101 -- no credential material here: relative file paths and env var names only.
	return harnessapi.ManagedHomeSpec{
		ID:               string(KimiCodeHarnessID),
		Alias:            "kimi",
		HomeTransport:    "env",
		EnvVar:           "KIMI_CODE_HOME",
		ManagedOnly:      true,
		NativeDefaultDir: ".kimi-code",
		RootTarget:       "AGENTS.md",
		SkillsDir:        "skills",
		Forms:            []string{"monolithic"},
		DefaultForm:      "monolithic",
		MCPFilename:      "mcp.json",
		MCPFormat:        "mcpServers-json",
		ForbiddenMCPEnvNames: []string{
			"KIMI_API_KEY",
			"MOONSHOT_API_KEY",
			"KIMI_CODE_HOME",
		},
		Passthrough: map[string][]harnessapi.PassthroughSpec{
			"default": {
				{
					Path:           "credentials", // #nosec G101 -- relative path descriptor, not a secret.
					Strategy:       "file-link",
					FileLinkTarget: "credentials", // #nosec G101 -- relative path descriptor, not a secret.
				},
				{
					Path:           "config.toml",
					Strategy:       "file-link",
					FileLinkTarget: "config.toml",
				},
				{
					Path:           "region",
					Strategy:       "file-link",
					FileLinkTarget: "region",
				},
				{
					Path:           "device_id",
					Strategy:       "file-link",
					FileLinkTarget: "device_id",
				},
			},
		},
		CredentialFileGuardPath: "credentials/kimi-code.json", // #nosec G101 -- relative path descriptor, not a secret.
		VerifiedRelease:         "2.8.1",
		Probe:                   harnessapi.VersionProbeArgs("kimi"),
		SizeAdvisoryBytes:       32768,
		GlobalContextCap:        "32768 bytes recommended for AGENTS.md (empirically observed on kimi 2.8.1)",
		CredentialScope:         "file-backed per KIMI_CODE_HOME (credentials/kimi-code.json + config.toml, verified on kimi 2.8.1)",
		FirstRunSteps:           "Launch via curator run kimi; for shared auth run `kimi login` once in native ~/.kimi-code; for isolated auth run `KIMI_CODE_HOME=<home> kimi login`.",
		StatusNote:              "Kimi Code CLI (kimi 2.8.1): managed-home via KIMI_CODE_HOME; root context at AGENTS.md, skills at skills/, MCP at mcp.json; OAuth credentials in credentials/kimi-code.json and provider config in config.toml",
		NativeLoginHint:         "run `kimi login` in the native ~/.kimi-code home before resolving shared Kimi Code",
	}
}
