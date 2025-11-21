package generator

import (
	"fmt"
	"log"

	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/generator/core"
	pc "github.com/gfunc/subconvergo/proxy/core"
)

// Generate converts proxies to target format
func Generate(proxies []pc.ProxyInterface, opts core.GeneratorOptions, baseConfig string) (string, error) {
	log.Printf("[generator.Generate] target=%s proxies=%d ruleGen=%t", opts.Target, len(proxies), opts.EnableRuleGen)

	// Execute pipeline
	var err error
	for _, t := range opts.Pipelines {
		proxies, err = t.Transform(proxies, config.Global)
		if err != nil {
			return "", fmt.Errorf("transformation failed: %v", err)
		}
	}

	// Append proxy type to remark if enabled
	if opts.AppendProxyType {
		for i := range proxies {
			proxies[i].SetRemark(fmt.Sprintf("%s [%s]", proxies[i].GetRemark(), proxies[i].GetType()))
		}
	}

	target := opts.Target
	if target == "clashr" {
		target = "clash"
	}

	gen, err := core.GetGenerator(target)
	if err != nil {
		return "", fmt.Errorf("unsupported target: %s", target)
	}

	// Populate additional fields for the generator
	opts.Base = baseConfig
	opts.Type = target
	opts.Rule = opts.EnableRuleGen

	return gen.Generate(proxies, opts.ProxyGroups, opts.RawRules, config.Global, opts)
}
