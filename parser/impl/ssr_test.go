package impl

import (
	"encoding/base64"
	"testing"

	"github.com/gfunc/subconvergo/proxy/impl"
	"github.com/stretchr/testify/assert"
)

func TestShadowsocksRParser_Parse(t *testing.T) {
	parser := &ShadowsocksRParser{}

	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(*testing.T, *impl.ShadowsocksRProxy)
	}{
		{
			name: "SSR Real",
			// ssr://example.com:8388:auth_sha1_v4:aes-256-gcm:http_simple:cGFzc3dvcmQ
			input: "ssr://" + base64.URLEncoding.EncodeToString([]byte("example.com:8388:auth_sha1_v4:aes-256-gcm:http_simple:"+base64.URLEncoding.EncodeToString([]byte("password")))),
			check: func(t *testing.T, p *impl.ShadowsocksRProxy) {
				assert.Equal(t, "ssr", p.Type)
				assert.Equal(t, "example.com", p.Server)
				assert.Equal(t, 8388, p.Port)
				assert.Equal(t, "auth_sha1_v4", p.Protocol)
				assert.Equal(t, "aes-256-gcm", p.EncryptMethod)
				assert.Equal(t, "http_simple", p.Obfs)
				assert.Equal(t, "password", p.Password)
			},
		},
		{
			name: "SSR with Params",
			// ssr://example.com:8388:auth_sha1_v4:aes-256-gcm:http_simple:cGFzc3dvcmQ/?remarks=Example&group=R3JvdXA&obfsparam=b2Jmc19wYXJhbQ&protoparam=cHJvdG9fcGFyYW0
			input: "ssr://" + base64.URLEncoding.EncodeToString([]byte("example.com:8388:auth_sha1_v4:aes-256-gcm:http_simple:"+base64.URLEncoding.EncodeToString([]byte("password"))+"/?remarks="+base64.URLEncoding.EncodeToString([]byte("Example"))+"&group="+base64.URLEncoding.EncodeToString([]byte("Group"))+"&obfsparam="+base64.URLEncoding.EncodeToString([]byte("obfs_param"))+"&protoparam="+base64.URLEncoding.EncodeToString([]byte("proto_param")))),
			check: func(t *testing.T, p *impl.ShadowsocksRProxy) {
				assert.Equal(t, "ssr", p.Type)
				assert.Equal(t, "Example", p.Remark)
				assert.Equal(t, "Group", p.Group)
				assert.Equal(t, "obfs_param", p.ObfsParam)
				assert.Equal(t, "proto_param", p.ProtocolParam)
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

			var ssrProxy *impl.ShadowsocksRProxy
			if mp, ok := proxy.(*impl.MihomoProxy); ok {
				ssrProxy = mp.ProxyInterface.(*impl.ShadowsocksRProxy)
			} else {
				ssrProxy = proxy.(*impl.ShadowsocksRProxy)
			}

			tt.check(t, ssrProxy)
		})
	}
}
