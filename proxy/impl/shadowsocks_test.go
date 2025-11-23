package impl

import (
	"testing"

	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/proxy/core"
	"github.com/stretchr/testify/assert"
)

func TestShadowsocksProxy_ToSingleConfig(t *testing.T) {
	proxy := &ShadowsocksProxy{
		BaseProxy: core.BaseProxy{
			Type:   "ss",
			Remark: "test-ss",
			Server: "1.2.3.4",
			Port:   8388,
		},
		Password:      "password",
		EncryptMethod: "aes-256-gcm",
	}

	link, err := proxy.ToSingleConfig(&config.ProxySetting{})
	assert.NoError(t, err)
	// aes-256-gcm:password -> YWVzLTI1Ni1nY206cGFzc3dvcmQ=
	assert.Equal(t, "ss://YWVzLTI1Ni1nY206cGFzc3dvcmQ=@1.2.3.4:8388#test-ss", link)

	// Test with Plugin
	t.Run("Test Plugin Params", func(t *testing.T) {
		proxy := &ShadowsocksProxy{
			BaseProxy: core.BaseProxy{
				Type:   "ss",
				Remark: "test-ss-plugin",
				Server: "1.2.3.4",
				Port:   8388,
			},
			Password:      "password",
			EncryptMethod: "aes-256-gcm",
			Plugin:        "simple-obfs",
			PluginOpts: map[string]interface{}{
				"obfs":      "http",
				"obfs-host": "example.com",
			},
		}
		link, err := proxy.ToSingleConfig(nil)
		assert.NoError(t, err)
		assert.Contains(t, link, "plugin=")
		// Plugin params are URL encoded
		assert.Contains(t, link, "obfs%3Dhttp")
		assert.Contains(t, link, "obfs-host%3Dexample.com")
	})

	// Test with special characters in remark
	proxySpecialChars := &ShadowsocksProxy{
		BaseProxy: core.BaseProxy{
			Type:   "ss",
			Remark: "test ss #1",
			Server: "1.2.3.4",
			Port:   8388,
		},
		Password:      "password",
		EncryptMethod: "aes-256-gcm",
	}
	linkSpecial, err := proxySpecialChars.ToSingleConfig(&config.ProxySetting{})
	assert.NoError(t, err)
	assert.Contains(t, linkSpecial, "#test%20ss%20%231")

	// Test with empty remark
	proxyEmptyRemark := &ShadowsocksProxy{
		BaseProxy: core.BaseProxy{
			Type:   "ss",
			Remark: "",
			Server: "1.2.3.4",
			Port:   8388,
		},
		Password:      "password",
		EncryptMethod: "aes-256-gcm",
	}
	linkEmpty, err := proxyEmptyRemark.ToSingleConfig(&config.ProxySetting{})
	assert.NoError(t, err)
	assert.Contains(t, linkEmpty, "#")
	assert.False(t, len(linkEmpty) > len("ss://YWVzLTI1Ni1nY206cGFzc3dvcmQ=@1.2.3.4:8388#"))
}

func TestShadowsocksProxy_ToClashConfig(t *testing.T) {
	proxy := &ShadowsocksProxy{
		BaseProxy: core.BaseProxy{
			Type:   "ss",
			Remark: "test-ss",
			Server: "1.2.3.4",
			Port:   8388,
		},
		Password:      "password",
		EncryptMethod: "aes-256-gcm",
	}

	clashConfig, err := proxy.ToClashConfig(&config.ProxySetting{})
	assert.NoError(t, err)
	assert.NotNil(t, clashConfig)
	assert.Equal(t, "ss", clashConfig["type"])
	assert.Equal(t, "test-ss", clashConfig["name"])
	assert.Equal(t, "1.2.3.4", clashConfig["server"])
	assert.Equal(t, 8388, clashConfig["port"])
	assert.Equal(t, "password", clashConfig["password"])
	assert.Equal(t, "aes-256-gcm", clashConfig["cipher"])
}
