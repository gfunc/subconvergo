package impl

import (
	"net/url"
	"testing"

	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/proxy/core"
	"github.com/stretchr/testify/assert"
)

func TestHysteriaProxy_ToShareLink(t *testing.T) {
	proxy := &HysteriaProxy{
		BaseProxy: core.BaseProxy{
			Type:   "hysteria2",
			Remark: "test-hysteria2",
			Server: "1.2.3.4",
			Port:   443,
		},
		Password:      "password",
		Obfs:          "salamander",
		AllowInsecure: true,
		Params:        url.Values{},
	}
	proxy.Params.Add("sni", "example.com")

	link, err := proxy.ToShareLink(&config.ProxySetting{})
	assert.NoError(t, err)
	// hysteria2://password@1.2.3.4:443?insecure=1&obfs=salamander&sni=example.com#test-hysteria2
	assert.Contains(t, link, "hysteria2://password@1.2.3.4:443")
	assert.Contains(t, link, "insecure=1")
	assert.Contains(t, link, "obfs=salamander")
	assert.Contains(t, link, "sni=example.com")
	assert.Contains(t, link, "#test-hysteria2")

	// Test Hysteria 1
	proxyH1 := &HysteriaProxy{
		BaseProxy: core.BaseProxy{
			Type:   "hysteria",
			Remark: "test-hysteria1",
			Server: "1.2.3.4",
			Port:   443,
		},
		Params: url.Values{},
	}
	proxyH1.Params.Add("up", "100")
	proxyH1.Params.Add("down", "100")
	proxyH1.Params.Add("auth", "myauth")

	linkH1, err := proxyH1.ToShareLink(&config.ProxySetting{})
	assert.NoError(t, err)
	assert.Contains(t, linkH1, "hysteria://@1.2.3.4:443") // Password is empty in struct, but auth param is present
	assert.Contains(t, linkH1, "up=100")
	assert.Contains(t, linkH1, "down=100")
	assert.Contains(t, linkH1, "auth=myauth")

	// Test with empty params
	proxyEmpty := &HysteriaProxy{
		BaseProxy: core.BaseProxy{
			Type:   "hysteria2",
			Remark: "test-empty",
			Server: "1.2.3.4",
			Port:   443,
		},
		Password: "pass",
		Params:   url.Values{},
	}
	linkEmpty, err := proxyEmpty.ToShareLink(&config.ProxySetting{})
	assert.NoError(t, err)
	assert.Contains(t, linkEmpty, "hysteria2://pass@1.2.3.4:443")
	assert.NotContains(t, linkEmpty, "?")
	assert.Contains(t, linkEmpty, "#test-empty")
}
