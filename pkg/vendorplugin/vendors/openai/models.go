package openai

import (
	"github.com/relux-works/curator-runtimes/pkg/vendorplugin"
	"github.com/relux-works/curator-runtimes/pkg/vendorplugin/benchdata"
)

// Model facts are declared once in benchdata; this adapter adds runtime types.
var models = vendorplugin.ModelsFromBenchdata(benchdata.OpenAIModels())
