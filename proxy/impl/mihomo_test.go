package impl

import (
	"testing"

	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/proxy/core"
	"github.com/stretchr/testify/assert"
)

func TestMihomoProxy_ToShareLink(t *testing.T) {
	ssProxy := &ShadowsocksProxy{
		BaseProxy: core.BaseProxy{
			Type:   "ss",
			Remark: "ss-proxy",
			Server: "1.2.3.4",
			Port:   8388,
		},
		Password:      "password",
		EncryptMethod: "aes-256-gcm",
	}

	mihomoProxy := &MihomoProxy{
		ProxyInterface: ssProxy,
	}

	link, err := mihomoProxy.ToShareLink(&config.ProxySetting{})
	assert.NoError(t, err)
	assert.Contains(t, link, "ss://")
	assert.Contains(t, link, "#ss-proxy")
}
