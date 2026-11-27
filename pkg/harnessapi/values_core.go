package harnessapi

import (
	"time"

	"github.com/relux-works/archon-pantheon/pkg/harnessvalue"
)

// The following constants define the closed contract vocabulary and bounds.
const (
	UnknownID             = "environment_unknown"
	APIVersionUnsupported = "harness_adapter_api_version_unsupported"
	RegistryMismatch      = "harness_adapter_registry_mismatch"
	CapabilityUnsupported = "harness_capability_unsupported"
	ProbeSpecUnsupported  = "harness_probe_spec_unsupported"
	RuntimeMissing        = "adapter_runtime_missing"
)

// Refusal carries a stable code without rendering untrusted input values.
type Refusal struct{ code string }

// NewRefusal bounds a stable diagnostic code without formatting input values.
func NewRefusal(code string) Refusal {
	if code == "" || harnessvalue.Text(code) != nil {
		code = "harness_refusal_invalid"
	}
	return Refusal{code}
}

// Code returns the stable refusal code.
func (r Refusal) Code() string { return r.code }

// Error returns the refusal code without input values.
func (r Refusal) Error() string { return r.code }

// CapabilitySet is an immutable closed set of declared operations.
type CapabilitySet struct{ bits uint64 }

// NewCapabilitySet validates operation identifiers and builds an immutable set.
func NewCapabilitySet(values ...Capability) (CapabilitySet, error) {
	var set CapabilitySet
	for _, value := range values {
		if value < ManagedHome || value > ReleaseProbing {
			return CapabilitySet{}, NewRefusal(CapabilityUnsupported)
		}
		set.bits |= uint64(1) << value
	}
	return set, nil
}

// Has reports whether the operation is explicitly declared.
func (s CapabilitySet) Has(value Capability) bool {
	return value >= ManagedHome && value <= ReleaseProbing && s.bits&(uint64(1)<<value) != 0
}

// Values returns declared operations in stable numeric order.
func (s CapabilitySet) Values() []Capability {
	var values []Capability
	for value := ManagedHome; value <= ReleaseProbing; value++ {
		if s.Has(value) {
			values = append(values, value)
		}
	}
	return values
}

// ProbeSpec remains untrusted data: only the registry's exact tuple admits it.
// Invalid oversized values retain an invalid marker, never a repaired tuple.
type ProbeSpec struct {
	alias       string
	arguments   []string
	timeout     time.Duration
	environment map[string]string
	invalid     bool
}

// NewProbeSpec copies a proposed tuple without granting probe permission.
func NewProbeSpec(alias string, args []string, timeout time.Duration, env map[string]string) ProbeSpec {
	spec := ProbeSpec{alias: alias, timeout: timeout}
	if harnessvalue.Text(alias) != nil || len(args) > harnessvalue.MaxItems || len(env) > harnessvalue.MaxItems {
		return ProbeSpec{invalid: true}
	}
	for _, arg := range args {
		if harnessvalue.Text(arg) != nil {
			return ProbeSpec{invalid: true}
		}
	}
	for key, value := range env {
		if harnessvalue.Text(key) != nil || harnessvalue.Text(value) != nil {
			return ProbeSpec{invalid: true}
		}
	}
	spec.arguments = append([]string(nil), args...)
	spec.environment = copyMap(env)
	return spec
}

// ExecutableAlias returns the proposed bare executable alias.
func (p ProbeSpec) ExecutableAlias() string { return p.alias }

// Arguments returns a separate argument slice.
func (p ProbeSpec) Arguments() []string { return append([]string(nil), p.arguments...) }

// Timeout returns the proposed deadline.
func (p ProbeSpec) Timeout() time.Duration { return p.timeout }

// Environment returns a separate map of proposed overrides.
func (p ProbeSpec) Environment() map[string]string { return copyMap(p.environment) }

// ValidValue reports constructor size validation, independently of probe permission.
func (p ProbeSpec) ValidValue() bool { return !p.invalid }

func copyMap(input map[string]string) map[string]string {
	if input == nil {
		return nil
	}
	output := make(map[string]string, len(input))
	for k, v := range input {
		output[k] = v
	}
	return output
}

// VerifiedRelease carries a bounded vendor release observation independently of the API version.
type VerifiedRelease struct{ version string }

// NewVerifiedRelease bounds a nonempty vendor release observation.
func NewVerifiedRelease(version string) (VerifiedRelease, error) {
	if version == "" || harnessvalue.Text(version) != nil {
		return VerifiedRelease{}, NewRefusal("harness_release_invalid")
	}
	return VerifiedRelease{version}, nil
}

// Version returns the vendor release observation.
func (v VerifiedRelease) Version() string { return v.version }

// ProbeOutput is an immutable harness contract value.
type ProbeOutput struct {
	stdout, stderr harnessvalue.Bytes
	exitCode       int
}

// NewProbeOutput bounds and copies process output and its observed exit code.
func NewProbeOutput(stdout, stderr []byte, exitCode int) (ProbeOutput, error) {
	out, err := harnessvalue.NewBytes(stdout)
	if err != nil {
		return ProbeOutput{}, err
	}
	errout, err := harnessvalue.NewBytes(stderr)
	if err != nil {
		return ProbeOutput{}, err
	}
	return ProbeOutput{out, errout, exitCode}, nil
}

// Stdout returns a separate copy of standard output.
func (p ProbeOutput) Stdout() []byte { return p.stdout.Copy() }

// Stderr returns a separate copy of standard error.
func (p ProbeOutput) Stderr() []byte { return p.stderr.Copy() }

// ExitCode returns the observed process exit code.
func (p ProbeOutput) ExitCode() int { return p.exitCode }

// Context carries explicit platform/scope inputs; it has no ambient lookup.
type Context struct {
	platform string
	scope    Scope
	release  VerifiedRelease
}

// NewContext validates the platform and scope and binds a release observation.
func NewContext(platform string, scope Scope, release VerifiedRelease) (Context, error) {
	if platform != "linux" && platform != "darwin" && platform != "windows" {
		return Context{}, NewRefusal("harness_platform_unsupported")
	}
	if scope != UserScope && scope != ProjectScope {
		return Context{}, NewRefusal("harness_scope_unsupported")
	}
	return Context{platform, scope, release}, nil
}

// Platform returns the explicit target platform.
func (c Context) Platform() string { return c.platform }

// Scope returns the explicit delivery scope.
func (c Context) Scope() Scope { return c.scope }

// Release returns the immutable vendor release observation.
func (c Context) Release() VerifiedRelease { return c.release }
