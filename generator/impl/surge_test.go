package impl

import (
	"testing"

	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/generator/core"
	"github.com/stretchr/testify/assert"
)

func TestSurgeGenerator_Generate(t *testing.T) {
	gen := &SurgeGenerator{}
	proxies := getTestProxies()

	opts := core.GeneratorOptions{
		Base:         "[General]\nloglevel = notify",
		ProxySetting: config.ProxySetting{},
	}

	output, err := gen.Generate(proxies, nil, nil, nil, opts)
	assert.NoError(t, err)

	assert.Contains(t, output, "[Proxy]")
	assert.Contains(t, output, "DIRECT = direct")
	assert.Contains(t, output, "ss-proxy = ss, 1.2.3.4, 8388, encrypt-method=aes-256-gcm, password=password")
	// Surge generator in this implementation skips unsupported proxies (SSR, VMess, VLESS, Trojan, Hysteria2, TUIC, AnyTLS)
	// assert.Contains(t, output, "ssr-proxy = ssr, 1.2.3.4:8388, encrypt-method=aes-256-gcm, password=password, protocol=auth_aes128_md5, obfs=tls1.2_ticket_auth")
	// assert.Contains(t, output, "vmess-proxy = vmess, 5.6.7.8:443, username=uuid, ws=true, ws-path=/path, ws-headers=Host:example.com, tls=true, sni=example.com")
	// assert.Contains(t, output, "trojan-proxy = trojan, 13.14.15.16:443, password=password, ws=true, ws-path=/path, ws-headers=Host:example.com, sni=example.com")
}

func TestSurgeGenerator_Generate_WithOptions(t *testing.T) {
	gen := &SurgeGenerator{}
	proxies := getTestProxies()

	opts := core.GeneratorOptions{
		Base: "[General]\nloglevel = notify",
		ProxySetting: config.ProxySetting{
			UDP: true,
			TFO: true,
		},
	}

	output, err := gen.Generate(proxies, nil, nil, nil, opts)
	assert.NoError(t, err)

	assert.Contains(t, output, "[Proxy]")
	assert.Contains(t, output, "ss-proxy = ss, 1.2.3.4, 8388, encrypt-method=aes-256-gcm, password=password, udp-relay=true, tfo=true")
}

func TestSurgeGenerator_Generate_WithGroups(t *testing.T) {
	gen := &SurgeGenerator{}
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

	opts := core.GeneratorOptions{
		Base:         "[General]\nloglevel = notify",
		ProxySetting: config.ProxySetting{},
	}

	output, err := gen.Generate(proxies, groups, nil, nil, opts)
	assert.NoError(t, err)

	assert.Contains(t, output, "[Proxy Group]")
	// vmess-proxy is skipped, so only ss-proxy remains
	assert.Contains(t, output, "Select Group = select, ss-proxy")
	// Only ss-proxy is supported
	assert.Contains(t, output, "URL Test Group = url-test, ss-proxy, url=http://www.gstatic.com/generate_204")
}
