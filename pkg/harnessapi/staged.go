package harnessapi

// OperationNotPorted distinguishes a staged surface from a vendor capability
// that has been qualified as unsupported. Migration replaces these methods one
// harness at a time. A staged declaration must never claim these operations.
const OperationNotPorted = "harness_adapter_operation_not_ported"

// UnportedOperations is the explicit staged API surface. It has no effect
// dependencies, callbacks or ambient inputs. It never returns a success plan.
type UnportedOperations struct{}

// ParseReleaseProbe refuses this operation until its harness surface is ported.
func (UnportedOperations) ParseReleaseProbe(Context, ProbeOutput) (VerifiedRelease, error) {
	return VerifiedRelease{}, NewRefusal(OperationNotPorted)
}

// Home refuses this operation until its harness surface is ported.
func (UnportedOperations) Home(Context) (HomeLayout, error) {
	return HomeLayout{}, NewRefusal(OperationNotPorted)
}

// Instructions refuses this operation until its harness surface is ported.
func (UnportedOperations) Instructions(Context, Form) (InstructionSpec, error) {
	return InstructionSpec{}, NewRefusal(OperationNotPorted)
}

// Skills refuses this operation until its harness surface is ported.
func (UnportedOperations) Skills(Context, Scope) (SkillDiscovery, error) {
	return SkillDiscovery{}, NewRefusal(OperationNotPorted)
}

// Settings refuses this operation until its harness surface is ported.
func (UnportedOperations) Settings(Context) (SettingsSpec, error) {
	return SettingsSpec{}, NewRefusal(OperationNotPorted)
}

// MCP refuses this operation until its harness surface is ported.
func (UnportedOperations) MCP(Context, Scope) (MCPChannel, error) {
	return MCPChannel{}, NewRefusal(OperationNotPorted)
}

// Credentials refuses this operation until its harness surface is ported.
func (UnportedOperations) Credentials(Context) (CredentialRoute, error) {
	return CredentialRoute{}, NewRefusal(OperationNotPorted)
}

// Trust refuses this operation until its harness surface is ported.
func (UnportedOperations) Trust(Context) (TrustPolicy, error) {
	return TrustPolicy{}, NewRefusal(OperationNotPorted)
}

// Preflight refuses this operation until its harness surface is ported.
func (UnportedOperations) Preflight(Context, Snapshot, Request) ([]Refusal, error) {
	return nil, NewRefusal(OperationNotPorted)
}

// PlanSeed refuses this operation until its harness surface is ported.
func (UnportedOperations) PlanSeed(Context, Snapshot, Request) (SeedPlan, error) {
	return SeedPlan{}, NewRefusal(OperationNotPorted)
}

// ProjectContext refuses this operation until its harness surface is ported.
func (UnportedOperations) ProjectContext(Context, Snapshot, Request) (ProjectContextPlan, error) {
	return ProjectContextPlan{}, NewRefusal(OperationNotPorted)
}

// Status refuses this operation until its harness surface is ported.
func (UnportedOperations) Status(Context, Snapshot) (Status, error) {
	return Status{}, NewRefusal(OperationNotPorted)
}

// Drift refuses this operation until its harness surface is ported.
func (UnportedOperations) Drift(Context, Snapshot) (DriftReport, error) {
	return DriftReport{}, NewRefusal(OperationNotPorted)
}
