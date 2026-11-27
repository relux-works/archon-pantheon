package agentic_test

import (
	"errors"
	"path"
	"slices"
	"strings"
	"testing"

	"github.com/relux-works/archon-pantheon/pkg/agentic"
	"github.com/relux-works/archon-pantheon/pkg/agentic/systems/agy"
	"github.com/relux-works/archon-pantheon/pkg/agentic/systems/claude"
)

func antigravityTestRegistry(t *testing.T) *agentic.Registry {
	t.Helper()
	reg := agentic.NewRegistry()
	if err := reg.Register(agy.NewWithRuntime(agy.Runtime{Executable: "/opt/agy/bin/agy", Version: "1.5.16"})); err != nil {
		t.Fatalf("register agy: %v", err)
	}
	if err := reg.Register(claude.New()); err != nil {
		t.Fatalf("register claude: %v", err)
	}
	return reg
}

func validAntigravityV7Context(home string) *agentic.CuratorContext {
	return &agentic.CuratorContext{
		Revision:    agentic.CuratorLaunchFragmentV7,
		Environment: "antigravity",
		Profile: agentic.CuratorProfilePin{
			Name:       "team-default",
			LockSHA256: strings.Repeat("a", 64),
		},
		Precedence: agentic.CuratorPrecedence{
			Winner:    "higher-weight",
			Placement: "winner-last",
		},
		Home: &agentic.CuratorHomeBinding{
			Path:      home,
			Transport: agentic.CuratorHomeTransportFlag,
		},
		Env: map[string]string{
			agentic.CuratorAntigravityForceFileTokenVariable: agentic.CuratorAntigravityForceFileTokenTrue,
		},
		Permissions: &agentic.CuratorPermissions{
			Mode:   agentic.CuratorPermissionNative,
			Locked: false,
			Source: agentic.CuratorPermissionSourceDefault,
		},
		CredentialStore: &agentic.CuratorCredentialStore{
			Selection: agentic.CuratorCredentialStoreFile,
		},
	}
}

func antigravityExecRequest(home string, ctx *agentic.CuratorContext) agentic.LaunchRequest {
	return agentic.LaunchRequest{
		System:  "antigravity",
		Model:   agentic.Model{ID: "gemini-4-carbon-low", Effort: agentic.EffortSupportNone},
		WorkDir: "/work/repo",
		Home:    home,
		Prompt:  []byte("hello"),
		Context: ctx,
		Env: []string{
			"HOME=/Users/operator",
			"PATH=/usr/bin:/bin",
			"AGY_CLI_FORCE_FILE_TOKEN_STORAGE=false",
			"GEMINI_API_KEY=leaked-ambient-key",
			"GITHUB_TOKEN=allowed-mcp-token",
		},
	}
}

func TestBuildPlanAntigravityCuratorV7Succeeds(t *testing.T) {
	reg := antigravityTestRegistry(t)
	home := "/Users/operator/.curator/environments/team-default/antigravity"
	ctx := validAntigravityV7Context(home)
	ctx.MCP = &agentic.CuratorMCPContext{
		Path:     path.Join(home, agentic.CuratorAntigravityMCPFile),
		EnvNames: []string{"GITHUB_TOKEN"},
		Channels: []agentic.CuratorChannelDescriptor{
			{Kind: agentic.CuratorDescriptorFile, Filename: agentic.CuratorAntigravityMCPFile},
		},
	}

	// Exec mode: --gemini_dir <home>, filtered env, AGY_CLI_FORCE_FILE_TOKEN_STORAGE=true, AGY_CLI_DISABLE_AUTO_UPDATE=true
	plan, err := agentic.BuildPlan(reg, antigravityExecRequest(home, ctx), agentic.LaunchModeExec)
	if err != nil {
		t.Fatalf("BuildPlan(exec) err = %v", err)
	}
	if len(plan.Argv) < 2 || plan.Argv[0] != "--gemini_dir" || plan.Argv[1] != home {
		t.Fatalf("exec argv = %v, want leading --gemini_dir %s", plan.Argv, home)
	}
	if !slices.Contains(plan.Env, "AGY_CLI_FORCE_FILE_TOKEN_STORAGE=true") {
		t.Fatalf("env missing AGY_CLI_FORCE_FILE_TOKEN_STORAGE=true: %v", plan.Env)
	}
	if !slices.Contains(plan.Env, "AGY_CLI_DISABLE_AUTO_UPDATE=true") {
		t.Fatalf("env missing AGY_CLI_DISABLE_AUTO_UPDATE=true: %v", plan.Env)
	}
	if !slices.Contains(plan.Env, "GITHUB_TOKEN=allowed-mcp-token") {
		t.Fatalf("env missing allowed MCP var GITHUB_TOKEN: %v", plan.Env)
	}
	for _, entry := range plan.Env {
		if strings.HasPrefix(entry, "GEMINI_API_KEY=") || entry == "AGY_CLI_FORCE_FILE_TOKEN_STORAGE=false" {
			t.Fatalf("env leaked forbidden ambient var %q", entry)
		}
	}
	prov, ok := plan.CuratorContextProvenanceSnapshot()
	if !ok || prov.ManagedHome != home || prov.ManagedHomeVariable != "" || prov.Fragment.Home == nil || prov.Fragment.Home.Path != home {
		t.Fatalf("unexpected provenance: ok=%v prov=%+v", ok, prov)
	}

	// Interactive mode (yolo)
	ireq := agentic.LaunchRequest{
		System:         "antigravity",
		Model:          agentic.Model{ID: "gemini-4-carbon-low", Effort: agentic.EffortSupportNone},
		WorkDir:        "/work/repo",
		Home:           home,
		PermissionMode: agentic.PermissionModeYolo,
		ToolRelease:    "1.5.16",
		NativeArgs:     []string{"--verbose"},
		Context:        ctx,
		Env:            []string{"HOME=/Users/operator", "PATH=/usr/bin"},
	}
	iplan, err := agentic.BuildPlan(reg, ireq, agentic.LaunchModeInteractive)
	if err != nil {
		t.Fatalf("BuildPlan(interactive) err = %v", err)
	}
	wantPrefix := []string{"--gemini_dir", home, "--model", "gemini-4-carbon-low", "--dangerously-skip-permissions", "--verbose"}
	if !slices.Equal(iplan.Argv, wantPrefix) {
		t.Fatalf("interactive argv = %v, want %v", iplan.Argv, wantPrefix)
	}
}

func TestBuildPlanCuratorContextRefusesInvalidFragmentV7(t *testing.T) {
	reg := antigravityTestRegistry(t)
	home := "/Users/operator/.curator/environments/team-default/antigravity"

	// DecodeCuratorFragment negative tests covering CuratorHomeBinding.UnmarshalJSON
	badJSONs := []string{
		`{"fragment":"launch-env-fragment-v7","environment":"antigravity","home":"not-an-object"}`,
		`{"fragment":"launch-env-fragment-v7","environment":"antigravity","home":{"path":"/a/antigravity","transport":"flag","extra":1}}`,
		`{"fragment":"launch-env-fragment-v7","environment":"antigravity","home":{"path":null,"transport":"flag"}}`,
		`{"fragment":"launch-env-fragment-v7","environment":"antigravity","home":{"transport":"flag"}}`,
		`{"fragment":"launch-env-fragment-v7","environment":"antigravity","home":{"path":"/a/antigravity"}}`,
	}
	for i, raw := range badJSONs {
		if _, err := agentic.DecodeCuratorFragment([]byte(raw)); !errors.Is(err, agentic.ErrCuratorContextMalformed) {
			t.Fatalf("badJSON[%d] err = %v, want ErrCuratorContextMalformed", i, err)
		}
	}

	cases := []struct {
		name    string
		mutate  func(req *agentic.LaunchRequest)
		mode    agentic.LaunchMode
		wantErr error
	}{
		{
			name: "v7-missing-permissions",
			mutate: func(req *agentic.LaunchRequest) {
				req.Context.Permissions = nil
			},
			mode:    agentic.LaunchModeExec,
			wantErr: agentic.ErrCuratorContextMalformed,
		},
		{
			name: "home-under-v4",
			mutate: func(req *agentic.LaunchRequest) {
				req.System = "claude-code"
				req.Context.Revision = agentic.CuratorLaunchFragmentV4
				req.Context.Environment = "claude_code"
				req.Context.Env = map[string]string{"CLAUDE_CONFIG_DIR": home}
				req.Context.CredentialStore = &agentic.CuratorCredentialStore{Selection: agentic.CuratorCredentialStoreNative}
			},
			mode:    agentic.LaunchModeExec,
			wantErr: agentic.ErrCuratorContextMalformed,
		},
		{
			name: "home-on-non-flag-env",
			mutate: func(req *agentic.LaunchRequest) {
				req.System = "claude-code"
				req.Context.Environment = "claude_code"
				req.Context.Env = map[string]string{"CLAUDE_CONFIG_DIR": home}
				req.Context.CredentialStore = &agentic.CuratorCredentialStore{Selection: agentic.CuratorCredentialStoreNative}
			},
			mode:    agentic.LaunchModeExec,
			wantErr: agentic.ErrCuratorContextMalformed,
		},
		{
			name: "antigravity-wrong-revision-v4",
			mutate: func(req *agentic.LaunchRequest) {
				req.Context.Revision = agentic.CuratorLaunchFragmentV4
				req.Context.Home = nil
			},
			mode:    agentic.LaunchModeExec,
			wantErr: agentic.ErrCuratorFragmentRevisionUnsupported,
		},
		{
			name: "antigravity-dot-gemini-basename",
			mutate: func(req *agentic.LaunchRequest) {
				bad := "/Users/operator/.curator/environments/team-default/.gemini"
				req.Home = bad
				req.Context.Home.Path = bad
			},
			mode:    agentic.LaunchModeExec,
			wantErr: agentic.ErrCuratorContextMalformed,
		},
		{
			name: "antigravity-bad-env",
			mutate: func(req *agentic.LaunchRequest) {
				req.Context.Env = map[string]string{"AGY_CLI_FORCE_FILE_TOKEN_STORAGE": "false"}
			},
			mode:    agentic.LaunchModeExec,
			wantErr: agentic.ErrCuratorContextMalformed,
		},
		{
			name: "antigravity-system-prompt-forbidden",
			mutate: func(req *agentic.LaunchRequest) {
				req.Context.SystemPrompt = &agentic.CuratorSystemPromptContext{Path: "/a/p.md", Intent: agentic.CuratorSystemPromptAppend}
			},
			mode:    agentic.LaunchModeExec,
			wantErr: agentic.ErrCuratorContextMalformed,
		},
		{
			name: "antigravity-mcp-wrong-path",
			mutate: func(req *agentic.LaunchRequest) {
				req.Context.MCP = &agentic.CuratorMCPContext{
					Path:     "/wrong/mcp_config.json",
					EnvNames: []string{},
					Channels: []agentic.CuratorChannelDescriptor{{Kind: agentic.CuratorDescriptorFile, Filename: agentic.CuratorAntigravityMCPFile}},
				}
			},
			mode:    agentic.LaunchModeExec,
			wantErr: agentic.ErrCuratorContextMalformed,
		},
		{
			name: "antigravity-mcp-unsorted-env-names",
			mutate: func(req *agentic.LaunchRequest) {
				req.Context.MCP = &agentic.CuratorMCPContext{
					Path:     path.Join(home, agentic.CuratorAntigravityMCPFile),
					EnvNames: []string{"Z_VAR", "A_VAR"},
					Channels: []agentic.CuratorChannelDescriptor{{Kind: agentic.CuratorDescriptorFile, Filename: agentic.CuratorAntigravityMCPFile}},
				}
			},
			mode:    agentic.LaunchModeExec,
			wantErr: agentic.ErrCuratorContextMalformed,
		},
		{
			name: "antigravity-mcp-wrong-channel",
			mutate: func(req *agentic.LaunchRequest) {
				req.Context.MCP = &agentic.CuratorMCPContext{
					Path:     path.Join(home, agentic.CuratorAntigravityMCPFile),
					EnvNames: []string{},
					Channels: []agentic.CuratorChannelDescriptor{{Kind: agentic.CuratorDescriptorFile, Filename: "other.json"}},
				}
			},
			mode:    agentic.LaunchModeExec,
			wantErr: agentic.ErrCuratorContextMalformed,
		},
		{
			name: "antigravity-keychain-credential-store",
			mutate: func(req *agentic.LaunchRequest) {
				req.Context.CredentialStore = &agentic.CuratorCredentialStore{Selection: agentic.CuratorCredentialStoreKeychain}
			},
			mode:    agentic.LaunchModeExec,
			wantErr: agentic.ErrCuratorContextMalformed,
		},
		{
			name: "antigravity-mismatched-environment",
			mutate: func(req *agentic.LaunchRequest) {
				req.Context = &agentic.CuratorContext{
					Revision:    agentic.CuratorLaunchFragmentV4,
					Environment: "claude_code",
					Profile:     agentic.CuratorProfilePin{Name: "p", LockSHA256: strings.Repeat("a", 64)},
					Precedence:  agentic.CuratorPrecedence{Winner: "higher-weight", Placement: "winner-last"},
					Env:         map[string]string{"CLAUDE_CONFIG_DIR": home},
					Permissions: &agentic.CuratorPermissions{Mode: agentic.CuratorPermissionNative, Source: agentic.CuratorPermissionSourceDefault},
				}
			},
			mode:    agentic.LaunchModeExec,
			wantErr: agentic.ErrCuratorContextUnsupported,
		},
		{
			name: "antigravity-legacy-composition-refused",
			mutate: func(req *agentic.LaunchRequest) {
				req.Composition = agentic.Composition{Prefix: []string{"--mcp-config", `{"mcpServers":{}}`}}
			},
			mode:    agentic.LaunchModeExec,
			wantErr: agentic.ErrCuratorContextUnsupported,
		},
		{
			name: "antigravity-exec-permission-mode-refused",
			mutate: func(req *agentic.LaunchRequest) {
				req.PermissionMode = agentic.PermissionModeYolo
			},
			mode:    agentic.LaunchModeExec,
			wantErr: agentic.ErrPermissionModeNotInteractive,
		},
		{
			name: "antigravity-exec-native-args-refused",
			mutate: func(req *agentic.LaunchRequest) {
				req.NativeArgs = []string{"--verbose"}
			},
			mode:    agentic.LaunchModeExec,
			wantErr: agentic.ErrNativeArgsNotInteractive,
		},
		{
			name: "antigravity-interactive-invalid-permission-mode",
			mutate: func(req *agentic.LaunchRequest) {
				req.Prompt = nil
				req.PermissionMode = agentic.PermissionMode("bogus")
			},
			mode:    agentic.LaunchModeInteractive,
			wantErr: agentic.ErrPermissionModeUnknown,
		},
		{
			name: "antigravity-interactive-conflicting-gemini-dir",
			mutate: func(req *agentic.LaunchRequest) {
				req.Prompt = nil
				req.ToolRelease = "1.5.16"
				req.NativeArgs = []string{"--gemini_dir=/other"}
			},
			mode:    agentic.LaunchModeInteractive,
			wantErr: agentic.ErrNativePolicyConflict,
		},
		{
			name: "antigravity-interactive-conflicting-cli-dir-name",
			mutate: func(req *agentic.LaunchRequest) {
				req.Prompt = nil
				req.ToolRelease = "1.5.16"
				req.NativeArgs = []string{"--cli_dir_name=other"}
			},
			mode:    agentic.LaunchModeInteractive,
			wantErr: agentic.ErrNativePolicyConflict,
		},
		{
			name: "antigravity-interactive-conflicting-settings",
			mutate: func(req *agentic.LaunchRequest) {
				req.Prompt = nil
				req.ToolRelease = "1.5.16"
				req.NativeArgs = []string{"--settings", `{"modelProvider":"gemini"}`}
			},
			mode:    agentic.LaunchModeInteractive,
			wantErr: agentic.ErrNativePolicyConflict,
		},
		{
			name: "antigravity-interactive-managed-native-unverified-release",
			mutate: func(req *agentic.LaunchRequest) {
				req.Prompt = nil
				req.PermissionMode = agentic.PermissionModeNative
				req.ToolRelease = "1.5.15"
			},
			mode:    agentic.LaunchModeInteractive,
			wantErr: agentic.ErrPermissionModeUnverifiedRelease,
		},
		{
			name: "antigravity-interactive-conflicting-effort",
			mutate: func(req *agentic.LaunchRequest) {
				req.Prompt = nil
				req.ToolRelease = "1.5.18"
				req.Model = agentic.Model{ID: "custom-model", Effort: agentic.EffortSupportRequired}
				req.Effort = "xhigh"
				req.NativeArgs = []string{"--effort", "low"}
			},
			mode:    agentic.LaunchModeInteractive,
			wantErr: agentic.ErrNativePolicyConflict,
		},
		{
			name: "antigravity-mcp-forbidden-adc-env",
			mutate: func(req *agentic.LaunchRequest) {
				req.Context.MCP = &agentic.CuratorMCPContext{
					Path:     path.Join(home, agentic.CuratorAntigravityMCPFile),
					EnvNames: []string{"AGY_ADC_AUTH"},
					Channels: []agentic.CuratorChannelDescriptor{{Kind: agentic.CuratorDescriptorFile, Filename: agentic.CuratorAntigravityMCPFile}},
				}
			},
			mode:    agentic.LaunchModeExec,
			wantErr: agentic.ErrCuratorContextMalformed,
		},
		{
			name: "antigravity-node-credential-source-unavailable",
			mutate: func(req *agentic.LaunchRequest) {
				req.Context.CredentialStore = &agentic.CuratorCredentialStore{Selection: agentic.CuratorCredentialStoreNative}
				req.Context.CredentialSource = &agentic.CuratorCredentialSource{
					Source:           &agentic.CuratorNodeSource{Mode: "node", SourceRef: "local:0123456789abcdef0123456789abcdef", HarnessChannel: "GEMINI_API_KEY"},
					SourceGeneration: 1,
					PolicyGeneration: 1,
				}
			},
			mode:    agentic.LaunchModeExec,
			wantErr: agentic.ErrCredentialInjectionUnavailable,
		},
		{
			name: "antigravity-node-credential-source-wrong-store",
			mutate: func(req *agentic.LaunchRequest) {
				req.Context.CredentialSource = &agentic.CuratorCredentialSource{
					Source:           &agentic.CuratorNodeSource{Mode: "node", SourceRef: "local:0123456789abcdef0123456789abcdef", HarnessChannel: "GEMINI_API_KEY"},
					SourceGeneration: 1,
					PolicyGeneration: 1,
				}
			},
			mode:    agentic.LaunchModeExec,
			wantErr: agentic.ErrCuratorContextMalformed,
		},
		{
			name: "v4-credential-source-malformed",
			mutate: func(req *agentic.LaunchRequest) {
				req.System = "claude-code"
				req.Context.Revision = agentic.CuratorLaunchFragmentV4
				req.Context.Environment = "claude_code"
				req.Context.Home = nil
				req.Context.Env = map[string]string{"CLAUDE_CONFIG_DIR": home}
				req.Context.CredentialStore = &agentic.CuratorCredentialStore{Selection: agentic.CuratorCredentialStoreNative}
				req.Context.CredentialSource = &agentic.CuratorCredentialSource{
					Source:           &agentic.CuratorNodeSource{Mode: "node", SourceRef: "local:0123456789abcdef0123456789abcdef", HarnessChannel: "CLAUDE_CODE_OAUTH_TOKEN"},
					SourceGeneration: 1,
					PolicyGeneration: 1,
				}
			},
			mode:    agentic.LaunchModeExec,
			wantErr: agentic.ErrCuratorContextMalformed,
		},
		{
			name: "v7-non-flag-credential-source-unavailable",
			mutate: func(req *agentic.LaunchRequest) {
				req.System = "claude-code"
				req.Context.Environment = "claude_code"
				req.Context.Home = nil
				req.Context.Env = map[string]string{"CLAUDE_CONFIG_DIR": home}
				req.Context.CredentialStore = &agentic.CuratorCredentialStore{Selection: agentic.CuratorCredentialStoreNative}
				req.Context.CredentialSource = &agentic.CuratorCredentialSource{
					Source:           &agentic.CuratorNodeSource{Mode: "node", SourceRef: "local:0123456789abcdef0123456789abcdef", HarnessChannel: "CLAUDE_CODE_OAUTH_TOKEN"},
					SourceGeneration: 1,
					PolicyGeneration: 1,
				}
			},
			mode:    agentic.LaunchModeExec,
			wantErr: agentic.ErrCredentialInjectionUnavailable,
		},
		{
			name: "antigravity-interactive-yolo-unverified-release",
			mutate: func(req *agentic.LaunchRequest) {
				req.Prompt = nil
				req.PermissionMode = agentic.PermissionModeYolo
				req.ToolRelease = "9.9.9"
			},
			mode:    agentic.LaunchModeInteractive,
			wantErr: agentic.ErrPermissionModeUnverifiedRelease,
		},
		{
			name: "antigravity-interactive-duplicate-yolo",
			mutate: func(req *agentic.LaunchRequest) {
				req.Prompt = nil
				req.PermissionMode = agentic.PermissionModeYolo
				req.ToolRelease = "1.5.16"
				req.NativeArgs = []string{"--dangerously-skip-permissions"}
			},
			mode:    agentic.LaunchModeInteractive,
			wantErr: agentic.ErrPermissionModeDuplicate,
		},
		{
			name: "antigravity-interactive-yolo-mode-conflict",
			mutate: func(req *agentic.LaunchRequest) {
				req.Prompt = nil
				req.PermissionMode = agentic.PermissionModeYolo
				req.ToolRelease = "1.5.16"
				req.NativeArgs = []string{"--mode=plan"}
			},
			mode:    agentic.LaunchModeInteractive,
			wantErr: agentic.ErrNativePolicyConflict,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := antigravityExecRequest(home, validAntigravityV7Context(home))
			tc.mutate(&req)
			if _, err := agentic.BuildPlan(reg, req, tc.mode); !errors.Is(err, tc.wantErr) {
				t.Fatalf("BuildPlan(%s) err = %v, want %v", tc.name, err, tc.wantErr)
			}
		})
	}
}
