package impl

import (
	"net/url"

	pc "github.com/gfunc/subconvergo/proxy/core"
	"github.com/gfunc/subconvergo/proxy/impl"
)

func getTestProxies() []pc.ProxyInterface {
	return []pc.ProxyInterface{
		&impl.ShadowsocksProxy{
			BaseProxy: pc.BaseProxy{
				Type:   "ss",
				Remark: "ss-proxy",
				Server: "1.2.3.4",
				Port:   8388,
			},
			Password:      "password",
			EncryptMethod: "aes-256-gcm",
		},
		&impl.ShadowsocksRProxy{
			BaseProxy: pc.BaseProxy{
				Type:   "ssr",
				Remark: "ssr-proxy",
				Server: "1.2.3.4",
				Port:   8388,
			},
			Password:      "password",
			EncryptMethod: "aes-256-gcm",
			Protocol:      "auth_aes128_md5",
			Obfs:          "tls1.2_ticket_auth",
		},
		&impl.VMessProxy{
			BaseProxy: pc.BaseProxy{
				Type:   "vmess",
				Remark: "vmess-proxy",
				Server: "5.6.7.8",
				Port:   443,
			},
			UUID:    "uuid",
			AlterID: 64,
			Network: "ws",
			Path:    "/path",
			Host:    "example.com",
			TLS:     true,
			SNI:     "example.com",
		},
		&impl.VLESSProxy{
			BaseProxy: pc.BaseProxy{
				Type:   "vless",
				Remark: "vless-proxy",
				Server: "9.10.11.12",
				Port:   443,
			},
			UUID:    "uuid",
			Network: "ws",
			Path:    "/path",
			TLS:     true,
			SNI:     "example.com",
			Flow:    "xtls-rprx-vision",
		},
		&impl.TrojanProxy{
			BaseProxy: pc.BaseProxy{
				Type:   "trojan",
				Remark: "trojan-proxy",
				Server: "13.14.15.16",
				Port:   443,
			},
			Password: "password",
			Network:  "ws",
			Path:     "/path",
			Host:     "example.com",
		},
		&impl.HysteriaProxy{
			BaseProxy: pc.BaseProxy{
				Type:   "hysteria2",
				Remark: "hysteria2-proxy",
				Server: "17.18.19.20",
				Port:   443,
			},
			Password: "password",
			Obfs:     "salamander",
			Params:   url.Values{"obfs-password": []string{"secret"}},
		},
		&impl.TUICProxy{
			BaseProxy: pc.BaseProxy{
				Type:   "tuic",
				Remark: "tuic-proxy",
				Server: "21.22.23.24",
				Port:   443,
			},
			UUID:     "uuid",
			Password: "password",
			Params:   url.Values{"congestion_control": []string{"bbr"}},
		},
		&impl.AnyTLSProxy{
			BaseProxy: pc.BaseProxy{
				Type:   "anytls",
				Remark: "anytls-proxy",
				Server: "25.26.27.28",
				Port:   443,
			},
			Password:                 "password",
			SNI:                      "example.com",
			Alpn:                     []string{"h2", "http/1.1"},
			Fingerprint:              "chrome",
			IdleSessionCheckInterval: 30,
			IdleSessionTimeout:       60,
			MinIdleSession:           5,
		},
	}
}
