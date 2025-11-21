package impl

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/generator/core"
	"github.com/stretchr/testify/assert"
)

func TestSingleGenerator_Generate(t *testing.T) {
	gen := &SingleGenerator{Target: "ss"}
	proxies := getTestProxies()

	opts := core.GeneratorOptions{
		ProxySetting: config.ProxySetting{},
	}

	output, err := gen.Generate(proxies, nil, nil, nil, opts)
	assert.NoError(t, err)

	decodedBytes, err := base64.StdEncoding.DecodeString(output)
	assert.NoError(t, err)
	decoded := string(decodedBytes)

	// Should only contain SS link
	assert.Contains(t, decoded, "ss://")
	assert.NotContains(t, decoded, "vmess://")

	assert.Contains(t, decoded, "#ss-proxy")
}

func TestSingleGenerator_Generate_V2Ray(t *testing.T) {
	gen := &SingleGenerator{Target: "v2ray"}
	proxies := getTestProxies()

	opts := core.GeneratorOptions{
		ProxySetting: config.ProxySetting{},
	}

	output, err := gen.Generate(proxies, nil, nil, nil, opts)
	assert.NoError(t, err)

	decodedBytes, err := base64.StdEncoding.DecodeString(output)
	assert.NoError(t, err)
	decoded := string(decodedBytes)

	// Should contain VMess and VLESS links
	assert.Contains(t, decoded, "vmess://")
	assert.Contains(t, decoded, "vless://")

	lines := strings.Split(decoded, "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		assert.False(t, strings.HasPrefix(line, "ss://"), "Line should not start with ss://: %s", line)
	}
}

func TestSingleGenerator_Generate_Trojan(t *testing.T) {
	gen := &SingleGenerator{Target: "trojan"}
	proxies := getTestProxies()

	opts := core.GeneratorOptions{
		ProxySetting: config.ProxySetting{},
	}

	output, err := gen.Generate(proxies, nil, nil, nil, opts)
	assert.NoError(t, err)

	decodedBytes, err := base64.StdEncoding.DecodeString(output)
	assert.NoError(t, err)
	decoded := string(decodedBytes)

	assert.Contains(t, decoded, "trojan://")
	assert.NotContains(t, decoded, "ss://")
}

func TestSingleGenerator_Generate_Mixed(t *testing.T) {
	gen := &SingleGenerator{Target: "mixed"}
	proxies := getTestProxies()

	opts := core.GeneratorOptions{
		ProxySetting: config.ProxySetting{},
	}

	output, err := gen.Generate(proxies, nil, nil, nil, opts)
	assert.NoError(t, err)

	decodedBytes, err := base64.StdEncoding.DecodeString(output)
	assert.NoError(t, err)
	decoded := string(decodedBytes)

	assert.Contains(t, decoded, "ss://")
	assert.Contains(t, decoded, "ssr://")
	assert.Contains(t, decoded, "vmess://")
	assert.Contains(t, decoded, "vless://")
	assert.Contains(t, decoded, "trojan://")
	assert.Contains(t, decoded, "hysteria2://")
	assert.Contains(t, decoded, "tuic://")
}
