package impl

import (
	"testing"

	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/proxy/core"
	"github.com/stretchr/testify/assert"
)

func TestVLESSProxy_ToSingleConfig(t *testing.T) {
	proxy := &VLESSProxy{
		BaseProxy: core.BaseProxy{
			Type:   "vless",
			Remark: "test-vless",
			Server: "1.2.3.4",
			Port:   443,
		},
		UUID:    "uuid",
		Network: "ws",
		Path:    "/path",
		Host:    "example.com",
		TLS:     true,
		SNI:     "example.com",
	}

	link, err := proxy.ToSingleConfig(&config.ProxySetting{})
	assert.NoError(t, err)
	// vless://uuid@1.2.3.4:443?host=example.com&path=%2Fpath&security=tls&sni=example.com&type=ws#test-vless
	assert.Contains(t, link, "vless://uuid@1.2.3.4:443")
	assert.Contains(t, link, "type=ws")
	assert.Contains(t, link, "security=tls")
	assert.Contains(t, link, "sni=example.com")
	assert.Contains(t, link, "path=%2Fpath")
	assert.Contains(t, link, "host=example.com")
	assert.Contains(t, link, "#test-vless")

	// Test No TLS
	proxyNoTLS := &VLESSProxy{
		BaseProxy: core.BaseProxy{
			Type:   "vless",
			Remark: "test-vless-notls",
			Server: "1.2.3.4",
			Port:   80,
		},
		UUID:    "uuid",
		Network: "tcp",
	}
	linkNoTLS, err := proxyNoTLS.ToSingleConfig(&config.ProxySetting{})
	assert.NoError(t, err)
	assert.Contains(t, linkNoTLS, "vless://uuid@1.2.3.4:80")
	assert.Contains(t, linkNoTLS, "type=tcp")
	assert.NotContains(t, linkNoTLS, "security=tls")

	// Test gRPC
	proxyGRPC := &VLESSProxy{
		BaseProxy: core.BaseProxy{
			Type:   "vless",
			Remark: "test-vless-grpc",
			Server: "1.2.3.4",
			Port:   443,
		},
		UUID:    "uuid",
		Network: "grpc",
		Path:    "serviceName",
		TLS:     true,
	}
	linkGRPC, err := proxyGRPC.ToSingleConfig(&config.ProxySetting{})
	assert.NoError(t, err)
	assert.Contains(t, linkGRPC, "type=grpc")
	// Verify that path is not included for gRPC as per current implementation.
	// Note: Standard VLESS links usually put serviceName in the query params, not path.
	assert.NotContains(t, linkGRPC, "path=")
}

func TestVLESSProxy_ToClashConfig_GlobalOverrides(t *testing.T) {
	proxy := &VLESSProxy{
		BaseProxy: core.BaseProxy{
			Type:   "vless",
			Remark: "test-vless",
			Server: "1.2.3.4",
			Port:   443,
		},
		UUID:    "uuid",
		Network: "ws",
		Path:    "/path",
		Host:    "example.com",
		TLS:     true,
		SNI:     "example.com",
	}

	// Test with global overrides
	udp := true
	tfo := true
	scv := true
	tls13 := true
	opts := &config.ProxySetting{
		UDP:   &udp,
		TFO:   &tfo,
		SCV:   &scv,
		TLS13: &tls13,
	}

	clashConfig, err := proxy.ToClashConfig(opts)
	assert.NoError(t, err)
	assert.NotNil(t, clashConfig)

	// Verify overrides
	assert.Equal(t, true, clashConfig["udp"])
	assert.Equal(t, true, clashConfig["tfo"])
	assert.Equal(t, true, clashConfig["skip-cert-verify"])
	assert.Equal(t, true, clashConfig["tls13"])
}

func TestVLESSProxy_ToClashConfig(t *testing.T) {
	proxy := &VLESSProxy{
		BaseProxy: core.BaseProxy{
			Type:   "vless",
			Remark: "test-vless",
			Server: "1.2.3.4",
			Port:   443,
		},
		UUID:    "uuid",
		Network: "ws",
		Path:    "/path",
		Host:    "example.com",
		TLS:     true,
		SNI:     "example.com",
	}

	clashConfig, err := proxy.ToClashConfig(&config.ProxySetting{})
	assert.NoError(t, err)
	assert.NotNil(t, clashConfig)
	assert.Equal(t, "vless", clashConfig["type"])
	assert.Equal(t, "test-vless", clashConfig["name"])
	assert.Equal(t, "1.2.3.4", clashConfig["server"])
	assert.Equal(t, 443, clashConfig["port"])
	assert.Equal(t, "uuid", clashConfig["uuid"])
	assert.Equal(t, "ws", clashConfig["network"])
	assert.Equal(t, true, clashConfig["tls"])
	assert.Equal(t, "example.com", clashConfig["servername"])

	wsOpts, ok := clashConfig["ws-opts"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "/path", wsOpts["path"])
	headers, ok := wsOpts["headers"].(map[string]string)
	assert.True(t, ok)
	assert.Equal(t, "example.com", headers["Host"])
}

func TestVLESSProxy_ToClashConfig_Full(t *testing.T) {
	proxy := &VLESSProxy{
		BaseProxy: core.BaseProxy{
			Type:   "vless",
			Remark: "test-vless-full",
			Server: "1.2.3.4",
			Port:   443,
		},
		UUID:            "uuid",
		Network:         "grpc",
		GRPCServiceName: "service",
		GRPCMode:        "multi",
		TLS:             true,
		SNI:             "sni.com",
		Alpn:            []string{"h2", "http/1.1"},
		Fingerprint:     "chrome",
		PublicKey:       "public-key",
		ShortID:         "short-id",
		Flow:            "xtls-rprx-vision",
		XTLS:            2,
	}

	clashConfig, err := proxy.ToClashConfig(&config.ProxySetting{})
	assert.NoError(t, err)
	assert.NotNil(t, clashConfig)

	assert.Equal(t, "vless", clashConfig["type"])
	assert.Equal(t, "test-vless-full", clashConfig["name"])
	assert.Equal(t, "1.2.3.4", clashConfig["server"])
	assert.Equal(t, 443, clashConfig["port"])
	assert.Equal(t, "uuid", clashConfig["uuid"])
	assert.Equal(t, "grpc", clashConfig["network"])
	assert.Equal(t, true, clashConfig["tls"])
	assert.Equal(t, "sni.com", clashConfig["servername"])
	assert.Equal(t, []string{"h2", "http/1.1"}, clashConfig["alpn"])
	assert.Equal(t, "chrome", clashConfig["client-fingerprint"])
	assert.Equal(t, "xtls-rprx-vision", clashConfig["flow"])

	grpcOpts, ok := clashConfig["grpc-opts"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "service", grpcOpts["grpc-service-name"])
	assert.Equal(t, "multi", grpcOpts["grpc-mode"])

	realityOpts, ok := clashConfig["reality-opts"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "public-key", realityOpts["public-key"])
	assert.Equal(t, "short-id", realityOpts["short-id"])
}

func TestVLESSProxy_ToSingleConfig_Reality(t *testing.T) {
	proxy := &VLESSProxy{
		BaseProxy: core.BaseProxy{
			Type:   "vless",
			Remark: "test-reality",
			Server: "1.2.3.4",
			Port:   443,
		},
		UUID:        "uuid",
		Network:     "tcp",
		TLS:         true,
		SNI:         "sni.com",
		Fingerprint: "chrome",
		PublicKey:   "public-key",
		ShortID:     "short-id",
		Flow:        "xtls-rprx-vision",
	}

	link, err := proxy.ToSingleConfig(nil)
	assert.NoError(t, err)
	assert.Contains(t, link, "vless://uuid@1.2.3.4:443")
	assert.Contains(t, link, "security=reality")
	assert.Contains(t, link, "pbk=public-key")
	assert.Contains(t, link, "sid=short-id")
	assert.Contains(t, link, "fp=chrome")
	assert.Contains(t, link, "sni=sni.com")
	assert.Contains(t, link, "flow=xtls-rprx-vision")
	assert.Contains(t, link, "#test-reality")
}

func TestVLESSProxy_ToClashConfig_MihomoParams(t *testing.T) {
	xudp := true
	packetAddr := true
	echEnable := true
	v2rayHttpUpgrade := true
	v2rayHttpUpgradeFastOpen := true

	proxy := &VLESSProxy{
		BaseProxy: core.BaseProxy{
			Type:   "vless",
			Remark: "test-vless-mihomo",
			Server: "1.2.3.4",
			Port:   443,
		},
		UUID:                     "uuid",
		Network:                  "ws",
		Path:                     "/path",
		Host:                     "example.com",
		TLS:                      true,
		SNI:                      "example.com",
		PacketEncoding:           "xudp",
		XUDP:                     &xudp,
		PacketAddr:               &packetAddr,
		IpVersion:                "ipv4",
		ClientFingerprint:        "chrome",
		EchEnable:                &echEnable,
		EchConfig:                "ech-config",
		Certificate:              "cert",
		PrivateKeyPem:            "key",
		VlessEncryption:          "none",
		WsMaxEarlyData:           2048,
		WsEarlyDataHeaderName:    "Sec-WebSocket-Protocol",
		V2rayHttpUpgrade:         &v2rayHttpUpgrade,
		V2rayHttpUpgradeFastOpen: &v2rayHttpUpgradeFastOpen,
		Flow:                     "xtls-rprx-vision",
		FlowSet:                  true,
	}

	config, err := proxy.ToClashConfig(nil)
	assert.NoError(t, err)

	assert.Equal(t, "xudp", config["packet-encoding"])
	assert.Equal(t, true, config["xudp"])
	assert.Equal(t, true, config["packet-addr"])
	assert.Equal(t, "ipv4", config["ip-version"])
	assert.Equal(t, "chrome", config["client-fingerprint"])

	echOpts, ok := config["ech-opts"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, true, echOpts["enable"])
	assert.Equal(t, "ech-config", echOpts["config"])

	assert.Equal(t, "cert", config["certificate"])
	assert.Equal(t, "key", config["private-key"])
	assert.Equal(t, "none", config["encryption"])
	assert.Equal(t, "xtls-rprx-vision", config["flow"])

	wsOpts, ok := config["ws-opts"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, 2048, wsOpts["max-early-data"])
	assert.Equal(t, "Sec-WebSocket-Protocol", wsOpts["early-data-header-name"])

	assert.Equal(t, true, config["v2ray-http-upgrade"])
	assert.Equal(t, true, config["v2ray-http-upgrade-fast-open"])
}
