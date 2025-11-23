package proxy

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
			name:    "Invalid Format",
			input:   "hysteria://invalid",
			wantErr: true,
		},
		{
			name:    "Invalid Port",
			input:   "hysteria://example.com:invalid",
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
