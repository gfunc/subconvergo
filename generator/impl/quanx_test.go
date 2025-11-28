package impl

import (
	"testing"

	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/generator/core"
	"github.com/stretchr/testify/assert"
)

func TestQuantumultXGenerator_Generate(t *testing.T) {
	gen := &QuantumultXGenerator{}
	proxies := getTestProxies()

	opts := core.GeneratorOptions{
		Base:         "[general]",
		ProxySetting: config.ProxySetting{},
	}

	output, err := gen.Generate(proxies, nil, nil, nil, opts)
	assert.NoError(t, err)

	assert.Contains(t, output, "[server_local]")
	assert.Contains(t, output, "shadowsocks=1.2.3.4:8388, method=aes-256-gcm, password=password, tag=ss-proxy")
	// TODO: Add SSR support for Quantumult X
	// assert.Contains(t, output, "shadowsocks=1.2.3.4:8388, method=aes-256-gcm, password=password, obfs=http, obfs-host=example.com, obfs-uri=/path, tag=ssr-proxy")
	assert.Contains(t, output, "vmess=5.6.7.8:443, method=aes-128-gcm, password=uuid, obfs=ws, obfs-uri=/path, obfs-host=example.com, tag=vmess-proxy")
	assert.Contains(t, output, "trojan=13.14.15.16:443, password=password, tls-host=example.com, tag=trojan-proxy")
}

func TestQuantumultXGenerator_Generate_WithGroups(t *testing.T) {
	gen := &QuantumultXGenerator{}
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
			Rule: []string{".*"},
		},
	}

	opts := core.GeneratorOptions{
		Base:         "[general]",
		ProxySetting: config.ProxySetting{},
	}

	output, err := gen.Generate(proxies, groups, nil, nil, opts)
	assert.NoError(t, err)

	assert.Contains(t, output, "[policy]")
	assert.Contains(t, output, "static=Select Group, ss-proxy, vmess-proxy, img-url=https://raw.githubusercontent.com/Koolson/Qure/master/IconSet/Proxy.png")
	// QuanX generator skips unsupported proxies (vless, hysteria2, tuic, anytls)
	assert.Contains(t, output, "available=URL Test Group, ss-proxy, ssr-proxy, vmess-proxy, trojan-proxy, img-url=https://raw.githubusercontent.com/Koolson/Qure/master/IconSet/Proxy.png")
}
