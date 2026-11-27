// Package moonshot is the vendor plugin for MoonshotAI Kimi Code models.
package moonshot

import (
	"fmt"

	"github.com/relux-works/archon-pantheon/pkg/agentic"
	"github.com/relux-works/archon-pantheon/pkg/vendorplugin"

	_ "github.com/relux-works/archon-pantheon/pkg/agentic/systems/kimi"
)

// VendorID is the normalized MoonshotAI vendor identifier.
const VendorID vendorplugin.VendorID = "moonshot"

// RuntimeID is the Kimi Code CLI runtime identifier.
const RuntimeID vendorplugin.RuntimeID = "kimi"

// Vendor is the MoonshotAI vendor plugin. It holds no mutable state.
type Vendor struct{}

// New returns a fresh MoonshotAI vendor plugin instance.
func New() *Vendor { return &Vendor{} }

// RuntimeDeclaration returns the runtime pairing Kimi Code CLI with MoonshotAI.
func RuntimeDeclaration() vendorplugin.RuntimeDeclaration {
	return vendorplugin.RuntimeDeclaration{
		ID:     RuntimeID,
		System: "kimi",
		Vendor: VendorID,
		Broker: vendorplugin.BrokerProvenance{
			Checked: []string{moonshotCatalogSource},
			Found:   "MoonshotAI Kimi Code CLI managed:kimi-code provider (observed on kimi 2.8.1)",
		},
	}
}

func init() {
	if err := vendorplugin.Register(New()); err != nil {
		panic(fmt.Sprintf("moonshot: registering the vendor plugin: %v", err))
	}
}

func (*Vendor) ID() vendorplugin.VendorID { return VendorID }

func (*Vendor) Models() []vendorplugin.Model { return vendorplugin.CloneModels(moonshotModels) }

func (*Vendor) Availability(vendorplugin.AvailabilityQuery) (vendorplugin.Availability, error) {
	return vendorplugin.Unchecked(), nil
}

func (*Vendor) Spawn(sc vendorplugin.SpawnContext) (agentic.LaunchRequest, error) {
	return vendorplugin.PassthroughLaunch(sc), nil
}
