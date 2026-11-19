package kimi

import "github.com/relux-works/skill-agents-management/internal/launchenv"

func resolveBinary(env []string) (string, error) {
	return launchenv.LookPath(env, executableName)
}
