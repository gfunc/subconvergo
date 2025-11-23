package impl

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/generator/core"
	pc "github.com/gfunc/subconvergo/proxy/core"
)

// SingBoxGenerator implements the Generator interface for sing-box
type SingBoxGenerator struct{}

func init() {
	core.RegisterGenerator(&SingBoxGenerator{})
}

// Name returns the generator name
func (g *SingBoxGenerator) Name() string {
	return "singbox"
}

// Generate produces the sing-box configuration
func (g *SingBoxGenerator) Generate(proxies []pc.ProxyInterface, groups []config.ProxyGroupConfig, rules []string, global *config.Settings, opts core.GeneratorOptions) (string, error) {
	log.Printf("Generating sing-box config for %d proxies", len(proxies))
	// Parse base configuration as JSON
	var base map[string]interface{}
	if err := json.Unmarshal([]byte(opts.Base), &base); err != nil {
		return "", fmt.Errorf("failed to parse base config: %w", err)
	}

	// Convert proxies to sing-box outbounds
	var outbounds []map[string]interface{}

	// Add DIRECT and REJECT outbounds
	outbounds = append(outbounds, map[string]interface{}{
		"type": "direct",
		"tag":  "DIRECT",
	})
	outbounds = append(outbounds, map[string]interface{}{
		"type": "block",
		"tag":  "REJECT",
	})

	// Add proxy outbounds
	for _, proxy := range proxies {
		outbound := convertToSingBox(proxy, opts)
		if outbound != nil {
			outbounds = append(outbounds, outbound)
		} else {
			log.Printf("Proxy %s skipped for sing-box (not supported)", proxy.GetRemark())
		}
	}

	base["outbounds"] = outbounds

	// Generate routing rules if enabled
	if opts.Rule && len(opts.Rulesets) > 0 {
		rules := generateSingBoxRules(opts.Rulesets)
		if route, ok := base["route"].(map[string]interface{}); ok {
			route["rules"] = rules
		}
	}

	// Add clash_mode if enabled
	if opts.ProxySetting.SingBoxAddClashMode {
		experimental, ok := base["experimental"].(map[string]interface{})
		if !ok {
			experimental = make(map[string]interface{})
			base["experimental"] = experimental
		}

		clashApi, ok := experimental["clash_api"].(map[string]interface{})
		if !ok {
			clashApi = make(map[string]interface{})
			experimental["clash_api"] = clashApi
		}

		clashApi["default_mode"] = "rule"
	}

	// Marshal back to JSON with proper indentation
	output, err := json.MarshalIndent(base, "", "    ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal output: %w", err)
	}

	return string(output), nil
}

func convertToSingBox(p pc.ProxyInterface, opts core.GeneratorOptions) map[string]interface{} {
	if mixin, ok := p.(pc.SingboxConvertableMixin); ok {
		config, err := mixin.ToSingboxConfig(&opts.ProxySetting)
		if err != nil {
			log.Printf("Failed to convert proxy %s to sing-box: %v", p.GetRemark(), err)
			return nil
		}
		return config
	}
	log.Printf("Proxy %s skipped for sing-box (not supported)", p.GetRemark())
	return nil
}

func generateSingBoxRules(rulesets []config.RulesetConfig) []map[string]interface{} {
	var rules []map[string]interface{}

	for _, ruleset := range rulesets {
		rule := map[string]interface{}{
			"rule_set": []string{ruleset.Ruleset},
			"outbound": ruleset.Group,
		}
		rules = append(rules, rule)
	}

	// Add final rule
	rules = append(rules, map[string]interface{}{
		"outbound": "DIRECT",
	})

	return rules
}
