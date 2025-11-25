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
