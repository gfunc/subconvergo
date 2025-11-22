package impl

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/proxy/core"
	"github.com/stretchr/testify/assert"
)

func TestShadowsocksRProxy_ToShareLink(t *testing.T) {
	proxy := &ShadowsocksRProxy{
		BaseProxy: core.BaseProxy{
			Type:   "ssr",
			Remark: "test-ssr",
			Server: "1.2.3.4",
			Port:   8388,
		},
		Password:      "password",
		EncryptMethod: "aes-256-cfb",
		Protocol:      "origin",
		Obfs:          "plain",
		ProtocolParam: "param1",
		ObfsParam:     "param2",
	}

	link, err := proxy.ToShareLink(&config.ProxySetting{})
	assert.NoError(t, err)
	assert.True(t, strings.HasPrefix(link, "ssr://"))

	decodedBytes, err := base64.URLEncoding.DecodeString(strings.TrimPrefix(link, "ssr://"))
	assert.NoError(t, err)
	decoded := string(decodedBytes)

	// Check main part
	assert.Contains(t, decoded, "1.2.3.4:8388:origin:aes-256-cfb:plain:")

	// Check params
	assert.Contains(t, decoded, "obfsparam=")
	assert.Contains(t, decoded, "protoparam=")
	assert.Contains(t, decoded, "remarks=")

	// Test with empty params
	proxyEmptyParams := &ShadowsocksRProxy{
		BaseProxy: core.BaseProxy{
			Type:   "ssr",
			Remark: "test-ssr-empty",
			Server: "1.2.3.4",
			Port:   8388,
		},
		Password:      "password",
		EncryptMethod: "aes-256-cfb",
		Protocol:      "origin",
		Obfs:          "plain",
	}
	linkEmpty, err := proxyEmptyParams.ToShareLink(&config.ProxySetting{})
	assert.NoError(t, err)
	decodedEmptyBytes, err := base64.URLEncoding.DecodeString(strings.TrimPrefix(linkEmpty, "ssr://"))
	assert.NoError(t, err)
	decodedEmpty := string(decodedEmptyBytes)
	assert.NotContains(t, decodedEmpty, "obfsparam=")
	assert.NotContains(t, decodedEmpty, "protoparam=")
	assert.Contains(t, decodedEmpty, "remarks=")

	// Test with special characters in remark
	proxySpecial := &ShadowsocksRProxy{
		BaseProxy: core.BaseProxy{
			Type:   "ssr",
			Remark: "test ssr #1",
			Server: "1.2.3.4",
			Port:   8388,
		},
		Password:      "password",
		EncryptMethod: "aes-256-cfb",
		Protocol:      "origin",
		Obfs:          "plain",
	}
	linkSpecial, err := proxySpecial.ToShareLink(&config.ProxySetting{})
	assert.NoError(t, err)
	decodedSpecialBytes, err := base64.URLEncoding.DecodeString(strings.TrimPrefix(linkSpecial, "ssr://"))
	assert.NoError(t, err)
	decodedSpecial := string(decodedSpecialBytes)
	// remarks should be base64 encoded
	// test ssr #1 -> dGVzdCBzc3IgIzE=
	assert.Contains(t, decodedSpecial, "remarks=dGVzdCBzc3IgIzE=")
}

func TestShadowsocksRProxy_ToClashConfig(t *testing.T) {
	proxy := &ShadowsocksRProxy{
		BaseProxy: core.BaseProxy{
			Type:   "ssr",
			Remark: "test-ssr",
			Server: "1.2.3.4",
			Port:   8388,
		},
		Password:      "password",
		EncryptMethod: "aes-256-cfb",
		Protocol:      "origin",
		Obfs:          "plain",
		ProtocolParam: "param1",
		ObfsParam:     "param2",
	}

	clashConfig := proxy.ToClashConfig(&config.ProxySetting{})
	assert.NotNil(t, clashConfig)
	assert.Equal(t, "ssr", clashConfig["type"])
	assert.Equal(t, "test-ssr", clashConfig["name"])
	assert.Equal(t, "1.2.3.4", clashConfig["server"])
	assert.Equal(t, 8388, clashConfig["port"])
	assert.Equal(t, "password", clashConfig["password"])
	assert.Equal(t, "aes-256-cfb", clashConfig["cipher"])
	assert.Equal(t, "origin", clashConfig["protocol"])
	assert.Equal(t, "plain", clashConfig["obfs"])
	assert.Equal(t, "param1", clashConfig["protocol-param"])
	assert.Equal(t, "param2", clashConfig["obfs-param"])
}
