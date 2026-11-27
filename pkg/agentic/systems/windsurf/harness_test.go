package windsurf

import (
	"errors"
	"reflect"
	"testing"

	"github.com/relux-works/archon-pantheon/pkg/harnessapi"
)

func code(err error) string {
	var refusal harnessapi.Refusal
	if errors.As(err, &refusal) {
		return refusal.Code()
	}
	return ""
}

func stringsFrom(t *testing.T, paths harnessapi.NamedPaths) []string {
	t.Helper()
	values := paths.Paths()
	result := make([]string, 0, len(values))
	for _, value := range values {
		if !value.Valid() {
			t.Fatalf("invalid returned path %q", value.String())
		}
		result = append(result, value.String())
	}
	return result
}

func TestRegisteredSurfaceAndSkillScopes(t *testing.T) {
	adapter := NewHarnessAdapter()
	if adapter.ID() != WindsurfHarnessID || adapter.APIVersion() != harnessapi.AdapterAPIVersionV1 {
		t.Fatalf("adapter metadata: %s v%d", adapter.ID(), adapter.APIVersion())
	}
	wantCapabilities, err := harnessapi.NewCapabilitySet(harnessapi.NativeSkillDiscovery, harnessapi.MCPInspection)
	if err != nil {
		t.Fatal(err)
	}
	if adapter.Capabilities() != wantCapabilities {
		t.Fatalf("capabilities: %v", adapter.Capabilities().Values())
	}
	for _, row := range []struct {
		scope  harnessapi.Scope
		mirror bool
	}{
		{harnessapi.UserScope, true},
		{harnessapi.ProjectScope, false},
	} {
		discovery, err := adapter.Skills(harnessapi.Context{}, row.scope)
		if err != nil {
			t.Fatalf("skills %s: %v", row.scope, err)
		}
		if got := stringsFrom(t, discovery.Paths()); !reflect.DeepEqual(got, []string{".agents/skills"}) {
			t.Fatalf("paths %s: %v", row.scope, got)
		}
		if got := stringsFrom(t, discovery.IgnorePaths()); !reflect.DeepEqual(got, []string{".agents"}) {
			t.Fatalf("ignore paths %s: %v", row.scope, got)
		}
		if !discovery.NativeDiscovery() || discovery.Mirror() != row.mirror || discovery.Scope() != row.scope {
			t.Fatalf("discovery policy %s: native=%t mirror=%t", row.scope, discovery.NativeDiscovery(), discovery.Mirror())
		}
	}
	home, err := adapter.MCP(harnessapi.Context{}, harnessapi.UserScope)
	if err != nil || !reflect.DeepEqual(stringsFrom(t, home.Paths()), []string{".codeium/windsurf/mcp_config.json"}) {
		t.Fatalf("home MCP: %#v, %v", home, err)
	}
	if _, err := adapter.ReleaseProbe(harnessapi.Context{}); code(err) != harnessapi.CapabilityUnsupported {
		t.Fatalf("release probe refusal: %v", err)
	}
}
