package alibaba

import (
	"github.com/relux-works/archon-pantheon/pkg/vendorplugin"
	"github.com/relux-works/archon-pantheon/pkg/vendorplugin/benchdata"
)

// Model facts are declared once in benchdata; this adapter adds runtime types.
var models = vendorplugin.ModelsFromBenchdata(benchdata.AlibabaModels())
