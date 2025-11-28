package sub

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClashSubscriptionParser(t *testing.T) {
	content := `
proxies:
  - name: "ss1"
    type: ss
    server: server
    port: 443
    cipher: aes-256-gcm
    password: password
  - name: "vmess1"
    type: vmess
    server: server
    port: 443
    uuid: uuid
    alterId: 0
    cipher: auto
    network: ws
    ws-opts:
      path: /path
`
	parser := &ClashSubscriptionParser{}
	assert.True(t, parser.CanParse(content))

	sub, err := parser.Parse(content)
	assert.NoError(t, err)
	assert.NotNil(t, sub)
	assert.Equal(t, 2, len(sub.Proxies))

	ss := sub.Proxies[0]
	assert.Equal(t, "ss", ss.GetType())
	assert.Equal(t, "ss1", ss.GetRemark())

	vmess := sub.Proxies[1]
	assert.Equal(t, "vmess", vmess.GetType())
	assert.Equal(t, "vmess1", vmess.GetRemark())
}

func TestSSDSubscriptionParser(t *testing.T) {
	// SSD is base64 encoded JSON
	jsonContent := `
{
  "airport": "TestAirport",
  "port": 443,
  "encryption": "aes-256-gcm",
  "password": "password",
  "servers": [
    {
      "server": "server1",
      "remarks": "ss1"
    }
  ]
}
`
	encoded := "ssd://" + base64.RawURLEncoding.EncodeToString([]byte(jsonContent))

	parser := &SSDSubscriptionParser{}
	assert.True(t, parser.CanParse(encoded))

	sub, err := parser.Parse(encoded)
	assert.NoError(t, err)
	assert.NotNil(t, sub)
	assert.Equal(t, 1, len(sub.Proxies))

	ss := sub.Proxies[0]
	assert.Equal(t, "ss", ss.GetType())
	assert.Equal(t, "ss1", ss.GetRemark())
	assert.Equal(t, "TestAirport", ss.GetGroup())
}

func TestSSConfSubscriptionParser(t *testing.T) {
	jsonContent := `
{
  "version": 1,
  "remarks": "TestGroup",
  "configs": [
    {
      "server": "server1",
      "server_port": "443",
      "password": "password",
      "method": "aes-256-gcm",
      "remarks": "ss1"
    }
  ]
}
`
	parser := &SSSubscriptionParser{}
	assert.True(t, parser.CanParse(jsonContent))

	sub, err := parser.Parse(jsonContent)
	assert.NoError(t, err)
	assert.NotNil(t, sub)
	assert.Equal(t, 1, len(sub.Proxies))

	ss := sub.Proxies[0]
	assert.Equal(t, "ss", ss.GetType())
	assert.Equal(t, "ss1", ss.GetRemark())
	assert.Equal(t, "TestGroup", ss.GetGroup())
}

func TestV2RaySubscriptionParser(t *testing.T) {
	jsonContent := `
{
  "outbounds": [
    {
      "protocol": "vmess",
      "settings": {
        "vnext": [
          {
            "address": "server1",
            "port": 443,
            "users": [
              {
                "id": "uuid",
                "alterId": 0,
                "security": "auto"
              }
            ]
          }
        ]
      },
      "streamSettings": {
        "network": "ws",
        "wsSettings": {
          "path": "/path"
        }
      }
    }
  ]
}
`
	parser := &V2RaySubscriptionParser{}
	assert.True(t, parser.CanParse(jsonContent))

	sub, err := parser.Parse(jsonContent)
	assert.NoError(t, err)
	assert.NotNil(t, sub)
	assert.Equal(t, 1, len(sub.Proxies))

	vmess := sub.Proxies[0]
	assert.Equal(t, "vmess", vmess.GetType())
	assert.Equal(t, "V2Ray Config", vmess.GetRemark())
}
