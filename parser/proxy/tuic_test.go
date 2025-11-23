package proxy

import (
	"testing"

	"github.com/gfunc/subconvergo/proxy/impl"
	"github.com/stretchr/testify/assert"
)

func TestTUICParser_Parse(t *testing.T) {
	parser := &TUICParser{}

	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(*testing.T, *impl.TUICProxy)
	}{
		{
			name:  "TUIC Standard",
			input: "tuic://uuid:password@example.com:443?allow_insecure=1#Example",
			check: func(t *testing.T, p *impl.TUICProxy) {
				assert.Equal(t, "tuic", p.Type)
				assert.Equal(t, "example.com", p.Server)
				assert.Equal(t, 443, p.Port)
				assert.Equal(t, "uuid", p.UUID)
				assert.Equal(t, "password", p.Password)
				assert.Equal(t, true, p.AllowInsecure)
				assert.Equal(t, "Example", p.Remark)
			},
		},
		{
			name:  "TUIC No Password",
			input: "tuic://uuid@example.com:443",
			check: func(t *testing.T, p *impl.TUICProxy) {
				assert.Equal(t, "tuic", p.Type)
				assert.Equal(t, "uuid", p.UUID)
				assert.Equal(t, "", p.Password)
			},
		},
		{
			name:  "TUIC with Params",
			input: "tuic://uuid:password@example.com:443?congestion_control=bbr&udp_relay_mode=native",
			check: func(t *testing.T, p *impl.TUICProxy) {
				assert.Equal(t, "tuic", p.Type)
				assert.Equal(t, "bbr", p.Params.Get("congestion_control"))
				assert.Equal(t, "native", p.Params.Get("udp_relay_mode"))
			},
		},
		{
			name:    "Invalid Port",
			input:   "tuic://uuid:password@example.com:invalid",
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

			var tProxy *impl.TUICProxy
			if mp, ok := proxy.(*impl.MihomoProxy); ok {
				tProxy = mp.ProxyInterface.(*impl.TUICProxy)
			} else {
				tProxy = proxy.(*impl.TUICProxy)
			}

			tt.check(t, tProxy)
		})
	}
}
