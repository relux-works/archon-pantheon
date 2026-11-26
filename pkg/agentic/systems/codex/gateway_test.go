package codex

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/pelletier/go-toml/v2"
	"github.com/relux-works/curator-runtimes/internal/gatewaybasis"
	"github.com/relux-works/curator-runtimes/pkg/agentic"
)

func gatewayPlan(t *testing.T, native []string) agentic.Plan {
	t.Helper()
	r := managedPrimaryRequest(t)
	r.ToolRelease = "0.167.0"
	r.NativeArgs = native
	return requireManagedPlan(t, r)
}

func gatewayBinding(t *testing.T, p agentic.Plan) agentic.CodexGatewayBinding {
	t.Helper()
	data, err := os.ReadFile(p.Binary)
	if err != nil {
		t.Fatal(err)
	}
	return agentic.CodexGatewayBinding{ProviderID: GatewayProviderID, BaseURL: "http://127.0.0.1:45678/chatgpt/backend-api/codex", HeaderName: "X-Netproxy-Token", TokenEnv: GatewayTokenEnv, WireAPI: "responses", RequiresOpenAIAuth: true, ToolRelease: "0.167.0", ExecutablePath: p.Binary, ExecutableSHA256: fmt.Sprintf("%x", sha256.Sum256(data))}
}

func gatewayConfigDocument(args []string, owned int) []byte {
	values := make([]string, 0, owned/2)
	for i := 0; i < owned; i += 2 {
		values = append(values, args[i+1])
	}
	return []byte(strings.Join(values, "\n"))
}

// A zero-value binding preserves the existing gateway argv without search settings.
// Fresh and exact resume retain every original argument and native suffix;
// decoded TOML proves auth, transport and env-header values independently of
// the renderer. No native runtime or actual credentials are used here.
func TestGatewayFreshAndResume(t *testing.T) {
	for _, suffix := range [][]string{nil, {"resume", managedThread}, {"--no-alt-screen", "resume", managedThread}} {
		p := gatewayPlan(t, suffix)
		b := gatewayBinding(t, p)
		a, err := gatewayAdmitFixture(p, agentic.CodexGatewayInputs{NativeArgs: slices.Clone(suffix), FragmentEnv: []string{"DISJOINT=yes"}})
		if err != nil {
			t.Fatal(err)
		}
		final, err := gatewayFinalizeFixture(a, b)
		if err != nil {
			t.Fatal(err)
		}
		if !slices.Equal(final.Argv[gatewayOwnedArgs:], p.Argv) || final.ModelIdentity != p.ModelIdentity || final.PermissionMode != p.PermissionMode {
			t.Fatal("original admission changed")
		}
		var decoded map[string]any
		if err := toml.Unmarshal(gatewayConfigDocument(final.Argv, gatewayOwnedArgs), &decoded); err != nil {
			t.Fatal(err)
		}
		providers := decoded["model_providers"].(map[string]any)
		provider := providers[GatewayProviderID].(map[string]any)
		if decoded["model_provider"] != GatewayProviderID || provider["base_url"] != b.BaseURL || provider["requires_openai_auth"] != true || provider["supports_websockets"] != false || provider["wire_api"] != "responses" {
			t.Fatalf("wrong route contract: %#v", decoded)
		}
		if _, ok := provider["supports_standalone_web_search"]; ok {
			t.Fatal("zero-value binding enabled provider search")
		}
		if features, ok := decoded["features"].(map[string]any); !ok {
			t.Fatal("gateway feature settings are missing")
		} else if _, ok := features["standalone_web_search"]; ok {
			t.Fatal("zero-value binding enabled standalone search")
		}
		if got := provider["env_http_headers"].(map[string]any)[b.HeaderName]; got != GatewayTokenEnv {
			t.Fatal("wrong header binding")
		}
		if err := final.VerifyCodexGatewayProcess(b.ExecutablePath, final.Argv, append(slices.Clone(final.Env), GatewayTokenEnv+"=synthetic-child-only")); err != nil {
			t.Fatal(err)
		}
		if _, err := gatewayFinalizeFixture(a, b); !errors.Is(err, agentic.ErrCodexGatewayRefused) {
			t.Fatal("second finalization accepted", err)
		}
	}
}

// An enabled binding emits both canonical search settings and preserves the suffix.
func TestGatewayStandaloneSearchOptIn(t *testing.T) {
	r := managedPrimaryRequest(t)
	r.ToolRelease = "0.167.0"
	p := requireManagedPlan(t, r)
	b := gatewayBinding(t, p)
	b.SupportsStandaloneWebSearch = true
	a, err := gatewayAdmitFixture(p, agentic.CodexGatewayInputs{})
	if err != nil {
		t.Fatal(err)
	}
	final, err := gatewayFinalizeFixture(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if len(final.Argv) < gatewayOwnedArgs+2 || !slices.Equal(final.Argv[gatewayOwnedArgs+2:], p.Argv) {
		t.Fatal("enabled gateway changed or truncated the original argv suffix")
	}
	if !slices.Equal(final.Argv[gatewayOwnedArgs:gatewayOwnedArgs+2], []string{"-c", gatewayStandaloneSearchFeature}) {
		t.Fatalf("search feature is not after the existing owned switches: %q", final.Argv)
	}
	var decoded map[string]any
	if err := toml.Unmarshal(gatewayConfigDocument(final.Argv, gatewayOwnedArgs+2), &decoded); err != nil {
		t.Fatal(err)
	}
	providers := decoded["model_providers"].(map[string]any)
	provider := providers[GatewayProviderID].(map[string]any)
	if provider["supports_standalone_web_search"] != true {
		t.Fatalf("provider capability is not true: %#v", provider)
	}
	features := decoded["features"].(map[string]any)
	if features["standalone_web_search"] != true {
		t.Fatalf("standalone-search feature is not true: %#v", features)
	}
	if features["apps"] != false || features["plugins"] != false {
		t.Fatalf("app/plugin switches changed: %#v", features)
	}
	if err := final.VerifyCodexGatewayProcess(b.ExecutablePath, final.Argv, append(slices.Clone(final.Env), GatewayTokenEnv+"=synthetic-child-only")); err != nil {
		t.Fatal(err)
	}
	g := (&System{}).CodexGatewayBasis(r, agentic.LaunchModeInteractive)
	if _, err := g.Render(b, append(slices.Clone(p.Argv), "unexpected")); !errors.Is(err, agentic.ErrCodexGatewayRefused) {
		t.Fatalf("opt-in renderer accepted a changed original suffix: %v", err)
	}
}

// Final process verification accepts only the exact two-setting opt-in form.
func TestGatewayStandaloneSearchVerifierRefusals(t *testing.T) {
	p := gatewayPlan(t, nil)
	b := gatewayBinding(t, p)
	b.SupportsStandaloneWebSearch = true
	a, err := gatewayAdmitFixture(p, agentic.CodexGatewayInputs{})
	if err != nil {
		t.Fatal(err)
	}
	final, err := gatewayFinalizeFixture(a, b)
	if err != nil {
		t.Fatal(err)
	}
	validEnv := append(slices.Clone(final.Env), GatewayTokenEnv+"=synthetic-child-only")
	mutations := map[string]func([]string) []string{
		"provider omitted": func(args []string) []string {
			args = slices.Clone(args)
			args[1] = strings.Replace(args[1], ",supports_standalone_web_search=true", "", 1)
			return args
		},
		"feature omitted": func(args []string) []string {
			out := slices.Clone(args[:gatewayOwnedArgs])
			return append(out, args[gatewayOwnedArgs+2:]...)
		},
		"provider false": func(args []string) []string {
			args = slices.Clone(args)
			args[1] = strings.Replace(args[1], "supports_standalone_web_search=true", "supports_standalone_web_search=false", 1)
			return args
		},
		"feature false": func(args []string) []string {
			args = slices.Clone(args)
			args[gatewayOwnedArgs+1] = "features.standalone_web_search=false"
			return args
		},
		"provider malformed": func(args []string) []string {
			args = slices.Clone(args)
			args[1] = strings.Replace(args[1], "supports_standalone_web_search=true", "supports_standalone_web_search=maybe", 1)
			return args
		},
		"provider duplicate": func(args []string) []string {
			out := append(slices.Clone(args[:2]), args[:2]...)
			return append(out, args[2:]...)
		},
		"feature malformed": func(args []string) []string {
			args = slices.Clone(args)
			args[gatewayOwnedArgs+1] = "features.standalone_web_search=maybe"
			return args
		},
		"feature duplicate": func(args []string) []string {
			out := slices.Clone(args[:gatewayOwnedArgs+2])
			out = append(out, "-c", gatewayStandaloneSearchFeature)
			return append(out, args[gatewayOwnedArgs+2:]...)
		},
		"feature misplaced": func(args []string) []string {
			out := slices.Clone(args[:4])
			out = append(out, args[gatewayOwnedArgs:gatewayOwnedArgs+2]...)
			out = append(out, args[4:gatewayOwnedArgs]...)
			return append(out, args[gatewayOwnedArgs+2:]...)
		},
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			if err := final.VerifyCodexGatewayProcess(b.ExecutablePath, mutate(final.Argv), validEnv); !errors.Is(err, agentic.ErrCodexGatewayRefused) {
				t.Fatalf("altered search argv was accepted: %v", err)
			}
		})
	}
}

// Direct argv stays byte-identical and caller-supplied gateway search keys refuse.
func TestGatewayStandaloneSearchDirectRefusals(t *testing.T) {
	baseline, err := Args(managedPrimaryRequest(t), agentic.LaunchModeInteractive)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(baseline, []string{"-m", "test-model"}) {
		t.Fatalf("direct argv changed: %q", baseline)
	}
	disabled := managedPrimaryRequest(t)
	disabled.NativeArgs = []string{"-c", "web_search_mode=disabled"}
	if _, err := Args(disabled, agentic.LaunchModeInteractive); err != nil {
		t.Fatalf("direct search-mode control was refused: %v", err)
	}
	settings := []string{
		"features.standalone_web_search=true",
		`"features"."standalone_web_search" = false`,
		`features = { standalone_web_search = true }`,
		"model_providers.curator-gateway.supports_standalone_web_search=true",
		`"model_providers"."curator-gateway"."supports_standalone_web_search"=false`,
		`model_providers = { curator-gateway = { supports_standalone_web_search = true } }`,
	}
	forms := []func(string) []string{
		func(setting string) []string { return []string{"-c", setting} },
		func(setting string) []string { return []string{"--config", setting} },
		func(setting string) []string { return []string{"--config=" + setting} },
		func(setting string) []string { return []string{"-c" + setting} },
		func(setting string) []string { return []string{"-c=" + setting} },
	}
	for _, setting := range settings {
		for _, form := range forms {
			req := managedPrimaryRequest(t)
			req.NativeArgs = form(setting)
			if _, err := Args(req, agentic.LaunchModeInteractive); !errors.Is(err, agentic.ErrCodexGatewayRefused) {
				t.Fatalf("direct gateway-search override %q was accepted: %v", req.NativeArgs, err)
			}
		}
	}
	unrelatedFeature := managedPrimaryRequest(t)
	unrelatedFeature.NativeArgs = []string{"-c", "features={apps=false}"}
	if _, err := Args(unrelatedFeature, agentic.LaunchModeInteractive); err != nil {
		t.Fatalf("direct unrelated feature override was refused: %v", err)
	}
	direct := gatewayPlan(t, nil)
	for _, setting := range settings {
		if err := direct.VerifyCodexGatewayProcess(direct.Binary, append(slices.Clone(direct.Argv), "-c", setting), nil); !errors.Is(err, agentic.ErrCodexGatewayRefused) {
			t.Fatalf("non-gateway verifier accepted %q: %v", setting, err)
		}
	}
}

// A disabled search mode in native args prevents an enabled gateway binding from finalizing.
func TestGatewayStandaloneSearchRefusesDisabledMode(t *testing.T) {
	r := managedPrimaryRequest(t)
	r.ToolRelease = "0.167.0"
	r.NativeArgs = []string{"-c", "web_search_mode=disabled"}
	p := requireManagedPlan(t, r)
	b := gatewayBinding(t, p)
	b.SupportsStandaloneWebSearch = true
	g := (&System{}).CodexGatewayBasis(r, agentic.LaunchModeInteractive)
	if _, err := g.Render(b, p.Argv); !errors.Is(err, agentic.ErrCodexGatewayRefused) {
		t.Fatalf("enabled search binding overrode disabled search mode: %v", err)
	}
}

// Binding fields cannot select a different provider, route, credential channel,
// transport or executable; release metadata is checked as plan identity elsewhere.
func TestGatewayBindingRefusals(t *testing.T) {
	for name, change := range map[string]func(*agentic.CodexGatewayBinding){
		"provider":         func(b *agentic.CodexGatewayBinding) { b.ProviderID = "openai" },
		"remote-base":      func(b *agentic.CodexGatewayBinding) { b.BaseURL = "https://example.invalid/chatgpt/backend-api/codex" },
		"double-base":      func(b *agentic.CodexGatewayBinding) { b.BaseURL += "/backend-api/codex" },
		"query":            func(b *agentic.CodexGatewayBinding) { b.BaseURL += "?secret=bad" },
		"auth-header":      func(b *agentic.CodexGatewayBinding) { b.HeaderName = "Authorization" },
		"header-injection": func(b *agentic.CodexGatewayBinding) { b.HeaderName = "X-Netproxy-Token\nAuthorization" },
		"token-env":        func(b *agentic.CodexGatewayBinding) { b.TokenEnv = "OPENAI_API_KEY" },
		"websocket":        func(b *agentic.CodexGatewayBinding) { b.SupportsWebSockets = true },
		"no-auth":          func(b *agentic.CodexGatewayBinding) { b.RequiresOpenAIAuth = false },
		"wire":             func(b *agentic.CodexGatewayBinding) { b.WireAPI = "chat" },
		"digest":           func(b *agentic.CodexGatewayBinding) { b.ExecutableSHA256 = strings.Repeat("0", 64) },
	} {
		t.Run(name, func(t *testing.T) {
			p := gatewayPlan(t, nil)
			a, err := gatewayAdmitFixture(p, agentic.CodexGatewayInputs{})
			if err != nil {
				t.Fatal(err)
			}
			b := gatewayBinding(t, p)
			b.AttemptID = a.AttemptID()
			if b.AttemptID == "" {
				t.Fatal("admission returned an empty attempt ID")
			}
			change(&b)
			final, receipt, err := a.Finalize(b)
			requireCodexGatewayReason(t, err, agentic.CodexGatewayBindingNotCanonical)
			if !reflect.DeepEqual(final, agentic.PlanWithEnvironment{}) {
				t.Fatalf("refused finalization returned a plan: %#v", final)
			}
			if receipt != nil {
				t.Fatalf("refused finalization returned a receipt: %#v", receipt)
			}

			valid := gatewayBinding(t, p)
			valid.AttemptID = a.AttemptID()
			if _, err := gatewayFinalizeFixture(a, valid); err != nil {
				t.Fatalf("valid binding refused after failed attempt: %v", err)
			}
		})
	}
}

// Input aliases, unsupported native commands and mutable pre-admission plans refuse.
func TestGatewayAdmissionRefusals(t *testing.T) {
	for _, native := range [][]string{{"app-server"}, {"--remote", "ws://127.0.0.1:1234"}, {"resume", "--last"}, {"-c", `model_provider="other"`}, {"--profile=alternate"}, {"--config", `"model_providers"."curator-gateway".base_url="http://127.0.0.1:1"`}} {
		p := gatewayPlan(t, native)
		_, err := gatewayAdmitFixture(p, agentic.CodexGatewayInputs{NativeArgs: native})
		requireCodexGatewayReason(t, err, agentic.CodexGatewayNativeArgumentsNotAdmitted)
	}
	for _, tc := range []struct {
		name, want agentic.CodexGatewayRefusalCode
		overlay    []string
	}{
		{"route", agentic.CodexGatewayProviderOverride, []string{"OPENAI_BASE_URL=http://127.0.0.1:1"}},
		{"route-alias", agentic.CodexGatewayProviderOverride, []string{"openai_base_url=http://127.0.0.1:1"}},
		{"token", agentic.CodexGatewayTokenOverride, []string{GatewayTokenEnv + "=synthetic"}},
		{"environment-alias", agentic.CodexGatewayProviderOverride, []string{"codex_home=/tmp/alias"}},
		{"duplicate", agentic.CodexGatewayEnvironmentNotAdmitted, []string{"X=one", "X=two"}},
	} {
		p := gatewayPlan(t, nil)
		_, err := gatewayAdmitFixture(p, agentic.CodexGatewayInputs{PromptEnv: tc.overlay})
		requireCodexGatewayReason(t, err, tc.want)
	}
	for _, names := range [][]string{{GatewayTokenEnv}, {strings.ToLower(GatewayTokenEnv)}, {"OPENAI_BASE_URL"}, {"CODEX_REMOTE_URL"}} {
		p := gatewayPlan(t, nil)
		_, err := gatewayAdmitFixture(p, agentic.CodexGatewayInputs{LookupNames: names})
		want := agentic.CodexGatewayProviderOverride
		if strings.Contains(strings.ToUpper(names[0]), "GATEWAY_TOKEN") {
			want = agentic.CodexGatewayTokenOverride
		}
		requireCodexGatewayReason(t, err, want)
	}
	p := gatewayPlan(t, nil)
	p.ModelIdentity.Launched = "changed"
	_, err := agentic.SnapshotGatewayCandidate(agentic.PlanWithEnvironment{Plan: p})
	requireCodexGatewayReason(t, err, agentic.CodexGatewayPlanIdentityChanged)
}

// The final checker rejects late argv, token alias/duplicate/missing, original
// executable and staged executable changes; no re-digest blesses new bytes.
func TestGatewayLateMutationRefusals(t *testing.T) {
	p := gatewayPlan(t, []string{"resume", managedThread})
	b := gatewayBinding(t, p)
	a, err := gatewayAdmitFixture(p, agentic.CodexGatewayInputs{NativeArgs: []string{"resume", managedThread}})
	if err != nil {
		t.Fatal(err)
	}
	final, err := gatewayFinalizeFixture(a, b)
	if err != nil {
		t.Fatal(err)
	}
	valid := append(slices.Clone(final.Env), GatewayTokenEnv+"=synthetic")
	args := slices.Clone(final.Argv)
	args[len(args)-1] = "01a14c36-65d0-7d80-9218-d374ede83a8a"
	if err := final.VerifyCodexGatewayProcess(b.ExecutablePath, args, valid); !errors.Is(err, agentic.ErrCodexGatewayRefused) {
		t.Fatal("resume mutation accepted", err)
	}
	for _, env := range [][]string{final.Env, append(slices.Clone(valid), GatewayTokenEnv+"=duplicate"), append(slices.Clone(final.Env), strings.ToLower(GatewayTokenEnv)+"=alias")} {
		if err := final.VerifyCodexGatewayProcess(b.ExecutablePath, final.Argv, env); !errors.Is(err, agentic.ErrCodexGatewayRefused) {
			t.Fatal("token mutation accepted", err)
		}
	}
	if err := os.WriteFile(p.Binary, []byte("#!/bin/sh\nexit 9\n"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := final.VerifyCodexGatewayProcess(b.ExecutablePath, final.Argv, valid); !errors.Is(err, agentic.ErrCodexGatewayRefused) {
		t.Fatal("changed executable accepted", err)
	}
}

// These pure runtime tests issue a fixture capability internally. Production
// issuance through the real limits store is tested in gatewayadmission.
func gatewayAdmitFixture(p agentic.Plan, in agentic.CodexGatewayInputs) (*agentic.CodexGatewayAdmission, error) {
	built := agentic.PlanWithEnvironment{Plan: p}
	candidate, err := agentic.SnapshotGatewayCandidate(built)
	if err != nil {
		return nil, err
	}
	basis, err := gatewaybasis.Issue(candidate)
	if err != nil {
		return nil, err
	}
	built.GatewayBasis = basis
	return agentic.AdmitCodexGateway(basis, built, in)
}
func gatewayFinalizeFixture(a *agentic.CodexGatewayAdmission, b agentic.CodexGatewayBinding) (agentic.Plan, error) {
	b.AttemptID = a.AttemptID()
	out, receipt, err := a.Finalize(b)
	if err != nil {
		return agentic.Plan{}, err
	}
	held, ok := receipt.BindingSnapshot()
	if !ok || held != b {
		return agentic.Plan{}, errors.New("receipt lost original attempt/binding")
	}
	return out.Plan, nil
}

// Canonical rendering refuses a different original suffix and unsupported mode;
// the ordinary fresh case beside each refusal remains renderable.
func TestGatewayRendererRefusesMismatch(t *testing.T) {
	r := managedPrimaryRequest(t)
	r.ToolRelease = "0.167.0"
	p := requireManagedPlan(t, r)
	b := gatewayBinding(t, p)
	g := (&System{}).CodexGatewayBasis(r, agentic.LaunchModeInteractive)
	if _, err := g.Render(b, p.Argv); err != nil {
		t.Fatal(err)
	}
	_, err := g.Render(b, append(slices.Clone(p.Argv), "different"))
	requireCodexGatewayReason(t, err, agentic.CodexGatewayBindingNotCanonical)
	g = (&System{}).CodexGatewayBasis(r, agentic.LaunchModeExec)
	_, err = g.Render(b, p.Argv)
	requireCodexGatewayReason(t, err, agentic.CodexGatewayBindingNotCanonical)
	requireCodexGatewayReason(t, g.Admit(agentic.CodexGatewayInputs{}, p.Env), agentic.CodexGatewayInteractiveModeRequired)
}

// Valid credentials and token injection pass; malformed, aliased, duplicate, missing-token and proxy inputs refuse.
func TestGatewayEnvironmentPredicate(t *testing.T) {
	g := gatewayPolicy{}
	for _, injected := range []bool{false, true} {
		env := []string{"OPENAI_API_KEY=synthetic", "OPENAI_ORG_ID=org", "OPENAI_PROJECT_ID=project", "CODEX_HOME=/private", "CODEX_MANAGED_BY_NPM=1", "DISJOINT=ok"}
		if injected {
			env = append(env, GatewayTokenEnv+"=synthetic")
		}
		if err := g.CheckEnvironment(env, injected); err != nil {
			t.Fatal(err)
		}
	}
	// Every selector/name predicate has the otherwise valid token-bearing control.
	for name, row := range map[string]struct {
		extra []string
		want  agentic.CodexGatewayRefusalCode
	}{
		"malformed":      {[]string{"malformed"}, agentic.CodexGatewayEnvironmentMalformed},
		"empty-name":     {[]string{"=empty"}, agentic.CodexGatewayEnvironmentMalformed},
		"newline":        {[]string{"X=line\nbreak"}, agentic.CodexGatewayEnvironmentMalformed},
		"fold-duplicate": {[]string{"X=one", "x=two"}, agentic.CodexGatewayEnvironmentDuplicate},
		"openai-route":   {[]string{"OPENAI_BASE_URL=other"}, agentic.CodexGatewayProviderOverride},
		"codex-route":    {[]string{"CODEX_REMOTE_URL=other"}, agentic.CodexGatewayProviderOverride},
	} {
		t.Run(name, func(t *testing.T) {
			control := []string{"DISJOINT=ok", GatewayTokenEnv + "=synthetic"}
			if err := g.CheckEnvironment(control, true); err != nil {
				t.Fatal(err)
			}
			err := g.CheckEnvironment(append(control, row.extra...), true)
			requireCodexGatewayReason(t, err, row.want)
		})
	}
	// Proxy refusal also applies to composed and token-injected environments, regardless of loopback bypass.
	for _, name := range []string{"HTTP_PROXY", "http_proxy", "HTTPS_PROXY", "https_proxy", "ALL_PROXY", "all_proxy"} {
		t.Run(name, func(t *testing.T) {
			for _, injected := range []bool{false, true} {
				for _, bypass := range []bool{false, true} {
					control := []string{"DISJOINT=ok", "CODEX_HOME=/private"}
					if injected {
						control = append(control, GatewayTokenEnv+"=synthetic")
					}
					if bypass {
						control = append(control, "NO_PROXY=127.0.0.1,::1,localhost")
					}
					if err := g.CheckEnvironment(control, injected); err != nil {
						t.Fatal(err)
					}
					err := g.CheckEnvironment(append(control, name+"=http://proxy.example.invalid:3128"), injected)
					requireCodexGatewayReason(t, err, agentic.CodexGatewayDirectProxyPresent)
				}
			}
		})
	}

	// Token rows isolate empty, aliased and missing injection from the name checks.
	for _, tc := range []struct {
		env  []string
		want agentic.CodexGatewayRefusalCode
	}{
		{[]string{GatewayTokenEnv + "="}, agentic.CodexGatewayTokenOverride},
		{[]string{strings.ToLower(GatewayTokenEnv) + "=alias"}, agentic.CodexGatewayTokenOverride},
		{[]string{"DISJOINT=ok"}, agentic.CodexGatewayInjectedTokenMissing},
	} {
		err := g.CheckEnvironment(tc.env, true)
		requireCodexGatewayReason(t, err, tc.want)
	}
}
