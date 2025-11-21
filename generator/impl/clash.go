package impl

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/generator/core"
	"github.com/gfunc/subconvergo/generator/utils"
	pc "github.com/gfunc/subconvergo/proxy/core"
	"github.com/gfunc/subconvergo/proxy/impl"
	C "github.com/metacubex/mihomo/config"
	R "github.com/metacubex/mihomo/rules"
	RC "github.com/metacubex/mihomo/rules/common"
	"gopkg.in/yaml.v3"
)

// ClashGenerator implements the Generator interface for Clash
type ClashGenerator struct{}

func init() {
	core.RegisterGenerator(&ClashGenerator{})
}

// Name returns the generator name
func (g *ClashGenerator) Name() string {
	return "clash"
}

// Generate produces the Clash configuration
func (g *ClashGenerator) Generate(proxies []pc.ProxyInterface, groups []config.ProxyGroupConfig, rules []string, global *config.Settings, opts core.GeneratorOptions) (string, error) {
	baseConfig := opts.Base
	var base = make(map[string]any)
	if err := yaml.Unmarshal([]byte(baseConfig), &base); err != nil {
		log.Printf("[ClashGenerator] failed to parse base config err=%v", err)
		return "", fmt.Errorf("failed to parse base config: %w", err)
	}
	// Parse clash configuration
	var clash = &C.RawConfig{}
	if err := yaml.Unmarshal([]byte(baseConfig), clash); err != nil {
		log.Printf("[ClashGenerator] failed to parse base config err=%v", err)
		return "", fmt.Errorf("failed to parse base config: %w", err)
	}

	// Convert proxies to Clash format
	var clashProxies []map[string]interface{}
	for _, p := range proxies {
		switch c := p.(type) {
		case *impl.MihomoProxy:
			clashProxies = append(clashProxies, c.ToClashConfig(&opts.ProxySetting))
		case pc.SubconverterProxy:
			clashProxies = append(clashProxies, c.ToClashConfig(&opts.ProxySetting))
		default:
			log.Printf("[ClashGenerator] unsupported proxy type=%T remark=%s", p, p.GetRemark())
		}
	}

	// Set proxies field name based on Clash config
	clash.Proxy = clashProxies
	base[utils.GetFieldTag("yaml", "Proxy", clash, "proxies")] = clashProxies

	// Generate proxy groups
	if len(groups) > 0 {
		proxyGroups := generateClashProxyGroups(proxies, groups)
		clash.ProxyGroup = proxyGroups
		base[utils.GetFieldTag("yaml", "ProxyGroup", clash, "proxy-groups")] = proxyGroups
	}

	// Generate rules if enabled
	if opts.Rule {
		allRules := make([]string, 0)
		if len(rules) > 0 {
			allRules = append(allRules, rules...)
		}
		if len(opts.Rulesets) > 0 {
			allRules = append(allRules, generateClashRules(opts.Rulesets)...)
		}
		clash.Rule = allRules
		base[utils.GetFieldTag("yaml", "Rule", clash, "rules")] = allRules
	}

	// Marshal back to YAML
	output, err := yaml.Marshal(base)
	if err != nil {
		return "", fmt.Errorf("failed to marshal output: %w", err)
	}

	return string(output), nil
}

func generateClashProxyGroups(proxies []pc.ProxyInterface, groups []config.ProxyGroupConfig) []map[string]interface{} {
	var result []map[string]interface{}

	for _, group := range groups {
		clashGroup := map[string]interface{}{
			"name": group.Name,
			"type": group.Type,
		}

		// Filter proxies based on group rules using advanced filtering
		filteredProxies := utils.FilterProxiesByRules(proxies, group.Rule)
		if len(filteredProxies) == 0 {
			// Add DIRECT or REJECT if no proxies matched
			filteredProxies = []string{"DIRECT"}
		}
		clashGroup["proxies"] = filteredProxies

		// Add type-specific fields
		groupType := strings.ToLower(group.Type)
		switch groupType {
		case "url-test", "fallback", "load-balance":
			if group.URL != "" {
				clashGroup["url"] = group.URL
			}
			if group.Interval > 0 {
				clashGroup["interval"] = group.Interval
			}
			if group.Tolerance > 0 {
				clashGroup["tolerance"] = group.Tolerance
			}
			if group.Timeout > 0 {
				clashGroup["timeout"] = group.Timeout
			}
			// Add strategy for load-balance
			if groupType == "load-balance" && group.Strategy != "" {
				clashGroup["strategy"] = group.Strategy
			}
		}

		// Add lazy option if specified
		if group.Lazy != nil {
			clashGroup["lazy"] = *group.Lazy
		}

		// Add disable-udp option if specified
		if group.DisableUDP != nil {
			clashGroup["disable-udp"] = *group.DisableUDP
		}

		result = append(result, clashGroup)
	}

	return result
}

// generateClashRules generates Clash rules from ruleset configs
func generateClashRules(rulesets []config.RulesetConfig) []string {
	var rules []string
	hasFinalRule := false

	for _, ruleset := range rulesets {
		// Check if it's a direct rule (not a ruleset)
		if ruleset.Rule != "" {
			// Process special rule formats
			rule := ruleset.Rule

			// Handle []FINAL -> MATCH
			if strings.HasPrefix(rule, "[]FINAL") {
				rule = "MATCH," + ruleset.Group
				hasFinalRule = true
			} else if strings.HasPrefix(rule, "[]") {
				// Handle []GEOIP,CN -> GEOIP,CN or other [] prefixed rules
				// Remove the [] prefix and append group
				rule = strings.TrimPrefix(rule, "[]")
				if ruleset.Group != "" && !strings.Contains(rule, ","+ruleset.Group) {
					rule = rule + "," + ruleset.Group
				}
			}

			// Validate with mihomo clash
			if validateClashRule(rule) {
				rules = append(rules, rule)
			}
			continue
		}

		// Fetch and convert ruleset
		if ruleset.Ruleset != "" {
			content, err := utils.FetchRuleset(ruleset.Ruleset)
			if err != nil {
				log.Printf("[ClashGenerator] ruleset fetch failed %s err=%v", ruleset.Ruleset, err)
				// If fetch fails, add as RULE-SET reference
				rule := fmt.Sprintf("RULE-SET,%s,%s", ruleset.Ruleset, ruleset.Group)
				rules = append(rules, rule)
				continue
			}

			// Convert ruleset content to rules
			rulesetRules := convertRulesetToClash(content)
			for _, r := range rulesetRules {
				// Append group if not already present
				if !strings.Contains(r, ","+ruleset.Group) && ruleset.Group != "" {
					// Count commas to determine where to add group

					parts := strings.Split(r, ",")
					if len(parts) == 2 {
						// Add group as last parameter
						r = r + "," + ruleset.Group
					} else if len(parts) > 2 {
						// Insert group before last parameter
						r = strings.Join(parts[:len(parts)-1], ",") + "," + ruleset.Group + "," + parts[len(parts)-1]
					}
				}
				if validateClashRule(r) {
					rules = append(rules, r)
				}
			}
		}
	}

	// Only add final MATCH rule if there's no explicit FINAL rule
	if !hasFinalRule {
		rules = append(rules, "MATCH,DIRECT")
	}
	return rules
}

func validateClashRule(rule string) bool {
	tp, payload, target, params := RC.ParseRulePayload(rule, true)
	if strings.Contains(tp, "GEO") {
		return true
	}
	if target != "" {
		_, err := R.ParseRule(tp, payload, target, params, nil)
		return err == nil
	}
	return false
}

// convertRulesetToClash converts ruleset content to Clash format
// Handles Clash payload format, QuanX format, and Surge format
func convertRulesetToClash(content string) []string {
	var rules []string
	lines := strings.Split(content, "\n")

	// Check if it's Clash payload format
	isPayload := false
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "payload:") {
			isPayload = true
			break
		}
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "//") {
			continue
		}

		if strings.HasPrefix(line, "payload:") {
			continue
		}

		// Handle Clash payload format
		if isPayload {
			// Remove leading dash and quotes
			line = strings.TrimPrefix(line, "-")
			line = strings.TrimSpace(line)
			line = strings.Trim(line, "'\"")

			// Check if line already has rule type
			if strings.Contains(line, ",") {
				rules = append(rules, line)
				continue
			}

			// Detect rule type from content
			if strings.Contains(line, "/") {
				// IP-CIDR
				if strings.Count(line, ":") > 1 {
					rules = append(rules, "IP-CIDR6,"+line)
				} else {
					rules = append(rules, "IP-CIDR,"+line)
				}
			} else if strings.HasPrefix(line, ".") || strings.HasPrefix(line, "+.") {
				// Domain suffix
				line = strings.TrimPrefix(line, "+.")
				line = strings.TrimPrefix(line, ".")
				if strings.HasSuffix(line, ".*") {
					// Domain keyword
					line = strings.TrimSuffix(line, ".*")
					rules = append(rules, "DOMAIN-KEYWORD,"+line)
				} else {
					rules = append(rules, "DOMAIN-SUFFIX,"+line)
				}
			} else {
				// Full domain
				rules = append(rules, "DOMAIN,"+line)
			}
		} else {
			// Surge/QuanX format - already has rule type
			// Convert QuanX HOST to DOMAIN
			line = regexp.MustCompile(`^(?i:host)`).ReplaceAllString(line, "DOMAIN")
			line = regexp.MustCompile(`^(?i:host-suffix)`).ReplaceAllString(line, "DOMAIN-SUFFIX")
			line = regexp.MustCompile(`^(?i:host-keyword)`).ReplaceAllString(line, "DOMAIN-KEYWORD")
			line = regexp.MustCompile(`^(?i:ip6-cidr)`).ReplaceAllString(line, "IP-CIDR6")
			// remove // comments at end of line
			if idx := strings.Index(line, "//"); idx != -1 {
				line = strings.TrimSpace(line[:idx])
			}
			rules = append(rules, line)
		}
	}

	return rules
}
