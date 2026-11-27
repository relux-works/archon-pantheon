package moonshot_test

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"testing"

	"github.com/relux-works/archon-pantheon/pkg/agentic"
	kimiSystem "github.com/relux-works/archon-pantheon/pkg/agentic/systems/kimi"
	"github.com/relux-works/archon-pantheon/pkg/runtimecatalog"
	"github.com/relux-works/archon-pantheon/pkg/vendorplugin"
	"github.com/relux-works/archon-pantheon/pkg/vendorplugin/vendors/moonshot"
)

func TestMoonshotKimiYoloAndCatalogLaunch(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell stub test requires unix")
	}
	dir := t.TempDir()
	stub := filepath.Join(dir, "kimi")
	if err := os.WriteFile(stub, []byte("#!/bin/sh\necho 2.8.1\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	const catalogTOML = `
schema_version = 1

[vendors.moonshot.models."kimi-custom-alias"]
description = "Operator custom Kimi Code alias"
systems = ["kimi"]
[vendors.moonshot.models."kimi-custom-alias".effort]
support = "none"
`
	cat, err := runtimecatalog.ParseModelCatalog([]byte(catalogTOML))
	if err != nil {
		t.Fatalf("ParseModelCatalog: %v", err)
	}
	systems := agentic.NewRegistry()
	if err := systems.Register(kimiSystem.New()); err != nil {
		t.Fatalf("register kimi: %v", err)
	}
	reg := vendorplugin.NewRegistry(systems)
	if err := reg.DeclareRuntime(moonshot.RuntimeDeclaration()); err != nil {
		t.Fatalf("DeclareRuntime: %v", err)
	}
	vendors, err := cat.ApplyAll([]vendorplugin.Vendor{moonshot.New()})
	if err != nil {
		t.Fatalf("ApplyAll: %v", err)
	}
	for _, v := range vendors {
		if err := reg.Register(v); err != nil {
			t.Fatalf("Register(%s): %v", v.ID(), err)
		}
	}

	home := filepath.Join(t.TempDir(), "kimi_code")
	env := []string{"PATH=" + dir}
	for _, modelID := range []vendorplugin.ModelID{"kimi-code/kimi-for-coding", "kimi-code/k3.1-pro", "kimi-custom-alias"} {
		plan, err := vendorplugin.BuildLaunch(context.Background(), reg, vendorplugin.SpawnRequest{
			Runtime:        moonshot.RuntimeID,
			Model:          modelID,
			ToolRelease:    "2.8.1",
			PermissionMode: agentic.PermissionModeYolo,
			Home:           home,
			Env:            env,
		}, agentic.LaunchModeInteractive)
		if err != nil {
			t.Fatalf("BuildLaunch(%s): %v", modelID, err)
		}
		want := []string{"--model", string(modelID), "--yolo"}
		if !slices.Equal(plan.Argv, want) {
			t.Fatalf("argv(%s) = %v, want %v", modelID, plan.Argv, want)
		}
	}
}
