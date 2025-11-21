package impl

import (
	"testing"

	"github.com/gfunc/subconvergo/proxy/impl"
	"github.com/stretchr/testify/assert"
)

func TestHysteriaParser_Parse(t *testing.T) {
	parser := &HysteriaParser{}

	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(*testing.T, *impl.HysteriaProxy)
	}{
		{
			name:  "Hysteria2 Standard",
			input: "hysteria2://password@example.com:443?insecure=1&sni=example.com#Example",
			check: func(t *testing.T, p *impl.HysteriaProxy) {
				assert.Equal(t, "hysteria2", p.Type)
				assert.Equal(t, "example.com", p.Server)
				assert.Equal(t, 443, p.Port)
				assert.Equal(t, "password", p.Password)
				assert.Equal(t, true, p.AllowInsecure)
				assert.Equal(t, "example.com", p.Params.Get("sni"))
				assert.Equal(t, "Example", p.Remark)
			},
		},
		{
			name:  "Hysteria2 Password in Params",
			input: "hysteria2://example.com:443?password=secret&insecure=0",
			check: func(t *testing.T, p *impl.HysteriaProxy) {
				assert.Equal(t, "hysteria2", p.Type)
				assert.Equal(t, "example.com", p.Server)
				assert.Equal(t, 443, p.Port)
				assert.Equal(t, "secret", p.Password)
				assert.Equal(t, false, p.AllowInsecure)
			},
		},
		{
			name:  "Hysteria2 with Up/Down",
			input: "hysteria2://password@example.com:443?up=100&down=200",
			check: func(t *testing.T, p *impl.HysteriaProxy) {
				assert.Equal(t, "hysteria2", p.Type)
				assert.Equal(t, "100", p.Params.Get("up"))
				assert.Equal(t, "200", p.Params.Get("down"))
			},
		},
		{
			name:  "Hysteria1",
			input: "hysteria://example.com:443?auth=secret&upmbps=50&downmbps=100&obfs=xplus",
			check: func(t *testing.T, p *impl.HysteriaProxy) {
				assert.Equal(t, "hysteria", p.Type)
				assert.Equal(t, "secret", p.Password)
				assert.Equal(t, "50", p.Params.Get("upmbps"))
				assert.Equal(t, "100", p.Params.Get("downmbps"))
				assert.Equal(t, "xplus", p.Obfs)
			},
		},
		{
			name:  "Hysteria2 with Obfs",
			input: "hysteria2://password@example.com:443?obfs=salamander&obfs-password=secret",
			check: func(t *testing.T, p *impl.HysteriaProxy) {
				assert.Equal(t, "hysteria2", p.Type)
				assert.Equal(t, "salamander", p.Obfs)
				assert.Equal(t, "secret", p.Params.Get("obfs-password"))
			},
		},
		{
			name:    "Invalid Format",
			input:   "hysteria2://invalid",
			wantErr: true,
		},
		{
			name:    "Invalid Port",
			input:   "hysteria2://example.com:invalid",
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

			var hProxy *impl.HysteriaProxy
			if mp, ok := proxy.(*impl.MihomoProxy); ok {
				hProxy = mp.ProxyInterface.(*impl.HysteriaProxy)
			} else {
				hProxy = proxy.(*impl.HysteriaProxy)
			}

			tt.check(t, hProxy)
		})
	}
}
