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

// SurgeGenerator implements the Generator interface for Surge
type SurgeGenerator struct{}

func init() {
	core.RegisterGenerator(&SurgeGenerator{})
}

// Name returns the generator name
func (g *SurgeGenerator) Name() string {
	return "surge"
}

// Generate produces the Surge configuration
func (g *SurgeGenerator) Generate(proxies []pc.ProxyInterface, groups []config.ProxyGroupConfig, rules []string, global *config.Settings, opts core.GeneratorOptions) (string, error) {
	log.Printf("Generating Surge config for %d proxies", len(proxies))
	var output strings.Builder

	// Add base configuration
	output.WriteString(opts.Base)
	output.WriteString("\n\n[Proxy]\n")

	// Generate proxy section
	var validProxies []pc.ProxyInterface
	for _, proxy := range proxies {
		line := convertToSurge(proxy, opts)
		if line != "" {
			output.WriteString(line)
			output.WriteString("\n")
			validProxies = append(validProxies, proxy)
		} else {
			log.Printf("Proxy %s skipped for Surge (not supported)", proxy.GetRemark())
		}
	}

	// Generate proxy groups
	if len(groups) > 0 {
		output.WriteString("\n[Proxy Group]\n")

		for _, group := range groups {
			line := convertGroupToSurge(group, validProxies)
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
			// Surge rules format: RULE-SET,ruleset-name,policy
			output.WriteString(fmt.Sprintf("RULE-SET,%s,%s\n", ruleset.Ruleset, ruleset.Group))
		}
		output.WriteString("FINAL,DIRECT\n")
	}

	return output.String(), nil
}

func convertToSurge(p pc.ProxyInterface, opts core.GeneratorOptions) string {
	if mixin, ok := p.(pc.SurgeConvertableMixin); ok {
		config, err := mixin.ToSurgeConfig(&opts.ProxySetting)
		if err != nil {
			log.Printf("Failed to convert proxy %s to Surge: %v", p.GetRemark(), err)
			return ""
		}
		return config
	}
	log.Printf("Proxy %s skipped for Surge (not supported)", p.GetRemark())
	return ""
}

func convertGroupToSurge(group config.ProxyGroupConfig, proxies []pc.ProxyInterface) string {
	var parts []string

	// Map group types to Surge types
	surgeType := strings.ToLower(group.Type)
	if surgeType == "selector" {
		surgeType = "select"
	}
	parts = append(parts, surgeType)

	// Filter proxies based on group rules using advanced filtering
	filtered := utils.FilterProxiesByRules(proxies, group.Rule)
	if len(filtered) == 0 {
		filtered = []string{"DIRECT"}
	}
	parts = append(parts, filtered...)

	// Add type-specific options
	switch surgeType {
	case "url-test", "fallback", "load-balance":
		if group.URL != "" {
			parts = append(parts, fmt.Sprintf("url=%s", group.URL))
		}
		if group.Interval > 0 {
			parts = append(parts, fmt.Sprintf("interval=%d", group.Interval))
		}
		if group.Tolerance > 0 {
			parts = append(parts, fmt.Sprintf("tolerance=%d", group.Tolerance))
		}
	}

	return fmt.Sprintf("%s = %s", group.Name, strings.Join(parts, ", "))
}
