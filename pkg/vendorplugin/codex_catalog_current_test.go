package vendorplugin_test

import (
	"context"
	"errors"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/relux-works/skill-agents-management/pkg/agentic"
	"github.com/relux-works/skill-agents-management/pkg/vendorplugin"
	_ "github.com/relux-works/skill-agents-management/pkg/vendorplugin/vendors/openai"
)

// This compatibility table proves the current Codex catalog's model ids are
// admitted through codex_cli with their exact ids, contexts, provenance, and
// declared efforts.
func TestCurrentCodexCatalogRowsResolveThroughCodexCLI(t *testing.T) {
	cases := []struct {
		model         vendorplugin.ModelID
		effort        string
		contextWindow int
		unmeasured    bool
	}{
		{"gpt-6.2-sol", "low", 500_000, true},
		{"gpt-6.2-astra", "low", 272_000, true},
		{"gpt-6.1-terra", "low", 500_000, true},
		{"gpt-reserve", "medium", 272_000, true},
		{"codex-auto-review", "medium", 272_000, true},
		{"gpt-6.1-sol", "medium", 272_000, false},
	}

	runtime, err := vendorplugin.Default.ResolveRuntime("codex")
	if err != nil {
		t.Fatalf("ResolveRuntime(codex): %v", err)
	}
	if runtime.SystemID != "codex" || runtime.VendorID != "openai" {
		t.Fatalf("codex runtime resolved to system/vendor %q/%q, want codex/openai", runtime.SystemID, runtime.VendorID)
	}

	for _, tc := range cases {
		t.Run(string(tc.model), func(t *testing.T) {
			req := vendorCuratorRequest(t, curatorVendorCase{
				runtime: "codex", model: tc.model, environment: "codex_cli", homeVariable: "CODEX_HOME",
			})
			req.Effort = tc.effort
			result, err := vendorplugin.BuildLaunchWithEnvironment(context.Background(), vendorplugin.Default, req, agentic.LaunchModeExec)
			if err != nil {
				t.Fatalf("BuildLaunchWithEnvironment(%q, effort=%q): %v", tc.model, tc.effort, err)
			}

			plan := result.Plan
			if plan.System != "codex" {
				t.Fatalf("plan resolved system %q, want codex", plan.System)
			}
			if plan.ModelIdentity.Requested != string(tc.model) || plan.ModelIdentity.Launched != string(tc.model) {
				t.Fatalf("model identity = %+v, want exact id %q with no alias", plan.ModelIdentity, tc.model)
			}
			argv := strings.Join(plan.Argv, " ")
			if !strings.Contains(argv, "-m "+string(tc.model)) || !strings.Contains(argv, `model_reasoning_effort="`+tc.effort+`"`) {
				t.Fatalf("argv %q does not carry the exact model and explicit catalog effort", plan.Argv)
			}
			provenance, ok := plan.CuratorContextProvenanceSnapshot()
			if !ok || provenance.Fragment.Environment != "codex_cli" {
				t.Fatalf("Curator context environment provenance = (%+v, %v), want codex_cli", provenance, ok)
			}

			var model vendorplugin.Model
			found := false
			for _, candidate := range runtime.Vendor.Models() {
				if candidate.ID == tc.model {
					model, found = candidate, true
					break
				}
			}
			if !found {
				t.Fatalf("resolved OpenAI vendor does not declare %q", tc.model)
			}
			if model.Effort.Recommended != tc.effort || !slices.Equal(model.Effort.Vocabulary, []string{"low", "medium", "high", "xhigh", "max"}) {
				t.Fatalf("effort declaration = %+v, want recommendation %q and the established non-ultra axis", model.Effort, tc.effort)
			}
			if model.ContextWindowTokens != tc.contextWindow || len(model.Systems) != 1 || model.Systems[0] != "codex" {
				t.Fatalf("model context/systems = %d/%v, want %d/[codex]", model.ContextWindowTokens, model.Systems, tc.contextWindow)
			}
			if model.AliasOf != "" {
				t.Fatalf("catalog row %q invents alias %q", tc.model, model.AliasOf)
			}
			if tc.unmeasured && (model.Unmeasured == nil || model.Rank != nil) {
				t.Fatalf("catalog-only row %q has rank/measurement claims: rank=%+v unmeasured=%+v", tc.model, model.Rank, model.Unmeasured)
			}
			if tc.unmeasured {
				if model.Unmeasured.Kind != vendorplugin.UnmeasuredClaimKind || model.Unmeasured.AsOf != "2026-11-24" || model.Unmeasured.Source != "OpenAI Codex CLI models_cache.json (codex-cli 0.167.0, read 2026-11-24)" {
					t.Fatalf("catalog provenance = %+v, want the current Codex CLI model-cache read", model.Unmeasured)
				}
				for _, want := range []string{
					string(tc.model),
					`default_reasoning_level "` + tc.effort + `"`,
					"context_window " + strconv.Itoa(tc.contextWindow),
				} {
					if !strings.Contains(model.Unmeasured.Observation, want) {
						t.Errorf("catalog observation %q omits source field %q", model.Unmeasured.Observation, want)
					}
				}
			}
			if !tc.unmeasured && (model.Unmeasured != nil || model.Rank == nil) {
				t.Fatalf("neighbor row %q changed its ranked declaration: rank=%+v unmeasured=%+v", tc.model, model.Rank, model.Unmeasured)
			}
		})
	}
}

// Unmeasured catalog rows use the vendor's declared recommendation when the
// caller omits effort. This table pins the source defaults through argv.
func TestCurrentCodexCatalogRowsUseDeclaredDefaultWhenEffortOmitted(t *testing.T) {
	for _, tc := range []struct {
		model  vendorplugin.ModelID
		effort string
	}{
		{"gpt-6.2-sol", "low"},
		{"gpt-6.2-astra", "low"},
		{"gpt-6.1-terra", "low"},
		{"gpt-reserve", "medium"},
		{"codex-auto-review", "medium"},
	} {
		t.Run(string(tc.model), func(t *testing.T) {
			req := vendorCuratorRequest(t, curatorVendorCase{
				runtime: "codex", model: tc.model, environment: "codex_cli", homeVariable: "CODEX_HOME",
			})
			req.Effort = ""
			result, err := vendorplugin.BuildLaunchWithEnvironment(context.Background(), vendorplugin.Default, req, agentic.LaunchModeExec)
			if err != nil {
				t.Fatalf("BuildLaunchWithEnvironment(%q, no effort): %v", tc.model, err)
			}
			argv := strings.Join(result.Plan.Argv, " ")
			if !strings.Contains(argv, `model_reasoning_effort="`+tc.effort+`"`) {
				t.Fatalf("omitted effort did not resolve to declared default %q: %v", tc.effort, result.Plan.Argv)
			}
		})
	}
}

// Ranked rows do not use Effort.Recommended as a fallback; the new unmeasured
// behavior must not change the neighboring gpt-6.1-sol contract.
func TestRankedCodexNeighborStillRequiresExplicitEffort(t *testing.T) {
	req := vendorCuratorRequest(t, curatorVendorCase{
		runtime: "codex", model: "gpt-6.1-sol", environment: "codex_cli", homeVariable: "CODEX_HOME",
	})
	req.Effort = ""
	result, err := vendorplugin.BuildLaunchWithEnvironment(context.Background(), vendorplugin.Default, req, agentic.LaunchModeExec)
	if !errors.Is(err, vendorplugin.ErrEffortMissing) {
		t.Fatalf("BuildLaunchWithEnvironment(gpt-6.1-sol, no effort) error = %v, want ErrEffortMissing", err)
	}
	if !reflect.DeepEqual(result, vendorplugin.BuildLaunchResult{}) {
		t.Fatalf("missing-effort refusal returned a partial plan: %+v", result)
	}
}

// The cache lists ultra for these models, but the module's accepted effort
// vocabulary excludes it because ultra is a delegation mode.
func TestCurrentCodexCatalogRowsRefuseUltraEffort(t *testing.T) {
	for _, model := range []vendorplugin.ModelID{"gpt-6.2-sol", "gpt-6.2-astra", "gpt-6.1-terra"} {
		t.Run(string(model), func(t *testing.T) {
			req := vendorCuratorRequest(t, curatorVendorCase{
				runtime: "codex", model: model, environment: "codex_cli", homeVariable: "CODEX_HOME",
			})
			req.Effort = "ultra"
			result, err := vendorplugin.BuildLaunchWithEnvironment(context.Background(), vendorplugin.Default, req, agentic.LaunchModeExec)
			if !errors.Is(err, vendorplugin.ErrEffortNotInVocabulary) {
				t.Fatalf("BuildLaunchWithEnvironment(%q, effort=ultra) error = %v, want ErrEffortNotInVocabulary", model, err)
			}
			if !reflect.DeepEqual(result, vendorplugin.BuildLaunchResult{}) {
				t.Fatalf("unsupported-effort refusal returned a partial plan: %+v", result)
			}
		})
	}
}

// The neighboring unknown id is a negative control: catalog refresh must not
// make version-shaped or prefix-matched spellings resolve.
func TestUnknownCodexCatalogModelStillRefuses(t *testing.T) {
	req := vendorCuratorRequest(t, curatorVendorCase{
		runtime: "codex", model: "gpt-6.9-sol", environment: "codex_cli", homeVariable: "CODEX_HOME",
	})
	req.Effort = "low"
	result, err := vendorplugin.BuildLaunchWithEnvironment(context.Background(), vendorplugin.Default, req, agentic.LaunchModeExec)
	if !errors.Is(err, vendorplugin.ErrUnknownModel) {
		t.Fatalf("BuildLaunchWithEnvironment(gpt-6.9-sol) error = %v, want ErrUnknownModel", err)
	}
	if !reflect.DeepEqual(result, vendorplugin.BuildLaunchResult{}) {
		t.Fatalf("unknown model refusal returned a partial plan: %+v", result)
	}
}
