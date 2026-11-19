package kimi

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"slices"
	"testing"

	"github.com/relux-works/skill-agents-management/pkg/agentic"
)

// TestDeclarationMatchesTheSourcedCLIContract pins the supported modes,
// home, and unsupported shared launch parameters against the Kimi Code CLI
// 2.6.3 contract. It does not establish installed-user configuration.
func TestDeclarationMatchesTheSourcedCLIContract(t *testing.T) {
	caps := New().Capabilities()
	if New().ID() != "kimi" {
		t.Fatalf("ID() = %q, want kimi", New().ID())
	}
	for _, mode := range []agentic.LaunchMode{agentic.LaunchModeExec, agentic.LaunchModeDryRun, agentic.LaunchModeInteractive} {
		if !caps.SupportsMode(mode) {
			t.Errorf("mode %s is not declared", mode)
		}
	}
	if caps.EffortTransport != agentic.EffortTransportNone || caps.SupportsGoal || caps.SupportsBudget || caps.SupportsServiceTier || caps.CompositionGrammar != agentic.GrammarNone {
		t.Fatalf("capabilities claim an unsupported launch parameter: %+v", caps)
	}
	if caps.HomeEnvVar != homeEnvVar || caps.DefaultHome != defaultHome {
		t.Fatalf("home declaration = (%q, %q), want (%q, %q)", caps.HomeEnvVar, caps.DefaultHome, homeEnvVar, defaultHome)
	}
}

// TestBuildPlanBuildsEachSupportedMode drives the real registry and argument
// builder. Dry-run uses a placeholder without reading the absent prompt file;
// model aliases remain opaque until Kimi reads its own configuration.
func TestBuildPlanBuildsEachSupportedMode(t *testing.T) {
	binDir := buildFakeKimi(t)
	workDir := t.TempDir()
	customHome := filepath.Join(t.TempDir(), "kimi-home")
	promptFile := filepath.Join(workDir, "task.md")
	if err := os.WriteFile(promptFile, []byte("prompt from file"), 0o600); err != nil {
		t.Fatalf("writing prompt file: %v", err)
	}
	cases := []struct {
		name string
		mode agentic.LaunchMode
		edit func(*agentic.LaunchRequest)
		want []string
	}{
		{
			name: "prompt",
			mode: agentic.LaunchModeExec,
			want: []string{"--model", "custom-alias", "--prompt", "task prompt"},
		},
		{
			name: "opaque model alias",
			mode: agentic.LaunchModeExec,
			edit: func(req *agentic.LaunchRequest) { req.Model.ID = "configured alias" },
			want: []string{"--model", "configured alias", "--prompt", "task prompt"},
		},
		{
			name: "prompt file",
			mode: agentic.LaunchModeExec,
			edit: func(req *agentic.LaunchRequest) { req.PromptPath = promptFile },
			want: []string{"--model", "custom-alias", "--prompt", "prompt from file"},
		},
		{
			name: "dry-run",
			mode: agentic.LaunchModeDryRun,
			edit: func(req *agentic.LaunchRequest) { req.PromptPath = filepath.Join(workDir, "not-created.md") },
			want: []string{"--model", "custom-alias", "--prompt", promptPlaceholder},
		},
		{
			name: "interactive resume by id",
			mode: agentic.LaunchModeInteractive,
			edit: func(req *agentic.LaunchRequest) {
				req.Prompt = nil
				req.NativeArgs = []string{"--session", "session-123"}
			},
			want: []string{"--model", "custom-alias", "--session", "session-123"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := newRequest(binDir, workDir)
			req.Home = customHome
			req.Env = []string{"PATH=" + binDir, homeEnvVar + "=" + customHome}
			if tc.mode == agentic.LaunchModeInteractive {
				req.Prompt = nil
			}
			if tc.edit != nil {
				tc.edit(&req)
			}
			plan, err := agentic.BuildPlan(agentic.Default, req, tc.mode)
			if err != nil {
				t.Fatalf("BuildPlan(%s): %v", tc.mode, err)
			}
			if !reflect.DeepEqual(plan.Argv, tc.want) {
				t.Errorf("Argv = %#v, want %#v", plan.Argv, tc.want)
			}
			if plan.Binary != filepath.Join(binDir, executableName) || plan.WorkDir != workDir || plan.Home != customHome {
				t.Errorf("plan target = (%q, %q, %q), want binary, work directory, and selected home", plan.Binary, plan.WorkDir, plan.Home)
			}
			if plan.Stdin.Attached || len(plan.Stdin.Bytes) != 0 {
				t.Errorf("Kimi prompt unexpectedly reached stdin: %+v", plan.Stdin)
			}
			if !slices.Contains(plan.Env, homeEnvVar+"="+customHome) {
				t.Errorf("supplied Kimi home was lost from the child environment: %v", plan.Env)
			}
		})
	}
}

// TestBuildPlanRefusesInvalidInputsWithoutMutation pins plugin-owned
// refusals and inherited admission gates. Every row checks that a rejected
// launch leaves the caller's request unchanged; the missing-binary row proves
// only PATH resolution, not installation state elsewhere on the machine.
func TestBuildPlanRefusesInvalidInputsWithoutMutation(t *testing.T) {
	binDir := buildFakeKimi(t)
	emptyDir := t.TempDir()
	workDir := t.TempDir()
	cases := []struct {
		name string
		mode agentic.LaunchMode
		edit func(*agentic.LaunchRequest)
		want error
	}{
		{"missing model", agentic.LaunchModeExec, func(r *agentic.LaunchRequest) { r.Model.ID = "" }, agentic.ErrModelMissing},
		{"whitespace-only model", agentic.LaunchModeExec, func(r *agentic.LaunchRequest) { r.Model.ID = " " }, agentic.ErrModelMissing},
		{"empty prompt", agentic.LaunchModeExec, func(r *agentic.LaunchRequest) { r.Prompt = nil }, ErrPromptInvalid},
		{"whitespace-only prompt", agentic.LaunchModeExec, func(r *agentic.LaunchRequest) { r.Prompt = []byte(" \t\n") }, ErrPromptInvalid},
		{"invalid UTF-8 prompt", agentic.LaunchModeExec, func(r *agentic.LaunchRequest) { r.Prompt = []byte{0xff} }, ErrPromptInvalid},
		{"NUL prompt", agentic.LaunchModeExec, func(r *agentic.LaunchRequest) { r.Prompt = []byte{'a', 0, 'b'} }, ErrPromptInvalid},
		{"unreadable prompt file", agentic.LaunchModeExec, func(r *agentic.LaunchRequest) { r.PromptPath = filepath.Join(workDir, "missing-prompt.md") }, nil},
		{"missing binary", agentic.LaunchModeExec, func(r *agentic.LaunchRequest) { r.Env = []string{"PATH=" + emptyDir} }, nil},
		{"bare session selector", agentic.LaunchModeInteractive, func(r *agentic.LaunchRequest) { r.NativeArgs = []string{"--session"} }, agentic.ErrSessionResumeInvalid},
		{"bare short session selector", agentic.LaunchModeInteractive, func(r *agentic.LaunchRequest) { r.NativeArgs = []string{"-S"} }, agentic.ErrSessionResumeInvalid},
		{"bare long resume alias", agentic.LaunchModeInteractive, func(r *agentic.LaunchRequest) { r.NativeArgs = []string{"--resume"} }, agentic.ErrSessionResumeInvalid},
		{"bare resume alias", agentic.LaunchModeInteractive, func(r *agentic.LaunchRequest) { r.NativeArgs = []string{"-r"} }, agentic.ErrSessionResumeInvalid},
		{"conflicting resume selectors", agentic.LaunchModeInteractive, func(r *agentic.LaunchRequest) { r.NativeArgs = []string{"--continue", "--session", "session-123"} }, agentic.ErrSessionResumeInvalid},
		{"conflicting short session selectors", agentic.LaunchModeInteractive, func(r *agentic.LaunchRequest) { r.NativeArgs = []string{"-S", "first", "-r", "second"} }, agentic.ErrSessionResumeInvalid},
		{"conflicting long and short session selectors", agentic.LaunchModeInteractive, func(r *agentic.LaunchRequest) { r.NativeArgs = []string{"--session", "first", "-S", "second"} }, agentic.ErrSessionResumeInvalid},
		{"empty identified session", agentic.LaunchModeInteractive, func(r *agentic.LaunchRequest) { r.NativeArgs = []string{"--session", ""} }, agentic.ErrSessionResumeInvalid},
		{"continue selector does not take a value", agentic.LaunchModeInteractive, func(r *agentic.LaunchRequest) { r.NativeArgs = []string{"--continue=session-123"} }, agentic.ErrSessionResumeInvalid},
		{"prompt is not an interactive channel", agentic.LaunchModeInteractive, func(r *agentic.LaunchRequest) { r.Prompt = []byte("task") }, agentic.ErrParameterNotInteractive},
		{"unknown permission mode", agentic.LaunchModeInteractive, func(r *agentic.LaunchRequest) { r.PermissionMode = "never" }, agentic.ErrPermissionModeUnknown},
		{"unverified yolo release", agentic.LaunchModeInteractive, func(r *agentic.LaunchRequest) { r.PermissionMode = agentic.PermissionModeYolo; r.ToolRelease = "2.6.4" }, agentic.ErrPermissionModeUnverifiedRelease},
		{"unrepresented auto posture", agentic.LaunchModeInteractive, func(r *agentic.LaunchRequest) { r.NativeArgs = []string{"--auto"} }, agentic.ErrPermissionModeUnsupported},
		{"canonical yolo under native posture", agentic.LaunchModeInteractive, func(r *agentic.LaunchRequest) { r.NativeArgs = []string{"--yolo"} }, agentic.ErrNativePolicyConflict},
		{"short yolo under native posture", agentic.LaunchModeInteractive, func(r *agentic.LaunchRequest) { r.NativeArgs = []string{"-y"} }, agentic.ErrNativePolicyConflict},
		{"clustered yolo and continue short options", agentic.LaunchModeInteractive, func(r *agentic.LaunchRequest) { r.NativeArgs = []string{"-yc"} }, agentic.ErrNativePolicyConflict},
		{"clustered continue and yolo short options", agentic.LaunchModeInteractive, func(r *agentic.LaunchRequest) { r.NativeArgs = []string{"-cy"} }, agentic.ErrNativePolicyConflict},
		{"clustered yolo and hidden continue short options", agentic.LaunchModeInteractive, func(r *agentic.LaunchRequest) { r.NativeArgs = []string{"-yC"} }, agentic.ErrNativePolicyConflict},
		{"clustered hidden continue and yolo short options", agentic.LaunchModeInteractive, func(r *agentic.LaunchRequest) { r.NativeArgs = []string{"-Cy"} }, agentic.ErrNativePolicyConflict},
		{"cluster ending in a value-taking short option", agentic.LaunchModeInteractive, func(r *agentic.LaunchRequest) { r.NativeArgs = []string{"-ym", "configured-alias"} }, agentic.ErrNativePolicyConflict},
		{"hidden yolo alias under native posture", agentic.LaunchModeInteractive, func(r *agentic.LaunchRequest) { r.NativeArgs = []string{"--yes"} }, agentic.ErrNativePolicyConflict},
		{"hidden auto-approve alias under native posture", agentic.LaunchModeInteractive, func(r *agentic.LaunchRequest) { r.NativeArgs = []string{"--auto-approve"} }, agentic.ErrNativePolicyConflict},
		{"duplicate yolo under yolo posture", agentic.LaunchModeInteractive, func(r *agentic.LaunchRequest) {
			r.PermissionMode = agentic.PermissionModeYolo
			r.ToolRelease = "2.6.3"
			r.NativeArgs = []string{"--yolo"}
		}, agentic.ErrPermissionModeDuplicate},
		{"duplicate short yolo alias", agentic.LaunchModeInteractive, func(r *agentic.LaunchRequest) {
			r.PermissionMode = agentic.PermissionModeYolo
			r.ToolRelease = "2.6.3"
			r.NativeArgs = []string{"-y"}
		}, agentic.ErrPermissionModeDuplicate},
		{"duplicate hidden yolo alias", agentic.LaunchModeInteractive, func(r *agentic.LaunchRequest) {
			r.PermissionMode = agentic.PermissionModeYolo
			r.ToolRelease = "2.6.3"
			r.NativeArgs = []string{"--yes"}
		}, agentic.ErrPermissionModeDuplicate},
		{"duplicate auto-approve alias", agentic.LaunchModeInteractive, func(r *agentic.LaunchRequest) {
			r.PermissionMode = agentic.PermissionModeYolo
			r.ToolRelease = "2.6.3"
			r.NativeArgs = []string{"--auto-approve"}
		}, agentic.ErrPermissionModeDuplicate},
		{"prompt override in interactive mode", agentic.LaunchModeInteractive, func(r *agentic.LaunchRequest) { r.NativeArgs = []string{"--prompt", "task"} }, agentic.ErrNativePolicyConflict},
		{"model override in interactive mode", agentic.LaunchModeInteractive, func(r *agentic.LaunchRequest) { r.NativeArgs = []string{"--model", "other"} }, agentic.ErrNativePolicyConflict},
		{"effort has no transport", agentic.LaunchModeExec, func(r *agentic.LaunchRequest) { r.Effort = "high" }, agentic.ErrEffortNotTransportable},
		{"goal has no transport", agentic.LaunchModeExec, func(r *agentic.LaunchRequest) { r.Goal = &agentic.Goal{ID: "goal"} }, agentic.ErrGoalUnsupported},
		{"budget has no transport", agentic.LaunchModeExec, func(r *agentic.LaunchRequest) { r.Budget = &agentic.Budget{USD: 1} }, agentic.ErrBudgetUnsupported},
		{"service tier has no transport", agentic.LaunchModeExec, func(r *agentic.LaunchRequest) { r.ServiceTier = "priority" }, agentic.ErrServiceTierUnsupported},
		{"composition is not interactive", agentic.LaunchModeInteractive, func(r *agentic.LaunchRequest) { r.Composition.Prefix = []string{"--mcp-config", `{}`} }, agentic.ErrCompositionNotInteractive},
		{"native suffix outside interactive", agentic.LaunchModeExec, func(r *agentic.LaunchRequest) { r.NativeArgs = []string{"--verbose"} }, agentic.ErrNativeArgsNotInteractive},
		{"permission posture outside interactive", agentic.LaunchModeExec, func(r *agentic.LaunchRequest) { r.PermissionMode = agentic.PermissionModeYolo }, agentic.ErrPermissionModeNotInteractive},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := newRequest(binDir, workDir)
			if tc.mode == agentic.LaunchModeInteractive {
				req.Prompt = nil
			}
			if tc.edit != nil {
				tc.edit(&req)
			}
			before := cloneRequest(req)
			_, err := agentic.BuildPlan(agentic.Default, req, tc.mode)
			if err == nil {
				t.Fatal("BuildPlan accepted an invalid request")
			}
			if tc.want != nil && !errors.Is(err, tc.want) {
				t.Errorf("BuildPlan error = %v, want %v", err, tc.want)
			}
			if !reflect.DeepEqual(req, before) {
				t.Errorf("refusal changed the request: got %#v, want %#v", req, before)
			}
		})
	}
}

// TestInteractiveResumeSelectorsAcceptSourcedForms checks sourced session and
// continue spellings through BuildPlan. It proves -c and -C remain accepted and
// forwarded, and a y in an attached session ID remains a value, not a flag.
// It does not prove that any session exists.
func TestInteractiveResumeSelectorsAcceptSourcedForms(t *testing.T) {
	binDir := buildFakeKimi(t)
	for _, tc := range []struct {
		name string
		args []string
	}{
		{"continue long", []string{"--continue"}},
		{"continue short", []string{"-c"}},
		{"continue hidden alias", []string{"-C"}},
		{"session long", []string{"--session", "session-123"}},
		{"session short", []string{"-S", "session-123"}},
		{"session short with attached value containing y", []string{"-Ssession-y"}},
		{"resume long alias", []string{"--resume", "session-123"}},
		{"resume short alias", []string{"-r", "session-123"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := newRequest(binDir, t.TempDir())
			req.Prompt = nil
			req.NativeArgs = append([]string(nil), tc.args...)
			plan, err := agentic.BuildPlan(agentic.Default, req, agentic.LaunchModeInteractive)
			if err != nil {
				t.Fatalf("BuildPlan: %v", err)
			}
			want := append([]string{"--model", "custom-alias"}, tc.args...)
			if !reflect.DeepEqual(plan.Argv, want) {
				t.Fatalf("Argv = %#v, want %#v", plan.Argv, want)
			}
		})
	}
}

// TestYoloMappingIsPinnedToTheDetectedRelease verifies the sourced 2.6.3
// bypass spelling and refuses drift. This proves the module's release table,
// not behavior of a separately installed CLI.
func TestYoloMappingIsPinnedToTheDetectedRelease(t *testing.T) {
	for _, tc := range []struct {
		release string
		mode    agentic.PermissionMode
		want    agentic.PermissionMapping
		wantErr error
	}{
		{release: "2.6.3", mode: agentic.PermissionModeNative, want: agentic.PermissionMapping{Grammar: agentic.PermissionGrammarV1}},
		{release: "2.6.3", mode: agentic.PermissionModeYolo, want: agentic.PermissionMapping{Grammar: agentic.PermissionGrammarV1, Flag: yoloFlag}},
		{release: "", mode: agentic.PermissionModeYolo, wantErr: agentic.ErrPermissionModeUnverifiedRelease},
		{release: "2.6.4", mode: agentic.PermissionModeYolo, wantErr: agentic.ErrPermissionModeUnverifiedRelease},
	} {
		got, err := New().PermissionMapping(tc.release, tc.mode)
		if tc.wantErr != nil {
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("PermissionMapping(%q, %q) error = %v, want %v", tc.release, tc.mode, err, tc.wantErr)
			}
			continue
		}
		if err != nil || got != tc.want {
			t.Errorf("PermissionMapping(%q, %q) = (%+v, %v), want (%+v, nil)", tc.release, tc.mode, got, err, tc.want)
		}
	}
	req := newRequest(buildFakeKimi(t), t.TempDir())
	req.Prompt = nil
	req.PermissionMode = agentic.PermissionModeYolo
	req.ToolRelease = "2.6.3"
	plan, err := agentic.BuildPlan(agentic.Default, req, agentic.LaunchModeInteractive)
	if err != nil {
		t.Fatalf("BuildPlan(yolo): %v", err)
	}
	if !slices.Contains(plan.Argv, yoloFlag) {
		t.Fatalf("verified yolo plan omitted %q: %v", yoloFlag, plan.Argv)
	}
}

// TestFakeKimiReceivesBuildPlanArgv builds a standard-library-only fake CLI in
// a temporary directory and executes plans for prompt, opaque-alias, and
// interactive modes. This proves argv and working-directory handoff through
// the real plan path; it does not establish alias validity or contact a model
// service.
func TestFakeKimiReceivesBuildPlanArgv(t *testing.T) {
	binDir := buildFakeKimi(t)
	for _, tc := range []struct {
		name string
		mode agentic.LaunchMode
		edit func(*agentic.LaunchRequest)
		want []string
	}{
		{
			name: "prompt",
			mode: agentic.LaunchModeExec,
			want: []string{"--model", "custom-alias", "--prompt", "task prompt"},
		},
		{
			name: "opaque model alias bytes",
			mode: agentic.LaunchModeExec,
			edit: func(r *agentic.LaunchRequest) { r.Model.ID = "configured alias" },
			want: []string{"--model", "configured alias", "--prompt", "task prompt"},
		},
		{
			name: "interactive",
			mode: agentic.LaunchModeInteractive,
			edit: func(r *agentic.LaunchRequest) { r.NativeArgs = []string{"--continue"} },
			want: []string{"--model", "custom-alias", "--continue"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			workDir := t.TempDir()
			capture := filepath.Join(t.TempDir(), "capture.json")
			req := newRequest(binDir, workDir)
			if tc.mode == agentic.LaunchModeInteractive {
				req.Prompt = nil
			}
			if tc.edit != nil {
				tc.edit(&req)
			}
			plan, err := agentic.BuildPlan(agentic.Default, req, tc.mode)
			if err != nil {
				t.Fatalf("BuildPlan: %v", err)
			}
			cmd := exec.Command(plan.Binary, plan.Argv...)
			cmd.Dir = plan.WorkDir
			cmd.Env = append(plan.Env, "KIMI_CAPTURE="+capture)
			if output, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("fake Kimi execution: %v; output: %s", err, output)
			}
			data, err := os.ReadFile(capture)
			if err != nil {
				t.Fatalf("reading fake CLI capture: %v", err)
			}
			var got struct {
				Args []string `json:"args"`
				Dir  string   `json:"dir"`
			}
			if err := json.Unmarshal(data, &got); err != nil {
				t.Fatalf("decoding fake CLI capture: %v", err)
			}
			canonicalWorkDir, err := filepath.EvalSymlinks(workDir)
			if err != nil {
				t.Fatalf("resolve work directory: %v", err)
			}
			if !reflect.DeepEqual(got.Args, tc.want) || got.Dir != canonicalWorkDir {
				t.Errorf("fake CLI received args/dir (%#v, %q), want (%#v, %q)", got.Args, got.Dir, tc.want, canonicalWorkDir)
			}
		})
	}
}

// TestProbeToolReleaseUsesTheResolvedFakeBinary proves that detection invokes
// the PATH-resolved executable with exactly --version and accepts a zero-exit
// exact release token. It does not prove an installed build or account works.
func TestProbeToolReleaseUsesTheResolvedFakeBinary(t *testing.T) {
	binDir := buildFakeKimi(t)
	capture := filepath.Join(t.TempDir(), "version-args.json")
	env := []string{"PATH=" + binDir, "KIMI_CAPTURE=" + capture}
	release, err := agentic.ProbeToolRelease(context.Background(), New(), env)
	if err != nil {
		t.Fatalf("ProbeToolRelease: %v", err)
	}
	if release != "2.6.3" {
		t.Fatalf("release = %q, want 2.6.3", release)
	}
	data, err := os.ReadFile(capture)
	if err != nil {
		t.Fatalf("reading probe capture: %v", err)
	}
	var got struct {
		Args []string `json:"args"`
	}
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("decoding probe capture: %v", err)
	}
	if !reflect.DeepEqual(got.Args, []string{"--version"}) {
		t.Fatalf("version probe args = %#v, want [--version]", got.Args)
	}
}

// TestProbeToolReleaseFailsClosed pins missing-binary, nonzero-exit, and
// unparseable-answer refusals. These controls exercise the probe contract with
// a fake executable and do not establish installed CLI health.
func TestProbeToolReleaseFailsClosed(t *testing.T) {
	binDir := buildFakeKimi(t)
	for _, tc := range []struct {
		name string
		env  []string
	}{
		{name: "missing binary", env: []string{"PATH=" + t.TempDir()}},
		{name: "nonzero exit", env: []string{"PATH=" + binDir, "KIMI_FAKE_EXIT=7"}},
		{name: "unparseable answer", env: []string{"PATH=" + binDir, "KIMI_FAKE_VERSION=latest"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := agentic.ProbeToolRelease(context.Background(), New(), tc.env); !errors.Is(err, agentic.ErrToolReleaseUndetected) {
				t.Fatalf("ProbeToolRelease error = %v, want ErrToolReleaseUndetected", err)
			}
		})
	}
}

func newRequest(binDir, workDir string) agentic.LaunchRequest {
	return agentic.LaunchRequest{
		System:  systemID,
		Model:   agentic.Model{ID: "custom-alias", Effort: agentic.EffortSupportNone},
		Prompt:  []byte("task prompt"),
		Env:     []string{"PATH=" + binDir, homeEnvVar + "=/custom/kimi"},
		WorkDir: workDir,
	}
}

func cloneRequest(req agentic.LaunchRequest) agentic.LaunchRequest {
	copy := req
	copy.Prompt = append([]byte(nil), req.Prompt...)
	copy.Env = append([]string(nil), req.Env...)
	copy.NativeArgs = append([]string(nil), req.NativeArgs...)
	copy.Composition.Prefix = append([]string(nil), req.Composition.Prefix...)
	copy.ContextDescriptors = append([]agentic.ContextDescriptor(nil), req.ContextDescriptors...)
	if req.Goal != nil {
		value := *req.Goal
		copy.Goal = &value
	}
	if req.Budget != nil {
		value := *req.Budget
		copy.Budget = &value
	}
	return copy
}

func buildFakeKimi(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module fake-kimi\n\ngo 1.25.5\n"), 0o600); err != nil {
		t.Fatalf("writing fake module: %v", err)
	}
	const source = `package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) == 2 && os.Args[1] == "--version" {
		version := os.Getenv("KIMI_FAKE_VERSION")
		if version == "" { version = "2.6.3" }
		fmt.Println(version)
		if os.Getenv("KIMI_FAKE_EXIT") != "" { os.Exit(7) }
	}
	capture := os.Getenv("KIMI_CAPTURE")
	if capture == "" {
		if len(os.Args) == 2 && os.Args[1] == "--version" { return }
		os.Exit(2)
	}
	data, err := json.Marshal(struct {
		Args []string ` + "`json:\"args\"`" + `
		Dir string ` + "`json:\"dir\"`" + `
	}{os.Args[1:], mustGetwd()})
	if err != nil { os.Exit(3) }
	if err := os.WriteFile(capture, data, 0o600); err != nil { os.Exit(4) }
}

func mustGetwd() string {
	value, err := os.Getwd()
	if err != nil { os.Exit(5) }
	return value
}
`
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte(source), 0o600); err != nil {
		t.Fatalf("writing fake CLI source: %v", err)
	}
	binDir := filepath.Join(dir, "bin")
	if err := os.MkdirAll(binDir, 0o700); err != nil {
		t.Fatalf("creating fake CLI bin: %v", err)
	}
	binary := filepath.Join(binDir, executableName)
	cmd := exec.Command("go", "build", "-o", binary, ".")
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("building fake CLI: %v; output: %s", err, output)
	}
	return binDir
}
