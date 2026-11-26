package runtimecatalog_test

import (
	"context"
	"slices"
	"testing"

	"github.com/relux-works/curator-runtimes/pkg/agentic"
	agySystem "github.com/relux-works/curator-runtimes/pkg/agentic/systems/agy"
	geminiSystem "github.com/relux-works/curator-runtimes/pkg/agentic/systems/gemini"
	pinativeSystem "github.com/relux-works/curator-runtimes/pkg/agentic/systems/pinative"
	"github.com/relux-works/curator-runtimes/pkg/runtimecatalog"
	"github.com/relux-works/curator-runtimes/pkg/vendorplugin"
	"github.com/relux-works/curator-runtimes/pkg/vendorplugin/vendors/google"
)

func TestCatalogTomlAntigravityCustomModelAndEffortYoloLaunch(t *testing.T) {
	t.Parallel()

	const catalogTOML = `
schema_version = 1

[vendors.google.models."gemini-4-neon-custom"]
description = "Operator-configured Gemini 4 Neon model with explicit effort"
systems = ["antigravity"]
[vendors.google.models."gemini-4-neon-custom".effort]
support = "required"
accepts = ["low", "medium", "high", "xhigh", "max"]
recommended = "xhigh"
`
	cat, err := runtimecatalog.ParseModelCatalog([]byte(catalogTOML))
	if err != nil {
		t.Fatalf("ParseModelCatalog: %v", err)
	}

	systems := agentic.NewRegistry()
	if err := systems.Register(geminiSystem.New()); err != nil {
		t.Fatalf("register gemini: %v", err)
	}
	if err := systems.Register(pinativeSystem.New()); err != nil {
		t.Fatalf("register pi-native: %v", err)
	}
	if err := systems.Register(agySystem.NewWithRuntime(agySystem.Runtime{
		Executable: "/opt/agy/bin/agy",
		Version:    "1.5.18",
	})); err != nil {
		t.Fatalf("register agy: %v", err)
	}

	reg := vendorplugin.NewRegistry(systems)
	if err := vendorplugin.SeedFrozenRuntimes(reg); err != nil {
		t.Fatalf("SeedFrozenRuntimes: %v", err)
	}
	vendors, err := cat.ApplyAll([]vendorplugin.Vendor{google.New()})
	if err != nil {
		t.Fatalf("ApplyAll: %v", err)
	}
	for _, v := range vendors {
		if err := reg.Register(v); err != nil {
			t.Fatalf("Register(%s): %v", v.ID(), err)
		}
	}

	// 1. Operator-declared catalog.toml model with explicit reasoning effort + yolo on agy 1.5.18.
	plan, err := vendorplugin.BuildLaunch(context.Background(), reg, vendorplugin.SpawnRequest{
		Runtime:        "agy",
		Model:          "gemini-4-neon-custom",
		Effort:         "xhigh",
		ToolRelease:    "1.5.18",
		PermissionMode: agentic.PermissionModeYolo,
	}, agentic.LaunchModeInteractive)
	if err != nil {
		t.Fatalf("BuildLaunch custom model: %v", err)
	}
	want := []string{"--model", "gemini-4-neon-custom", "--effort", "xhigh", "--dangerously-skip-permissions"}
	if !slices.Equal(plan.Argv, want) {
		t.Fatalf("custom model argv = %v, want %v", plan.Argv, want)
	}

	// 2. Built-in gemini-4-neon-xhigh (effort-none) + yolo on agy 1.5.18 emits no --effort flag.
	neonPlan, err := vendorplugin.BuildLaunch(context.Background(), reg, vendorplugin.SpawnRequest{
		Runtime:        "agy",
		Model:          "gemini-4-neon-xhigh",
		ToolRelease:    "1.5.18",
		PermissionMode: agentic.PermissionModeYolo,
	}, agentic.LaunchModeInteractive)
	if err != nil {
		t.Fatalf("BuildLaunch gemini-4-neon-xhigh: %v", err)
	}
	wantNeon := []string{"--model", "gemini-4-neon-xhigh", "--dangerously-skip-permissions"}
	if !slices.Equal(neonPlan.Argv, wantNeon) {
		t.Fatalf("neon argv = %v, want %v", neonPlan.Argv, wantNeon)
	}
}
