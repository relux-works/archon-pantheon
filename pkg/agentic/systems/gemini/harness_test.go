package gemini

import (
	"errors"
	"reflect"
	"testing"

	"github.com/relux-works/curator-runtimes/pkg/harnessapi"
)

func contextFor(t *testing.T, scope harnessapi.Scope) harnessapi.Context {
	t.Helper()
	context, err := harnessapi.NewContext("darwin", scope, harnessapi.VerifiedRelease{})
	if err != nil {
		t.Fatal(err)
	}
	return context
}

func pathsOf(paths harnessapi.NamedPaths) []string {
	values := paths.Paths()
	result := make([]string, len(values))
	for i, value := range values {
		result[i] = value.String()
	}
	return result
}

func refusalCode(err error) string {
	var refusal harnessapi.Refusal
	if errors.As(err, &refusal) {
		return refusal.Code()
	}
	return ""
}

func TestHarnessAdapterMetadataAndUnsupportedProbe(t *testing.T) {
	adapter := NewHarnessAdapter()
	if adapter.ID() != GeminiHarnessID || adapter.APIVersion() != harnessapi.AdapterAPIVersionV1 {
		t.Fatalf("adapter metadata = (%q, %d)", adapter.ID(), adapter.APIVersion())
	}
	want, err := harnessapi.NewCapabilitySet(harnessapi.SkillProjection, harnessapi.MCPInspection)
	if err != nil {
		t.Fatal(err)
	}
	if adapter.Capabilities() != want || !HarnessCapabilities().Has(harnessapi.SkillProjection) ||
		!HarnessCapabilities().Has(harnessapi.MCPInspection) || HarnessCapabilities().Has(harnessapi.NativeMCPWrite) {
		t.Fatalf("capabilities = %v", adapter.Capabilities().Values())
	}
	if _, err := adapter.ReleaseProbe(contextFor(t, harnessapi.UserScope)); refusalCode(err) != harnessapi.CapabilityUnsupported {
		t.Fatalf("release probe refusal = %v", err)
	}
	for _, scope := range []harnessapi.Scope{harnessapi.ProjectScope, harnessapi.UserScope} {
		discovery, err := adapter.Skills(contextFor(t, scope), scope)
		if err != nil || !reflect.DeepEqual(pathsOf(discovery.Paths()), []string{".gemini/skills"}) {
			t.Fatalf("skills(%s) = %#v, %v", scope, discovery, err)
		}
		channel, err := adapter.MCP(contextFor(t, scope), scope)
		if err != nil || !reflect.DeepEqual(pathsOf(channel.Paths()), []string{".gemini/settings.json"}) {
			t.Fatalf("mcp(%s) = %#v, %v", scope, channel, err)
		}
	}
}
