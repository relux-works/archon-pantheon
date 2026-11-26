package moonshot

import (
	"github.com/relux-works/curator-runtimes/pkg/agentic"
	"github.com/relux-works/curator-runtimes/pkg/vendorplugin"
)

const moonshotCatalogSource = "MoonshotAI Kimi Code CLI 2.8.1 config.toml (managed:kimi-code) & https://www.kimi.com/code/docs/en/ (checked 2026-11-26)"

func moonshotUnmeasured(id, note string) *vendorplugin.UnmeasuredClaim {
	return &vendorplugin.UnmeasuredClaim{
		Kind:        "unmeasured",
		AsOf:        "2026-11-26",
		Source:      moonshotCatalogSource,
		Observation: id + ": " + note + "; capability is unmeasured on the Bug Hunt Bench",
	}
}

var moonshotModels = []vendorplugin.Model{
	{
		ID:          "kimi-code/kimi-for-coding",
		Description: "Default Kimi Code CLI coding model alias",
		Unmeasured:  moonshotUnmeasured("kimi-code/kimi-for-coding", "default managed:kimi-code alias in kimi 2.8.1"),
		Lifecycle:   vendorplugin.LifecycleCurrent,
		Effort:      vendorplugin.EffortDeclaration{Support: agentic.EffortSupportNone},
		Systems:     []agentic.SystemID{"kimi"},
		Publisher:   "moonshot",
	},
	{
		ID:          "kimi-code/kimi-for-coding-highspeed",
		Description: "Kimi Code CLI high-speed coding model alias",
		Unmeasured:  moonshotUnmeasured("kimi-code/kimi-for-coding-highspeed", "high-speed managed:kimi-code alias in kimi 2.8.1"),
		Lifecycle:   vendorplugin.LifecycleCurrent,
		Effort:      vendorplugin.EffortDeclaration{Support: agentic.EffortSupportNone},
		Systems:     []agentic.SystemID{"kimi"},
		Publisher:   "moonshot",
	},
	{
		ID:          "kimi-code/k3.1-pro",
		Description: "MoonshotAI Kimi K3.1 Pro model via Kimi Code CLI",
		Unmeasured:  moonshotUnmeasured("kimi-code/k3.1-pro", "k3.1-pro managed:kimi-code alias in kimi 2.8.1"),
		Lifecycle:   vendorplugin.LifecycleCurrent,
		Effort:      vendorplugin.EffortDeclaration{Support: agentic.EffortSupportNone},
		Systems:     []agentic.SystemID{"kimi"},
		Publisher:   "moonshot",
	},
	{
		ID:          "kimi-code/k3.1-pro-highspeed",
		Description: "MoonshotAI Kimi K3.1 Pro high-speed model via Kimi Code CLI",
		Unmeasured:  moonshotUnmeasured("kimi-code/k3.1-pro-highspeed", "k3.1-pro-highspeed managed:kimi-code alias in kimi 2.8.1"),
		Lifecycle:   vendorplugin.LifecycleCurrent,
		Effort:      vendorplugin.EffortDeclaration{Support: agentic.EffortSupportNone},
		Systems:     []agentic.SystemID{"kimi"},
		Publisher:   "moonshot",
	},
	{
		ID:          "kimi-code/k3",
		Description: "MoonshotAI Kimi K3 model via Kimi Code CLI",
		Unmeasured:  moonshotUnmeasured("kimi-code/k3", "k3 managed:kimi-code alias in kimi 2.8.1"),
		Lifecycle:   vendorplugin.LifecycleCurrent,
		Effort:      vendorplugin.EffortDeclaration{Support: agentic.EffortSupportNone},
		Systems:     []agentic.SystemID{"kimi"},
		Publisher:   "moonshot",
	},
	{
		ID:          "kimi-code/k3-256k",
		Description: "MoonshotAI Kimi K3 256k context model via Kimi Code CLI",
		Unmeasured:  moonshotUnmeasured("kimi-code/k3-256k", "k3-256k managed:kimi-code alias in kimi 2.8.1"),
		Lifecycle:   vendorplugin.LifecycleCurrent,
		Effort:      vendorplugin.EffortDeclaration{Support: agentic.EffortSupportNone},
		Systems:     []agentic.SystemID{"kimi"},
		Publisher:   "moonshot",
	},
}
