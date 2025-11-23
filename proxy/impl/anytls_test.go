package impl

import (
	"testing"

	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/proxy/core"
	"github.com/stretchr/testify/assert"
)

func TestAnyTLSProxy_ToSingleConfig(t *testing.T) {
	proxy := &AnyTLSProxy{
		BaseProxy: core.BaseProxy{
			Type:   "anytls",
			Remark: "anytls-proxy",
			Server: "1.2.3.4",
			Port:   443,
		},
		Password:                 "password",
		SNI:                      "example.com",
		Alpn:                     []string{"h2", "http/1.1"},
		Fingerprint:              "chrome",
		IdleSessionCheckInterval: 30,
		IdleSessionTimeout:       60,
		MinIdleSession:           5,
		TFO:                      true,
		AllowInsecure:            true,
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
	proxy := &AnyTLSProxy{
		BaseProxy: core.BaseProxy{
			Type:   "anytls",
			Remark: "anytls-proxy",
			Server: "1.2.3.4",
			Port:   443,
		},
		Password:                 "password",
		SNI:                      "example.com",
		Alpn:                     []string{"h2", "http/1.1"},
		Fingerprint:              "chrome",
		IdleSessionCheckInterval: 30,
		IdleSessionTimeout:       60,
		MinIdleSession:           5,
		TFO:                      true,
		AllowInsecure:            true,
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
