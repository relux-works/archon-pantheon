package kimi

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/relux-works/curator-runtimes/internal/toolprobe"
	"github.com/relux-works/curator-runtimes/pkg/agentic"
)

// ProbeToolRelease runs the resolved CLI's version flag and returns an exact
// release token. Any failure remains undetected; callers must not synthesize
// a release to enable yolo.
func (*System) ProbeToolRelease(ctx context.Context, env []string) (string, error) {
	binary, err := resolveBinary(env)
	if err != nil {
		return "", errors.Join(agentic.ErrToolReleaseUndetected, fmt.Errorf("kimi: resolving the binary: %w", err))
	}
	out, err := toolprobe.VersionOutput(ctx, binary, env)
	if err != nil {
		return "", errors.Join(agentic.ErrToolReleaseUndetected, fmt.Errorf("kimi: probing the binary: %w", err))
	}
	line, _, _ := strings.Cut(string(out), "\n")
	fields := strings.Fields(line)
	if len(fields) == 0 || !toolprobe.IsReleaseTriple(fields[0]) {
		return "", errors.Join(agentic.ErrToolReleaseUndetected, fmt.Errorf("kimi: parsing the version answer: expected an exact release triple"))
	}
	return fields[0], nil
}
