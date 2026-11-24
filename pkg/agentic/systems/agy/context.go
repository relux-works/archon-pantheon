package agy

import (
	"fmt"

	"github.com/relux-works/skill-agents-management/pkg/agentic"
)

// ValidateCuratorContext is the Antigravity channel-layer gate for a typed
// Curator launch-env-fragment-v7 context (CIP-0033). BuildPlan invokes it
// before any launch surface is built; Argv and ChildEnv repeat it for direct
// plugin callers.
func (*System) ValidateCuratorContext(req agentic.LaunchRequest, _ agentic.LaunchMode) error {
	return validateAgyCuratorContext(req)
}

func validateAgyCuratorContext(req agentic.LaunchRequest) error {
	if req.Context == nil {
		return nil
	}
	if err := agentic.ValidateCuratorContext(req.Context, req.Home); err != nil {
		return err
	}
	if !agentic.CuratorFragmentUsesFlagHome(req.Context.Environment) {
		return fmt.Errorf("%w: Antigravity requires the antigravity fragment environment", agentic.ErrCuratorContextUnsupported)
	}
	if !req.Composition.IsZero() {
		return fmt.Errorf("%w: Antigravity Curator context cannot combine with legacy MCP composition", agentic.ErrCuratorContextUnsupported)
	}
	return nil
}
