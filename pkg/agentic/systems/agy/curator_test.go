package agy

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/relux-works/archon-pantheon/pkg/agentic"
)

func TestCuratorV7DirectCallerParity(t *testing.T) {
	t.Parallel()
	sys := NewWithRuntime(Runtime{Executable: "/opt/agy/bin/agy", Version: "1.5.16"})
	home := "/Users/operator/.curator/environments/team/antigravity"
	ctx := &agentic.CuratorContext{
		Revision:    agentic.CuratorLaunchFragmentV7,
		Environment: "antigravity",
		Profile:     agentic.CuratorProfilePin{Name: "team", LockSHA256: strings.Repeat("b", 64)},
		Precedence:  agentic.CuratorPrecedence{Winner: "higher-weight", Placement: "winner-last"},
		Home:        &agentic.CuratorHomeBinding{Path: home, Transport: agentic.CuratorHomeTransportFlag},
		Env:         map[string]string{agentic.CuratorAntigravityForceFileTokenVariable: agentic.CuratorAntigravityForceFileTokenTrue},
		Permissions: &agentic.CuratorPermissions{Mode: agentic.CuratorPermissionNative, Source: agentic.CuratorPermissionSourceDefault},
		CredentialStore: &agentic.CuratorCredentialStore{
			Selection: agentic.CuratorCredentialStoreFile,
		},
	}
	req := agentic.LaunchRequest{
		System:  "antigravity",
		Model:   agentic.Model{ID: "gemini-4-carbon-low"},
		Home:    home,
		Prompt:  []byte("test"),
		Context: ctx,
	}

	argv, err := sys.Argv(req, agentic.LaunchModeExec)
	if err != nil {
		t.Fatalf("Argv(exec) err = %v", err)
	}
	if argv[0] != "--gemini_dir" || argv[1] != home {
		t.Fatalf("argv = %v, want leading --gemini_dir %s", argv, home)
	}

	env, err := sys.ChildEnv([]string{"HOME=/Users/operator", "PATH=/usr/bin", "GEMINI_API_KEY=secret", "AGY_CLI_FORCE_FILE_TOKEN_STORAGE=false"}, req)
	if err != nil {
		t.Fatalf("ChildEnv err = %v", err)
	}
	if !slices.Contains(env, "AGY_CLI_FORCE_FILE_TOKEN_STORAGE=true") || !slices.Contains(env, "AGY_CLI_DISABLE_AUTO_UPDATE=true") {
		t.Fatalf("env missing required literals: %v", env)
	}
	if slices.Contains(env, "GEMINI_API_KEY=secret") {
		t.Fatalf("ChildEnv leaked GEMINI_API_KEY: %v", env)
	}

	// Direct caller with invalid context fails in both Argv and ChildEnv
	badReq := req
	badCtx := *ctx
	badCtx.Env = map[string]string{"AGY_CLI_FORCE_FILE_TOKEN_STORAGE": "false"}
	badReq.Context = &badCtx
	if _, err := sys.Argv(badReq, agentic.LaunchModeExec); !errors.Is(err, agentic.ErrCuratorContextMalformed) {
		t.Fatalf("direct Argv bad context err = %v, want ErrCuratorContextMalformed", err)
	}
	if _, err := sys.ChildEnv(nil, badReq); !errors.Is(err, agentic.ErrCuratorContextMalformed) {
		t.Fatalf("direct ChildEnv bad context err = %v, want ErrCuratorContextMalformed", err)
	}

	// Interactive mode requires preflight
	if _, err := New().Argv(agentic.LaunchRequest{System: "antigravity", Home: home, Context: ctx}, agentic.LaunchModeInteractive); !errors.Is(err, ErrRuntimeNotPreflighted) {
		t.Fatalf("interactive without preflight err = %v, want ErrRuntimeNotPreflighted", err)
	}
}
