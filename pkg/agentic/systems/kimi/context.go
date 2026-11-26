package kimi

import (
	"fmt"

	"github.com/relux-works/curator-runtimes/pkg/agentic"
)

const curatorEnvironment = "kimi_code"

// ValidateCuratorContext enforces Kimi Code CLI's Curator managed-home
// fragment contract (env transport KIMI_CODE_HOME).
func (*System) ValidateCuratorContext(req agentic.LaunchRequest, _ agentic.LaunchMode) error {
	return validateKimiCuratorContext(req)
}

func validateKimiCuratorContext(req agentic.LaunchRequest) error {
	if req.Context == nil {
		return nil
	}
	if err := agentic.ValidateCuratorContext(req.Context, req.Home); err != nil {
		return err
	}
	if req.Context.Environment != curatorEnvironment {
		return fmt.Errorf("%w: Kimi Code requires the kimi_code fragment environment", agentic.ErrCuratorContextUnsupported)
	}
	if req.Context.SystemPrompt != nil {
		return fmt.Errorf("%w: Kimi Code does not support system-prompt fragment channels", agentic.ErrCuratorContextUnsupported)
	}
	if req.Context.MCP != nil {
		if len(req.Context.MCP.Channels) != 1 || req.Context.MCP.Channels[0].Kind != agentic.CuratorDescriptorFile {
			return fmt.Errorf("%w: Kimi Code MCP channel must be a managed-home file", agentic.ErrCuratorContextUnsupported)
		}
	}
	return nil
}
