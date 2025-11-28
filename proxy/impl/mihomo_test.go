package impl

import (
	"testing"

	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/proxy/core"
	"github.com/stretchr/testify/assert"
)

func TestMihomoProxy_ToSingleConfig(t *testing.T) {
	ssProxy := &ShadowsocksProxy{
		BaseProxy: core.BaseProxy{
			Type:   "ss",
			Remark: "ss-proxy",
			Server: "1.2.3.4",
			Port:   8388,
		},
		Password:      "password",
		EncryptMethod: "aes-256-gcm",
	}

	mihomoProxy := &MihomoProxy{
		ProxyInterface: ssProxy,
	}

	link, err := mihomoProxy.ToSingleConfig(&config.ProxySetting{})
	assert.NoError(t, err)
	assert.Contains(t, link, "ss://")
	assert.Contains(t, link, "#ss-proxy")
}

func TestMihomoProxy_ToClashConfig(t *testing.T) {
	ssProxy := &ShadowsocksProxy{
		BaseProxy: core.BaseProxy{
			Type:   "ss",
			Remark: "ss-proxy",
			Server: "1.2.3.4",
			Port:   8388,
		},
		Password:      "password",
		EncryptMethod: "aes-256-gcm",
	}

	mihomoProxy := &MihomoProxy{
		ProxyInterface: ssProxy,
	}

	clashConfig, err := mihomoProxy.ToClashConfig(&config.ProxySetting{})
	assert.NoError(t, err)
	assert.NotNil(t, clashConfig)
	assert.Equal(t, "ss", clashConfig["type"])
	assert.Equal(t, "ss-proxy", clashConfig["name"])
	assert.Equal(t, "1.2.3.4", clashConfig["server"])
	assert.Equal(t, 8388, clashConfig["port"])
	assert.Equal(t, "password", clashConfig["password"])
	assert.Equal(t, "aes-256-gcm", clashConfig["cipher"])
}
