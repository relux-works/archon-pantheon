// Package harnessapi defines the versioned value-only harness contract.
// This package owns no filesystem, process, network, or credential-store effects.
package harnessapi

import "time"

// ID is a canonical harness identifier.
type ID string

// APIVersion identifies the internal adapter contract revision.
type APIVersion uint16

// AdapterAPIVersionV1 identifies the first internal adapter contract.
const AdapterAPIVersionV1 APIVersion = 1

// ProbeTimeoutV1 pins the baseline read-only probe deadline.
const ProbeTimeoutV1 = 5 * time.Second

// Form names a supported instruction representation.
type Form string

// The following constants define the closed contract vocabulary and bounds.
const (
	Monolithic Form = "monolithic"
	Referenced Form = "referenced"
)

// Scope identifies user or project delivery.
type Scope string

// The following constants define the closed contract vocabulary and bounds.
const (
	UserScope    Scope = "user"
	ProjectScope Scope = "project"
)

// Capability names a separately declared operation.
type Capability uint16

// The following constants define the closed contract vocabulary and bounds.
const (
	ManagedHome Capability = iota + 1
	InstructionOutput
	SkillProjection
	NativeSkillDiscovery
	MCPInspection
	NativeMCPWrite
	ProjectContextOutput
	HomeLease
	CredentialMetadata
	TrustObservation
	ReleaseProbing
)

// Adapter exposes the staged harness surface with explicit operational refusals.
type Adapter interface {
	ID() ID
	APIVersion() APIVersion
	Capabilities() CapabilitySet
	ReleaseProbe(Context) (ProbeSpec, error)
	ParseReleaseProbe(Context, ProbeOutput) (VerifiedRelease, error)
	Home(Context) (HomeLayout, error)
	Instructions(Context, Form) (InstructionSpec, error)
	Skills(Context, Scope) (SkillDiscovery, error)
	Settings(Context) (SettingsSpec, error)
	MCP(Context, Scope) (MCPChannel, error)
	Credentials(Context) (CredentialRoute, error)
	Trust(Context) (TrustPolicy, error)
	Preflight(Context, Snapshot, Request) ([]Refusal, error)
	PlanSeed(Context, Snapshot, Request) (SeedPlan, error)
	ProjectContext(Context, Snapshot, Request) (ProjectContextPlan, error)
	Status(Context, Snapshot) (Status, error)
	Drift(Context, Snapshot) (DriftReport, error)
}
