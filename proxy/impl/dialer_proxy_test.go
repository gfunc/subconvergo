package impl

import (
	"testing"

	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/proxy/core"
	"github.com/stretchr/testify/assert"
)

func TestDialerProxy_ToClashConfig(t *testing.T) {
	tests := []struct {
		name     string
		proxy    core.ProxyInterface
		expected string
	}{
		{
			name: "Shadowsocks with dialer-proxy",
			proxy: &ShadowsocksProxy{
				BaseProxy: core.BaseProxy{
					Type:            "ss",
					Remark:          "ss-dialer",
					Server:          "1.2.3.4",
					Port:            8388,
					UnderlyingProxy: "relay-proxy",
				},
				Password:      "password",
				EncryptMethod: "aes-256-gcm",
			},
			expected: "relay-proxy",
		},
		{
			name: "VMess with dialer-proxy",
			proxy: &VMessProxy{
				BaseProxy: core.BaseProxy{
					Type:            "vmess",
					Remark:          "vmess-dialer",
					Server:          "1.2.3.4",
					Port:            443,
					UnderlyingProxy: "relay-proxy",
				},
				UUID:    "uuid",
				AlterID: 0,
				Cipher:  "auto",
			},
			expected: "relay-proxy",
		},
		{
			name: "Trojan with dialer-proxy",
			proxy: &TrojanProxy{
				BaseProxy: core.BaseProxy{
					Type:            "trojan",
					Remark:          "trojan-dialer",
					Server:          "1.2.3.4",
					Port:            443,
					UnderlyingProxy: "relay-proxy",
				},
				Password: "password",
			},
			expected: "relay-proxy",
		},
		{
			name: "VLESS with dialer-proxy",
			proxy: &VLESSProxy{
				BaseProxy: core.BaseProxy{
					Type:            "vless",
					Remark:          "vless-dialer",
					Server:          "1.2.3.4",
					Port:            443,
					UnderlyingProxy: "relay-proxy",
				},
				UUID: "uuid",
			},
			expected: "relay-proxy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if p, ok := tt.proxy.(core.ClashConvertableMixin); ok {
				clashConfig, err := p.ToClashConfig(&config.ProxySetting{})
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, clashConfig["dialer-proxy"])
			} else {
				t.Fatalf("Proxy does not implement ClashConvertableMixin")
			}
		})
	}
}
