package agy

import "github.com/relux-works/archon-pantheon/pkg/harnessapi"

// AntigravityHarnessID is the canonical Curator harness identifier for Antigravity CLI.
const AntigravityHarnessID harnessapi.ID = "antigravity"

// HarnessAdapter exposes the staged harness surface for Antigravity CLI (CIP-0033).
type HarnessAdapter struct{ harnessapi.UnportedOperations }

// NewHarnessAdapter constructs the staged harness adapter for Antigravity CLI.
func NewHarnessAdapter() harnessapi.Adapter { return HarnessAdapter{} }

// ID returns the canonical harness identifier.
func (HarnessAdapter) ID() harnessapi.ID { return AntigravityHarnessID }

// APIVersion returns the implemented internal contract revision.
func (HarnessAdapter) APIVersion() harnessapi.APIVersion { return harnessapi.AdapterAPIVersionV1 }

// Capabilities returns only actually ported operations.
func (HarnessAdapter) Capabilities() harnessapi.CapabilitySet {
	set, _ := harnessapi.NewCapabilitySet(harnessapi.ReleaseProbing)
	return set
}

// ReleaseProbe returns the read-only release probe tuple for agy.
func (HarnessAdapter) ReleaseProbe(harnessapi.Context) (harnessapi.ProbeSpec, error) {
	return harnessapi.NewProbeSpec("agy", []string{"--version"}, harnessapi.ProbeTimeoutV1, nil), nil
}

// ManagedHomeSpec returns the pure CIP-0033 managed-home specification for
// Antigravity CLI (verified against agy 1.5.16 on macOS ARM64 and Linux x86_64).
func ManagedHomeSpec() harnessapi.ManagedHomeSpec {
	// #nosec G101 -- no credential material here: relative file paths and environment variable names only.
	return harnessapi.ManagedHomeSpec{
		ID:                    string(AntigravityHarnessID),
		Alias:                 "agy",
		HomeTransport:         "flag",
		ManagedOnly:           true,
		NativeDefaultDir:      ".gemini",
		NativeHomeIgnoresEnv:  true,
		ForbiddenHomeBasename: ".gemini",
		RootTarget:            "config/AGENTS.md",
		SkillsDir:             "config/skills",
		Forms:                 []string{"monolithic"},
		DefaultForm:           "monolithic",
		MCPFilename:           "config/mcp_config.json",
		MCPFormat:             "mcpServers-json",
		ForbiddenMCPEnvNames: []string{
			"GEMINI_API_KEY",
			"GOOGLE_API_KEY",
			"AGY_ADC_AUTH",
			"GOOGLE_APPLICATION_CREDENTIALS",
			"GOOGLE_CLOUD_QUOTA_PROJECT",
			"AGY_CLI_FORCE_FILE_TOKEN_STORAGE",
			"AGY_CLI_DISABLE_AUTO_UPDATE",
		},
		Passthrough: map[string][]harnessapi.PassthroughSpec{
			"default": {
				{
					Path:           "antigravity-cli/antigravity-oauth-token", // #nosec G101 -- relative path descriptor, not a secret.
					Strategy:       "file-link",
					FileLinkTarget: "antigravity-cli/antigravity-oauth-token", // #nosec G101 -- relative path descriptor, not a secret.
				},
			},
		},
		CredentialFileGuardPath:           "antigravity-cli/antigravity-oauth-token", // #nosec G101 -- relative path descriptor, not a secret.
		ForcedCredentialStore:             "file",
		ForcedCredentialEnvVar:            "AGY_CLI_FORCE_FILE_TOKEN_STORAGE", // #nosec G101 -- environment variable name, not a secret.
		ForcedCredentialEnvValue:          "true",
		RequireLiveNativeCredentialTarget: true,
		Shadows: []harnessapi.ShadowSpec{
			{Path: "config/GEMINI.md", Surface: "root-context"},
			{Path: "config/rules", Surface: "root-context"},
			{Path: "GEMINI.md", Surface: "root-context"},
			{Path: "AGENTS.md", Surface: "root-context"},
		},
		VerifiedRelease:      "1.5.16",
		Probe:                harnessapi.VersionProbeArgs("agy"),
		SizeAdvisoryBytes:    24000,
		CredentialScope:      "file-backed per --gemini_dir under AGY_CLI_FORCE_FILE_TOKEN_STORAGE=true (verified on agy 1.5.16 macOS ARM64 & Linux x86_64)",
		GlobalContextCap:     "24000 bytes per rule file (empirically verified on agy 1.5.16)",
		FirstRunSteps:        "Launch via curator-run or SAM with --gemini_dir; for isolated auth run AGY_CLI_FORCE_FILE_TOKEN_STORAGE=true agy acp --login --gemini-dir <home>; for shared auth log in once in the native ~/.gemini home with AGY_CLI_FORCE_FILE_TOKEN_STORAGE=true.",
		StatusNote:           "Antigravity CLI (agy 1.5.16): managed-home-only via flag home transport (--gemini_dir, launch-env-fragment-v7); root context at config/AGENTS.md (24000-byte per-file cap), skills at config/skills/ (non-.gemini basename skips ~/.agents/skills), MCP at config/mcp_config.json; file-backed OAuth enforced via AGY_CLI_FORCE_FILE_TOKEN_STORAGE=true; headless credential-injection/1 remains unqualified on agy 1.5.16 due to stdio MCP GEMINI_API_KEY inheritance (CIP-0033); project-context/Werkstatt certification deferred to Phase 2",
		NativeLoginHint:      "native credential target is absent; run AGY_CLI_FORCE_FILE_TOKEN_STORAGE=true agy acp --login in the native ~/.gemini home before resolving shared Antigravity",
		ForcedStoreGuidance:  "managed Antigravity requires file storage; select native or file (global macOS Keychain cannot isolate managed homes)",
		CredentialPolicyNote: "managed Antigravity maps configured native (including unset) to effective file storage via AGY_CLI_FORCE_FILE_TOKEN_STORAGE=true; the stored selector is unchanged",
	}
}
