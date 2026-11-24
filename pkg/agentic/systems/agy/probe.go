package agy

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/relux-works/skill-agents-management/internal/launchenv"
	"github.com/relux-works/skill-agents-management/internal/toolprobe"
	"github.com/relux-works/skill-agents-management/pkg/agentic"
)

var _ agentic.ToolReleaseProber = (*System)(nil)

// ProbeRuntime locates agy on env's PATH and runs `agy --version` under the
// bounded shared toolprobe, returning preflight Runtime evidence for
// NewWithRuntime. System.ResolveBinary and System.Argv retain their strict
// no-PATH-fallback contract; launchers call ProbeRuntime explicitly before
// constructing a preflighted System.
func ProbeRuntime(ctx context.Context, env []string) (Runtime, error) {
	binary, err := launchenv.LookPath(env, displayPlaceholder)
	if err != nil {
		return Runtime{}, errors.Join(agentic.ErrToolReleaseUndetected, fmt.Errorf("agy: resolving the binary: %w", err))
	}
	var release string
	if parsed, err := probeVersion(ctx, binary, env); err != nil {
		return Runtime{}, err
	} else {
		release = parsed
	}
	return Runtime{Executable: binary, Version: release}, nil
}

// ProbeToolRelease runs `<binary> --version` against the preflighted binary
// when present, or locates agy on env's PATH otherwise. Any failure remains
// undetected; callers never synthesize a release to enable yolo.
func (s *System) ProbeToolRelease(ctx context.Context, env []string) (string, error) {
	if !s.runtime.IsZero() {
		return probeVersion(ctx, s.binary(), env)
	}
	var rt Runtime
	if probed, err := ProbeRuntime(ctx, env); err != nil {
		return "", err
	} else {
		rt = probed
	}
	return rt.Version, nil
}

func probeVersion(ctx context.Context, binary string, env []string) (string, error) {
	out, err := toolprobe.VersionOutput(ctx, binary, env)
	if err != nil {
		return "", errors.Join(agentic.ErrToolReleaseUndetected, fmt.Errorf("agy: probing the binary: %w", err))
	}
	line, _, _ := strings.Cut(string(out), "\n")
	fields := strings.Fields(line)
	if len(fields) == 0 || !toolprobe.IsReleaseTriple(fields[0]) {
		return "", errors.Join(agentic.ErrToolReleaseUndetected, fmt.Errorf("agy: parsing the version answer: expected an exact release triple"))
	}
	return fields[0], nil
}
