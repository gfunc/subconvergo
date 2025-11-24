package sub

import (
	"testing"

	"github.com/gfunc/subconvergo/proxy/impl"
	"github.com/stretchr/testify/assert"
)

func TestSurgeSubscriptionParser(t *testing.T) {
	parser := &SurgeSubscriptionParser{}

	t.Run("CanParse", func(t *testing.T) {
		assert.True(t, parser.CanParse("[Proxy]\nProxyA = ss, ..."))
		assert.False(t, parser.CanParse("ss://..."))
	})

	t.Run("Parse", func(t *testing.T) {
		content := `
[General]
loglevel = notify

[Proxy]
ProxyA = ss, 1.2.3.4, 8388, encrypt-method=aes-256-gcm, password=password
ProxyB = vmess, v2ray.cool, 10086, username=a3482e88-686a-4a58-8126-99c9df64b7bf, ws=true, tls=true, ws-path=/v2ray
ProxyC = http, 1.2.3.4, 8080, username=user, password=pass
ProxyD = trojan, 1.2.3.4, 443, password=pass, sni=example.com
ProxyE = socks5, 1.2.3.4, 1080, username=user, password=pass
ProxyF = custom, 1.2.3.4, 8388, aes-256-gcm, password, https://github.com/..., obfs=http, obfs-host=bing.com
`
		sub, err := parser.Parse(content)
		assert.NoError(t, err)
		assert.NotNil(t, sub)
		assert.Len(t, sub.Proxies, 6)

		// Check ProxyA (SS)
		var ss *impl.ShadowsocksProxy
		if mp, ok := sub.Proxies[0].(*impl.MihomoProxy); ok {
			ss = mp.ProxyInterface.(*impl.ShadowsocksProxy)
		} else {
			ss = sub.Proxies[0].(*impl.ShadowsocksProxy)
		}
		assert.Equal(t, "ss", ss.Type)
		assert.Equal(t, "ProxyA", ss.Remark)
		assert.Equal(t, "1.2.3.4", ss.Server)
		assert.Equal(t, 8388, ss.Port)
		assert.Equal(t, "aes-256-gcm", ss.EncryptMethod)
		assert.Equal(t, "password", ss.Password)

		// Check ProxyB (VMess)
		var vmess *impl.VMessProxy
		if mp, ok := sub.Proxies[1].(*impl.MihomoProxy); ok {
			vmess = mp.ProxyInterface.(*impl.VMessProxy)
		} else {
			vmess = sub.Proxies[1].(*impl.VMessProxy)
		}
		assert.Equal(t, "vmess", vmess.Type)
		assert.Equal(t, "ProxyB", vmess.Remark)
		assert.Equal(t, "v2ray.cool", vmess.Server)
		assert.Equal(t, 10086, vmess.Port)
		assert.Equal(t, "a3482e88-686a-4a58-8126-99c9df64b7bf", vmess.UUID)
		assert.Equal(t, "ws", vmess.Network)
		assert.True(t, vmess.TLS)
		assert.Equal(t, "/v2ray", vmess.Path)

		// Check ProxyC (HTTP)
		var http *impl.HttpProxy
		if mp, ok := sub.Proxies[2].(*impl.MihomoProxy); ok {
			http = mp.ProxyInterface.(*impl.HttpProxy)
		} else {
			http = sub.Proxies[2].(*impl.HttpProxy)
		}
		assert.Equal(t, "http", http.Type)
		assert.Equal(t, "ProxyC", http.Remark)
		assert.Equal(t, "1.2.3.4", http.Server)
		assert.Equal(t, 8080, http.Port)
		assert.Equal(t, "user", http.Username)
		assert.Equal(t, "pass", http.Password)

		// Check ProxyD (Trojan)
		var trojan *impl.TrojanProxy
		if mp, ok := sub.Proxies[3].(*impl.MihomoProxy); ok {
			trojan = mp.ProxyInterface.(*impl.TrojanProxy)
		} else {
			trojan = sub.Proxies[3].(*impl.TrojanProxy)
		}
		assert.Equal(t, "trojan", trojan.Type)
		assert.Equal(t, "ProxyD", trojan.Remark)
		assert.Equal(t, "1.2.3.4", trojan.Server)
		assert.Equal(t, 443, trojan.Port)
		assert.Equal(t, "pass", trojan.Password)
		assert.Equal(t, "example.com", trojan.Host)

		// Check ProxyE (Socks5)
		var socks *impl.Socks5Proxy
		if mp, ok := sub.Proxies[4].(*impl.MihomoProxy); ok {
			socks = mp.ProxyInterface.(*impl.Socks5Proxy)
		} else {
			socks = sub.Proxies[4].(*impl.Socks5Proxy)
		}
		assert.Equal(t, "socks5", socks.Type)
		assert.Equal(t, "ProxyE", socks.Remark)
		assert.Equal(t, "1.2.3.4", socks.Server)
		assert.Equal(t, 1080, socks.Port)
		assert.Equal(t, "user", socks.Username)
		assert.Equal(t, "pass", socks.Password)

		// Check ProxyF (Custom SS)
		var custom *impl.ShadowsocksProxy
		if mp, ok := sub.Proxies[5].(*impl.MihomoProxy); ok {
			custom = mp.ProxyInterface.(*impl.ShadowsocksProxy)
		} else {
			custom = sub.Proxies[5].(*impl.ShadowsocksProxy)
		}
		assert.Equal(t, "ss", custom.Type)
		assert.Equal(t, "ProxyF", custom.Remark)
		assert.Equal(t, "1.2.3.4", custom.Server)
		assert.Equal(t, 8388, custom.Port)
		assert.Equal(t, "aes-256-gcm", custom.EncryptMethod)
		assert.Equal(t, "password", custom.Password)
		assert.Equal(t, "obfs", custom.Plugin)
		assert.Equal(t, "bing.com", custom.PluginOpts["obfs-host"])
	})
}
