package impl

import (
	"encoding/json"
	"testing"

	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/generator/core"
	"github.com/stretchr/testify/assert"
)

func TestSingBoxGenerator_Generate(t *testing.T) {
	gen := &SingBoxGenerator{}
	proxies := getTestProxies()

	opts := core.GeneratorOptions{
		Base:         "{}",
		ProxySetting: config.ProxySetting{},
	}

	output, err := gen.Generate(proxies, nil, nil, nil, opts)
	assert.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal([]byte(output), &result)
	assert.NoError(t, err)

	outbounds := result["outbounds"].([]interface{})
	// DIRECT, REJECT, + 7 proxies
	assert.GreaterOrEqual(t, len(outbounds), 9)

	checkProxy := func(tag, pType string) {
		var found bool
		for _, o := range outbounds {
			outbound := o.(map[string]interface{})
			if outbound["tag"] == tag {
				found = true
				assert.Equal(t, pType, outbound["type"])
			}
		}
		assert.True(t, found, "Proxy %s not found", tag)
	}

	checkProxy("ss-proxy", "shadowsocks")
	checkProxy("ssr-proxy", "ssr") // SingBox might not support SSR natively or uses different type
	checkProxy("vmess-proxy", "vmess")
	checkProxy("vless-proxy", "vless")
	checkProxy("trojan-proxy", "trojan")
	checkProxy("hysteria2-proxy", "hysteria2")
	checkProxy("tuic-proxy", "tuic")
}

func TestSingBoxGenerator_Generate_WithClashMode(t *testing.T) {
	gen := &SingBoxGenerator{}
	proxies := getTestProxies()

	opts := core.GeneratorOptions{
		Base: "{}",
		ProxySetting: config.ProxySetting{
			SingBoxAddClashMode: true,
		},
	}

	output, err := gen.Generate(proxies, nil, nil, nil, opts)
	assert.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal([]byte(output), &result)
	assert.NoError(t, err)

	experimental, ok := result["experimental"].(map[string]interface{})
	assert.True(t, ok)
	clashAPI, ok := experimental["clash_api"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "rule", clashAPI["default_mode"])
}

func TestSingBoxGenerator_Generate_WithRulesAndClashMode(t *testing.T) {
	gen := &SingBoxGenerator{}
	proxies := getTestProxies()

	opts := core.GeneratorOptions{
		Base: `{"route": {}}`,
		ProxySetting: config.ProxySetting{
			SingBoxAddClashMode: true,
		},
		Rule: true,
		Rulesets: []config.RulesetConfig{
			{
				Ruleset: "geosite:google",
				Group:   "Select Group",
			},
		},
	}

	output, err := gen.Generate(proxies, nil, nil, nil, opts)
	assert.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal([]byte(output), &result)
	assert.NoError(t, err)

	// Verify rules
	route := result["route"].(map[string]interface{})
	rules := route["rules"].([]interface{})
	assert.Len(t, rules, 2) // 1 custom + 1 final

	rule1 := rules[0].(map[string]interface{})
	assert.Equal(t, "geosite:google", rule1["rule_set"].([]interface{})[0])
	assert.Equal(t, "Select Group", rule1["outbound"])

	// Verify clash mode
	experimental := result["experimental"].(map[string]interface{})
	clashApi := experimental["clash_api"].(map[string]interface{})
	assert.Equal(t, "rule", clashApi["default_mode"])
}
