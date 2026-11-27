package agentic_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/relux-works/archon-pantheon/pkg/agentic"
)

// Gateway refusal codes preserve the sentinel, expose only known check classes, and never echo caller text.
func TestCodexGatewayRefusalCodesAreBounded(t *testing.T) {
	codes := []agentic.CodexGatewayRefusalCode{
		agentic.CodexGatewayBasisInvalid,
		agentic.CodexGatewayInputNotAdmitted,
		agentic.CodexGatewayEnvironmentNotAdmitted,
		agentic.CodexGatewayBindingNotCanonical,
		agentic.CodexGatewayPlanIdentityChanged,
		agentic.CodexGatewayExecutableIdentityFailed,
		agentic.CodexGatewayInteractiveModeRequired,
		agentic.CodexGatewayOtherRouteSelected,
		agentic.CodexGatewayNativeArgumentsNotAdmitted,
		agentic.CodexGatewayEnvironmentMalformed,
		agentic.CodexGatewayEnvironmentDuplicate,
		agentic.CodexGatewayDirectProxyPresent,
		agentic.CodexGatewayTokenOverride,
		agentic.CodexGatewayProviderOverride,
		agentic.CodexGatewayCredentialNameNotAdmitted,
		agentic.CodexGatewayInjectedTokenMissing,
		agentic.CodexGatewayEntryNotInteractive,
		agentic.CodexGatewayOperationCancelled,
	}
	for _, code := range codes {
		err := agentic.RefuseCodexGateway(code)
		var refusal *agentic.CodexGatewayRefusalError
		if !errors.Is(err, agentic.ErrCodexGatewayRefused) || !errors.As(err, &refusal) || refusal.ReasonCode() != code {
			t.Fatalf("refusal %q did not preserve its bounded typed code: %v", code, err)
		}
	}
	err := agentic.RefuseCodexGateway(agentic.CodexGatewayRefusalCode("secret-input"))
	var refusal *agentic.CodexGatewayRefusalError
	if !errors.As(err, &refusal) || refusal.ReasonCode() != agentic.CodexGatewayInputNotAdmitted || strings.Contains(err.Error(), "secret-input") {
		t.Fatalf("unknown reason was not reduced to a safe check class: %v", err)
	}
}
