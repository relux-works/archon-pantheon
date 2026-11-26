package harnessregistry

import (
	"errors"
	"testing"

	"github.com/relux-works/curator-runtimes/pkg/harnessapi"
)

func code(err error) string {
	var refusal harnessapi.Refusal
	if errors.As(err, &refusal) {
		return refusal.Code()
	}
	return ""
}

type adapterSpy struct {
	harnessapi.UnportedOperations
	id                   harnessapi.ID
	version              harnessapi.APIVersion
	capabilities         harnessapi.CapabilitySet
	metadata, operations *int
}

func TestAPIVersionRefusal(t *testing.T) {
	for _, row := range []struct {
		name                   string
		registered, reported   harnessapi.APIVersion
		id                     harnessapi.ID
		capabilities           harnessapi.CapabilitySet
		want                   string
		constructors, metadata int
	}{
		{"registered-v2", 2, 2, "fixture", harnessapi.CapabilitySet{}, harnessapi.APIVersionUnsupported, 0, 0},
		{"reported-v2", 1, 2, "fixture", harnessapi.CapabilitySet{}, harnessapi.APIVersionUnsupported, 1, 1},
		{"correct-v1", 1, 1, "fixture", harnessapi.CapabilitySet{}, "", 1, 3},
		{"wrong-id", 1, 1, "other", harnessapi.CapabilitySet{}, harnessapi.RegistryMismatch, 1, 2},
	} {
		t.Run(row.name, func(t *testing.T) {
			constructors, metadata, operations := 0, 0, 0
			adapter, err := Verify(row.registered, "fixture", harnessapi.CapabilitySet{}, func() harnessapi.Adapter {
				constructors++
				return adapterSpy{id: row.id, version: row.reported, capabilities: row.capabilities, metadata: &metadata, operations: &operations}
			})
			if code(err) != row.want || constructors != row.constructors || metadata != row.metadata || operations != 0 {
				t.Fatalf("refusal=%v constructors=%d metadata=%d operations=%d", err, constructors, metadata, operations)
			}
			if row.want != "" && adapter != nil {
				t.Fatal("refused adapter escaped")
			}
			if row.want == "" {
				if adapter == nil {
					t.Fatal("valid adapter absent")
				}
				_, _ = adapter.ReleaseProbe(harnessapi.Context{})
				if operations != 1 {
					t.Fatal("positive operational spy not wired")
				}
			}
		})
	}
	metadata, operations := 0, 0
	claimed, _ := harnessapi.NewCapabilitySet(harnessapi.ManagedHome)
	_, err := Verify(1, "fixture", harnessapi.CapabilitySet{}, func() harnessapi.Adapter {
		return adapterSpy{id: "fixture", version: 1, capabilities: claimed, metadata: &metadata, operations: &operations}
	})
	if code(err) != harnessapi.RegistryMismatch || operations != 0 {
		t.Fatalf("capability mismatch: %v operations=%d", err, operations)
	}
	for alias, id := range map[harnessapi.ID]harnessapi.ID{"claude": "claude_code", "codex": "codex_cli"} {
		adapter, err := Lookup(alias)
		if err != nil || adapter.ID() != id {
			t.Fatalf("alias %s: %v", alias, err)
		}
	}
	for _, id := range []harnessapi.ID{"", "unknown", "CLAUDE_CODE", "kimi_code"} {
		if adapter, err := Lookup(id); code(err) != harnessapi.UnknownID || adapter != nil {
			t.Fatalf("unknown ID accepted: %s %v", id, err)
		}
	}
	for _, id := range IDs() {
		adapter, err := Lookup(id)
		if err != nil || adapter.ID() != id || adapter.APIVersion() != 1 {
			t.Fatalf("registration %s: %v", id, err)
		}
	}
	if DefaultID != "codex_cli" {
		t.Fatal("default changed")
	}
	for alias, want := range map[harnessapi.ID]string{
		"agy":         "antigravity",
		"antigravity": "antigravity",
		"claude":      "claude-code",
		"claude_code": "claude-code",
		"codex":       "codex",
		"codex_cli":   "codex",
		"kimi":        "kimi",
		"kimi_code":   "kimi",
		"gemini":      "gemini-cli",
		"muse":        "muse",
	} {
		got, err := ResolveSystemID(alias)
		if err != nil || string(got) != want {
			t.Fatalf("ResolveSystemID(%q) = %q, %v; want %q", alias, got, err, want)
		}
	}
	if specs := ManagedHomeSpecs(); len(specs) != 2 || specs[0].ID != "antigravity" || specs[1].ID != "kimi_code" {
		t.Fatalf("ManagedHomeSpecs() = %+v", specs)
	}
}

func (s adapterSpy) ID() harnessapi.ID                        { *s.metadata++; return s.id }
func (s adapterSpy) APIVersion() harnessapi.APIVersion        { *s.metadata++; return s.version }
func (s adapterSpy) Capabilities() harnessapi.CapabilitySet   { *s.metadata++; return s.capabilities }
func (s adapterSpy) ReleaseProbe(harnessapi.Context) (harnessapi.ProbeSpec, error) {
	*s.operations++
	return harnessapi.ProbeSpec{}, nil
}
func (s adapterSpy) ParseReleaseProbe(c harnessapi.Context, o harnessapi.ProbeOutput) (harnessapi.VerifiedRelease, error) {
	*s.operations++
	return s.UnportedOperations.ParseReleaseProbe(c, o)
}
func (s adapterSpy) Home(c harnessapi.Context) (harnessapi.HomeLayout, error) {
	*s.operations++
	return s.UnportedOperations.Home(c)
}
func (s adapterSpy) Instructions(c harnessapi.Context, f harnessapi.Form) (harnessapi.InstructionSpec, error) {
	*s.operations++
	return s.UnportedOperations.Instructions(c, f)
}
func (s adapterSpy) Skills(c harnessapi.Context, v harnessapi.Scope) (harnessapi.SkillDiscovery, error) {
	*s.operations++
	return s.UnportedOperations.Skills(c, v)
}
func (s adapterSpy) Settings(c harnessapi.Context) (harnessapi.SettingsSpec, error) {
	*s.operations++
	return s.UnportedOperations.Settings(c)
}
func (s adapterSpy) MCP(c harnessapi.Context, v harnessapi.Scope) (harnessapi.MCPChannel, error) {
	*s.operations++
	return s.UnportedOperations.MCP(c, v)
}
func (s adapterSpy) Credentials(c harnessapi.Context) (harnessapi.CredentialRoute, error) {
	*s.operations++
	return s.UnportedOperations.Credentials(c)
}
func (s adapterSpy) Trust(c harnessapi.Context) (harnessapi.TrustPolicy, error) {
	*s.operations++
	return s.UnportedOperations.Trust(c)
}
func (s adapterSpy) Preflight(c harnessapi.Context, v harnessapi.Snapshot, r harnessapi.Request) ([]harnessapi.Refusal, error) {
	*s.operations++
	return s.UnportedOperations.Preflight(c, v, r)
}
func (s adapterSpy) PlanSeed(c harnessapi.Context, v harnessapi.Snapshot, r harnessapi.Request) (harnessapi.SeedPlan, error) {
	*s.operations++
	return s.UnportedOperations.PlanSeed(c, v, r)
}
func (s adapterSpy) ProjectContext(c harnessapi.Context, v harnessapi.Snapshot, r harnessapi.Request) (harnessapi.ProjectContextPlan, error) {
	*s.operations++
	return s.UnportedOperations.ProjectContext(c, v, r)
}
func (s adapterSpy) Status(c harnessapi.Context, v harnessapi.Snapshot) (harnessapi.Status, error) {
	*s.operations++
	return s.UnportedOperations.Status(c, v)
}
func (s adapterSpy) Drift(c harnessapi.Context, v harnessapi.Snapshot) (harnessapi.DriftReport, error) {
	*s.operations++
	return s.UnportedOperations.Drift(c, v)
}
