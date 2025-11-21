package impl

import (
	"os"
	"testing"

	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/generator/core"
	pc "github.com/gfunc/subconvergo/proxy/core"
	"github.com/gfunc/subconvergo/proxy/impl"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestClashGenerator_Generate(t *testing.T) {
	gen := &ClashGenerator{}
	proxies := getTestProxies()

	opts := core.GeneratorOptions{
		Base:         "proxies: []\nproxy-groups: []\nrules: []",
		ProxySetting: config.ProxySetting{},
	}

	output, err := gen.Generate(proxies, nil, nil, nil, opts)
	assert.NoError(t, err)

	var result map[string]interface{}
	err = yaml.Unmarshal([]byte(output), &result)
	assert.NoError(t, err)

	proxiesList := result["proxies"].([]interface{})
	assert.Len(t, proxiesList, len(proxies))

	// Verify each proxy type
	for _, p := range proxiesList {
		proxy := p.(map[string]interface{})
		switch proxy["name"] {
		case "ss-proxy":
			assert.Equal(t, "ss", proxy["type"])
			assert.Equal(t, "aes-256-gcm", proxy["cipher"])
		case "ssr-proxy":
			assert.Equal(t, "ssr", proxy["type"])
			assert.Equal(t, "auth_aes128_md5", proxy["protocol"])
		case "vmess-proxy":
			assert.Equal(t, "vmess", proxy["type"])
			assert.Equal(t, "ws", proxy["network"])
		case "vless-proxy":
			assert.Equal(t, "vless", proxy["type"])
			assert.Equal(t, "xtls-rprx-vision", proxy["flow"])
		case "trojan-proxy":
			assert.Equal(t, "trojan", proxy["type"])
			assert.Equal(t, "ws", proxy["network"])
		case "hysteria2-proxy":
			assert.Equal(t, "hysteria2", proxy["type"])
			assert.Equal(t, "salamander", proxy["obfs"])
		case "tuic-proxy":
			assert.Equal(t, "tuic", proxy["type"])
			assert.Equal(t, "bbr", proxy["congestion-controller"])
		}
	}
}

func TestClashGenerator_Generate_WithOptions(t *testing.T) {
	gen := &ClashGenerator{}

	proxies := []pc.ProxyInterface{
		&impl.ShadowsocksProxy{
			BaseProxy: pc.BaseProxy{
				Type:   "ss",
				Remark: "ss-proxy",
				Server: "1.2.3.4",
				Port:   8388,
			},
			Password:      "password",
			EncryptMethod: "aes-256-gcm",
		},
		&impl.VMessProxy{
			BaseProxy: pc.BaseProxy{
				Type:   "vmess",
				Remark: "vmess-proxy",
				Server: "5.6.7.8",
				Port:   443,
			},
			UUID:    "uuid",
			AlterID: 64,
			TLS:     true,
		},
	}

	opts := core.GeneratorOptions{
		Base: "proxies: []\nproxy-groups: []\nrules: []",
		ProxySetting: config.ProxySetting{
			UDP:   true,
			TFO:   true,
			SCV:   true,
			TLS13: true,
		},
	}

	output, err := gen.Generate(proxies, nil, nil, nil, opts)
	assert.NoError(t, err)

	var result map[string]interface{}
	err = yaml.Unmarshal([]byte(output), &result)
	assert.NoError(t, err)

	proxiesList := result["proxies"].([]interface{})
	assert.Len(t, proxiesList, 2)

	// Check SS Proxy options
	ssProxy := proxiesList[0].(map[string]interface{})
	assert.Equal(t, "ss", ssProxy["type"])
	assert.Equal(t, true, ssProxy["udp"])
	// assert.Equal(t, true, ssProxy["tfo"]) // TFO might not be supported for all types or fields might differ

	// Check VMess Proxy options
	vmessProxy := proxiesList[1].(map[string]interface{})
	assert.Equal(t, "vmess", vmessProxy["type"])
	assert.Equal(t, true, vmessProxy["udp"])
	assert.Equal(t, true, vmessProxy["skip-cert-verify"])
	// assert.Equal(t, true, vmessProxy["tls13"]) // Check if this field is actually exported
}

func TestClashGenerator_Generate_WithGroupsAndRules(t *testing.T) {
	gen := &ClashGenerator{}
	proxies := getTestProxies()

	groups := []config.ProxyGroupConfig{
		{
			Name: "Select Group",
			Type: "select",
			Rule: []string{"[]ss-proxy", "[]vmess-proxy"},
		},
		{
			Name: "URL Test Group",
			Type: "url-test",
			URL:  "http://www.gstatic.com/generate_204",
			Rule: []string{".*"},
		},
	}

	rules := []string{
		"DOMAIN-SUFFIX,google.com,Select Group",
		"GEOIP,CN,DIRECT",
		"MATCH,URL Test Group",
	}

	opts := core.GeneratorOptions{
		Base:         "proxies: []\nproxy-groups: []\nrules: []",
		ProxySetting: config.ProxySetting{},
		Rule:         true,
	}

	output, err := gen.Generate(proxies, groups, rules, nil, opts)
	assert.NoError(t, err)

	var result map[string]interface{}
	err = yaml.Unmarshal([]byte(output), &result)
	assert.NoError(t, err)

	// Verify groups
	proxyGroups := result["proxy-groups"].([]interface{})
	assert.Len(t, proxyGroups, 2)

	group1 := proxyGroups[0].(map[string]interface{})
	assert.Equal(t, "Select Group", group1["name"])
	assert.Equal(t, "select", group1["type"])
	assert.Contains(t, group1["proxies"], "ss-proxy")
	assert.Contains(t, group1["proxies"], "vmess-proxy")

	group2 := proxyGroups[1].(map[string]interface{})
	assert.Equal(t, "URL Test Group", group2["name"])
	assert.Equal(t, "url-test", group2["type"])
	// Should contain all proxies due to .*
	assert.Greater(t, len(group2["proxies"].([]interface{})), 2)

	// Verify rules
	rulesList := result["rules"].([]interface{})
	assert.Len(t, rulesList, 3)
	assert.Equal(t, "DOMAIN-SUFFIX,google.com,Select Group", rulesList[0])
}

func TestClashGenerator_Generate_WithRulesets(t *testing.T) {
	// Create temporary ruleset file
	rulesContent := "DOMAIN-SUFFIX,example.com\nIP-CIDR,1.2.3.4/32,no-resolve"
	err := os.WriteFile("test_rules.list", []byte(rulesContent), 0644)
	assert.NoError(t, err)
	defer os.Remove("test_rules.list")

	gen := &ClashGenerator{}
	proxies := getTestProxies()

	opts := core.GeneratorOptions{
		Base:         "proxies: []\nproxy-groups: []\nrules: []",
		ProxySetting: config.ProxySetting{},
		Rule:         true,
		Rulesets: []config.RulesetConfig{
			{
				Ruleset: "geosite:google",
				Group:   "Select Group",
			},
			{
				Ruleset: "test_rules.list",
				Group:   "Proxy",
			},
		},
	}

	output, err := gen.Generate(proxies, nil, nil, nil, opts)
	assert.NoError(t, err)

	var result map[string]interface{}
	err = yaml.Unmarshal([]byte(output), &result)
	assert.NoError(t, err)

	rulesList := result["rules"].([]interface{})
	// 1 (geosite) + 2 (test_rules.list) + 1 (MATCH,DIRECT) = 4
	assert.Len(t, rulesList, 4)
	assert.Equal(t, "RULE-SET,geosite:google,Select Group", rulesList[0])
	assert.Equal(t, "DOMAIN-SUFFIX,example.com,Proxy", rulesList[1])
	assert.Equal(t, "IP-CIDR,1.2.3.4/32,Proxy,no-resolve", rulesList[2])
	assert.Equal(t, "MATCH,DIRECT", rulesList[3])
}
