package kimi

import "github.com/relux-works/curator-runtimes/internal/launchenv"

func resolveBinary(env []string) (string, error) {
	return launchenv.LookPath(env, executableName)
}
