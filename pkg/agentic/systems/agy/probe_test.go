package agy

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/relux-works/curator-runtimes/pkg/agentic"
)

func TestAgyPermissionMapping(t *testing.T) {
	sys := New()
	native, err := sys.PermissionMapping(agyVerifiedRelease, agentic.PermissionModeNative)
	if err != nil || native.Grammar != agentic.PermissionGrammarV1 || native.Flag != "" {
		t.Fatalf("native mapping = (%+v, %v), want v1 with empty flag", native, err)
	}
	yolo, err := sys.PermissionMapping(agyVerifiedRelease, agentic.PermissionModeYolo)
	if err != nil || yolo.Grammar != agentic.PermissionGrammarV1 || yolo.Flag != agyYoloFlag {
		t.Fatalf("yolo mapping = (%+v, %v), want v1 with %s", yolo, err, agyYoloFlag)
	}
	if _, err := sys.PermissionMapping("1.5.17", agentic.PermissionModeYolo); !errors.Is(err, agentic.ErrPermissionModeUnverifiedRelease) {
		t.Fatalf("unverified release err = %v, want ErrPermissionModeUnverifiedRelease", err)
	}
	if _, err := sys.PermissionMapping(agyVerifiedRelease, agentic.PermissionMode("auto")); !errors.Is(err, agentic.ErrPermissionModeUnknown) {
		t.Fatalf("unknown mode err = %v, want ErrPermissionModeUnknown", err)
	}
}

func TestAgyProbeRuntimeAndToolRelease(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell stub test requires unix")
	}
	dir := t.TempDir()
	stub := filepath.Join(dir, "agy")
	if err := os.WriteFile(stub, []byte("#!/bin/sh\necho 1.5.16\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	env := []string{"PATH=" + dir}
	rt, err := ProbeRuntime(context.Background(), env)
	if err != nil {
		t.Fatalf("ProbeRuntime: %v", err)
	}
	if rt.Executable != stub || rt.Version != "1.5.16" {
		t.Fatalf("ProbeRuntime = %+v, want {%s 1.5.16}", rt, stub)
	}
	rel, err := agentic.ProbeToolRelease(context.Background(), New(), env)
	if err != nil || rel != "1.5.16" {
		t.Fatalf("ProbeToolRelease(New) = (%q, %v), want (1.5.16, nil)", rel, err)
	}
	relPreflighted, err := agentic.ProbeToolRelease(context.Background(), NewWithRuntime(rt), nil)
	if err != nil || relPreflighted != "1.5.16" {
		t.Fatalf("ProbeToolRelease(NewWithRuntime) = (%q, %v), want (1.5.16, nil)", relPreflighted, err)
	}

	// Missing binary refuses with ErrToolReleaseUndetected.
	if _, err := ProbeRuntime(context.Background(), []string{"PATH=" + t.TempDir()}); !errors.Is(err, agentic.ErrToolReleaseUndetected) {
		t.Fatalf("missing binary err = %v, want ErrToolReleaseUndetected", err)
	}

	// Malformed version answer refuses with ErrToolReleaseUndetected.
	badDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(badDir, "agy"), []byte("#!/bin/sh\necho not-a-version\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := ProbeRuntime(context.Background(), []string{"PATH=" + badDir}); !errors.Is(err, agentic.ErrToolReleaseUndetected) {
		t.Fatalf("malformed version err = %v, want ErrToolReleaseUndetected", err)
	}

	// Failing binary refuses with ErrToolReleaseUndetected.
	failDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(failDir, "agy"), []byte("#!/bin/sh\nexit 2\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := ProbeRuntime(context.Background(), []string{"PATH=" + failDir}); !errors.Is(err, agentic.ErrToolReleaseUndetected) {
		t.Fatalf("failing binary err = %v, want ErrToolReleaseUndetected", err)
	}
}
