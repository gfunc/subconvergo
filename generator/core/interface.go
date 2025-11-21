package core

import (
	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/proxy/core"
)

// Generator defines the interface for generating configuration output
type Generator interface {
	// Name returns the generator name (e.g., "Clash", "Surge")
	Name() string
	// Generate produces the configuration content
	Generate(proxies []core.ProxyInterface, groups []config.ProxyGroupConfig, rules []string, global *config.Settings, opts GeneratorOptions) (string, error)
}
