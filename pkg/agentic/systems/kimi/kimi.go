// Package kimi implements the Kimi Code CLI launch surface.
package kimi

import (
	"fmt"

	"github.com/relux-works/skill-agents-management/pkg/agentic"
)

const (
	systemID       agentic.SystemID = "kimi"
	executableName                  = "kimi"
	homeEnvVar                      = "KIMI_CODE_HOME"
	defaultHome                     = "~/.kimi-code"
)

// System is the Kimi Code CLI plugin. It is stateless; all launch inputs come
// from the request and its supplied environment.
type System struct{}

// New returns a system instance for registries that are assembled explicitly.
func New() *System { return &System{} }

func init() {
	if err := agentic.Register(New()); err != nil {
		panic(fmt.Sprintf("kimi: registering the plugin: %v", err))
	}
}

func (*System) ID() agentic.SystemID { return systemID }

// Capabilities lists the CLI surfaces this plugin can express. Kimi model
// aliases are configuration-defined, so the plugin transports them without
// reading or inventing an alias catalog.
func (*System) Capabilities() agentic.Capabilities {
	return agentic.Capabilities{
		LaunchModes: []agentic.LaunchMode{
			agentic.LaunchModeExec,
			agentic.LaunchModeDryRun,
			agentic.LaunchModeInteractive,
		},
		EffortTransport:     agentic.EffortTransportNone,
		SupportsGoal:        false,
		SupportsBudget:      false,
		SupportsServiceTier: false,
		CompositionGrammar:  agentic.GrammarNone,
		HomeEnvVar:          homeEnvVar,
		DefaultHome:         defaultHome,
	}
}

// ResolveBinary resolves the executable from the request's PATH.
func (*System) ResolveBinary(req agentic.LaunchRequest) (string, error) {
	return resolveBinary(req.Env)
}

// Argv builds the exact CLI suffix for a declared launch mode.
func (s *System) Argv(req agentic.LaunchRequest, mode agentic.LaunchMode) ([]string, error) {
	if !s.Capabilities().SupportsMode(mode) {
		return nil, fmt.Errorf("kimi: launch mode %s is not declared by this system", mode)
	}
	return Args(req, mode)
}

// ChildEnv preserves the supplied environment and adds the shared run context.
// The caller selects an alternate Kimi home through KIMI_CODE_HOME.
func (*System) ChildEnv(parent []string, req agentic.LaunchRequest) ([]string, error) {
	return agentic.WithRunContext(parent, req), nil
}

// Stdin is unused: Kimi Code receives prompt mode text through --prompt.
func (*System) Stdin(agentic.LaunchRequest) (agentic.StdinPayload, error) {
	return agentic.StdinPayload{}, nil
}

// ValidateComposition refuses composition because the CLI has no sourced
// per-launch MCP composition flag.
func (*System) ValidateComposition(c agentic.Composition) error {
	if c.IsZero() {
		return nil
	}
	return fmt.Errorf("kimi: launch composition is unsupported")
}
