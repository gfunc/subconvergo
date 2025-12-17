package impl

import (
	"testing"

	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/proxy/core"
	"github.com/gfunc/subconvergo/utils"
	"github.com/stretchr/testify/assert"
)

func TestSnellProxy_ToSingleConfig(t *testing.T) {
	proxy := &SnellProxy{
		BaseProxy: core.BaseProxy{
			Remark: "test-snell",
			Server: "1.2.3.4",
			Port:   8080,
		},
		Psk:       "psk",
		Obfs:      "http",
		ObfsParam: "example.com",
		Version:   2,
	}

	link, err := proxy.ToSingleConfig(&config.ProxySetting{})
	assert.NoError(t, err)
	// snell://1.2.3.4:8080?obfs=http&obfs-host=example.com&psk=psk&version=2#test-snell
	assert.Contains(t, link, "snell://1.2.3.4:8080")
	assert.Contains(t, link, "psk=psk")
	assert.Contains(t, link, "obfs=http")
	assert.Contains(t, link, "obfs-host=example.com")
	assert.Contains(t, link, "version=2")
	assert.Contains(t, link, "#test-snell")
}

func TestSnellProxy_ToClashConfig(t *testing.T) {
	proxy := &SnellProxy{
		BaseProxy: core.BaseProxy{
			Remark: "test-snell",
			Server: "1.2.3.4",
			Port:   8080,
		},
		Psk:       "psk",
		Obfs:      "http",
		ObfsParam: "example.com",
		Version:   2,
	}

	clashConfig, err := proxy.ToClashConfig(&config.ProxySetting{})
	assert.NoError(t, err)
	assert.NotNil(t, clashConfig)
	assert.Equal(t, "snell", clashConfig["type"])
	assert.Equal(t, "test-snell", clashConfig["name"])
	assert.Equal(t, "1.2.3.4", clashConfig["server"])
	assert.Equal(t, 8080, clashConfig["port"])
	assert.Equal(t, "psk", clashConfig["psk"])
	assert.Equal(t, 2, clashConfig["version"])

	obfsOpts, ok := clashConfig["obfs-opts"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "http", obfsOpts["mode"])
	assert.Equal(t, "example.com", obfsOpts["host"])
}

func TestSnellProxy_ToSurgeConfig(t *testing.T) {
	proxy := &SnellProxy{
		BaseProxy: core.BaseProxy{
			Remark: "test-snell",
			Server: "1.2.3.4",
			Port:   8080,
		},
		Psk:     "psk",
		Obfs:    "http",
		Version: 2,
	}

	surgeConfig, err := proxy.ToSurgeConfig(&config.ProxySetting{})
	assert.NoError(t, err)
	// test-snell = snell, 1.2.3.4, 8080, psk=psk, obfs=http, version=2
	assert.Contains(t, surgeConfig, "test-snell = snell, 1.2.3.4, 8080")
	assert.Contains(t, surgeConfig, "psk=psk")
	assert.Contains(t, surgeConfig, "obfs=http")
	assert.Contains(t, surgeConfig, "version=2")
}

func TestSnellProxy_ToClashConfig_GlobalOverrides(t *testing.T) {
	proxy := &SnellProxy{
		BaseProxy: core.BaseProxy{
			Remark: "test-snell",
		},
		Psk:     "psk",
		Obfs:    "http",
		Version: 2,
	}
	proxy.Server = "1.2.3.4"
	proxy.Port = 8080

	globalSettings := &config.ProxySetting{
		UDP:   utils.BoolPtr(true),
		TFO:   utils.BoolPtr(true),
		SCV:   utils.BoolPtr(true),
		TLS13: utils.BoolPtr(true), // Snell doesn't support TLS13
	}

	clashConfig, err := proxy.ToClashConfig(globalSettings)
	assert.NoError(t, err)
	assert.NotNil(t, clashConfig)

	assert.Equal(t, true, clashConfig["udp"])
	assert.Equal(t, true, clashConfig["tfo"])
	assert.Equal(t, true, clashConfig["skip-cert-verify"])
	assert.Nil(t, clashConfig["tls13"])
}
