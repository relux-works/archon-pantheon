package agy

import (
	"strings"

	"github.com/relux-works/skill-agents-management/pkg/agentic"
)

const agyDisableAutoUpdateEnv = "AGY_CLI_DISABLE_AUTO_UPDATE"

// agyCuratorParentEnvAllowlist is the exact parent-environment allowlist
// applied for Curator-managed Antigravity launches (CIP-0033). Keeping
// XDG_CONFIG_HOME / HOME intact preserves operator git/gh/gcloud config inside
// child commands while excluding ambient credentials and competing auth vars
// (GEMINI_API_KEY, GOOGLE_API_KEY, AGY_ADC_AUTH, GOOGLE_APPLICATION_CREDENTIALS,
// AGY_CLI_FORCE_FILE_TOKEN_STORAGE, etc.).
var agyCuratorParentEnvAllowlist = map[string]struct{}{
	"HOME":            {},
	"LANG":            {},
	"LC_ALL":          {},
	"LC_COLLATE":      {},
	"LC_CTYPE":        {},
	"LC_MESSAGES":     {},
	"LC_MONETARY":     {},
	"LC_NUMERIC":      {},
	"LC_TIME":         {},
	"LOGNAME":         {},
	"PATH":            {},
	"SHELL":           {},
	"SSL_CERT_DIR":    {},
	"SSL_CERT_FILE":   {},
	"TERM":            {},
	"TMPDIR":          {},
	"TZ":              {},
	"USER":            {},
	"XDG_CACHE_HOME":  {},
	"XDG_CONFIG_HOME": {},
	"XDG_DATA_HOME":   {},
	"XDG_RUNTIME_DIR": {},
	"XDG_STATE_HOME":  {},
}

func filterAgyCuratorParentEnv(parent []string, mcpEnvNames []string) []string {
	mcpAllowed := make(map[string]struct{}, len(mcpEnvNames))
	for _, name := range mcpEnvNames {
		mcpAllowed[name] = struct{}{}
	}
	filtered := make([]string, 0, len(parent))
	for _, entry := range parent {
		name, _, hasValue := strings.Cut(entry, "=")
		if !hasValue {
			continue
		}
		if _, ok := agyCuratorParentEnvAllowlist[name]; ok {
			filtered = append(filtered, entry)
			continue
		}
		if _, ok := mcpAllowed[name]; ok {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

// childEnv builds the Antigravity child environment.
//
// For legacy unmanaged launches (req.Context == nil), it retains the
// source-parity unfiltered contract pinned by parity_test.go and
// TestTheEmptyFilterIsDeliberate.
//
// For Curator-managed launches (req.Context != nil, CIP-0033), it filters the
// parent through agyCuratorParentEnvAllowlist (plus validated MCP env_names),
// overlays the run context, writes AGY_CLI_FORCE_FILE_TOKEN_STORAGE=true from
// the validated v7 fragment (defeating any ambient false), and pins
// AGY_CLI_DISABLE_AUTO_UPDATE=true as this plugin's own literal.
func childEnv(parent []string, req agentic.LaunchRequest) []string {
	if req.Context == nil {
		return agentic.WithRunContext(parent, req)
	}
	var mcpEnvNames []string
	if req.Context.MCP != nil {
		mcpEnvNames = req.Context.MCP.EnvNames
	}
	env := agentic.WithRunContext(filterAgyCuratorParentEnv(parent, mcpEnvNames), req)
	if forceFile, present := req.Context.Env[agentic.CuratorAntigravityForceFileTokenVariable]; present {
		env = agentic.SetEnvValue(env, agentic.CuratorAntigravityForceFileTokenVariable, forceFile)
	}
	return agentic.SetEnvValue(env, agyDisableAutoUpdateEnv, "true")
}
