package harnessapi

import (
	"strings"

	"github.com/relux-works/curator-runtimes/pkg/harnessvalue"
)

// Property is approved nonsecret configuration, never a raw native file.
type Property struct{ name, value string }

// NewProperty validates a closed nonsecret observation name and value.
func NewProperty(name, value string) (Property, error) {
	var allowed string
	switch name {
	case "instruction_form":
		allowed = "|monolithic|referenced|"
	case "settings_mode":
		allowed = "|managed|read-only|"
	case "seed_revision":
		allowed = "|A|B|C|"
	case "mcp_format":
		allowed = "|json|jsonc|toml|none|"
	case "credential_route":
		allowed = "|native|file-link|keyring-preferred|unsupported|"
	case "trust_observed", "owned", "present":
		allowed = "|true|false|unknown|"
	default:
		return Property{}, NewRefusal("harness_snapshot_field_unsupported")
	}
	if value == "" || strings.Contains(value, "|") || !strings.Contains(allowed, "|"+value+"|") {
		return Property{}, NewRefusal("harness_snapshot_value_invalid")
	}
	return Property{name, value}, nil
}

// Name returns the approved observation name.
func (p Property) Name() string { return p.name }

// Value returns the approved observation value.
func (p Property) Value() string { return p.value }

// Snapshot is an immutable harness contract value.
type Snapshot struct{ properties []Property }

// NewSnapshot bounds and copies unique approved observations.
func NewSnapshot(properties []Property) (Snapshot, error) {
	if len(properties) > harnessvalue.MaxItems {
		return Snapshot{}, NewRefusal("harness_value_too_large")
	}
	seen := map[string]bool{}
	for _, p := range properties {
		if _, err := NewProperty(p.name, p.value); err != nil {
			return Snapshot{}, err
		}
		if seen[p.name] {
			return Snapshot{}, NewRefusal("harness_snapshot_field_duplicate")
		}
		seen[p.name] = true
	}
	return Snapshot{append([]Property(nil), properties...)}, nil
}

// Properties returns a separate observation slice.
func (s Snapshot) Properties() []Property { return append([]Property(nil), s.properties...) }

// Request carries an explicit operation, instruction form and scope.
type Request struct {
	operation Capability
	form      Form
	scope     Scope
}

// NewRequest validates operation, form and scope without executing them.
func NewRequest(operation Capability, form Form, scope Scope) (Request, error) {
	if _, err := NewCapabilitySet(operation); err != nil {
		return Request{}, err
	}
	if form != Monolithic && form != Referenced {
		return Request{}, NewRefusal("environment_form_unsupported")
	}
	if scope != UserScope && scope != ProjectScope {
		return Request{}, NewRefusal("harness_scope_unsupported")
	}
	return Request{operation, form, scope}, nil
}

// Operation returns the requested operation.
func (r Request) Operation() Capability { return r.operation }

// Form returns the declared instruction form.
func (r Request) Form() Form { return r.form }

// Scope returns the explicit delivery scope.
func (r Request) Scope() Scope { return r.scope }

// NamedPaths describes an ordered native surface without allowing path effects.
type NamedPaths struct{ paths []harnessvalue.RelativePath }

// NewNamedPaths bounds and copies validated relative paths.
func NewNamedPaths(paths []harnessvalue.RelativePath) (NamedPaths, error) {
	if len(paths) > harnessvalue.MaxItems {
		return NamedPaths{}, NewRefusal("harness_value_too_large")
	}
	for _, p := range paths {
		if !p.Valid() {
			return NamedPaths{}, NewRefusal("harness_path_invalid")
		}
	}
	return NamedPaths{append([]harnessvalue.RelativePath(nil), paths...)}, nil
}

// Paths returns a separate ordered path slice.
func (p NamedPaths) Paths() []harnessvalue.RelativePath {
	return append([]harnessvalue.RelativePath(nil), p.paths...)
}

// HomeLayout describes environment selection and ordered relative roots.
type HomeLayout struct {
	environment string
	roots       NamedPaths
	parent      bool
}

// NewHomeLayout binds an environment selector to immutable relative roots.
func NewHomeLayout(environment string, roots NamedPaths, parent bool) (HomeLayout, error) {
	if environment == "" || harnessvalue.Text(environment) != nil {
		return HomeLayout{}, NewRefusal("harness_home_invalid")
	}
	return HomeLayout{environment, roots, parent}, nil
}

// EnvironmentVariable returns the native selector name.
func (h HomeLayout) EnvironmentVariable() string { return h.environment }

// Roots returns the immutable relative-root description.
func (h HomeLayout) Roots() NamedPaths { return h.roots }

// ParentValued reports whether the selector names a parent directory.
func (h HomeLayout) ParentValued() bool { return h.parent }

// InstructionSpec describes destinations, form and ownership notice bytes.
type InstructionSpec struct {
	paths  NamedPaths
	form   Form
	notice harnessvalue.Bytes
}

// NewInstructionSpec validates the form and copies ownership notice bytes.
func NewInstructionSpec(paths NamedPaths, form Form, notice []byte) (InstructionSpec, error) {
	if form != Monolithic && form != Referenced {
		return InstructionSpec{}, NewRefusal("environment_form_unsupported")
	}
	value, err := harnessvalue.NewBytes(notice)
	if err != nil {
		return InstructionSpec{}, err
	}
	return InstructionSpec{paths, form, value}, nil
}

// Paths returns the immutable path description.
func (s InstructionSpec) Paths() NamedPaths { return s.paths }

// Form returns the declared instruction form.
func (s InstructionSpec) Form() Form { return s.form }

// Notice returns a separate copy of ownership notice bytes.
func (s InstructionSpec) Notice() []byte { return s.notice.Copy() }
