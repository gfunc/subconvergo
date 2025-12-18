package impl

import (
	"testing"

	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/proxy/core"
	"github.com/stretchr/testify/assert"
)

func TestHysteria2Proxy_ToClashConfig(t *testing.T) {
	proxy := &Hysteria2Proxy{
		BaseProxy: core.BaseProxy{
			Type:   "hysteria2",
			Remark: "test-hysteria2",
			Server: "1.2.3.4",
			Port:   443,
		},
		Password:       "password",
		Sni:            "example.com",
		SkipCertVerify: true,
		Obfs:           "salamander",
		ObfsPassword:   "secret",
	}

	clashConfig, err := proxy.ToClashConfig(&config.ProxySetting{})
	assert.NoError(t, err)
	assert.NotNil(t, clashConfig)

	assert.Equal(t, "hysteria2", clashConfig["type"])
	assert.Equal(t, "test-hysteria2", clashConfig["name"])
	assert.Equal(t, "1.2.3.4", clashConfig["server"])
	assert.Equal(t, 443, clashConfig["port"])
	assert.Equal(t, "password", clashConfig["password"])
	assert.Equal(t, "password", clashConfig["auth"]) // Check for auth field
	assert.Equal(t, "example.com", clashConfig["sni"])
	assert.Equal(t, true, clashConfig["skip-cert-verify"])
	assert.Equal(t, "salamander", clashConfig["obfs"])
	assert.Equal(t, "secret", clashConfig["obfs-password"])
}
