package claude

import (
	claudedriver "github.com/relux-works/curator-runtimes/pkg/agentic/systems/claude/sessiondriver"
	"github.com/relux-works/curator-runtimes/pkg/sessiondriver"
)

// SessionDriverDescriptor returns the wire-protocol session driver descriptor for Claude stream-json.
func SessionDriverDescriptor() sessiondriver.Descriptor {
	return claudedriver.Descriptor()
}

// OpenSessionDriver opens or restores a wire-protocol session driver for Claude stream-json.
func OpenSessionDriver(cfg sessiondriver.Config, saved []byte) (*claudedriver.Driver, uint64, error) {
	return claudedriver.OpenDriver(cfg, saved)
}
