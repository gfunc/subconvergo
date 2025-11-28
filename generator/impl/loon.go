package impl

import (
	"fmt"
	"log"
	"strings"

	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/generator/core"
	"github.com/gfunc/subconvergo/generator/utils"
	pc "github.com/gfunc/subconvergo/proxy/core"
)

// LoonGenerator implements the Generator interface for Loon
type LoonGenerator struct{}

func init() {
	core.RegisterGenerator(&LoonGenerator{})
}

// Name returns the generator name
func (g *LoonGenerator) Name() string {
	return "loon"
}

// Generate produces the Loon configuration
func (g *LoonGenerator) Generate(proxies []pc.ProxyInterface, groups []config.ProxyGroupConfig, rules []string, global *config.Settings, opts core.GeneratorOptions) (string, error) {
	log.Printf("Generating Loon config for %d proxies", len(proxies))
	// Loon format is very similar to Surge
	var output strings.Builder

	output.WriteString(opts.Base)
	output.WriteString("\n\n[Proxy]\n")

	// Generate proxy section (same format as Surge)
	var validProxies []pc.ProxyInterface
	for _, proxy := range proxies {
		line := convertToLoon(proxy, opts)
		if line != "" {
			output.WriteString(line)
			output.WriteString("\n")
			validProxies = append(validProxies, proxy)
		} else {
			log.Printf("Proxy %s skipped for Loon (not supported)", proxy.GetRemark())
		}
	}

	// Generate proxy groups
	if len(groups) > 0 {
		output.WriteString("\n[Proxy Group]\n")

		for _, group := range groups {
			line := convertGroupToLoon(group, validProxies)
			if line != "" {
				output.WriteString(line)
				output.WriteString("\n")
			}
		}
	}

	// Generate rules
	if opts.Rule && len(opts.Rulesets) > 0 {
		output.WriteString("\n[Rule]\n")
		for _, ruleset := range opts.Rulesets {
			output.WriteString(fmt.Sprintf("RULE-SET,%s,%s\n", ruleset.Ruleset, ruleset.Group))
		}
		output.WriteString("FINAL,DIRECT\n")
	}

	return output.String(), nil
}

func convertGroupToLoon(group config.ProxyGroupConfig, proxies []pc.ProxyInterface) string {
	var sb strings.Builder
	sb.WriteString(group.Name)
	sb.WriteString(" = ")
	sb.WriteString(strings.ToLower(group.Type))

	filtered := utils.FilterProxiesByRules(proxies, group.Rule)
	if len(filtered) == 0 {
		filtered = []string{"DIRECT"}
	}

	for _, p := range filtered {
		sb.WriteString(", ")
		sb.WriteString(p)
	}

	sb.WriteString(", img-url=https://raw.githubusercontent.com/Koolson/Qure/master/IconSet/Proxy.png")

	return sb.String()
}

// convertToLoon converts a single proxy to Loon format
func convertToLoon(proxy pc.ProxyInterface, opts core.GeneratorOptions) string {
	if mixin, ok := proxy.(pc.LoonConvertableMixin); ok {
		config, err := mixin.ToLoonConfig(&opts.ProxySetting)
		if err != nil {
			log.Printf("Failed to convert proxy %s to Loon: %v", proxy.GetRemark(), err)
			return ""
		}
		return config
	}
	log.Printf("Proxy %s skipped for Loon (not supported)", proxy.GetRemark())
	return ""
}
