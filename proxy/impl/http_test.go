package impl

import (
	"testing"

	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/proxy/core"
	"github.com/gfunc/subconvergo/utils"
	"github.com/stretchr/testify/assert"
)

func TestHttpProxy_ToSingleConfig(t *testing.T) {
	// Test HTTP
	proxy := &HttpProxy{
		BaseProxy: core.BaseProxy{
			Remark: "test-http",
			Server: "1.2.3.4",
			Port:   8080,
		},
		Username: "user",
		Password: "password",
	}
	link, err := proxy.ToSingleConfig(&config.ProxySetting{})
	assert.NoError(t, err)
	assert.Contains(t, link, "http://user:password@1.2.3.4:8080")
	assert.Contains(t, link, "#test-http")

	// Test HTTPS
	proxyHttps := &HttpProxy{
		BaseProxy: core.BaseProxy{
			Remark: "test-https",
			Server: "1.2.3.4",
			Port:   443,
		},
		Username: "user",
		Password: "password",
		Tls:      true,
	}
	linkHttps, err := proxyHttps.ToSingleConfig(&config.ProxySetting{})
	assert.NoError(t, err)
	assert.Contains(t, linkHttps, "https://user:password@1.2.3.4:443")

	// Test No Auth
	proxyNoAuth := &HttpProxy{
		BaseProxy: core.BaseProxy{
			Remark: "test-noauth",
			Server: "1.2.3.4",
			Port:   80,
		},
	}
	linkNoAuth, err := proxyNoAuth.ToSingleConfig(&config.ProxySetting{})
	assert.NoError(t, err)
	assert.Contains(t, linkNoAuth, "http://1.2.3.4:80")
}

func TestHttpProxy_ToClashConfig(t *testing.T) {
	proxy := &HttpProxy{
		BaseProxy: core.BaseProxy{
			Remark: "test-http",
			Server: "1.2.3.4",
			Port:   8080,
		},
		Username:       "user",
		Password:       "password",
		Tls:            true,
		SkipCertVerify: true,
	}

	clashConfig, err := proxy.ToClashConfig(&config.ProxySetting{})
	assert.NoError(t, err)
	assert.NotNil(t, clashConfig)
	assert.Equal(t, "http", clashConfig["type"])
	assert.Equal(t, "test-http", clashConfig["name"])
	assert.Equal(t, "1.2.3.4", clashConfig["server"])
	assert.Equal(t, 8080, clashConfig["port"])
	assert.Equal(t, "user", clashConfig["username"])
	assert.Equal(t, "password", clashConfig["password"])
	assert.Equal(t, true, clashConfig["tls"])
	assert.Equal(t, true, clashConfig["skip-cert-verify"])
}

func TestHttpProxy_ToSurgeConfig(t *testing.T) {
	proxy := &HttpProxy{
		BaseProxy: core.BaseProxy{
			Remark: "test-http",
			Server: "1.2.3.4",
			Port:   8080,
		},
		Username: "user",
		Password: "password",
		Tls:      true,
	}

	surgeConfig, err := proxy.ToSurgeConfig(&config.ProxySetting{})
	assert.NoError(t, err)
	// test-http = http, 1.2.3.4, 8080, username=user, password=password, tls=true
	assert.Contains(t, surgeConfig, "test-http = http, 1.2.3.4, 8080")
	assert.Contains(t, surgeConfig, "username=user")
	assert.Contains(t, surgeConfig, "password=password")
	assert.Contains(t, surgeConfig, "tls=true")
}

func TestHttpProxy_ToClashConfig_GlobalOverrides(t *testing.T) {
	proxy := &HttpProxy{
		BaseProxy: core.BaseProxy{
			Remark: "test-http",
		},
		Username: "user",
		Password: "password",
		Tls:      true,
	}
	proxy.Server = "1.2.3.4"
	proxy.Port = 8080

	globalSettings := &config.ProxySetting{
		UDP:   utils.BoolPtr(true), // HTTP proxy usually doesn't support UDP, but let's see if it's ignored
		TFO:   utils.BoolPtr(true),
		SCV:   utils.BoolPtr(true),
		TLS13: utils.BoolPtr(true),
	}

	clashConfig, err := proxy.ToClashConfig(globalSettings)
	assert.NoError(t, err)
	assert.NotNil(t, clashConfig)

	assert.Nil(t, clashConfig["udp"]) // Should be nil/false
	assert.Equal(t, true, clashConfig["tfo"])
	assert.Equal(t, true, clashConfig["skip-cert-verify"])
	assert.Equal(t, true, clashConfig["tls13"])
}
