package pi

import "github.com/relux-works/curator-runtimes/pkg/agentic"

func stdin(agentic.LaunchRequest) (agentic.StdinPayload, error) {
	return agentic.StdinPayload{}, nil
}
