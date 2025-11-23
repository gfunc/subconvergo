package proxy

import (
	"testing"

	"github.com/gfunc/subconvergo/proxy/impl"
	"github.com/stretchr/testify/assert"
)

func TestTrojanParser_Parse(t *testing.T) {
	parser := &TrojanParser{}

	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(*testing.T, *impl.TrojanProxy)
	}{
		{
			name:  "Trojan Standard",
			input: "trojan://password@example.com:443?sni=example.com&allowInsecure=1#Example",
			check: func(t *testing.T, p *impl.TrojanProxy) {
				assert.Equal(t, "trojan", p.Type)
				assert.Equal(t, "example.com", p.Server)
				assert.Equal(t, 443, p.Port)
				assert.Equal(t, "password", p.Password)
				assert.Equal(t, "example.com", p.Host)
				assert.Equal(t, true, p.AllowInsecure)
				assert.Equal(t, "Example", p.Remark)
			},
		},
		{
			name:  "Trojan WS",
			input: "trojan://password@example.com:443?type=ws&path=/ws&sni=example.com",
			check: func(t *testing.T, p *impl.TrojanProxy) {
				assert.Equal(t, "trojan", p.Type)
				assert.Equal(t, "ws", p.Network)
				assert.Equal(t, "/ws", p.Path)
			},
		},
		{
			name:  "Trojan gRPC",
			input: "trojan://password@example.com:443?type=grpc&serviceName=grpc-service",
			check: func(t *testing.T, p *impl.TrojanProxy) {
				assert.Equal(t, "trojan", p.Type)
				assert.Equal(t, "grpc", p.Network)
				assert.Equal(t, "grpc-service", p.Path) // Path maps to serviceName for grpc
			},
		},
		{
			name:    "Invalid URL",
			input:   "trojan://invalid-url",
			wantErr: true,
		},
		{
			name:    "Missing Password",
			input:   "trojan://@example.com:443",
			wantErr: false,
			check: func(t *testing.T, p *impl.TrojanProxy) {
				assert.Equal(t, "trojan", p.Type)
				assert.Equal(t, "", p.Password)
			},
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

			var tProxy *impl.TrojanProxy
			if mp, ok := proxy.(*impl.MihomoProxy); ok {
				tProxy = mp.ProxyInterface.(*impl.TrojanProxy)
			} else {
				tProxy = proxy.(*impl.TrojanProxy)
			}

			tt.check(t, tProxy)
		})
	}
}
