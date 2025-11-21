package impl

import (
	"net/url"
	"testing"

	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/proxy/core"
	"github.com/stretchr/testify/assert"
)

func TestTUICProxy_ToShareLink(t *testing.T) {
	proxy := &TUICProxy{
		BaseProxy: core.BaseProxy{
			Type:   "tuic",
			Remark: "test-tuic",
			Server: "1.2.3.4",
			Port:   443,
		},
		UUID:          "uuid",
		Password:      "password",
		AllowInsecure: true,
		Params:        url.Values{},
	}
	proxy.Params.Add("sni", "example.com")

	link, err := proxy.ToShareLink(&config.ProxySetting{})
	assert.NoError(t, err)
	// tuic://uuid:password@1.2.3.4:443?allow_insecure=1&sni=example.com#test-tuic
	assert.Contains(t, link, "tuic://uuid:password@1.2.3.4:443")
	assert.Contains(t, link, "allow_insecure=1")
	assert.Contains(t, link, "sni=example.com")
	assert.Contains(t, link, "#test-tuic")

	// Test No Password
	proxyNoPass := &TUICProxy{
		BaseProxy: core.BaseProxy{
			Type:   "tuic",
			Remark: "test-tuic-nopass",
			Server: "1.2.3.4",
			Port:   443,
		},
		UUID:   "uuid",
		Params: url.Values{},
	}
	linkNoPass, err := proxyNoPass.ToShareLink(&config.ProxySetting{})
	assert.NoError(t, err)
	assert.Contains(t, linkNoPass, "tuic://uuid@1.2.3.4:443")

	// Test Params
	proxyParams := &TUICProxy{
		BaseProxy: core.BaseProxy{
			Type:   "tuic",
			Remark: "test-tuic-params",
			Server: "1.2.3.4",
			Port:   443,
		},
		UUID:   "uuid",
		Params: url.Values{},
	}
	proxyParams.Params.Add("congestion_control", "bbr")
	proxyParams.Params.Add("udp_relay_mode", "native")
	linkParams, err := proxyParams.ToShareLink(&config.ProxySetting{})
	assert.NoError(t, err)
	assert.Contains(t, linkParams, "congestion_control=bbr")
	assert.Contains(t, linkParams, "udp_relay_mode=native")
}
