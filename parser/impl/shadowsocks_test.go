package impl

import (
	"encoding/base64"
	"testing"

	"github.com/gfunc/subconvergo/proxy/impl"
	"github.com/stretchr/testify/assert"
)

func TestShadowsocksParser_Parse(t *testing.T) {
	parser := &ShadowsocksParser{}

	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(*testing.T, *impl.ShadowsocksProxy)
	}{
		{
			name:  "SS SIP002",
			input: "ss://" + base64.URLEncoding.EncodeToString([]byte("aes-256-gcm:password")) + "@example.com:8388#Example",
			check: func(t *testing.T, p *impl.ShadowsocksProxy) {
				assert.Equal(t, "ss", p.Type)
				assert.Equal(t, "example.com", p.Server)
				assert.Equal(t, 8388, p.Port)
				assert.Equal(t, "aes-256-gcm", p.EncryptMethod)
				assert.Equal(t, "password", p.Password)
				assert.Equal(t, "Example", p.Remark)
			},
		},
		{
			name:  "SS Legacy",
			input: "ss://" + base64.URLEncoding.EncodeToString([]byte("aes-256-gcm:password@example.com:8388")) + "#Example",
			check: func(t *testing.T, p *impl.ShadowsocksProxy) {
				assert.Equal(t, "ss", p.Type)
				assert.Equal(t, "example.com", p.Server)
				assert.Equal(t, 8388, p.Port)
				assert.Equal(t, "aes-256-gcm", p.EncryptMethod)
				assert.Equal(t, "password", p.Password)
				assert.Equal(t, "Example", p.Remark)
			},
		},
		{
			name:  "SS with Plugin",
			input: "ss://" + base64.URLEncoding.EncodeToString([]byte("aes-256-gcm:password")) + "@example.com:8388?plugin=obfs-local%3Bobfs%3Dhttp%3Bobfs-host%3Dexample.com",
			check: func(t *testing.T, p *impl.ShadowsocksProxy) {
				assert.Equal(t, "ss", p.Type)
				assert.Equal(t, "obfs-local", p.Plugin)
				assert.Equal(t, "http", p.PluginOpts["obfs"])
				assert.Equal(t, "example.com", p.PluginOpts["obfs-host"])
			},
		},
		{
			name:    "Invalid Base64",
			input:   "ss://invalid-base64@example.com:8388",
			wantErr: true,
		},
		{
			name:    "Invalid Scheme",
			input:   "vmess://example.com:8388",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proxy, err := parser.Parse(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			var sProxy *impl.ShadowsocksProxy
			if mp, ok := proxy.(*impl.MihomoProxy); ok {
				sProxy = mp.ProxyInterface.(*impl.ShadowsocksProxy)
			} else {
				sProxy = proxy.(*impl.ShadowsocksProxy)
			}

			tt.check(t, sProxy)
		})
	}
}
