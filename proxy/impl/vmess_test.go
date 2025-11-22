package impl

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"

	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/proxy/core"
	"github.com/stretchr/testify/assert"
)

func TestVMessProxy_ToShareLink(t *testing.T) {
	proxy := &VMessProxy{
		BaseProxy: core.BaseProxy{
			Type:   "vmess",
			Remark: "test-vmess",
			Server: "1.2.3.4",
			Port:   443,
		},
		UUID:    "uuid",
		AlterID: 64,
		Network: "ws",
		Path:    "/path",
		Host:    "example.com",
		TLS:     true,
	}

	link, err := proxy.ToShareLink(&config.ProxySetting{})
	assert.NoError(t, err)
	assert.True(t, strings.HasPrefix(link, "vmess://"))

	decodedBytes, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(link, "vmess://"))
	assert.NoError(t, err)

	var data map[string]interface{}
	err = json.Unmarshal(decodedBytes, &data)
	assert.NoError(t, err)

	assert.Equal(t, "2", data["v"])
	assert.Equal(t, "test-vmess", data["ps"])
	assert.Equal(t, "1.2.3.4", data["add"])
	assert.Equal(t, "443", data["port"])
	assert.Equal(t, "uuid", data["id"])
	assert.Equal(t, "64", data["aid"])
	assert.Equal(t, "ws", data["net"])
	assert.Equal(t, "example.com", data["host"])
	assert.Equal(t, "/path", data["path"])
	assert.Equal(t, "tls", data["tls"])

	// Test No TLS and TCP
	proxyNoTLS := &VMessProxy{
		BaseProxy: core.BaseProxy{
			Type:   "vmess",
			Remark: "test-vmess-notls",
			Server: "1.2.3.4",
			Port:   80,
		},
		UUID:    "uuid",
		AlterID: 0,
		Network: "tcp",
	}
	linkNoTLS, err := proxyNoTLS.ToShareLink(&config.ProxySetting{})
	assert.NoError(t, err)
	decodedNoTLSBytes, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(linkNoTLS, "vmess://"))
	assert.NoError(t, err)
	var dataNoTLS map[string]interface{}
	err = json.Unmarshal(decodedNoTLSBytes, &dataNoTLS)
	assert.NoError(t, err)
	assert.Equal(t, "", dataNoTLS["tls"])
	assert.Equal(t, "tcp", dataNoTLS["net"])
	assert.Equal(t, "0", dataNoTLS["aid"])

	// Test gRPC
	proxyGRPC := &VMessProxy{
		BaseProxy: core.BaseProxy{
			Type:   "vmess",
			Remark: "test-vmess-grpc",
			Server: "1.2.3.4",
			Port:   443,
		},
		UUID:    "uuid",
		Network: "grpc",
		Path:    "serviceName",
		TLS:     true,
	}
	linkGRPC, err := proxyGRPC.ToShareLink(&config.ProxySetting{})
	assert.NoError(t, err)
	decodedGRPCBytes, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(linkGRPC, "vmess://"))
	assert.NoError(t, err)
	var dataGRPC map[string]interface{}
	err = json.Unmarshal(decodedGRPCBytes, &dataGRPC)
	assert.NoError(t, err)
	assert.Equal(t, "grpc", dataGRPC["net"])
	assert.Equal(t, "serviceName", dataGRPC["path"])
}

func TestVMessProxy_ToClashConfig(t *testing.T) {
	proxy := &VMessProxy{
		BaseProxy: core.BaseProxy{
			Type:   "vmess",
			Remark: "test-vmess",
			Server: "1.2.3.4",
			Port:   443,
		},
		UUID:    "uuid",
		AlterID: 64,
		Network: "ws",
		Path:    "/path",
		Host:    "example.com",
		TLS:     true,
		SNI:     "example.com",
	}

	clashConfig := proxy.ToClashConfig(&config.ProxySetting{})
	assert.NotNil(t, clashConfig)
	assert.Equal(t, "vmess", clashConfig["type"])
	assert.Equal(t, "test-vmess", clashConfig["name"])
	assert.Equal(t, "1.2.3.4", clashConfig["server"])
	assert.Equal(t, 443, clashConfig["port"])
	assert.Equal(t, "uuid", clashConfig["uuid"])
	assert.Equal(t, 64, clashConfig["alterId"])
	assert.Equal(t, "auto", clashConfig["cipher"])
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
