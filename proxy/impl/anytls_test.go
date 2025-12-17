package impl

import (
	"testing"

	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/proxy/core"
	"github.com/gfunc/subconvergo/utils"
	"github.com/stretchr/testify/assert"
)

func TestAnyTLSProxy_ToSingleConfig(t *testing.T) {
	tfo := true
	scv := true
	proxy := &AnyTLSProxy{
		BaseProxy: core.BaseProxy{
			Type:   "anytls",
			Remark: "anytls-proxy",
			Server: "1.2.3.4",
			Port:   443,
			TFO:    &tfo,
			SCV:    &scv,
		},
		Password:                 "password",
		SNI:                      "example.com",
		Alpn:                     []string{"h2", "http/1.1"},
		Fingerprint:              "chrome",
		IdleSessionCheckInterval: 30,
		IdleSessionTimeout:       60,
		MinIdleSession:           5,
	}

	link, err := proxy.ToSingleConfig(&config.ProxySetting{})
	assert.NoError(t, err)

	assert.Contains(t, link, "anytls://password@1.2.3.4:443")
	assert.Contains(t, link, "peer=example.com")
	assert.Contains(t, link, "alpn=h2%2Chttp%2F1.1")
	assert.Contains(t, link, "hpkp=chrome")
	assert.Contains(t, link, "tfo=1")
	assert.Contains(t, link, "insecure=1")
	assert.Contains(t, link, "idle_session_check_interval=30")
	assert.Contains(t, link, "idle_session_timeout=60")
	assert.Contains(t, link, "min_idle_session=5")
	assert.Contains(t, link, "#anytls-proxy")
}

func TestAnyTLSProxy_ToClashConfig(t *testing.T) {
	tfo := true
	scv := true
	proxy := &AnyTLSProxy{
		BaseProxy: core.BaseProxy{
			Type:   "anytls",
			Remark: "anytls-proxy",
			Server: "1.2.3.4",
			Port:   443,
			TFO:    &tfo,
			SCV:    &scv,
		},
		Password:                 "password",
		SNI:                      "example.com",
		Alpn:                     []string{"h2", "http/1.1"},
		Fingerprint:              "chrome",
		IdleSessionCheckInterval: 30,
		IdleSessionTimeout:       60,
		MinIdleSession:           5,
	}

	config, err := proxy.ToClashConfig(&config.ProxySetting{})
	assert.NoError(t, err)

	assert.Equal(t, "anytls", config["type"])
	assert.Equal(t, "anytls-proxy", config["name"])
	assert.Equal(t, "1.2.3.4", config["server"])
	assert.Equal(t, 443, config["port"])
	assert.Equal(t, "password", config["password"])
	assert.Equal(t, "example.com", config["sni"])
	assert.Equal(t, []string{"h2", "http/1.1"}, config["alpn"])
	assert.Equal(t, "chrome", config["fingerprint"])
	assert.Equal(t, 30, config["idle-session-check-interval"])
	assert.Equal(t, 60, config["idle-session-timeout"])
	assert.Equal(t, 5, config["min-idle-session"])
	assert.Equal(t, true, config["skip-cert-verify"])
	assert.Equal(t, true, config["tfo"])
}

func TestAnyTLSProxy_ToClashConfig_GlobalOverrides(t *testing.T) {
	proxy := &AnyTLSProxy{
		BaseProxy: core.BaseProxy{
			Remark: "test-anytls",
		},
		Password: "password",
	}
	proxy.Server = "1.2.3.4"
	proxy.Port = 443

	globalSettings := &config.ProxySetting{
		UDP:   utils.BoolPtr(true), // Should be ignored
		TFO:   utils.BoolPtr(true),
		SCV:   utils.BoolPtr(true),
		TLS13: utils.BoolPtr(true), // Should be ignored
	}

	clashConfig, err := proxy.ToClashConfig(globalSettings)
	assert.NoError(t, err)
	assert.NotNil(t, clashConfig)

	assert.Nil(t, clashConfig["udp"])
	assert.Equal(t, true, clashConfig["tfo"])
	assert.Equal(t, true, clashConfig["skip-cert-verify"])
	assert.Nil(t, clashConfig["tls13"])
}
