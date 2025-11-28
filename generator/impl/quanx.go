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

// QuantumultXGenerator implements the Generator interface for Quantumult X
type QuantumultXGenerator struct{}

func init() {
	core.RegisterGenerator(&QuantumultXGenerator{})
}

// Name returns the generator name
func (g *QuantumultXGenerator) Name() string {
	return "quanx"
}

// Generate produces the Quantumult X configuration
func (g *QuantumultXGenerator) Generate(proxies []pc.ProxyInterface, groups []config.ProxyGroupConfig, rules []string, global *config.Settings, opts core.GeneratorOptions) (string, error) {
	log.Printf("Generating Quantumult X config for %d proxies", len(proxies))
	var output strings.Builder

	// Add base configuration
	output.WriteString(opts.Base)
	output.WriteString("\n\n[server_local]\n")

	// Generate proxy section
	var validProxies []pc.ProxyInterface
	for _, proxy := range proxies {
		line := convertToQuantumultX(proxy, opts)
		if line != "" {
			output.WriteString(line)
			output.WriteString("\n")
			validProxies = append(validProxies, proxy)
		} else {
			log.Printf("Proxy %s skipped for Quantumult X (not supported)", proxy.GetRemark())
		}
	}

	// Generate policy section
	if len(groups) > 0 {
		output.WriteString("\n[policy]\n")

		for _, group := range groups {
			line := convertGroupToQuantumultX(group, validProxies)
			if line != "" {
				output.WriteString(line)
				output.WriteString("\n")
			}
		}
	}

	// Generate filter rules
	if opts.Rule && len(opts.Rulesets) > 0 {
		output.WriteString("\n[filter_local]\n")
		for _, ruleset := range opts.Rulesets {
			output.WriteString(fmt.Sprintf("# %s rules\n", ruleset.Ruleset))
		}
		output.WriteString(fmt.Sprintf("FINAL,%s\n", "DIRECT"))
	}

	return output.String(), nil
}

func convertToQuantumultX(p pc.ProxyInterface, opts core.GeneratorOptions) string {
	if mixin, ok := p.(pc.QuantumultXConvertableMixin); ok {
		config, err := mixin.ToQuantumultXConfig(&opts.ProxySetting)
		if err != nil {
			log.Printf("Failed to convert proxy %s to Quantumult X: %v", p.GetRemark(), err)
			return ""
		}
		return config
	}
	log.Printf("Proxy %s skipped for Quantumult X (not supported)", p.GetRemark())
	return ""
}

func convertGroupToQuantumultX(group config.ProxyGroupConfig, proxies []pc.ProxyInterface) string {
	groupType := strings.ToLower(group.Type)
	if groupType == "select" {
		groupType = "static"
	} else if groupType == "url-test" {
		groupType = "available"
	}

	var parts []string
	parts = append(parts, fmt.Sprintf("%s=%s", groupType, group.Name))

	// Filter proxies using advanced filtering
	filtered := utils.FilterProxiesByRules(proxies, group.Rule)
	if len(filtered) == 0 {
		filtered = []string{"direct"}
	}
	parts = append(parts, filtered...)

	// Add img-url if needed
	parts = append(parts, "img-url=https://raw.githubusercontent.com/Koolson/Qure/master/IconSet/Proxy.png")

	return strings.Join(parts, ", ")
}
