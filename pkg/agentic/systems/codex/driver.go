package codex

import (
	codexdriver "github.com/relux-works/archon-pantheon/pkg/agentic/systems/codex/sessiondriver"
	"github.com/relux-works/archon-pantheon/pkg/sessiondriver"
)

// SessionDriverDescriptor returns the wire-protocol session driver descriptor for Codex app-server.
func SessionDriverDescriptor() sessiondriver.Descriptor {
	return codexdriver.Descriptor()
}

// OpenSessionDriver opens or restores a wire-protocol session driver for Codex app-server.
func OpenSessionDriver(cfg sessiondriver.Config, saved []byte) (*codexdriver.Driver, uint64, error) {
	return codexdriver.OpenDriver(cfg, saved)
}
