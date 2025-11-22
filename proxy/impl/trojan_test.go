package impl

import (
	"testing"

	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/proxy/core"
	"github.com/stretchr/testify/assert"
)

func TestTrojanProxy_ToShareLink(t *testing.T) {
	proxy := &TrojanProxy{
		BaseProxy: core.BaseProxy{
			Type:   "trojan",
			Remark: "test-trojan",
			Server: "1.2.3.4",
			Port:   443,
		},
		Password:      "password",
		Network:       "ws",
		Path:          "/path",
		Host:          "example.com",
		TLS:           true,
		AllowInsecure: true,
	}

	link, err := proxy.ToShareLink(&config.ProxySetting{})
	assert.NoError(t, err)
	// trojan://password@1.2.3.4:443?allowInsecure=1&path=%2Fpath&sni=example.com&type=ws#test-trojan
	assert.Contains(t, link, "trojan://password@1.2.3.4:443")
	assert.Contains(t, link, "sni=example.com")
	assert.Contains(t, link, "type=ws")
	assert.Contains(t, link, "path=%2Fpath")
	assert.Contains(t, link, "allowInsecure=1")
	assert.Contains(t, link, "#test-trojan")

	// Test TCP (default) and no SNI
	proxyTCP := &TrojanProxy{
		BaseProxy: core.BaseProxy{
			Type:   "trojan",
			Remark: "test-trojan-tcp",
			Server: "1.2.3.4",
			Port:   443,
		},
		Password: "password",
	}
	linkTCP, err := proxyTCP.ToShareLink(&config.ProxySetting{})
	assert.NoError(t, err)
	assert.Contains(t, linkTCP, "trojan://password@1.2.3.4:443")
	assert.NotContains(t, linkTCP, "type=")
	assert.NotContains(t, linkTCP, "sni=")
	assert.Contains(t, linkTCP, "#test-trojan-tcp")

	// Test special chars in path
	proxySpecialPath := &TrojanProxy{
		BaseProxy: core.BaseProxy{
			Type:   "trojan",
			Remark: "test-trojan-path",
			Server: "1.2.3.4",
			Port:   443,
		},
		Password: "password",
		Network:  "ws",
		Path:     "/path with spaces",
	}
	linkSpecialPath, err := proxySpecialPath.ToShareLink(&config.ProxySetting{})
	assert.NoError(t, err)
	assert.Contains(t, linkSpecialPath, "path=%2Fpath+with+spaces")
}

func TestTrojanProxy_ToClashConfig(t *testing.T) {
	proxy := &TrojanProxy{
		BaseProxy: core.BaseProxy{
			Type:   "trojan",
			Remark: "test-trojan",
			Server: "1.2.3.4",
			Port:   443,
		},
		Password:      "password",
		Network:       "ws",
		Path:          "/path",
		Host:          "example.com",
		TLS:           true,
		AllowInsecure: true,
	}

	clashConfig := proxy.ToClashConfig(&config.ProxySetting{})
	assert.NotNil(t, clashConfig)
	assert.Equal(t, "trojan", clashConfig["type"])
	assert.Equal(t, "test-trojan", clashConfig["name"])
	assert.Equal(t, "1.2.3.4", clashConfig["server"])
	assert.Equal(t, 443, clashConfig["port"])
	assert.Equal(t, "password", clashConfig["password"])
	assert.Equal(t, "ws", clashConfig["network"])
	assert.Equal(t, "example.com", clashConfig["sni"])
	assert.Equal(t, true, clashConfig["skip-cert-verify"])

	wsOpts, ok := clashConfig["ws-opts"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "/path", wsOpts["path"])
}
