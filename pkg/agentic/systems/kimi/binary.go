package kimi

import "github.com/relux-works/archon-pantheon/internal/launchenv"

func resolveBinary(env []string) (string, error) {
	return launchenv.LookPath(env, executableName)
}
