// Package kimi implements the Kimi Code CLI launch surface.
package kimi

import (
	"fmt"

	"github.com/relux-works/archon-pantheon/pkg/agentic"
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
	if err := validateKimiCuratorContext(req); err != nil {
		return nil, err
	}
	return Args(req, mode)
}

// ChildEnv preserves the supplied environment, enforces KIMI_CODE_HOME from
// Curator context when present, and adds the shared run context.
func (*System) ChildEnv(parent []string, req agentic.LaunchRequest) ([]string, error) {
	if err := validateKimiCuratorContext(req); err != nil {
		return nil, err
	}
	env := agentic.WithRunContext(parent, req)
	if req.Context != nil && req.Home != "" {
		env = setEnvVar(env, homeEnvVar, req.Home)
	}
	return env, nil
}

func setEnvVar(env []string, key, value string) []string {
	prefix := key + "="
	out := make([]string, 0, len(env)+1)
	replaced := false
	for _, entry := range env {
		if len(entry) >= len(prefix) && entry[:len(prefix)] == prefix {
			if !replaced {
				out = append(out, prefix+value)
				replaced = true
			}
			continue
		}
		out = append(out, entry)
	}
	if !replaced {
		out = append(out, prefix+value)
	}
	return out
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
