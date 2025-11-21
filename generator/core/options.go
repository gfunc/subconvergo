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

	// Legacy options (to be deprecated)
	// SortProxies     bool
	// RenameNodes     []config.RenameNodeConfig
	// Emoji           config.EmojiConfig

	// Additional options from legacy core
	Type      string
	Sort      string
	Filter    string
	Include   string
	Exclude   string
	Config    string
	Upload    bool
	Token     string
	EmojiFlag bool // Renamed from Emoji to avoid conflict
	List      bool
	Expand    bool
	Classic   bool
	NewName   bool
	Append    string
	Insert    string
	Fd        bool
	SortFlag  bool
	Rename    string

	Base     string
	IncludeR string
	ExcludeR string
	Rule     bool
	Style    string
	Inserts  string
}
