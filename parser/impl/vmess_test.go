package impl

import (
	"testing"

	"github.com/gfunc/subconvergo/proxy/impl"
	"github.com/stretchr/testify/assert"
)

func TestVMessParser_Parse(t *testing.T) {
	parser := &VMessParser{}

	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(*testing.T, *impl.VMessProxy)
	}{
		{
			name:  "VMess JSON",
			input: "vmess://ew0KICAidiI6ICIyIiwNCiAgInBzIjogInJlbWFyayIsDQogICJhZGQiOiAiMTI3LjAuMC4xIiwNCiAgInBvcnQiOiAiODA4MCIsDQogICJpZCI6ICJ1dWlkIiwNCiAgImFpZCI6ICIwIiwNCiAgInNjeSI6ICJhdXRvIiwNCiAgIm5ldCI6ICJ3cyIsDQogICJ0eXBlIjogIm5vbmUiLA0KICAiaG9zdCI6ICJleGFtcGxlLmNvbSIsDQogICJwYXRoIjogIi8iLA0KICAidGxzIjogInRscyIsDQogICJzbmkiOiAiIg0KfQ==",
			check: func(t *testing.T, p *impl.VMessProxy) {
				assert.Equal(t, "vmess", p.Type)
				assert.Equal(t, "remark", p.Remark)
				assert.Equal(t, "127.0.0.1", p.Server)
				assert.Equal(t, 8080, p.Port)
				assert.Equal(t, "uuid", p.UUID)
				assert.Equal(t, "ws", p.Network)
				assert.Equal(t, true, p.TLS)
			},
		},
		{
			name:  "VMess Standard",
			input: "vmess://uuid@127.0.0.1:8080?remark=test&network=ws&tls=1",
			check: func(t *testing.T, p *impl.VMessProxy) {
				assert.Equal(t, "vmess", p.Type)
				assert.Equal(t, "test", p.Remark)
				assert.Equal(t, "127.0.0.1", p.Server)
				assert.Equal(t, 8080, p.Port)
				assert.Equal(t, "uuid", p.UUID)
				assert.Equal(t, "ws", p.Network)
				assert.Equal(t, true, p.TLS)
			},
		},
		{
			name:    "Invalid Base64",
			input:   "vmess://invalid-base64",
			wantErr: true,
		},
		{
			name:    "Invalid JSON",
			input:   "vmess://e30=", // {}
			wantErr: true,           // Should fail validation if fields are missing
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

			var vProxy *impl.VMessProxy
			if mp, ok := proxy.(*impl.MihomoProxy); ok {
				vProxy = mp.ProxyInterface.(*impl.VMessProxy)
			} else {
				vProxy = proxy.(*impl.VMessProxy)
			}

			tt.check(t, vProxy)
		})
	}
}
