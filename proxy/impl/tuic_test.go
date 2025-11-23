package impl

import (
	"net/url"
	"testing"

	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/proxy/core"
	"github.com/stretchr/testify/assert"
)

func TestTUICProxy_ToSingleConfig(t *testing.T) {
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

	link, err := proxy.ToSingleConfig(&config.ProxySetting{})
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
	linkNoPass, err := proxyNoPass.ToSingleConfig(&config.ProxySetting{})
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
	linkParams, err := proxyParams.ToSingleConfig(&config.ProxySetting{})
	assert.NoError(t, err)
	assert.Contains(t, linkParams, "congestion_control=bbr")
	assert.Contains(t, linkParams, "udp_relay_mode=native")
}

func TestTUICProxy_ToClashConfig(t *testing.T) {
	params := url.Values{}
	params.Set("sni", "example.com")
	params.Set("congestion_control", "bbr")
	params.Set("udp_relay_mode", "native")

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
		Params:        params,
	}

	clashConfig, err := proxy.ToClashConfig(&config.ProxySetting{})
	assert.NoError(t, err)
	assert.NotNil(t, clashConfig)
	assert.Equal(t, "tuic", clashConfig["type"])
	assert.Equal(t, "test-tuic", clashConfig["name"])
	assert.Equal(t, "1.2.3.4", clashConfig["server"])
	assert.Equal(t, 443, clashConfig["port"])
	assert.Equal(t, "uuid", clashConfig["uuid"])
	assert.Equal(t, "password", clashConfig["password"])
	assert.Equal(t, true, clashConfig["skip-cert-verify"])
	assert.Equal(t, "example.com", clashConfig["sni"])
	assert.Equal(t, "bbr", clashConfig["congestion-controller"])
	assert.Equal(t, "native", clashConfig["udp-relay-mode"])
}
