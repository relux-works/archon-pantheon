package codex

import (
	"slices"
	"strings"
	"testing"

	"github.com/pelletier/go-toml/v2"
	"github.com/relux-works/skill-agents-management/pkg/agentic"
)

// Claim: every admitted gateway launch (fresh, exact resume, with or without
// the display switch) carries `-c features.apps=false` as its third owned
// pair, directly after the provider table and the provider selection, and the
// original arguments follow unchanged. Decoded as TOML it is the boolean
// false at features.apps and nothing else. The process verifier accepts the
// final arguments and refuses them once the pair is removed or switched on,
// so a launcher that calls it cannot start the child without the switch.
// The fixtures never launch Codex: this proves what is rendered and verified,
// not what a Codex release does with the setting.
func TestGatewayTurnsTheAppsConnectionOff(t *testing.T) {
	for _, suffix := range [][]string{nil, {"--no-alt-screen"}, {"resume", managedThread}, {"--no-alt-screen", "resume", managedThread}} {
		p := gatewayPlan(t, suffix)
		b := gatewayBinding(t, p)
		a, err := gatewayAdmitFixture(p, agentic.CodexGatewayInputs{NativeArgs: slices.Clone(suffix)})
		if err != nil {
			t.Fatal(err)
		}
		final, err := gatewayFinalizeFixture(a, b)
		if err != nil {
			t.Fatal(err)
		}
		if len(final.Argv) < 6 || !slices.Equal(final.Argv[4:6], []string{"-c", "features.apps=false"}) {
			t.Fatalf("gateway arguments carry no apps switch after the provider selection: %q", final.Argv[:min(len(final.Argv), 6)])
		}
		if !slices.Equal(final.Argv[gatewayOwnedArgs:], p.Argv) {
			t.Fatal("original arguments changed")
		}
		var decoded map[string]any
		if err := toml.Unmarshal([]byte(final.Argv[5]), &decoded); err != nil {
			t.Fatal(err)
		}
		features, ok := decoded["features"].(map[string]any)
		if !ok || len(decoded) != 1 || len(features) != 1 || features["apps"] != false {
			t.Fatalf("apps switch decodes to %#v", decoded)
		}
		count := 0
		for _, arg := range final.Argv {
			if strings.Contains(arg, "apps") {
				count++
			}
		}
		if count != 1 {
			t.Fatalf("%d arguments name apps, want the one owned switch", count)
		}
		env := append(slices.Clone(final.Env), GatewayTokenEnv+"=synthetic-child-only")
		if err := final.VerifyCodexGatewayProcess(b.ExecutablePath, final.Argv, env); err != nil {
			t.Fatal(err)
		}
		removed := append(slices.Clone(final.Argv[:4]), final.Argv[6:]...)
		switched := slices.Clone(final.Argv)
		switched[5] = "features.apps=true"
		appended := append(slices.Clone(final.Argv), "-c", "features.apps=true")
		for name, argv := range map[string][]string{"removed": removed, "switched-on": switched, "overridden-later": appended} {
			requireGatewayRefusedWithoutReason(t, name, final.VerifyCodexGatewayProcess(b.ExecutablePath, argv, env))
		}
	}
}

func requireGatewayRefusedWithoutReason(t *testing.T, name string, err error) {
	t.Helper()
	if err == nil {
		t.Fatalf("%s: process arguments accepted", name)
	}
}

// Claim: a launch without a gateway binding renders no apps setting, no other
// feature selector and none of the other settings a gateway launch owns, in
// every mode and with the request shapes that add `-c` pairs. The gateway
// switches are not defaults of the Codex plugin.
func TestNonGatewayLaunchesRenderNoAppsSetting(t *testing.T) {
	interactive := managedPrimaryRequest(t)
	interactive.ToolRelease = "0.167.0"
	interactive.Effort = "high"
	interactive.AdditionalWritableRoots = []string{t.TempDir()}
	plain := agentic.LaunchRequest{System: New().ID(), Model: agentic.Model{ID: parityModel}, Effort: "high", ServiceTier: "priority", WorkDir: "/tmp/project"}
	profiled := plain
	profiled.Profile = "delivery-profile"
	for name, tc := range map[string]struct {
		req  agentic.LaunchRequest
		mode agentic.LaunchMode
	}{
		"interactive":     {interactive, agentic.LaunchModeInteractive},
		"exec":            {plain, agentic.LaunchModeExec},
		"exec-profile":    {profiled, agentic.LaunchModeExec},
		"dry-run":         {plain, agentic.LaunchModeDryRun},
		"managed-session": {plain, agentic.LaunchModeManagedSession},
	} {
		t.Run(name, func(t *testing.T) {
			args, err := Args(tc.req, tc.mode)
			if err != nil {
				t.Fatal(err)
			}
			for _, arg := range args {
				if strings.Contains(arg, "features") || strings.Contains(arg, "apps") || arg == "--enable" || arg == "--disable" {
					t.Fatalf("non-gateway launch renders a feature setting: %q", arg)
				}
				for _, owned := range []string{"plugins", "analytics", "otel", "check_for_update"} {
					if strings.Contains(arg, owned) {
						t.Fatalf("non-gateway launch renders a gateway-owned setting: %q", arg)
					}
				}
			}
			if selectsOwnedSetting(args) {
				t.Fatalf("non-gateway launch sets what a gateway launch owns: %q", args)
			}
		})
	}
	// The built plan of an interactive launch is what the gateway later
	// prefixes: it holds no apps setting either.
	if p := requireManagedPlan(t, interactive); selectsOwnedSetting(p.Argv) || slices.ContainsFunc(p.Argv, func(arg string) bool { return strings.Contains(arg, "features") }) {
		t.Fatal("interactive plan renders a gateway-owned setting before any gateway binding")
	}
}

// Claim: a caller cannot switch the Apps connection back on. Every spelling
// of an apps selection in the native arguments is REFUSED at admission, with
// the native-arguments reason, and the argument constructor refuses the same
// request with a binding: nothing is overridden or silently dropped. The
// spellings are the override under both option names and the joined form,
// the whole table, a quoted key, the legacy feature name and the enable and
// disable flags, each alone and after the admitted display switch.
func TestGatewayRefusesACallerAppsSelection(t *testing.T) {
	for _, selection := range [][]string{
		{"-c", "features.apps=true"}, {"--config", "features.apps=true"}, {"--config=features.apps=true"},
		{"-c", `features={apps=true}`}, {"-c", `"features".apps=true`}, {"-c", "features.connectors=true"},
		{"--enable", "apps"}, {"--enable=apps"}, {"--disable", "apps"}, {"-c", "features.apps=false"},
	} {
		for _, lead := range [][]string{nil, {"--no-alt-screen"}} {
			native := append(slices.Clone(lead), selection...)
			t.Run(strings.Join(native, " "), func(t *testing.T) {
				p := gatewayPlan(t, native)
				_, err := gatewayAdmitFixture(p, agentic.CodexGatewayInputs{NativeArgs: slices.Clone(native)})
				requireCodexGatewayReason(t, err, agentic.CodexGatewayNativeArgumentsNotAdmitted)
				r := managedPrimaryRequest(t)
				r.ToolRelease = "0.167.0"
				r.NativeArgs = native
				b := gatewayBinding(t, p)
				r.CodexGateway = &b
				if args, err := Args(r, agentic.LaunchModeInteractive); err == nil {
					t.Fatalf("constructor rendered a gateway launch with a caller's selection: %q", args)
				}
			})
		}
	}
}
