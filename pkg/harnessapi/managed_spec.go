package harnessapi

// ShadowSpec names a home-relative path that shadows or additively merges with
// a managed surface inside an adapter home.
type ShadowSpec struct {
	Path    string
	Surface string
}

// PassthroughSpec describes one native credential passthrough route declared by
// an adapter.
type PassthroughSpec struct {
	Path           string
	Strategy       string
	FileLinkTarget string
}

// ManagedHomeSpec is the pure, adapter-owned descriptor for a managed-home-only
// environment registered through harnessregistry. It carries zero filesystem,
// process, or credential-store effects.
type ManagedHomeSpec struct {
	ID                                string
	Alias                             string
	HomeTransport                     string
	EnvVar                            string
	ManagedOnly                       bool
	NativeDefaultDir                  string
	NativeHomeIgnoresEnv              bool
	ForbiddenHomeBasename             string
	RootTarget                        string
	SkillsDir                         string
	Forms                             []string
	DefaultForm                       string
	MCPFilename                       string
	MCPFormat                         string
	ForbiddenMCPEnvNames              []string
	Passthrough                       map[string][]PassthroughSpec
	CredentialFileGuardPath           string
	ForcedCredentialStore             string
	ForcedCredentialEnvVar            string
	ForcedCredentialEnvValue          string
	RequireLiveNativeCredentialTarget bool
	Shadows                           []ShadowSpec
	VerifiedRelease                   string
	Probe                             []string
	SizeAdvisoryBytes                 int64
	CredentialScope                   string
	GlobalContextCap                  string
	FirstRunSteps                     string
	StatusNote                        string
	NativeLoginHint                   string
	ForcedStoreGuidance               string
	CredentialPolicyNote              string
}

// VersionProbeArgs returns a fresh "<binary> --version" probe tuple.
func VersionProbeArgs(binary string) []string {
	return append([]string(nil), binary, "--version")
}
