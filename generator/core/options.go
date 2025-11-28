package core

import (
	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/generator/transformers"
)

// GeneratorOptions contains options for proxy generation
type GeneratorOptions struct {
	config.ProxySetting

	// Core options
	Target          string
	ProxyGroups     []config.ProxyGroupConfig
	Rulesets        []config.RulesetConfig
	RawRules        []string
	AppendProxyType bool
	EnableRuleGen   bool
	Pipelines       []transformers.Transformer

	// Additional options from legacy core
	Type     string
	Base     string
	Rule     bool
	SurgeVer int
}
