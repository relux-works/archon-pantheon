package harnessapi

import (
	"fmt"

	"github.com/relux-works/archon-pantheon/pkg/harnessvalue"
)

// SkillDiscovery describes native discovery and projection for an explicit scope.
type SkillDiscovery struct {
	paths, ignores NamedPaths
	scope          Scope
	native, mirror bool
}

// NewSkillDiscovery binds discovery paths and mirror policy to an explicit scope.
func NewSkillDiscovery(paths, ignores NamedPaths, scope Scope, native, mirror bool) (SkillDiscovery, error) {
	if scope != UserScope && scope != ProjectScope {
		return SkillDiscovery{}, NewRefusal("harness_scope_unsupported")
	}
	return SkillDiscovery{paths, ignores, scope, native, mirror}, nil
}

// Paths returns the immutable path description.
func (s SkillDiscovery) Paths() NamedPaths { return s.paths }

// IgnorePaths returns the generated-path ignore description.
func (s SkillDiscovery) IgnorePaths() NamedPaths { return s.ignores }

// Scope returns the explicit delivery scope.
func (s SkillDiscovery) Scope() Scope { return s.scope }

// NativeDiscovery reports native discovery support in this description.
func (s SkillDiscovery) NativeDiscovery() bool { return s.native }

// Mirror reports whether projection is requested.
func (s SkillDiscovery) Mirror() bool { return s.mirror }

// Format identifies an in-memory native serialization.
type Format string

// The following constants define the closed contract vocabulary and bounds.
const (
	JSON     Format = "json"
	TOML     Format = "toml"
	JSONC    Format = "jsonc"
	Markdown Format = "markdown"
)

func validFormat(f Format) bool { return f == JSON || f == TOML || f == JSONC || f == Markdown }

// SettingsSpec describes destinations, format and approved fields.
type SettingsSpec struct {
	paths  NamedPaths
	format Format
	fields []Property
}

// NewSettingsSpec binds a known format to approved fields and destinations.
func NewSettingsSpec(paths NamedPaths, format Format, fields []Property) (SettingsSpec, error) {
	if !validFormat(format) {
		return SettingsSpec{}, NewRefusal("harness_format_unsupported")
	}
	snapshot, err := NewSnapshot(fields)
	if err != nil {
		return SettingsSpec{}, err
	}
	return SettingsSpec{paths, format, snapshot.Properties()}, nil
}

// Paths returns the immutable path description.
func (s SettingsSpec) Paths() NamedPaths { return s.paths }

// Format returns the declared serialization format.
func (s SettingsSpec) Format() Format { return s.format }

// Fields returns a separate approved-field slice.
func (s SettingsSpec) Fields() []Property { return append([]Property(nil), s.fields...) }

// MCPChannel separates inspection, writing and project trust requirements.
type MCPChannel struct {
	paths                 NamedPaths
	format                Format
	inspect, write, trust bool
}

// NewMCPChannel validates inspection and write declarations independently of trust.
func NewMCPChannel(paths NamedPaths, format Format, inspect, write, trust bool) (MCPChannel, error) {
	if !validFormat(format) || write && !inspect {
		return MCPChannel{}, NewRefusal("harness_mcp_invalid")
	}
	return MCPChannel{paths, format, inspect, write, trust}, nil
}

// Paths returns the immutable path description.
func (m MCPChannel) Paths() NamedPaths { return m.paths }

// Format returns the declared serialization format.
func (m MCPChannel) Format() Format { return m.format }

// Inspect reports declared inspection support.
func (m MCPChannel) Inspect() bool { return m.inspect }

// Write reports declared native-write support.
func (m MCPChannel) Write() bool { return m.write }

// RequiresTrust reports whether project admission is required.
func (m MCPChannel) RequiresTrust() bool { return m.trust }

// CredentialRoute describes custody metadata only; it cannot request a copy.
type CredentialRoute struct {
	paths     NamedPaths
	strategy  string
	qualified bool
}

// NewCredentialRoute describes a closed metadata route without credential values or copy actions.
func NewCredentialRoute(paths NamedPaths, strategy string, qualified bool) (CredentialRoute, error) {
	switch strategy {
	case "native", "file-link", "keyring-preferred", "unsupported":
	default:
		return CredentialRoute{}, NewRefusal("environment_credential_unsupported")
	}
	return CredentialRoute{paths, strategy, qualified}, nil
}

// Paths returns the immutable path description.
func (c CredentialRoute) Paths() NamedPaths { return c.paths }

// Strategy returns the closed credential metadata route.
func (c CredentialRoute) Strategy() string { return c.strategy }

// Qualified reports the route's supplied qualification fact.
func (c CredentialRoute) Qualified() bool { return c.qualified }

// TrustPolicy describes required admission without granting it.
type TrustPolicy struct{ required, observationOnly bool }

// NewTrustPolicy describes an observation-only trust requirement.
func NewTrustPolicy(required bool) TrustPolicy { return TrustPolicy{required, true} }

// Required reports whether trust admission is required.
func (p TrustPolicy) Required() bool { return p.required }

// ObservationOnly reports that the policy grants no trust.
func (p TrustPolicy) ObservationOnly() bool { return p.observationOnly }

// FileAction is a proposed owned-file replacement, not an executable callback.
type FileAction struct {
	path     harnessvalue.RelativePath
	bytes    harnessvalue.Bytes
	preimage string
	mode     uint32
}

// NewFileAction bounds replacement bytes and validates path, mode and expected preimage.
func NewFileAction(path harnessvalue.RelativePath, bytes []byte, preimage string, mode uint32) (FileAction, error) {
	if !path.Valid() || mode & ^uint32(0o777) != 0 {
		return FileAction{}, NewRefusal("harness_plan_invalid")
	}
	if preimage != "absent" {
		if len(preimage) != 64 {
			return FileAction{}, NewRefusal("harness_preimage_invalid")
		}
		for _, c := range preimage {
			if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
				return FileAction{}, NewRefusal("harness_preimage_invalid")
			}
		}
	}
	value, err := harnessvalue.NewBytes(bytes)
	if err != nil {
		return FileAction{}, err
	}
	return FileAction{path, value, preimage, mode}, nil
}

// Path returns the validated relative destination.
func (f FileAction) Path() harnessvalue.RelativePath { return f.path }

// Bytes returns a separate copy of proposed file bytes.
func (f FileAction) Bytes() []byte { return f.bytes.Copy() }

// Preimage returns the expected preceding identity.
func (f FileAction) Preimage() string { return f.preimage }

// Mode returns the proposed permission bits.
func (f FileAction) Mode() uint32 { return f.mode }

// SeedPlan contains bounded file actions or typed refusals.
type SeedPlan struct {
	actions  []FileAction
	refusals []Refusal
}

// NewSeedPlan rejects duplicate destinations and mixed writes and refusals.
func NewSeedPlan(actions []FileAction, refusals []Refusal) (SeedPlan, error) {
	if len(actions) > harnessvalue.MaxItems || len(refusals) > harnessvalue.MaxItems {
		return SeedPlan{}, NewRefusal("harness_value_too_large")
	}
	seen := map[string]bool{}
	total := 0
	for _, action := range actions {
		if _, err := NewFileAction(action.path, action.bytes.Copy(), action.preimage, action.mode); err != nil {
			return SeedPlan{}, err
		}
		if seen[action.path.String()] {
			return SeedPlan{}, NewRefusal("harness_plan_path_duplicate")
		}
		seen[action.path.String()] = true
		total += action.bytes.Len()
		if total > harnessvalue.MaxTextBytes {
			return SeedPlan{}, NewRefusal("harness_value_too_large")
		}
	}
	if len(actions) > 0 && len(refusals) > 0 {
		return SeedPlan{}, NewRefusal("harness_plan_refused_with_actions")
	}
	return SeedPlan{append([]FileAction(nil), actions...), append([]Refusal(nil), refusals...)}, nil
}

// Actions returns a separate file-action slice.
func (s SeedPlan) Actions() []FileAction { return append([]FileAction(nil), s.actions...) }

// Refusals returns a separate refusal slice.
func (s SeedPlan) Refusals() []Refusal { return append([]Refusal(nil), s.refusals...) }

// ProjectContextPlan binds a seed plan to its protected-field description.
type ProjectContextPlan struct {
	plan      SeedPlan
	protected NamedPaths
}

// NewProjectContextPlan binds immutable output and protected-field descriptions.
func NewProjectContextPlan(plan SeedPlan, protected NamedPaths) ProjectContextPlan {
	return ProjectContextPlan{plan, protected}
}

// Plan returns the immutable output description.
func (p ProjectContextPlan) Plan() SeedPlan { return p.plan }

// ProtectedFields returns the protected native-field description.
func (p ProjectContextPlan) ProtectedFields() NamedPaths { return p.protected }

// Status contains approved read-only observations.
type Status struct{ properties []Property }

// NewStatus copies approved status observations.
func NewStatus(properties []Property) (Status, error) {
	s, err := NewSnapshot(properties)
	return Status{s.Properties()}, err
}

// Properties returns a separate observation slice.
func (s Status) Properties() []Property { return append([]Property(nil), s.properties...) }

// DriftReport describes owned paths and typed refusals.
type DriftReport struct {
	paths    NamedPaths
	refusals []Refusal
}

// NewDriftReport bounds and copies drift refusals and path descriptions.
func NewDriftReport(paths NamedPaths, refusals []Refusal) (DriftReport, error) {
	if len(refusals) > harnessvalue.MaxItems {
		return DriftReport{}, fmt.Errorf("harness_value_too_large")
	}
	return DriftReport{paths, append([]Refusal(nil), refusals...)}, nil
}

// Paths returns the immutable path description.
func (d DriftReport) Paths() NamedPaths { return d.paths }

// Refusals returns a separate refusal slice.
func (d DriftReport) Refusals() []Refusal { return append([]Refusal(nil), d.refusals...) }
