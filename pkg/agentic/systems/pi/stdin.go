package pi

import "github.com/relux-works/archon-pantheon/pkg/agentic"

func stdin(agentic.LaunchRequest) (agentic.StdinPayload, error) {
	return agentic.StdinPayload{}, nil
}
