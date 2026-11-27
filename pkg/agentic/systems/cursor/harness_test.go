package cursor

import (
	"errors"
	"reflect"
	"testing"

	"github.com/relux-works/archon-pantheon/pkg/harnessapi"
)

func refusalCode(err error) string {
	var refusal harnessapi.Refusal
	if errors.As(err, &refusal) {
		return refusal.Code()
	}
	return ""
}

func pathStrings(paths harnessapi.NamedPaths) []string {
	values := paths.Paths()
	got := make([]string, 0, len(values))
	for _, path := range values {
		got = append(got, path.String())
	}
	return got
}

func TestCursorAdapterDeclaresOnlyMigratedSurfaces(t *testing.T) {
	adapter := NewHarnessAdapter()
	if adapter.ID() != CursorHarnessID {
		t.Fatalf("ID() = %q, want %q", adapter.ID(), CursorHarnessID)
	}
	if adapter.APIVersion() != harnessapi.AdapterAPIVersionV1 {
		t.Fatalf("APIVersion() = %d, want %d", adapter.APIVersion(), harnessapi.AdapterAPIVersionV1)
	}
	wantCapabilities, err := harnessapi.NewCapabilitySet(harnessapi.SkillProjection, harnessapi.MCPInspection)
	if err != nil {
		t.Fatal(err)
	}
	if got := adapter.Capabilities().Values(); !reflect.DeepEqual(got, wantCapabilities.Values()) {
		t.Fatalf("Capabilities() = %v, want %v", got, wantCapabilities.Values())
	}
	if probe, err := adapter.ReleaseProbe(harnessapi.Context{}); err == nil || refusalCode(err) != harnessapi.CapabilityUnsupported || probe.ExecutableAlias() != "" {
		t.Fatalf("ReleaseProbe() = %#v, %v; want capability refusal", probe, err)
	}
	for _, scope := range []harnessapi.Scope{harnessapi.UserScope, harnessapi.ProjectScope} {
		skills, err := adapter.Skills(harnessapi.Context{}, scope)
		if err != nil {
			t.Fatalf("Skills(%q): %v", scope, err)
		}
		if got := pathStrings(skills.Paths()); !reflect.DeepEqual(got, []string{".cursor/rules"}) {
			t.Fatalf("Skills(%q).Paths() = %v", scope, got)
		}
		mcp, err := adapter.MCP(harnessapi.Context{}, scope)
		if err != nil {
			t.Fatalf("MCP(%q): %v", scope, err)
		}
		if got := pathStrings(mcp.Paths()); !reflect.DeepEqual(got, []string{".cursor/mcp.json"}) {
			t.Fatalf("MCP(%q).Paths() = %v", scope, got)
		}
	}
}
