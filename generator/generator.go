package generator

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/gfunc/subconvergo/proxy"

	"github.com/gfunc/subconvergo/config"
	C "github.com/metacubex/mihomo/config"
	R "github.com/metacubex/mihomo/rules"
	RC "github.com/metacubex/mihomo/rules/common"
	"gopkg.in/yaml.v3"
)

// GeneratorOptions contains options for proxy generation
type GeneratorOptions struct {
	Target              string
	ProxyGroups         []config.ProxyGroupConfig
	Rulesets            []config.RulesetConfig
	RawRules            []string
	AppendProxyType     bool
	EnableRuleGen       bool
	ClashProxiesStyle   string
	ClashGroupsStyle    string
	SingBoxAddClashMode bool
	UDP                 *bool
	TFO                 *bool
	SkipCertVerify      *bool
	TLS13               *bool
	NodeList            bool
}

// Generate converts proxies to target format
func Generate(proxies []proxy.ProxyInterface, opts GeneratorOptions, baseConfig string) (string, error) {
	switch opts.Target {
	case "clash", "clashr":
		return generateClash(proxies, opts, baseConfig)
	case "surge":
		return generateSurge(proxies, opts, baseConfig)
	case "quanx":
		return generateQuantumultX(proxies, opts, baseConfig)
	case "loon":
		return generateLoon(proxies, opts, baseConfig)
	case "singbox":
		return generateSingBox(proxies, opts, baseConfig)
	case "ss", "ssr", "v2ray", "trojan":
		return generateSingle(proxies, opts.Target)
	default:
		return "", fmt.Errorf("unsupported target: %s", opts.Target)
	}
}

func generateClash(proxies []proxy.ProxyInterface, opts GeneratorOptions, baseConfig string) (string, error) {
	// Parse base configuration
	var base = C.RawConfig{}
	if err := yaml.Unmarshal([]byte(baseConfig), &base); err != nil {
		return "", fmt.Errorf("failed to parse base config: %w", err)
	}
	// Convert proxies to Clash format
	var clashProxies []map[string]interface{}
	for _, p := range proxies {
		switch c := p.(type) {
		case *proxy.MihomoProxy:
			clashProxies = append(clashProxies, c.ProxyOptions())
		case proxy.SubconverterProxy:
			clashProxies = append(clashProxies, c.ProxyOptions())
		default:
			log.Printf("unsupported proxy type for clash generation: %T", p)
		}
	}

	// Set proxies field name based on Clash version
	base.Proxy = clashProxies

	// Generate proxy groups
	if len(opts.ProxyGroups) > 0 {
		proxyGroups := generateClashProxyGroups(proxies, opts.ProxyGroups, opts)
		base.ProxyGroup = proxyGroups
	}

	// Generate rules if enabled
	if opts.EnableRuleGen {
		allRules := make([]string, 0)
		if len(opts.RawRules) > 0 {
			allRules = append(allRules, opts.RawRules...)
		}
		if len(opts.Rulesets) > 0 {
			allRules = append(allRules, generateClashRules(opts.Rulesets)...)
		}
		base.Rule = allRules
	}

	// Marshal back to YAML
	output, err := yaml.Marshal(base)
	if err != nil {
		return "", fmt.Errorf("failed to marshal output: %w", err)
	}

	return string(output), nil
}

func generateClashProxyGroups(proxies []proxy.ProxyInterface, groups []config.ProxyGroupConfig, opts GeneratorOptions) []map[string]interface{} {
	var result []map[string]interface{}

	for _, group := range groups {
		clashGroup := map[string]interface{}{
			"name": group.Name,
			"type": group.Type,
		}

		// Filter proxies based on group rules using advanced filtering
		filteredProxies := filterProxiesByRules(proxies, group.Rule)
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

// applyMatcher checks if a proxy matches a rule pattern
// Supports special matchers: !!GROUP=, !!GROUPID=, !!TYPE=, !!PORT=, !!SERVER=
// Returns true if the proxy matches, and extracts the real regex pattern if present
func applyMatcher(rule string, proxy proxy.ProxyInterface) (bool, string) {
	// Handle special [] syntax for direct/reject nodes
	if strings.HasPrefix(rule, "[]") {
		return false, rule[2:]
	}

	// Handle !!GROUP= matcher
	if strings.HasPrefix(rule, "!!GROUP=") {
		parts := strings.SplitN(rule, "!!", 3)
		if len(parts) >= 2 {
			groupPattern := strings.TrimPrefix(parts[1], "GROUP=")
			realRule := ""
			if len(parts) > 2 {
				realRule = parts[2]
			}
			matched, _ := regexp.MatchString(groupPattern, proxy.GetGroup())
			return matched, realRule
		}
	}

	// Handle !!GROUPID= matcher
	// Note: GroupID is not currently tracked in parser.ProxyInterface, default to 0
	if strings.HasPrefix(rule, "!!GROUPID=") || strings.HasPrefix(rule, "!!INSERT=") {
		parts := strings.SplitN(rule, "!!", 3)
		if len(parts) >= 2 {
			var groupIDPattern string
			if strings.HasPrefix(rule, "!!GROUPID=") {
				groupIDPattern = strings.TrimPrefix(parts[1], "GROUPID=")
			} else {
				groupIDPattern = strings.TrimPrefix(parts[1], "INSERT=")
			}
			realRule := ""
			if len(parts) > 2 {
				realRule = parts[2]
			}
			// Parse range (e.g., "0", "0-5", "1,2,3")
			// Default GroupID to 0 since it's not tracked in current implementation
			matched := matchRange(groupIDPattern, 0)
			return matched, realRule
		}
	}

	// Handle !!TYPE= matcher
	if strings.HasPrefix(rule, "!!TYPE=") {
		parts := strings.SplitN(rule, "!!", 3)
		if len(parts) >= 2 {
			typePattern := strings.TrimPrefix(parts[1], "TYPE=")
			realRule := ""
			if len(parts) > 2 {
				realRule = parts[2]
			}
			proxyType := strings.ToUpper(proxy.GetType())
			// Use case-insensitive matching
			matched, _ := regexp.MatchString("(?i)^("+typePattern+")$", proxyType)
			return matched, realRule
		}
	}

	// Handle !!PORT= matcher
	if strings.HasPrefix(rule, "!!PORT=") {
		parts := strings.SplitN(rule, "!!", 3)
		if len(parts) >= 2 {
			portPattern := strings.TrimPrefix(parts[1], "PORT=")
			realRule := ""
			if len(parts) > 2 {
				realRule = parts[2]
			}
			matched := matchRange(portPattern, proxy.GetPort())
			return matched, realRule
		}
	}

	// Handle !!SERVER= matcher
	if strings.HasPrefix(rule, "!!SERVER=") {
		parts := strings.SplitN(rule, "!!", 3)
		if len(parts) >= 2 {
			serverPattern := strings.TrimPrefix(parts[1], "SERVER=")
			realRule := ""
			if len(parts) > 2 {
				realRule = parts[2]
			}
			matched, _ := regexp.MatchString(serverPattern, proxy.GetServer())
			return matched, realRule
		}
	}

	// No special matcher, return the rule as-is
	return true, rule
}

// matchRange checks if a value matches a range pattern (e.g., "1-5", "1,3,5", "10")
func matchRange(pattern string, value int) bool {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" {
		return true
	}

	// Handle comma-separated values: "1,2,3"
	if strings.Contains(pattern, ",") {
		parts := strings.Split(pattern, ",")
		for _, part := range parts {
			if matchRange(strings.TrimSpace(part), value) {
				return true
			}
		}
		return false
	}

	// Handle ranges: "1-5"
	if strings.Contains(pattern, "-") {
		parts := strings.Split(pattern, "-")
		if len(parts) == 2 {
			start, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
			end, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
			if err1 == nil && err2 == nil {
				return value >= start && value <= end
			}
		}
	}

	// Handle single value: "5"
	if num, err := strconv.Atoi(pattern); err == nil {
		return value == num
	}

	return false
}

// filterProxiesByRules filters proxies based on rule patterns
// Supports regex matching and special matchers
func filterProxiesByRules(proxies []proxy.ProxyInterface, rules []string) []string {
	var result []string
	seen := make(map[string]bool)

	for _, rule := range rules {
		// Handle [] prefix for direct inclusion
		if strings.HasPrefix(rule, "[]") {
			name := rule[2:]
			if !seen[name] {
				result = append(result, name)
				seen[name] = true
			}
			continue
		}

		// Filter proxies by rule
		for _, proxy := range proxies {
			matched, realRule := applyMatcher(rule, proxy)
			if !matched {
				continue
			}

			// If there's a real regex rule, check if proxy remark matches
			if realRule != "" {
				if match, _ := regexp.MatchString(realRule, proxy.GetRemark()); !match {
					continue
				}
			}

			// Add proxy if not already in result
			if !seen[proxy.GetRemark()] {
				result = append(result, proxy.GetRemark())
				seen[proxy.GetRemark()] = true
			}
		}
	}

	return result
}

// fetchRuleset fetches ruleset content from URL or local file
func fetchRuleset(path string) (string, error) {
	// Check if it's a URL
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		resp, err := http.Get(path)
		if err != nil {
			return "", fmt.Errorf("failed to fetch ruleset: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("failed to fetch ruleset: status %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("failed to read ruleset: %w", err)
		}
		return string(body), nil
	}

	// Try as local file
	basePath := config.Global.Common.BasePath
	if basePath == "" {
		basePath = "base"
	}

	// Try multiple possible paths
	paths := []string{
		path,
		filepath.Join(basePath, path),
		filepath.Join(basePath, "rules", path),
	}

	for _, p := range paths {
		if data, err := os.ReadFile(p); err == nil {
			return string(data), nil
		}
	}

	return "", fmt.Errorf("ruleset not found: %s", path)
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
			content, err := fetchRuleset(ruleset.Ruleset)
			if err != nil {
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
					}	else if len(parts) > 2 {
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

func generateSurge(proxies []proxy.ProxyInterface, opts GeneratorOptions, baseConfig string) (string, error) {
	var output strings.Builder

	// Add base configuration
	output.WriteString(baseConfig)
	output.WriteString("\n\n[Proxy]\n")

	// Generate proxy section
	for _, proxy := range proxies {
		line := convertToSurge(proxy, opts)
		if line != "" {
			output.WriteString(line)
			output.WriteString("\n")
		}
	}

	// Generate proxy groups
	if len(opts.ProxyGroups) > 0 {
		output.WriteString("\n[Proxy Group]\n")

		for _, group := range opts.ProxyGroups {
			line := convertGroupToSurge(group, proxies)
			if line != "" {
				output.WriteString(line)
				output.WriteString("\n")
			}
		}
	}

	// Generate rules
	if opts.EnableRuleGen && len(opts.Rulesets) > 0 {
		output.WriteString("\n[Rule]\n")
		for _, ruleset := range opts.Rulesets {
			// Surge rules format: RULE-SET,ruleset-name,policy
			output.WriteString(fmt.Sprintf("RULE-SET,%s,%s\n", ruleset.Ruleset, ruleset.Group))
		}
		output.WriteString("FINAL,DIRECT\n")
	}

	return output.String(), nil
}

func convertToSurge(proxy proxy.ProxyInterface, opts GeneratorOptions) string {
	var parts []string
	parts = append(parts, proxy.GetRemark())

	switch proxy.GetType() {
	case "ss":
		parts = append(parts, "ss")
		parts = append(parts, fmt.Sprintf("%s:%d", proxy.GetServer(), proxy.GetPort()))
		parts = append(parts, fmt.Sprintf("encrypt-method=%s", ""))
		parts = append(parts, fmt.Sprintf("password=%s", ""))
		if opts.UDP != nil && *opts.UDP {
			parts = append(parts, "udp-relay=true")
		}
		if opts.TFO != nil && *opts.TFO {
			parts = append(parts, "tfo=true")
		}

	case "vmess":
		parts = append(parts, "vmess")
		parts = append(parts, fmt.Sprintf("%s:%d", proxy.GetServer(), proxy.GetPort()))
		parts = append(parts, fmt.Sprintf("username=%s", ""))
		if "" == "ws" {
			parts = append(parts, "ws=true")
			if "" != "" {
				parts = append(parts, fmt.Sprintf("ws-path=%s", ""))
			}
			if "" != "" {
				parts = append(parts, fmt.Sprintf("ws-headers=Host:%s", ""))
			}
		}
		if false {
			parts = append(parts, "tls=true")
			if opts.SkipCertVerify != nil && *opts.SkipCertVerify {
				parts = append(parts, "skip-cert-verify=true")
			}
		}

	case "trojan":
		parts = append(parts, "trojan")
		parts = append(parts, fmt.Sprintf("%s:%d", proxy.GetServer(), proxy.GetPort()))
		parts = append(parts, fmt.Sprintf("password=%s", ""))
		if opts.SkipCertVerify != nil && *opts.SkipCertVerify {
			parts = append(parts, "skip-cert-verify=true")
		}
		if "" == "ws" {
			parts = append(parts, "ws=true")
			if "" != "" {
				parts = append(parts, fmt.Sprintf("ws-path=%s", ""))
			}
		}
	default:
		return ""
	}

	return strings.Join(parts, ", ")
}

func convertGroupToSurge(group config.ProxyGroupConfig, proxies []proxy.ProxyInterface) string {
	var parts []string
	parts = append(parts, group.Name)
	parts = append(parts, strings.ToLower(group.Type))

	// Filter proxies based on group rules using advanced filtering
	filtered := filterProxiesByRules(proxies, group.Rule)
	if len(filtered) == 0 {
		filtered = []string{"DIRECT"}
	}
	parts = append(parts, filtered...)

	// Add type-specific options
	switch strings.ToLower(group.Type) {
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

	return strings.Join(parts, ", ")
}

func generateQuantumultX(proxies []proxy.ProxyInterface, opts GeneratorOptions, baseConfig string) (string, error) {
	var output strings.Builder

	// Add base configuration
	output.WriteString(baseConfig)
	output.WriteString("\n\n[server_local]\n")

	// Generate proxy section
	for _, proxy := range proxies {
		line := convertToQuantumultX(proxy, opts)
		if line != "" {
			output.WriteString(line)
			output.WriteString("\n")
		}
	}

	// Generate policy section
	if len(opts.ProxyGroups) > 0 {
		output.WriteString("\n[policy]\n")

		for _, group := range opts.ProxyGroups {
			line := convertGroupToQuantumultX(group, proxies)
			if line != "" {
				output.WriteString(line)
				output.WriteString("\n")
			}
		}
	}

	// Generate filter rules
	if opts.EnableRuleGen && len(opts.Rulesets) > 0 {
		output.WriteString("\n[filter_local]\n")
		for _, ruleset := range opts.Rulesets {
			output.WriteString(fmt.Sprintf("# %s rules\n", ruleset.Ruleset))
		}
		output.WriteString(fmt.Sprintf("FINAL,%s\n", "DIRECT"))
	}

	return output.String(), nil
}

func convertToQuantumultX(proxy proxy.ProxyInterface, opts GeneratorOptions) string {
	var parts []string

	switch proxy.GetType() {
	case "ss", "shadowsocks":
		// Format: shadowsocks=server:port, method=encrypt-method, password=password, tag=name
		parts = append(parts, "shadowsocks="+fmt.Sprintf("%s:%d", proxy.GetServer(), proxy.GetPort()))
		parts = append(parts, fmt.Sprintf("method=%s", ""))
		parts = append(parts, fmt.Sprintf("password=%s", ""))
		if opts.UDP != nil && *opts.UDP {
			parts = append(parts, "udp-relay=true")
		}
		if opts.TFO != nil && *opts.TFO {
			parts = append(parts, "fast-open=true")
		}
		parts = append(parts, fmt.Sprintf("tag=%s", proxy.GetRemark()))

	case "vmess":
		// Format: vmess=server:port, method=encrypt-method, password=uuid, tag=name
		parts = append(parts, "vmess="+fmt.Sprintf("%s:%d", proxy.GetServer(), proxy.GetPort()))
		parts = append(parts, "method=aes-128-gcm")
		parts = append(parts, fmt.Sprintf("password=%s", ""))
		if "" == "ws" {
			parts = append(parts, "obfs=ws")
			if "" != "" {
				parts = append(parts, fmt.Sprintf("obfs-uri=%s", ""))
			}
			if "" != "" {
				parts = append(parts, fmt.Sprintf("obfs-host=%s", ""))
			}
		}
		if false {
			parts = append(parts, "obfs=over-tls")
		}
		parts = append(parts, fmt.Sprintf("tag=%s", proxy.GetRemark()))

	case "trojan":
		// Format: trojan=server:port, password=password, tag=name
		parts = append(parts, "trojan="+fmt.Sprintf("%s:%d", proxy.GetServer(), proxy.GetPort()))
		parts = append(parts, fmt.Sprintf("password=%s", ""))
		if opts.SkipCertVerify != nil && *opts.SkipCertVerify {
			parts = append(parts, "tls-verification=false")
		}
		parts = append(parts, fmt.Sprintf("tag=%s", proxy.GetRemark()))

	default:
		return ""
	}

	return strings.Join(parts, ", ")
}

func convertGroupToQuantumultX(group config.ProxyGroupConfig, proxies []proxy.ProxyInterface) string {
	groupType := strings.ToLower(group.Type)
	if groupType == "select" {
		groupType = "static"
	} else if groupType == "url-test" {
		groupType = "available"
	}

	var parts []string
	parts = append(parts, fmt.Sprintf("%s=%s", groupType, group.Name))

	// Filter proxies using advanced filtering
	filtered := filterProxiesByRules(proxies, group.Rule)
	if len(filtered) == 0 {
		filtered = []string{"direct"}
	}
	parts = append(parts, filtered...)

	// Add img-url if needed
	parts = append(parts, "img-url=https://raw.githubusercontent.com/Koolson/Qure/master/IconSet/Proxy.png")

	return strings.Join(parts, ", ")
}

func generateLoon(proxies []proxy.ProxyInterface, opts GeneratorOptions, baseConfig string) (string, error) {
	// Loon format is very similar to Surge
	var output strings.Builder

	output.WriteString(baseConfig)
	output.WriteString("\n\n[Proxy]\n")

	// Generate proxy section (same format as Surge)
	for _, proxy := range proxies {
		line := convertToSurge(proxy, opts)
		if line != "" {
			output.WriteString(line)
			output.WriteString("\n")
		}
	}

	// Generate proxy groups
	if len(opts.ProxyGroups) > 0 {
		output.WriteString("\n[Proxy Group]\n")

		for _, group := range opts.ProxyGroups {
			line := convertGroupToLoon(group, proxies)
			if line != "" {
				output.WriteString(line)
				output.WriteString("\n")
			}
		}
	}

	// Generate rules
	if opts.EnableRuleGen && len(opts.Rulesets) > 0 {
		output.WriteString("\n[Rule]\n")
		for _, ruleset := range opts.Rulesets {
			output.WriteString(fmt.Sprintf("RULE-SET,%s,%s\n", ruleset.Ruleset, ruleset.Group))
		}
		output.WriteString("FINAL,DIRECT\n")
	}

	return output.String(), nil
}

func convertGroupToLoon(group config.ProxyGroupConfig, proxies []proxy.ProxyInterface) string {
	var parts []string
	parts = append(parts, group.Name)
	parts = append(parts, "=")
	parts = append(parts, strings.ToLower(group.Type))

	// Filter proxies using advanced filtering
	filtered := filterProxiesByRules(proxies, group.Rule)
	if len(filtered) == 0 {
		filtered = []string{"direct"}
	}
	parts = append(parts, strings.Join(filtered, ","))

	// Add img-url
	parts = append(parts, "img-url=https://raw.githubusercontent.com/Koolson/Qure/master/IconSet/Proxy.png")

	return strings.Join(parts, ",")
}

func generateSingBox(proxies []proxy.ProxyInterface, opts GeneratorOptions, baseConfig string) (string, error) {
	// Parse base configuration as JSON
	var base map[string]interface{}
	if err := json.Unmarshal([]byte(baseConfig), &base); err != nil {
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
		}
	}

	base["outbounds"] = outbounds

	// Generate routing rules if enabled
	if opts.EnableRuleGen && len(opts.Rulesets) > 0 {
		rules := generateSingBoxRules(opts.Rulesets)
		if route, ok := base["route"].(map[string]interface{}); ok {
			route["rules"] = rules
		}
	}

	// Add clash_mode if enabled
	if opts.SingBoxAddClashMode {
		if experimental, ok := base["experimental"].(map[string]interface{}); ok {
			if clashAPI, ok := experimental["clash_api"].(map[string]interface{}); ok {
				clashAPI["default_mode"] = "rule"
			}
		}
	}

	// Marshal back to JSON with proper indentation
	output, err := json.MarshalIndent(base, "", "    ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal output: %w", err)
	}

	return string(output), nil
}

func convertToSingBox(proxy proxy.ProxyInterface, opts GeneratorOptions) map[string]interface{} {
	outbound := map[string]interface{}{
		"tag":         proxy.GetRemark(),
		"type":        proxy.GetType(),
		"server":      proxy.GetServer(),
		"server_port": proxy.GetPort(),
	}

	switch proxy.GetType() {
	case "ss", "shadowsocks":
		outbound["type"] = "shadowsocks"
		outbound["method"] = ""
		outbound["password"] = ""

	case "vmess":
		outbound["uuid"] = ""
		outbound["alter_id"] = 0
		outbound["security"] = "auto"
		if false {
			outbound["tls"] = map[string]interface{}{
				"enabled": true,
			}
		}
		if "" == "ws" {
			outbound["transport"] = map[string]interface{}{
				"type": "ws",
				"path": "",
				"headers": map[string]string{
					"Host": "",
				},
			}
		}

	case "trojan":
		outbound["password"] = ""
		outbound["tls"] = map[string]interface{}{
			"enabled": true,
		}
		if opts.SkipCertVerify != nil && *opts.SkipCertVerify {
			if tls, ok := outbound["tls"].(map[string]interface{}); ok {
				tls["insecure"] = true
			}
		}
	}

	return outbound
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

func generateSingle(proxies []proxy.ProxyInterface, format string) (string, error) {
	// Generate simple subscription (base64 encoded links)
	var lines []string
	for _, p := range proxies {
		// if SubconverterProxy
		if mixin, ok := p.(proxy.SubconverterProxy); ok {

			// Only include proxies matching the requested format
			if format == "v2ray" && (p.GetType() == "vmess" || p.GetType() == "vless") {
				link, err := mixin.GenerateLink()
				if err != nil {
					log.Panicln(`Failed to generate link for proxy:`, err)
				}
				lines = append(lines, link)
			} else if format == p.GetType() {
				link, err := mixin.GenerateLink()
				if err != nil {
					log.Panicln(`Failed to generate link for proxy:`, err)
				}
				lines = append(lines, link)
			}
		}
	}

	// Base64 encode the entire subscription
	subscription := strings.Join(lines, "\n")
	encoded := base64.StdEncoding.EncodeToString([]byte(subscription))
	return encoded, nil
}
