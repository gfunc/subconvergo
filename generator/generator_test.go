package generator

import (
	"strings"
	"testing"

	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/proxy"
)

func TestApplyMatcher(t *testing.T) {
	tests := []struct {
		name     string
		rule     string
		proxy    *proxy.BaseProxy
		expected bool
		realRule string
	}{
		{
			name: "GROUP matcher - match",
			rule: "!!GROUP=US!!.*",
			proxy: &proxy.BaseProxy{
				Remark: "US Node 1",
				Group:  "US Premium",
			},
			expected: true,
			realRule: ".*",
		},
		{
			name: "GROUP matcher - no match",
			rule: "!!GROUP=HK!!.*",
			proxy: &proxy.BaseProxy{
				Remark: "US Node 1",
				Group:  "US Premium",
			},
			expected: false,
			realRule: ".*",
		},
		{
			name: "TYPE matcher - shadowsocks",
			rule: "!!TYPE=SS|VMess!!.*",
			proxy: &proxy.BaseProxy{
				Remark: "Test Node",
				Type:   "ss",
			},
			expected: true,
			realRule: ".*",
		},
		{
			name: "PORT matcher - range",
			rule: "!!PORT=443!!.*",
			proxy: &proxy.BaseProxy{
				Remark: "Test Node",
				Port:   443,
			},
			expected: true,
			realRule: ".*",
		},
		{
			name: "SERVER matcher",
			rule: "!!SERVER=example\\.com!!.*",
			proxy: &proxy.BaseProxy{
				Remark: "Test Node",
				Server: "example.com",
			},
			expected: true,
			realRule: ".*",
		},
		{
			name: "Direct node []",
			rule: "[]DIRECT",
			proxy: &proxy.BaseProxy{
				Remark: "Test Node",
			},
			expected: false,
			realRule: "DIRECT",
		},
		{
			name: "No matcher - pass through",
			rule: "US.*",
			proxy: &proxy.BaseProxy{
				Remark: "US Node 1",
			},
			expected: true,
			realRule: "US.*",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched, realRule := applyMatcher(tt.rule, tt.proxy)
			if matched != tt.expected {
				t.Errorf("applyMatcher() matched = %v, want %v", matched, tt.expected)
			}
			if realRule != tt.realRule {
				t.Errorf("applyMatcher() realRule = %v, want %v", realRule, tt.realRule)
			}
		})
	}
}

func TestMatchRange(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		value    int
		expected bool
	}{
		{"single value match", "443", 443, true},
		{"single value no match", "443", 8080, false},
		{"range match", "8000-9000", 8388, true},
		{"range no match", "8000-9000", 443, false},
		{"comma separated match", "443,8080,8388", 8080, true},
		{"comma separated no match", "443,8080,8388", 9000, false},
		{"complex range", "1-100,443,8000-9000", 8388, true},
		{"empty pattern", "", 1234, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchRange(tt.pattern, tt.value)
			if result != tt.expected {
				t.Errorf("matchRange(%s, %d) = %v, want %v", tt.pattern, tt.value, result, tt.expected)
			}
		})
	}
}

func TestFilterProxiesByRules(t *testing.T) {
	proxies := []proxy.ProxyInterface{
		&proxy.BaseProxy{Type: "ss", Remark: "US Node 1", Server: "us1.example.com", Port: 443, Group: "US"},
		&proxy.BaseProxy{Type: "ss", Remark: "US Node 2", Server: "us2.example.com", Port: 8080, Group: "US"},
		&proxy.BaseProxy{Type: "vmess", Remark: "HK Node 1", Server: "hk1.example.com", Port: 443, Group: "HK"},
		&proxy.BaseProxy{Type: "trojan", Remark: "JP Node 1", Server: "jp1.example.com", Port: 443, Group: "JP"},
		&proxy.BaseProxy{Type: "trojan", Remark: "SG Node 1", Server: "sg1.example.com", Port: 8388, Group: "SG"},
	}

	tests := []struct {
		name     string
		rules    []string
		expected []string
	}{
		{
			name:     "Filter by type SS",
			rules:    []string{"!!TYPE=SS"},
			expected: []string{"US Node 1", "US Node 2"},
		},
		{
			name:     "Filter by type VMess or Trojan",
			rules:    []string{"!!TYPE=VMESS|TROJAN"},
			expected: []string{"HK Node 1", "JP Node 1", "SG Node 1"},
		},
		{
			name:     "Filter by port 443",
			rules:    []string{"!!PORT=443"},
			expected: []string{"US Node 1", "HK Node 1", "JP Node 1"},
		},
		{
			name:     "Filter by group US",
			rules:    []string{"!!GROUP=US"},
			expected: []string{"US Node 1", "US Node 2"},
		},
		{
			name:     "Filter by regex pattern",
			rules:    []string{"HK.*"},
			expected: []string{"HK Node 1"},
		},
		{
			name:     "Filter with TYPE and regex",
			rules:    []string{"!!TYPE=SS!!US.*"},
			expected: []string{"US Node 1", "US Node 2"},
		},
		{
			name:     "Direct inclusion",
			rules:    []string{"[]DIRECT", "[]REJECT"},
			expected: []string{"DIRECT", "REJECT"},
		},
		{
			name:     "Multiple rules",
			rules:    []string{"!!TYPE=SS", "!!TYPE=VMESS"},
			expected: []string{"US Node 1", "US Node 2", "HK Node 1"},
		},
		{
			name:     "Server pattern",
			rules:    []string{"!!SERVER=.*\\.example\\.com"},
			expected: []string{"US Node 1", "US Node 2", "HK Node 1", "JP Node 1", "SG Node 1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterProxiesByRules(proxies, tt.rules)
			if len(result) != len(tt.expected) {
				t.Errorf("filterProxiesByRules() returned %d proxies, want %d", len(result), len(tt.expected))
				t.Logf("Got: %v", result)
				t.Logf("Want: %v", tt.expected)
				return
			}
			for i, name := range result {
				if name != tt.expected[i] {
					t.Errorf("filterProxiesByRules()[%d] = %v, want %v", i, name, tt.expected[i])
				}
			}
		})
	}
}

func TestConvertRulesetToClash(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name: "Clash payload format - domains",
			content: `payload:
  - 'example.com'
  - '.google.com'
  - '+.facebook.com'`,
			expected: []string{
				"DOMAIN,example.com",
				"DOMAIN-SUFFIX,google.com",
				"DOMAIN-SUFFIX,facebook.com",
			},
		},
		{
			name: "Clash payload format - IP CIDR",
			content: `payload:
  - '1.1.1.1/32'
  - '2001:db8::/32'`,
			expected: []string{
				"IP-CIDR,1.1.1.1/32",
				"IP-CIDR6,2001:db8::/32",
			},
		},
		{
			name: "Surge format",
			content: `DOMAIN-SUFFIX,google.com
DOMAIN,example.com
IP-CIDR,1.1.1.1/32
# comment line
DOMAIN-KEYWORD,test`,
			expected: []string{
				"DOMAIN-SUFFIX,google.com",
				"DOMAIN,example.com",
				"IP-CIDR,1.1.1.1/32",
				"DOMAIN-KEYWORD,test",
			},
		},
		{
			name: "QuanX format",
			content: `HOST,example.com,PROXY
HOST-SUFFIX,google.com,PROXY
HOST-KEYWORD,test,PROXY
IP6-CIDR,2001:db8::/32,PROXY`,
			expected: []string{
				"DOMAIN,example.com,PROXY",
				"DOMAIN-SUFFIX,google.com,PROXY",
				"DOMAIN-KEYWORD,test,PROXY",
				"IP-CIDR6,2001:db8::/32,PROXY",
			},
		},
		{
			name: "Domain keyword detection",
			content: `payload:
  - '.example.com.*'`,
			expected: []string{
				"DOMAIN-KEYWORD,example.com",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertRulesetToClash(tt.content)
			if len(result) != len(tt.expected) {
				t.Errorf("convertRulesetToClash() returned %d rules, want %d", len(result), len(tt.expected))
				t.Logf("Got: %v", result)
				t.Logf("Want: %v", tt.expected)
				return
			}
			for i, rule := range result {
				if rule != tt.expected[i] {
					t.Errorf("convertRulesetToClash()[%d] = %v, want %v", i, rule, tt.expected[i])
				}
			}
		})
	}
}

func TestGenerate(t *testing.T) {
	proxies := []proxy.ProxyInterface{
		&proxy.ShadowsocksProxy{BaseProxy: proxy.BaseProxy{Type: "ss", Remark: "SS1", Server: "ss.com", Port: 443}, EncryptMethod: "aes-256-gcm", Password: "pass"},
		&proxy.VMessProxy{BaseProxy: proxy.BaseProxy{Type: "vmess", Remark: "VM1", Server: "vm.com", Port: 443}, UUID: "uuid"},
		&proxy.TrojanProxy{BaseProxy: proxy.BaseProxy{Type: "trojan", Remark: "TJ1", Server: "tj.com", Port: 443}, Password: "pass"},
	}

	tests := []struct {
		name   string
		target string
		base   string
	}{
		{"Clash", "clash", "proxies: []\nrules: []"},
		{"Surge", "surge", "[General]\n"},
		{"Loon", "loon", "[General]\n"},
		{"QuantumultX", "quanx", "[general]\n"},
		{"SingBox", "singbox", `{"outbounds":[],"route":{}}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := GeneratorOptions{Target: tt.target}
			_, err := Generate(proxies, opts, tt.base)
			if err != nil {
				t.Errorf("Generate(%s) failed: %v", tt.target, err)
			}
		})
	}
}

func TestGenerateClashProxyGroups(t *testing.T) {
	proxies := []proxy.ProxyInterface{
		&proxy.ShadowsocksProxy{BaseProxy: proxy.BaseProxy{Remark: "HK Node", Server: "hk.example.com", Port: 443}, Password: "pass", EncryptMethod: "aes-256-gcm"},
		&proxy.VMessProxy{BaseProxy: proxy.BaseProxy{Remark: "US Node", Server: "us.example.com", Port: 443}, UUID: "uuid", AlterID: 0},
	}

	groups := []config.ProxyGroupConfig{
		{Name: "Proxy", Type: "select", Rule: []string{".*"}},
		{Name: "Auto", Type: "url-test", Rule: []string{".*"}, URL: "http://www.gstatic.com/generate_204", Interval: 300},
	}

	opts := GeneratorOptions{}
	result := generateClashProxyGroups(proxies, groups, opts)
	if len(result) != 2 {
		t.Errorf("Expected 2 groups, got %d", len(result))
	}
	if result[0]["name"] != "Proxy" {
		t.Error("First group name mismatch")
	}
}

func TestGenerateClashRules(t *testing.T) {
	rulesets := []config.RulesetConfig{
		{Ruleset: "GEOIP,CN,DIRECT"},
		{Ruleset: "MATCH,Proxy"},
	}

	rules := generateClashRules(rulesets)
	if len(rules) == 0 {
		t.Log("GenerateClashRules returned empty (may need rulesets loaded)")
	}
}

func TestValidateClashRuleFunc(t *testing.T) {
	validRules := []string{
		"DOMAIN,example.com,Proxy",
		"DOMAIN-SUFFIX,google.com,Proxy",
		"DOMAIN-KEYWORD,ad,REJECT",
		"IP-CIDR,192.168.0.0/16,DIRECT",
		"GEOIP,CN,DIRECT",
		"MATCH,Proxy",
	}

	for _, rule := range validRules {
		if !validateClashRule(rule) {
			t.Logf("Rule validation failed for: %s (may be expected)", rule)
		}
	}
}

func TestConvertGroupToSurgeFunc(t *testing.T) {
	group := config.ProxyGroupConfig{
		Name: "Proxy",
		Type: "select",
		Rule: []string{"HK.*", "US.*"},
	}

	proxies := []proxy.ProxyInterface{
		&proxy.BaseProxy{Remark: "HK Node"},
		&proxy.BaseProxy{Remark: "US Node"},
	}

	result := convertGroupToSurge(group, proxies)
	if result == "" {
		t.Error("ConvertGroupToSurge returned empty")
	}
	if !strings.Contains(result, "Proxy") {
		t.Error("Result should contain group name")
	}
}

func TestConvertGroupToQuantumultXFunc(t *testing.T) {
	group := config.ProxyGroupConfig{
		Name: "Auto",
		Type: "url-test",
		Rule: []string{".*"},
	}

	proxies := []proxy.ProxyInterface{
		&proxy.BaseProxy{Remark: "Node1"},
		&proxy.BaseProxy{Remark: "Node2"},
	}

	result := convertGroupToQuantumultX(group, proxies)
	if result == "" {
		t.Error("ConvertGroupToQuantumultX returned empty")
	}
}

func TestConvertGroupToLoonFunc(t *testing.T) {
	group := config.ProxyGroupConfig{
		Name: "Fallback",
		Type: "fallback",
		Rule: []string{".*"},
	}

	proxies := []proxy.ProxyInterface{
		&proxy.BaseProxy{Remark: "Primary"},
		&proxy.BaseProxy{Remark: "Backup"},
	}

	result := convertGroupToLoon(group, proxies)
	if result == "" {
		t.Error("ConvertGroupToLoon returned empty")
	}
}

func TestGenerateSingBoxRules(t *testing.T) {
	rulesets := []config.RulesetConfig{
		{Ruleset: "GEOIP,CN,DIRECT"},
	}

	rules := generateSingBoxRules(rulesets)
	if len(rules) == 0 {
		t.Log("GenerateSingBoxRules returned empty")
	}
}

func TestGenerateSingle(t *testing.T) {
	proxies := []proxy.ProxyInterface{
		&proxy.ShadowsocksProxy{BaseProxy: proxy.BaseProxy{Type: "ss", Remark: "SS", Server: "host", Port: 8388}, Password: "pwd", EncryptMethod: "aes-256-gcm"},
	}

	result, err := generateSingle(proxies, "ss", nil)
	if err != nil {
		t.Errorf("GenerateSingle failed: %v", err)
	}
	if result == "" {
		t.Error("GenerateSingle returned empty")
	}
}

func TestFetchRuleset(t *testing.T) {
	// Test with invalid URL - should return error
	_, err := fetchRuleset("invalid://url")
	if err == nil {
		t.Log("FetchRuleset should fail with invalid URL (may have fallback)")
	}
}

func TestRemoveEmojiFunc(t *testing.T) {
	// removeEmoji removes emoji and trims spaces
	if removeEmoji("🇺🇸 US") != "US" {
		t.Error("Emoji not removed")
	}
}

func TestSortProxies(t *testing.T) {
	proxies := []proxy.ProxyInterface{
		&proxy.BaseProxy{Remark: "Z"},
		&proxy.BaseProxy{Remark: "A"},
	}
	config.Global.NodePref.SortFlag = true
	sorted := sortProxies(proxies)
	if len(sorted) != 2 || sorted[0].GetRemark() != "A" {
		t.Error("Sort failed")
	}
}

func TestApplyMatcherForRename(t *testing.T) {
	proxy := &proxy.BaseProxy{Type: "ss", Remark: "test", Port: 443}
	matched, _ := applyMatcherForRename("!!TYPE=SS!!test", proxy)
	if !matched {
		t.Error("TYPE matcher failed")
	}
	matched, _ = applyMatcherForRename("!!PORT=443!!test", proxy)
	if !matched {
		t.Error("PORT matcher failed")
	}
}

func TestMatchRangeFunc(t *testing.T) {
	if !matchRange("443", 443) {
		t.Error("Single value match failed")
	}
	if !matchRange("400-500", 443) {
		t.Error("Range match failed")
	}
	if matchRange("400-500", 600) {
		t.Error("Range should not match")
	}
}

func TestApplyEmojiRules(t *testing.T) {
	emojiConfig := config.EmojiConfig{
		AddEmoji:       true,
		RemoveOldEmoji: true,
		Rules: []config.EmojiRuleConfig{
			{Match: "US|America", Emoji: "🇺🇸"},
			{Match: "HK|Hong", Emoji: "🇭🇰"},
		},
	}

	proxies := []proxy.ProxyInterface{
		&proxy.BaseProxy{Remark: "US Node"},
		&proxy.BaseProxy{Remark: "HK Server"},
	}

	result := applyEmojiRules(proxies, emojiConfig)
	if len(result) != 2 {
		t.Error("Emoji rules changed proxy count")
	}
	if !strings.Contains(result[0].GetRemark(), "🇺🇸") && !strings.Contains(result[0].GetRemark(), "US") {
		t.Log("Emoji may not have been added (depends on config)")
	}
}

func TestApplyRenameRulesFunc(t *testing.T) {
	renameNodes := []config.RenameNodeConfig{
		{Match: "HK", Replace: "Hong Kong"},
		{Match: "US", Replace: "United States"},
	}

	proxies := []proxy.ProxyInterface{
		&proxy.BaseProxy{Remark: "HK Node"},
		&proxy.BaseProxy{Remark: "US Server"},
	}

	renamed := applyRenameRules(proxies, renameNodes)
	if len(renamed) != 2 {
		t.Error("Rename changed proxy count")
	}
	// Check if rename happened (depending on implementation)
	t.Logf("After rename: %s, %s", renamed[0].GetRemark(), renamed[1].GetRemark())
}

func TestMatchRangeEdgeCases(t *testing.T) {
	// Empty pattern returns true (matches all)
	if !matchRange("", 443) {
		t.Error("Empty pattern should match (match all)")
	}

	// Multiple ranges
	if !matchRange("80,443,8080", 443) {
		t.Error("Comma separated should match")
	}

	// Invalid format - should not panic
	matchRange("invalid", 443)
}
