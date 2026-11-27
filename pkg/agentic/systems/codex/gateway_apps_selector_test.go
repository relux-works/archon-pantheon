package codex

import (
	"slices"
	"testing"

	"github.com/relux-works/archon-pantheon/pkg/agentic"
)

// Claim: selectsOwnedSetting recognises every argument form that switches a
// Codex feature (the enable and disable flags, and an override whose top-level
// key is features, quoted or spaced, under either option name and joined) and
// nothing else: keys that merely contain the word or spell it in another
// case, values that mention it, a feature key under another table, and the
// profile flag that the Curator MCP channel renders in gateway launches all
// pass. TestSelectsOwnedTelemetryKeys holds the same for the other owned
// keys.
func TestSelectsFeature(t *testing.T) {
	for _, args := range [][]string{
		{"-c", "features.apps=true"}, {"-c", "features.apps=false"}, {"--config", "features.apps=true"},
		{"--config=features.apps=true"}, {"-cfeatures.apps=true"}, {"-c", "features={apps=true}"},
		{"-c", `"features".apps=true`}, {"-c", `'features'.apps=true`}, {"-c", " features . apps = true"},
		{"-c", "features.connectors=true"}, {"-c", "features.sleep_tool.mode=\"on\""}, {"-c", "features"},
		{"--enable", "apps"}, {"--enable=apps"}, {"--disable", "apps"}, {"--disable=apps"},
		{"-m", "model", "-c", "features.apps=true"}, {"-p", "managed", "-c", "features.apps=true"},
		{"-m", "model", "--enable", "apps"}, {"-c", "model=\"m\"", "--config", "features.apps=true"},
	} {
		if !selectsOwnedSetting(args) {
			t.Errorf("not recognised: %q", args)
		}
	}
	for _, args := range [][]string{
		nil, {"-m", "features"}, {"-c", `model="features.apps=true"`}, {"-c", `developer_instructions="set features.apps=true"`},
		{"-c", "mcp_servers.features.command=\"x\""}, {"-c", "mcp_servers.apps.enabled=true"}, {"-c", "feature.apps=true"},
		{"-c", "featuresx.apps=true"}, {"-c", "xfeatures.apps=true"}, {"-c", "Features.apps=true"}, {"-c", "model_reasoning_effort=\"high\""},
		{"-c", "sandbox_workspace_write.writable_roots=[\"/tmp/features\"]"}, {"--no-alt-screen"}, {"resume", "features"},
		{"--search", "--sandbox", "danger-full-access"}, {"-c"}, {"-p", "features"}, {"--profile", "managed"},
		{"-c", `profile="features"`}, {"-c", "profiles.x.features.apps=true"},
	} {
		if selectsOwnedSetting(args) {
			t.Errorf("wrongly recognised: %q", args)
		}
	}
}

// Claim: selectsOwnedSetting recognises an override of each other owned key
// (analytics, the update check, and otel, which names another exporter than
// the built-in one `analytics.enabled` concerns) in the same forms as a feature
// override: either option name, joined, the whole table, quoted or spaced.
// It recognises nothing else: a key that merely contains or extends the
// word, a value that mentions it, the same word under another table, and a
// model or profile of that name all pass. This proves the predicate only;
// no request reaches it with such an argument (TestGatewayRenderedTailSelectsNoFeature).
func TestSelectsOwnedTelemetryKeys(t *testing.T) {
	for _, args := range [][]string{
		{"-c", "analytics.enabled=true"}, {"-c", "analytics.enabled=false"}, {"--config", "analytics.enabled=true"},
		{"--config=analytics.enabled=true"}, {"-canalytics.enabled=true"}, {"-c", "analytics={enabled=true}"},
		{"-c", `"analytics".enabled=true`}, {"-c", `'analytics'.enabled=true`}, {"-c", " analytics . enabled = true"},
		{"-c", "analytics"}, {"-c", `otel.metrics_exporter="statsig"`}, {"-c", `otel.exporter="otlp-http"`},
		{"-c", `otel={metrics_exporter="statsig"}`}, {"--config=otel.environment=\"x\""}, {"-c", `"otel".trace_exporter="none"`},
		{"-c", "check_for_update_on_startup=true"}, {"-c", "check_for_update_on_startup=false"},
		{"--config", "check_for_update_on_startup=true"}, {"--config=check_for_update_on_startup=true"},
		{"-ccheck_for_update_on_startup=true"}, {"-c", `"check_for_update_on_startup"=true`},
		{"-c", " check_for_update_on_startup = true"}, {"-c", "features.plugins=true"}, {"--enable", "plugins"},
		{"-m", "model", "-c", "analytics.enabled=true"}, {"-p", "managed", "-c", "check_for_update_on_startup=true"},
		{"-c", "model=\"m\"", "--config", `otel.metrics_exporter="statsig"`},
	} {
		if !selectsOwnedSetting(args) {
			t.Errorf("not recognised: %q", args)
		}
	}
	for _, args := range [][]string{
		{"-m", "analytics"}, {"-p", "otel"}, {"--profile", "check_for_update_on_startup"},
		{"-c", `model="analytics.enabled=true"`}, {"-c", `developer_instructions="set check_for_update_on_startup=true"`},
		{"-c", "mcp_servers.analytics.command=\"x\""}, {"-c", "mcp_servers.otel.enabled=true"},
		{"-c", "analyticsx.enabled=true"}, {"-c", "xanalytics.enabled=true"}, {"-c", "analytic.enabled=true"},
		{"-c", "otelx.exporter=\"none\""}, {"-c", "hotel.exporter=\"none\""},
		{"-c", "check_for_update=true"}, {"-c", "check_for_update_on_startup_x=true"},
		{"-c", "tui.check_for_update_on_startup=true"}, {"-c", "profiles.x.analytics.enabled=true"},
		{"-c", `profile="analytics"`}, {"resume", "analytics"},
		{"-c", "Analytics.enabled=true"}, {"-c", `OTEL.metrics_exporter="statsig"`}, {"-c", "Check_For_Update_On_Startup=true"},
	} {
		if selectsOwnedSetting(args) {
			t.Errorf("wrongly recognised: %q", args)
		}
	}
}

// Claim: selectsOwnedSetting recognises an override of each owned key in the
// spelling `-c=key=value`, where a command-line parser strips the equals sign
// after the short option, alone and after other arguments. Each recognised
// row is paired with one that differs only in the key and is not owned: an
// unrelated key, the owned word extended (`featuresx`), and the owned word in
// another case, since the client's keys are case-sensitive. A second equals
// sign leaves an empty key, which is no owned key. This proves the predicate
// only: no test launches Codex, so what the client's own parser accepts in
// this spelling is not established here.
func TestSelectsOwnedSettingAfterShortEquals(t *testing.T) {
	for _, row := range []struct{ owned, other string }{
		{"-c=features.apps=true", "-c=model=\"m\""},
		{"-c=features.plugins=true", "-c=featuresx.plugins=true"},
		{"-c=features={plugins=true}", "-c=Features.plugins=true"},
		{"-c=analytics.enabled=true", "-c=analyticsx.enabled=true"},
		{"-c=\"analytics\".enabled=true", "-c=Analytics.enabled=true"},
		{"-c=otel.metrics_exporter=\"statsig\"", "-c=hotel.metrics_exporter=\"statsig\""},
		{"-c=check_for_update_on_startup=true", "-c=check_for_update=true"},
		{"-c= check_for_update_on_startup = true", "-c=CHECK_FOR_UPDATE_ON_STARTUP=true"},
		{"-c=features.plugins=false", "-c==features.plugins=false"},
		{"-c=analytics", "-c=mcp_servers.analytics.command=\"x\""},
	} {
		for _, before := range [][]string{nil, {"-m", "model"}, {"-c", "model=\"m\""}} {
			if args := append(slices.Clone(before), row.owned); !selectsOwnedSetting(args) {
				t.Errorf("not recognised: %q", args)
			}
			if args := append(slices.Clone(before), row.other); selectsOwnedSetting(args) {
				t.Errorf("wrongly recognised: %q", args)
			}
		}
	}
}

// Claim: the arguments a gateway launch renders after its owned prefix hold
// no feature selector and no other owned setting for the request shapes that
// add `-c` pairs (effort, and writable roots on a fresh launch) and for the
// admitted native suffixes, and the renderer accepts them. An original that
// differs by an appended override of an owned key is refused as a different
// original; no request reaches the renderer's own selector guard, which is
// why that guard is recorded as defensive.
func TestGatewayRenderedTailSelectsNoFeature(t *testing.T) {
	r := managedPrimaryRequest(t)
	r.ToolRelease = "0.167.0"
	r.Effort = "high"
	roots := []string{t.TempDir()}
	for _, suffix := range [][]string{nil, {"--no-alt-screen"}, {"resume", managedThread}} {
		r.NativeArgs = suffix
		// Writable roots belong to a fresh launch; with native arguments they refuse.
		r.AdditionalWritableRoots = nil
		if suffix == nil {
			r.AdditionalWritableRoots = roots
		}
		p := requireManagedPlan(t, r)
		b := gatewayBinding(t, p)
		g := (&System{}).CodexGatewayBasis(r, agentic.LaunchModeInteractive)
		args, err := g.Render(b, p.Argv)
		if err != nil {
			t.Fatal(err)
		}
		if len(args) != gatewayOwnedArgs+len(p.Argv) || !slices.Equal(args[4:gatewayOwnedArgs], ownedSwitchPairs()) {
			t.Fatalf("owned prefix is not %d arguments ending in the four switches", gatewayOwnedArgs)
		}
		if selectsOwnedSetting(args[gatewayOwnedArgs:]) || selectsOwnedSetting(args[:4]) {
			t.Fatal("an owned setting outside the four switches")
		}
		for i := 4; i < gatewayOwnedArgs; i += 2 {
			if !selectsOwnedSetting(args[i : i+2]) {
				t.Fatalf("owned switch %q is not recognised as one", args[i+1])
			}
		}
		for _, override := range []string{"features.apps=true", "features.plugins=true", "analytics.enabled=true", "check_for_update_on_startup=true", `otel.metrics_exporter="statsig"`} {
			_, err = g.Render(b, append(slices.Clone(p.Argv), "-c", override))
			requireCodexGatewayReason(t, err, agentic.CodexGatewayBindingNotCanonical)
		}
	}
}
