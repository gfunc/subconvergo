package proxy

import (
	"testing"

	"github.com/gfunc/subconvergo/proxy/impl"
	"github.com/stretchr/testify/assert"
)

func TestHysteria2Parser_Parse(t *testing.T) {
	parser := &Hysteria2Parser{}

	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(*testing.T, *impl.Hysteria2Proxy)
	}{
		{
			name:  "Hysteria2 Standard",
			input: "hysteria2://password@example.com:443?insecure=1&sni=example.com#Example",
			check: func(t *testing.T, p *impl.Hysteria2Proxy) {
				assert.Equal(t, "hysteria2", p.Type)
				assert.Equal(t, "example.com", p.Server)
				assert.Equal(t, 443, p.Port)
				assert.Equal(t, "password", p.Password)
				assert.Equal(t, true, p.SkipCertVerify)
				assert.Equal(t, "example.com", p.Sni)
				assert.Equal(t, "Example", p.Remark)
			},
		},
		{
			name:  "Hysteria2 with Obfs",
			input: "hysteria2://password@example.com:443?obfs=salamander&obfs-password=secret",
			check: func(t *testing.T, p *impl.Hysteria2Proxy) {
				assert.Equal(t, "hysteria2", p.Type)
				assert.Equal(t, "salamander", p.Obfs)
				assert.Equal(t, "secret", p.ObfsPassword)
			},
		},
		{
			name:    "Invalid Format",
			input:   "hysteria2://invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proxy, err := parser.ParseSingle(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			hProxy := proxy.(*impl.Hysteria2Proxy)
			tt.check(t, hProxy)
		})
	}
}
