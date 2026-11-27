package codex

import (
	"slices"
	"strings"
	"testing"

	"github.com/pelletier/go-toml/v2"
	"github.com/relux-works/archon-pantheon/pkg/agentic"
)

// ownedSwitchPairs is the expected text of the four owned switches, spelled
// here independently of the constants the renderer uses.
func ownedSwitchPairs() []string {
	return []string{
		"-c", "features.apps=false",
		"-c", "features.plugins=false",
		"-c", "analytics.enabled=false",
		"-c", "check_for_update_on_startup=false",
	}
}

// Claim: every admitted gateway launch (fresh, exact resume, with or without
// the display switch) carries the four owned switches as its third to sixth
// `-c` pairs, in one fixed order directly after the provider table and the
// provider selection, and the original arguments follow unchanged. Each
// switch decodes as TOML to the boolean false at its own key and nothing
// else, and each key is named by exactly one argument. The process verifier
// accepts the final arguments and refuses them once any one switch is
// removed, switched on, moved or overridden by a later pair, so a launcher
// that calls it cannot start the child without all four.
// The fixtures never launch Codex: this proves what is rendered and verified,
// not what a Codex release does with the settings.
func TestGatewayRendersTheOwnedSwitches(t *testing.T) {
	want := ownedSwitchPairs()
	paths := [][]string{{"features", "apps"}, {"features", "plugins"}, {"analytics", "enabled"}, {"check_for_update_on_startup"}}
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
		if len(final.Argv) != 4+len(want)+len(p.Argv) || !slices.Equal(final.Argv[4:4+len(want)], want) {
			t.Fatalf("gateway arguments do not carry the four switches after the provider selection: %q", final.Argv[:min(len(final.Argv), 4+len(want))])
		}
		if !slices.Equal(final.Argv[4+len(want):], p.Argv) {
			t.Fatal("original arguments changed")
		}
		if !strings.HasPrefix(final.Argv[1], "model_providers.") || !strings.HasPrefix(final.Argv[3], "model_provider=") {
			t.Fatalf("the switches do not follow the provider table and selection: %q", final.Argv[:4])
		}
		for i, path := range paths {
			var decoded any
			var table map[string]any
			if err := toml.Unmarshal([]byte(final.Argv[5+2*i]), &table); err != nil {
				t.Fatal(err)
			}
			decoded = table
			for _, key := range path {
				level, ok := decoded.(map[string]any)
				if !ok || len(level) != 1 {
					t.Fatalf("switch %q decodes to more than its own key: %#v", final.Argv[5+2*i], table)
				}
				decoded = level[key]
			}
			if decoded != false {
				t.Fatalf("switch %q does not decode to false at %v: %#v", final.Argv[5+2*i], path, table)
			}
			named := 0
			for _, arg := range final.Argv {
				if strings.Contains(arg, strings.Join(path, ".")) {
					named++
				}
			}
			if named != 1 {
				t.Fatalf("%d arguments name %v, want the one owned switch", named, path)
			}
		}
		env := append(slices.Clone(final.Env), GatewayTokenEnv+"=synthetic-child-only")
		if err := final.VerifyCodexGatewayProcess(b.ExecutablePath, final.Argv, env); err != nil {
			t.Fatal(err)
		}
		for i := range paths {
			at := 4 + 2*i
			removed := append(slices.Clone(final.Argv[:at]), final.Argv[at+2:]...)
			switched := slices.Clone(final.Argv)
			switched[at+1] = strings.TrimSuffix(switched[at+1], "false") + "true"
			overridden := append(slices.Clone(final.Argv), "-c", switched[at+1])
			moved := slices.Clone(final.Argv)
			other := 4 + 2*((i+1)%len(paths))
			moved[at+1], moved[other+1] = moved[other+1], moved[at+1]
			last := append(slices.Clone(removed), final.Argv[at:at+2]...)
			for name, argv := range map[string][]string{"removed": removed, "switched-on": switched, "overridden-later": overridden, "moved": moved, "moved-to-the-end": last} {
				requireGatewayRefusedWithoutReason(t, want[2*i+1]+" "+name, final.VerifyCodexGatewayProcess(b.ExecutablePath, argv, env))
			}
		}
	}
}

// Claim: a caller cannot set what a gateway launch owns besides the Apps
// switch. Every spelling of a plugins, analytics, exporter or update-check
// selection in the native arguments is REFUSED at admission, with the
// native-arguments reason, and the argument constructor refuses the same
// request with a binding: nothing is overridden or silently dropped. That
// holds for the value the launch renders itself as well. The refusal comes
// from the admitted native suffix, which lists what may follow and so names
// none of these keys; TestGatewayRefusesACallerAppsSelection is the same
// claim for the Apps switch.
func TestGatewayRefusesACallerSelectionOfAnOwnedSetting(t *testing.T) {
	for _, selection := range [][]string{
		{"-c", "analytics.enabled=true"}, {"--config", "analytics.enabled=true"}, {"--config=analytics.enabled=true"},
		{"-c", `analytics={enabled=true}`}, {"-c", `"analytics".enabled=true`}, {"-c", "analytics.enabled=false"},
		{"-c", `otel.metrics_exporter="statsig"`}, {"-c", `otel.exporter="otlp-http"`}, {"--config=otel.trace_exporter=\"none\""},
		{"-c", "check_for_update_on_startup=true"}, {"--config", "check_for_update_on_startup=true"},
		{"--config=check_for_update_on_startup=true"}, {"-c", "check_for_update_on_startup=false"},
		{"-c", "features.plugins=true"}, {"-c", `features={plugins=true}`}, {"-c", "features.plugins=false"},
		{"--enable", "plugins"}, {"--enable=plugins"}, {"--disable", "plugins"},
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
