package harnessapi

import (
	"reflect"
	"strings"
	"testing"

	"github.com/relux-works/archon-pantheon/pkg/harnessvalue"
)

func TestImmutableProbeSpec(t *testing.T) {
	args := []string{"--version"}
	env := map[string]string{"fixture": "fixed"}
	p := NewProbeSpec("fixture", args, ProbeTimeoutV1, env)
	args[0] = "--other"
	env["fixture"] = "other"
	p.Arguments()[0] = "--mutated"
	p.Environment()["fixture"] = "mutated"
	if !reflect.DeepEqual(p.Arguments(), []string{"--version"}) || p.Environment()["fixture"] != "fixed" {
		t.Fatal("probe aliases caller storage")
	}
	bad := NewProbeSpec(strings.Repeat("x", harnessvalue.MaxPathBytes+1), nil, ProbeTimeoutV1, nil)
	if bad.ValidValue() || len(bad.ExecutableAlias()) != 0 {
		t.Fatal("oversized probe retained or repaired")
	}
}

func TestSnapshotBoundary(t *testing.T) {
	property, err := NewProperty("seed_revision", "C")
	if err != nil {
		t.Fatal(err)
	}
	input := []Property{property}
	snapshot, err := NewSnapshot(input)
	if err != nil {
		t.Fatal(err)
	}
	input[0] = Property{}
	output := snapshot.Properties()
	output[0] = Property{}
	if snapshot.Properties()[0].Value() != "C" {
		t.Fatal("snapshot aliases storage")
	}
	for _, row := range [][2]string{{"access_token", "fixture"}, {"settings_mode", "raw-native-content"}, {"credential_route", "session-material"}, {"owned", "true|false"}} {
		if _, err := NewProperty(row[0], row[1]); err == nil {
			t.Fatal("unapproved observation accepted")
		}
	}
	if _, err := NewSnapshot([]Property{property, property}); err == nil {
		t.Fatal("duplicate field accepted")
	}
	if _, err := NewSnapshot([]Property{{}}); err == nil {
		t.Fatal("invalid zero property accepted")
	}
}

func TestPlanBoundary(t *testing.T) {
	path, err := harnessvalue.NewRelativePath("instructions/root.md")
	if err != nil {
		t.Fatal(err)
	}
	bytes := []byte("fixed")
	action, err := NewFileAction(path, bytes, "absent", 0o600)
	if err != nil {
		t.Fatal(err)
	}
	bytes[0] = 'X'
	action.Bytes()[0] = 'Y'
	if string(action.Bytes()) != "fixed" {
		t.Fatal("action aliases bytes")
	}
	actions := []FileAction{action}
	plan, err := NewSeedPlan(actions, nil)
	if err != nil {
		t.Fatal(err)
	}
	actions[0] = FileAction{}
	plan.Actions()[0] = FileAction{}
	if string(plan.Actions()[0].Bytes()) != "fixed" {
		t.Fatal("plan aliases actions")
	}
	if _, err := NewFileAction(harnessvalue.RelativePath{}, nil, "absent", 0o600); err == nil {
		t.Fatal("zero path accepted")
	}
	if _, err := NewFileAction(path, nil, "unknown", 0o600); err == nil {
		t.Fatal("invalid preimage accepted")
	}
	if _, err := NewFileAction(path, nil, "absent", 0o4600); err == nil {
		t.Fatal("special mode accepted")
	}
	if _, err := NewSeedPlan([]FileAction{action, action}, nil); err == nil {
		t.Fatal("duplicate write accepted")
	}
	if _, err := NewSeedPlan([]FileAction{action}, []Refusal{NewRefusal("refused")}); err == nil {
		t.Fatal("refused plan carries writes")
	}
}

func TestValueTypesHaveNoHandles(t *testing.T) {
	var inspect func(reflect.Type)
	seen := map[reflect.Type]bool{}
	inspect = func(typ reflect.Type) {
		if seen[typ] {
			return
		}
		seen[typ] = true
		switch typ.Kind() {
		case reflect.Func, reflect.Interface, reflect.Pointer, reflect.Chan, reflect.UnsafePointer:
			t.Fatalf("handle in value type: %v", typ)
		case reflect.Slice, reflect.Array:
			inspect(typ.Elem())
		case reflect.Map:
			inspect(typ.Key())
			inspect(typ.Elem())
		case reflect.Struct:
			for i := 0; i < typ.NumField(); i++ {
				f := typ.Field(i)
				if f.IsExported() {
					t.Fatalf("exported field: %v.%s", typ, f.Name)
				}
				inspect(f.Type)
			}
		}
	}
	for _, value := range []any{Context{}, Snapshot{}, Request{}, ProbeSpec{}, ProbeOutput{}, HomeLayout{}, InstructionSpec{}, SkillDiscovery{}, SettingsSpec{}, MCPChannel{}, CredentialRoute{}, TrustPolicy{}, SeedPlan{}, ProjectContextPlan{}, Status{}, DriftReport{}} {
		inspect(reflect.TypeOf(value))
	}
}

func TestContextAndRequestRows(t *testing.T) {
	release, err := NewVerifiedRelease("1.2.3")
	if err != nil {
		t.Fatal(err)
	}
	for _, platform := range []string{"darwin", "linux", "windows"} {
		for _, scope := range []Scope{UserScope, ProjectScope} {
			ctx, err := NewContext(platform, scope, release)
			if err != nil || ctx.Platform() != platform || ctx.Scope() != scope || ctx.Release().Version() != "1.2.3" {
				t.Fatalf("valid context: %v", err)
			}
			for _, form := range []Form{Monolithic, Referenced} {
				req, err := NewRequest(ManagedHome, form, scope)
				if err != nil || req.Operation() != ManagedHome || req.Form() != form || req.Scope() != scope {
					t.Fatalf("valid request: %v", err)
				}
			}
		}
	}
	for _, platform := range []string{"", "ios", "LINUX"} {
		if _, err := NewContext(platform, UserScope, release); err == nil {
			t.Fatal("unknown platform accepted")
		}
	}
	if _, err := NewContext("linux", Scope("unknown"), release); err == nil {
		t.Fatal("unknown context scope accepted")
	}
	for _, row := range []struct {
		operation Capability
		form      Form
		scope     Scope
	}{{0, Monolithic, UserScope}, {ReleaseProbing + 1, Monolithic, UserScope}, {ManagedHome, Form("unknown"), UserScope}, {ManagedHome, Monolithic, Scope("unknown")}} {
		if _, err := NewRequest(row.operation, row.form, row.scope); err == nil {
			t.Fatal("unknown request dimension accepted")
		}
	}
	set, err := NewCapabilitySet(ReleaseProbing)
	if err != nil || !set.Has(ReleaseProbing) || set.Has(ManagedHome) || set.Has(0) {
		t.Fatalf("closed capability set: %v", err)
	}
	values := set.Values()
	values[0] = ManagedHome
	if !set.Has(ReleaseProbing) || set.Has(ManagedHome) {
		t.Fatal("capability accessor mutated set")
	}
}
