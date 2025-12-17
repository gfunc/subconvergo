package impl

import (
	"testing"

	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/proxy/core"
	"github.com/gfunc/subconvergo/utils"
	"github.com/stretchr/testify/assert"
)

func TestWireGuardProxy_ToSingleConfig(t *testing.T) {
	proxy := &WireGuardProxy{
		BaseProxy: core.BaseProxy{
			Remark: "test-wg",
		},
	}
	_, err := proxy.ToSingleConfig(&config.ProxySetting{})
	assert.Error(t, err)
}

func TestWireGuardProxy_ToClashConfig(t *testing.T) {
	proxy := &WireGuardProxy{
		BaseProxy: core.BaseProxy{
			Remark: "test-wg",
			Server: "1.2.3.4",
			Port:   51820,
		},
		Ip:           "10.0.0.1",
		Ipv6:         "2001:db8::1",
		PrivateKey:   "private",
		PublicKey:    "public",
		PreSharedKey: "preshared",
		Dns:          []string{"1.1.1.1", "8.8.8.8"},
		Mtu:          1420,
		Udp:          true,
	}

	clashConfig, err := proxy.ToClashConfig(&config.ProxySetting{})
	assert.NoError(t, err)
	assert.NotNil(t, clashConfig)
	assert.Equal(t, "wireguard", clashConfig["type"])
	assert.Equal(t, "test-wg", clashConfig["name"])
	assert.Equal(t, "1.2.3.4", clashConfig["server"])
	assert.Equal(t, 51820, clashConfig["port"])
	assert.Equal(t, "10.0.0.1", clashConfig["ip"])
	assert.Equal(t, "2001:db8::1", clashConfig["ipv6"])
	assert.Equal(t, "private", clashConfig["private-key"])
	assert.Equal(t, "public", clashConfig["public-key"])
	assert.Equal(t, "preshared", clashConfig["pre-shared-key"])
	assert.Equal(t, []string{"1.1.1.1", "8.8.8.8"}, clashConfig["dns"])
	assert.Equal(t, 1420, clashConfig["mtu"])
	assert.Equal(t, true, clashConfig["udp"])
}

func TestWireGuardProxy_ToClashConfig_GlobalOverrides(t *testing.T) {
	proxy := &WireGuardProxy{
		BaseProxy: core.BaseProxy{
			Remark: "test-wg",
		},
		Ip:         "10.0.0.1",
		PrivateKey: "private",
		PublicKey:  "public",
	}
	proxy.Server = "1.2.3.4"
	proxy.Port = 51820

	globalSettings := &config.ProxySetting{
		UDP:   utils.BoolPtr(true),
		TFO:   utils.BoolPtr(true), // Should be ignored
		SCV:   utils.BoolPtr(true), // Should be ignored
		TLS13: utils.BoolPtr(true), // Should be ignored
	}

	clashConfig, err := proxy.ToClashConfig(globalSettings)
	assert.NoError(t, err)
	assert.NotNil(t, clashConfig)

	assert.Equal(t, true, clashConfig["udp"])
	assert.Nil(t, clashConfig["tfo"])
	assert.Nil(t, clashConfig["skip-cert-verify"])
	assert.Nil(t, clashConfig["tls13"])
}
