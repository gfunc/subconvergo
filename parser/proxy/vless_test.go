package proxy

import (
	"testing"

	"github.com/gfunc/subconvergo/proxy/impl"
	"github.com/stretchr/testify/assert"
)

func TestVLESSParser_Parse(t *testing.T) {
	parser := &VLESSParser{}

	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(*testing.T, *impl.VLESSProxy)
	}{
		{
			name:  "VLESS Standard",
			input: "vless://uuid@example.com:443?security=tls&type=ws&path=/ws&sni=example.com#Example",
			check: func(t *testing.T, p *impl.VLESSProxy) {
				assert.Equal(t, "vless", p.Type)
				assert.Equal(t, "example.com", p.Server)
				assert.Equal(t, 443, p.Port)
				assert.Equal(t, "uuid", p.UUID)
				assert.Equal(t, true, p.TLS)
				assert.Equal(t, "ws", p.Network)
				assert.Equal(t, "/ws", p.Path)
				assert.Equal(t, "example.com", p.SNI)
				assert.Equal(t, "Example", p.Remark)
			},
		},
		{
			name:  "VLESS gRPC",
			input: "vless://uuid@example.com:443?security=reality&type=grpc&serviceName=grpc-service&mode=gun&sni=example.com&pbk=publickey&sid=shortid&fp=chrome",
			check: func(t *testing.T, p *impl.VLESSProxy) {
				assert.Equal(t, "vless", p.Type)
				assert.Equal(t, true, p.TLS)
				assert.Equal(t, "grpc", p.Network)
				assert.Equal(t, "grpc-service", p.GRPCServiceName)
				assert.Equal(t, "gun", p.GRPCMode)
				assert.Equal(t, "example.com", p.SNI)
				assert.Equal(t, "publickey", p.PublicKey)
				assert.Equal(t, "shortid", p.ShortID)
				assert.Equal(t, "chrome", p.Fingerprint)
			},
		},
		{
			name:  "VLESS XTLS-Reality",
			input: "vless://uuid@example.com:443?security=reality&flow=xtls-rprx-vision&sni=example.com&pbk=publickey&sid=shortid&fp=random",
			check: func(t *testing.T, p *impl.VLESSProxy) {
				assert.Equal(t, "vless", p.Type)
				assert.Equal(t, true, p.TLS)
				assert.Equal(t, "xtls-rprx-vision", p.Flow)
				assert.Equal(t, "example.com", p.SNI)
				assert.Equal(t, "publickey", p.PublicKey)
				assert.Equal(t, "shortid", p.ShortID)
				assert.Equal(t, "random", p.Fingerprint)
			},
		},
		{
			name:  "VLESS Reality Repro",
			input: "vless://8ee113c3-e5dd-4cda-bec6-14df6a12f654@8.8.8.8:443?encryption=none&security=reality&flow=xtls-rprx-vision&type=tcp&sni=aws.amazon.com&pbk=testpublickey&fp=chrome#test1234",
			check: func(t *testing.T, p *impl.VLESSProxy) {
				assert.Equal(t, "vless", p.Type)
				assert.Equal(t, "8.8.8.8", p.Server)
				assert.Equal(t, 443, p.Port)
				assert.Equal(t, "8ee113c3-e5dd-4cda-bec6-14df6a12f654", p.UUID)
				assert.Equal(t, true, p.TLS)
				assert.Equal(t, "tcp", p.Network)
				assert.Equal(t, "xtls-rprx-vision", p.Flow)
				assert.Equal(t, "aws.amazon.com", p.SNI)
				assert.Equal(t, "testpublickey", p.PublicKey)
				assert.Equal(t, "chrome", p.Fingerprint)
				assert.Equal(t, "test1234", p.Remark)
			},
		},
		{
			name:  "VLESS No TLS",
			input: "vless://uuid@example.com:80?security=none&type=tcp",
			check: func(t *testing.T, p *impl.VLESSProxy) {
				assert.Equal(t, "vless", p.Type)
				assert.Equal(t, false, p.TLS)
				assert.Equal(t, "tcp", p.Network)
			},
		},
		{
			name:    "Missing UUID",
			input:   "vless://@example.com:443",
			wantErr: false,
			check: func(t *testing.T, p *impl.VLESSProxy) {
				assert.Equal(t, "vless", p.Type)
				assert.Equal(t, "", p.UUID)
			},
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

			var vProxy *impl.VLESSProxy
			if mp, ok := proxy.(*impl.MihomoProxy); ok {
				vProxy = mp.ProxyInterface.(*impl.VLESSProxy)
			} else {
				vProxy = proxy.(*impl.VLESSProxy)
			}

			tt.check(t, vProxy)
		})
	}
}
