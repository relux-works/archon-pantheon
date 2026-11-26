package claude

import (
	"reflect"
	"testing"

	"github.com/relux-works/curator-runtimes/pkg/harnessapi"
)

func claudeHarnessContext(t *testing.T, platform string, scope harnessapi.Scope) harnessapi.Context {
	t.Helper()
	release, err := harnessapi.NewVerifiedRelease("2.1.261")
	if err != nil {
		t.Fatal(err)
	}
	context, err := harnessapi.NewContext(platform, scope, release)
	if err != nil {
		t.Fatal(err)
	}
	return context
}

func harnessPaths(t *testing.T, named harnessapi.NamedPaths) []string {
	t.Helper()
	result := []string{}
	for _, path := range named.Paths() {
		result = append(result, path.String())
	}
	return result
}

func TestHarnessAdapterMetadataAndProbe(t *testing.T) {
	adapter := HarnessAdapter{}
	if adapter.ID() != "claude_code" || adapter.APIVersion() != harnessapi.AdapterAPIVersionV1 {
		t.Fatalf("adapter identity/version = %q/%d", adapter.ID(), adapter.APIVersion())
	}
	capabilities := adapter.Capabilities().Values()
	want := []harnessapi.Capability{
		harnessapi.InstructionOutput,
		harnessapi.SkillProjection,
		harnessapi.MCPInspection,
		harnessapi.NativeMCPWrite,
		harnessapi.CredentialMetadata,
		harnessapi.TrustObservation,
		harnessapi.ReleaseProbing,
	}
	if !reflect.DeepEqual(capabilities, want) {
		t.Fatalf("capabilities = %v, want %v", capabilities, want)
	}
	probe, err := adapter.ReleaseProbe(harnessapi.Context{})
	if err != nil || probe.ExecutableAlias() != "claude" || !reflect.DeepEqual(probe.Arguments(), []string{"--version"}) || probe.Timeout() != harnessapi.ProbeTimeoutV1 || len(probe.Environment()) != 0 {
		t.Fatalf("release probe = %#v, error %v", probe, err)
	}
}

func TestHarnessAdapterHomeInstructionsSkillsAndSettings(t *testing.T) {
	adapter := HarnessAdapter{}
	home, err := adapter.Home(harnessapi.Context{})
	if err != nil || home.EnvironmentVariable() != "CLAUDE_CONFIG_DIR" || !reflect.DeepEqual(harnessPaths(t, home.Roots()), []string{".claude"}) || home.ParentValued() {
		t.Fatalf("home = %#v, error %v", home, err)
	}
	for _, form := range []harnessapi.Form{harnessapi.Monolithic, harnessapi.Referenced} {
		instructions, err := adapter.Instructions(harnessapi.Context{}, form)
		if err != nil || instructions.Form() != form || !reflect.DeepEqual(harnessPaths(t, instructions.Paths()), []string{"CLAUDE.md"}) {
			t.Fatalf("instructions for %q = %#v, error %v", form, instructions, err)
		}
	}
	userSkills, err := adapter.Skills(harnessapi.Context{}, harnessapi.UserScope)
	if err != nil || !reflect.DeepEqual(harnessPaths(t, userSkills.Paths()), []string{"skills"}) || len(harnessPaths(t, userSkills.IgnorePaths())) != 0 || userSkills.NativeDiscovery() || !userSkills.Mirror() {
		t.Fatalf("user skills = %#v, error %v", userSkills, err)
	}
	projectSkills, err := adapter.Skills(harnessapi.Context{}, harnessapi.ProjectScope)
	if err != nil || !reflect.DeepEqual(harnessPaths(t, projectSkills.Paths()), []string{".claude/skills"}) || !reflect.DeepEqual(harnessPaths(t, projectSkills.IgnorePaths()), []string{".claude/skills"}) {
		t.Fatalf("project skills = %#v, error %v", projectSkills, err)
	}
	userSettings, err := adapter.Settings(claudeHarnessContext(t, "linux", harnessapi.UserScope))
	if err != nil || userSettings.Format() != harnessapi.JSON || !reflect.DeepEqual(harnessPaths(t, userSettings.Paths()), []string{"settings.json"}) {
		t.Fatalf("user settings = %#v, error %v", userSettings, err)
	}
	projectSettings, err := adapter.Settings(claudeHarnessContext(t, "linux", harnessapi.ProjectScope))
	if err != nil || projectSettings.Format() != harnessapi.JSON || !reflect.DeepEqual(harnessPaths(t, projectSettings.Paths()), []string{".claude/settings.json", ".claude/settings.local.json"}) {
		t.Fatalf("project settings = %#v, error %v", projectSettings, err)
	}
}
