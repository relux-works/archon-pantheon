package agy

import (
	"fmt"

	"github.com/relux-works/curator-runtimes/pkg/agentic"
)

var _ agentic.PermissionMappingCapability = (*System)(nil)

var verifiedReleases = []agentic.ReleaseCapability{
	{
		MinRelease:          agyMinVerifiedRelease,
		MaxReleaseExclusive: agyMaxReleaseExclusive,
		Grammar:             agentic.PermissionGrammarV1,
		YoloSupported:       true,
	},
}

// PermissionMapping returns the release-verified interactive permission mapping
// for Antigravity CLI (CIP-0033, verified on agy >=1.5.16,<2.0.0).
func (*System) PermissionMapping(toolRelease string, mode agentic.PermissionMode) (agentic.PermissionMapping, error) {
	return permissionMapping(toolRelease, mode)
}

func permissionMapping(toolRelease string, mode agentic.PermissionMode) (agentic.PermissionMapping, error) {
	var effective agentic.PermissionMode
	if value, err := mode.Resolve(); err != nil {
		return agentic.PermissionMapping{}, fmt.Errorf("agy: %w", err)
	} else {
		effective = value
	}
	var capability agentic.ReleaseCapability
	if value, err := agentic.LookupReleaseCapability(verifiedReleases, toolRelease); err != nil {
		return agentic.PermissionMapping{}, fmt.Errorf("agy: %w", err)
	} else {
		capability = value
	}
	mapping := agentic.PermissionMapping{Grammar: capability.Grammar}
	if effective == agentic.PermissionModeNative {
		return mapping, nil
	}
	mapping.Flag = agyYoloFlag
	return mapping, nil
}
