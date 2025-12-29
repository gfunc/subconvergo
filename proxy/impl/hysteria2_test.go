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

func TestHysteria2Proxy_ToClashConfig_MihomoParams(t *testing.T) {
	echEnable := true
	proxy := &Hysteria2Proxy{
		BaseProxy: core.BaseProxy{
			Type:   "hysteria2",
			Remark: "test-hysteria2-mihomo",
			Server: "1.2.3.4",
			Port:   443,
		},
		Password:                       "password",
		Mport:                          "443,8443",
		InitialStreamReceiveWindow:     1024,
		MaxStreamReceiveWindow:         2048,
		InitialConnectionReceiveWindow: 4096,
		MaxConnectionReceiveWindow:     8192,
		UdpMTU:                         1350,
		IpVersion:                      "ipv6",
		ClientFingerprint:              "ios",
		EchEnable:                      &echEnable,
		EchConfig:                      "ech-config",
		Certificate:                    "cert",
		PrivateKeyPem:                  "key",
	}

	config, err := proxy.ToClashConfig(nil)
	assert.NoError(t, err)

	assert.Equal(t, "443,8443", config["mport"])
	assert.Equal(t, uint64(1024), config["initial-stream-receive-window"])
	assert.Equal(t, uint64(2048), config["max-stream-receive-window"])
	assert.Equal(t, uint64(4096), config["initial-connection-receive-window"])
	assert.Equal(t, uint64(8192), config["max-connection-receive-window"])
	assert.Equal(t, 1350, config["udp-mtu"])
	assert.Equal(t, "ipv6", config["ip-version"])
	assert.Equal(t, "ios", config["client-fingerprint"])

	echOpts, ok := config["ech-opts"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, true, echOpts["enable"])
	assert.Equal(t, "ech-config", echOpts["config"])

	assert.Equal(t, "cert", config["certificate"])
	assert.Equal(t, "key", config["private-key"])
}
