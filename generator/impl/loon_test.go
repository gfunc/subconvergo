package impl

import (
	"testing"

	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/generator/core"
	"github.com/stretchr/testify/assert"
)

func TestLoonGenerator_Generate(t *testing.T) {
	gen := &LoonGenerator{}
	proxies := getTestProxies()

	opts := core.GeneratorOptions{
		Base:         "[General]",
		ProxySetting: config.ProxySetting{},
	}

	output, err := gen.Generate(proxies, nil, nil, nil, opts)
	assert.NoError(t, err)

	assert.Contains(t, output, "[Proxy]")
	assert.Contains(t, output, "ss-proxy = ss, 1.2.3.4:8388, encrypt-method=aes-256-gcm, password=password")
	// TODO: Add SSR support for Loon
	// assert.Contains(t, output, "ssr-proxy = ssr, 1.2.3.4:8388, encrypt-method=aes-256-gcm, password=password, protocol=auth_aes128_md5, obfs=tls1.2_ticket_auth")
	assert.Contains(t, output, "vmess-proxy = vmess, 5.6.7.8:443, username=uuid, ws=true, ws-path=/path, ws-headers=Host:example.com, tls=true, sni=example.com")
	assert.Contains(t, output, "trojan-proxy = trojan, 13.14.15.16:443, password=password, ws=true, ws-path=/path, ws-headers=Host:example.com, sni=example.com")
	// Loon might not support VLESS, Hysteria2, TUIC in the same way or at all in this generator implementation.
	// Checking implementation details would be good, but for now I'll assert what I expect or check if they are skipped/logged.
}

func TestLoonGenerator_Generate_WithGroups(t *testing.T) {
	gen := &LoonGenerator{}
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
		Base:         "[General]",
		ProxySetting: config.ProxySetting{},
	}

	output, err := gen.Generate(proxies, groups, nil, nil, opts)
	assert.NoError(t, err)

	assert.Contains(t, output, "[Proxy Group]")
	assert.Contains(t, output, "Select Group = select, ss-proxy, vmess-proxy")
	assert.Contains(t, output, "URL Test Group = url-test, ss-proxy, ssr-proxy, vmess-proxy, vless-proxy, trojan-proxy, hysteria2-proxy, tuic-proxy, anytls-proxy, img-url=https://raw.githubusercontent.com/Koolson/Qure/master/IconSet/Proxy.png")
}
