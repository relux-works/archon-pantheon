package kimi

import (
	"fmt"

	"github.com/relux-works/archon-pantheon/pkg/agentic"
)

var verifiedReleases = []agentic.ReleaseCapability{
	{MinRelease: "2.6.3", MaxReleaseExclusive: "3.0.0", Grammar: agentic.PermissionGrammarV1, YoloSupported: true},
}

// PermissionMapping returns the release-pinned interactive posture mapping.
func (*System) PermissionMapping(toolRelease string, mode agentic.PermissionMode) (agentic.PermissionMapping, error) {
	return permissionMapping(toolRelease, mode)
}

func permissionMapping(toolRelease string, mode agentic.PermissionMode) (agentic.PermissionMapping, error) {
	var effective agentic.PermissionMode
	if effectiveValue, err := mode.Resolve(); err != nil {
		return agentic.PermissionMapping{}, fmt.Errorf("kimi: %w", err)
	} else {
		effective = effectiveValue
	}
	var capability agentic.ReleaseCapability
	if capabilityValue, err := agentic.LookupReleaseCapability(verifiedReleases, toolRelease); err != nil {
		return agentic.PermissionMapping{}, fmt.Errorf("kimi: %w", err)
	} else {
		capability = capabilityValue
	}
	mapping := agentic.PermissionMapping{Grammar: capability.Grammar}
	if effective == agentic.PermissionModeNative {
		return mapping, nil
	}
	if !capability.YoloSupported {
		return agentic.PermissionMapping{}, fmt.Errorf("kimi: refusing yolo: %w: tool release %q documents no bypass flag", agentic.ErrPermissionModeUnsupported, capability.Release)
	}
	mapping.Flag = yoloFlag
	return mapping, nil
}
