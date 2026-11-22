package codex

import (
	"slices"
	"testing"

	"github.com/relux-works/skill-agents-management/pkg/agentic"
)

// Claim: selectsFeature recognises every argument form that switches a Codex
// feature (the enable and disable flags, and an override whose top-level key
// is features, quoted or spaced, under either option name and joined) and
// nothing else: keys that merely contain the word, values that mention it, a
// feature key under another table, and the profile flag that the Curator MCP
// channel renders in gateway launches all pass.
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
		if !selectsFeature(args) {
			t.Errorf("not recognised: %q", args)
		}
	}
	for _, args := range [][]string{
		nil, {"-m", "features"}, {"-c", `model="features.apps=true"`}, {"-c", `developer_instructions="set features.apps=true"`},
		{"-c", "mcp_servers.features.command=\"x\""}, {"-c", "mcp_servers.apps.enabled=true"}, {"-c", "feature.apps=true"},
		{"-c", "featuresx.apps=true"}, {"-c", "xfeatures.apps=true"}, {"-c", "model_reasoning_effort=\"high\""},
		{"-c", "sandbox_workspace_write.writable_roots=[\"/tmp/features\"]"}, {"--no-alt-screen"}, {"resume", "features"},
		{"--search", "--sandbox", "danger-full-access"}, {"-c"}, {"-p", "features"}, {"--profile", "managed"},
		{"-c", `profile="features"`}, {"-c", "profiles.x.features.apps=true"},
	} {
		if selectsFeature(args) {
			t.Errorf("wrongly recognised: %q", args)
		}
	}
}

// Claim: the arguments a gateway launch renders after its owned prefix hold
// no feature selector for the request shapes that add `-c` pairs (effort,
// and writable roots on a fresh launch) and for the admitted native suffixes,
// and the renderer accepts them. An original that differs by an appended apps override is refused as
// a different original; no request reaches the renderer's own selector guard,
// which is why that guard is recorded as defensive.
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
		if len(args) != gatewayOwnedArgs+len(p.Argv) || !slices.Equal(args[gatewayOwnedArgs-2:gatewayOwnedArgs], []string{"-c", gatewayAppsOff}) {
			t.Fatalf("owned prefix is not %d arguments ending in the apps switch", gatewayOwnedArgs)
		}
		if selectsFeature(args[gatewayOwnedArgs:]) || !selectsFeature(args[:gatewayOwnedArgs]) {
			t.Fatal("a feature selector outside the owned prefix, or none inside it")
		}
		_, err = g.Render(b, append(slices.Clone(p.Argv), "-c", "features.apps=true"))
		requireCodexGatewayReason(t, err, agentic.CodexGatewayBindingNotCanonical)
	}
}
