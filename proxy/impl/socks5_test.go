package impl

import (
	"testing"

	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/proxy/core"
	"github.com/gfunc/subconvergo/utils"
	"github.com/stretchr/testify/assert"
)

func TestSocks5Proxy_ToSingleConfig(t *testing.T) {
	// With Auth
	proxy := &Socks5Proxy{
		BaseProxy: core.BaseProxy{
			Remark: "test-socks5",
			Server: "1.2.3.4",
			Port:   1080,
		},
		Username: "user",
		Password: "password",
	}
	link, err := proxy.ToSingleConfig(&config.ProxySetting{})
	assert.NoError(t, err)
	assert.Contains(t, link, "socks5://user:password@1.2.3.4:1080")
	assert.Contains(t, link, "#test-socks5")

	// No Auth
	proxyNoAuth := &Socks5Proxy{
		BaseProxy: core.BaseProxy{
			Remark: "test-noauth",
			Server: "1.2.3.4",
			Port:   1080,
		},
	}
	linkNoAuth, err := proxyNoAuth.ToSingleConfig(&config.ProxySetting{})
	assert.NoError(t, err)
	assert.Contains(t, linkNoAuth, "socks5://1.2.3.4:1080")
}

func TestSocks5Proxy_ToClashConfig(t *testing.T) {
	proxy := &Socks5Proxy{
		BaseProxy: core.BaseProxy{
			Remark: "test-socks5",
			Server: "1.2.3.4",
			Port:   1080,
		},
		Username: "user",
		Password: "password",
		TLS:      true,
	}

	clashConfig, err := proxy.ToClashConfig(&config.ProxySetting{})
	assert.NoError(t, err)
	assert.NotNil(t, clashConfig)
	assert.Equal(t, "socks5", clashConfig["type"])
	assert.Equal(t, "test-socks5", clashConfig["name"])
	assert.Equal(t, "1.2.3.4", clashConfig["server"])
	assert.Equal(t, 1080, clashConfig["port"])
	assert.Equal(t, "user", clashConfig["username"])
	assert.Equal(t, "password", clashConfig["password"])
	assert.Equal(t, true, clashConfig["tls"])
}

func TestSocks5Proxy_ToSurgeConfig(t *testing.T) {
	proxy := &Socks5Proxy{
		BaseProxy: core.BaseProxy{
			Remark: "test-socks5",
			Server: "1.2.3.4",
			Port:   1080,
		},
		Username: "user",
		Password: "password",
		TLS:      true,
	}

	surgeConfig, err := proxy.ToSurgeConfig(&config.ProxySetting{})
	assert.NoError(t, err)
	// test-socks5 = socks5-tls, 1.2.3.4, 1080, username=user, password=password
	assert.Contains(t, surgeConfig, "test-socks5 = socks5-tls, 1.2.3.4, 1080")
	assert.Contains(t, surgeConfig, "username=user")
	assert.Contains(t, surgeConfig, "password=password")
}

func TestSocks5Proxy_ToClashConfig_GlobalOverrides(t *testing.T) {
	proxy := &Socks5Proxy{
		BaseProxy: core.BaseProxy{
			Remark: "test-socks5",
		},
		Username: "user",
		Password: "password",
		TLS:      true,
	}
	proxy.Server = "1.2.3.4"
	proxy.Port = 1080

	globalSettings := &config.ProxySetting{
		UDP:   utils.BoolPtr(true),
		TFO:   utils.BoolPtr(true),
		SCV:   utils.BoolPtr(true),
		TLS13: utils.BoolPtr(true),
	}

	clashConfig, err := proxy.ToClashConfig(globalSettings)
	assert.NoError(t, err)
	assert.NotNil(t, clashConfig)

	assert.Equal(t, true, clashConfig["udp"])
	assert.Equal(t, true, clashConfig["tfo"])
	assert.Equal(t, true, clashConfig["skip-cert-verify"])
	assert.Equal(t, true, clashConfig["tls13"])
}
