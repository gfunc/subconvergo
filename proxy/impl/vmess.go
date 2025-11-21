package impl

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/proxy/core"
)

// VMessProxy represents a VMess proxy
type VMessProxy struct {
	core.BaseProxy `yaml:",inline"`
	UUID           string `yaml:"uuid" json:"uuid"`
	AlterID        int    `yaml:"alter_id" json:"alter_id"`
	Network        string `yaml:"network" json:"network"`
	Path           string `yaml:"path" json:"path"`
	Host           string `yaml:"host" json:"host"`
	TLS            bool   `yaml:"tls" json:"tls"`
	SNI            string `yaml:"sni" json:"sni"`
}

func (p *VMessProxy) ToShareLink(ext *config.ProxySetting) (string, error) {
	// VMess JSON format
	vmessData := map[string]interface{}{
		"v":    "2",
		"ps":   p.Remark,
		"add":  p.Server,
		"port": fmt.Sprintf("%d", p.Port),
		"id":   p.UUID,
		"aid":  fmt.Sprintf("%d", p.AlterID),
		"net":  p.Network,
		"type": "none",
		"host": p.Host,
		"path": p.Path,
		"tls":  "",
	}

	if p.TLS {
		vmessData["tls"] = "tls"
	}

	jsonBytes, err := json.Marshal(vmessData)
	if err != nil {
		return "", err
	}

	encoded := base64.StdEncoding.EncodeToString(jsonBytes)
	return "vmess://" + encoded, nil
}

func (p *VMessProxy) ToClashConfig(ext *config.ProxySetting) map[string]interface{} {
	options := map[string]interface{}{
		"type":    "vmess",
		"name":    p.Remark,
		"server":  p.Server,
		"port":    p.Port,
		"uuid":    p.UUID,
		"alterId": p.AlterID,
		"cipher":  "auto",
		"network": p.Network,
	}

	if p.TLS {
		options["tls"] = true
		if p.SNI != "" {
			options["servername"] = p.SNI
		}
	}
	switch p.Network {
	case "ws":
		wsOpts := make(map[string]interface{})
		if p.Path == "" {
			p.Path = "/"
		}
		wsOpts["path"] = p.Path
		if p.Host != "" {
			headers := make(map[string]string)
			headers["Host"] = p.Host
			wsOpts["headers"] = headers
		}
		options["ws-opts"] = wsOpts
	case "httpupgrade":
		options["network"] = "ws"
		wsOpts := make(map[string]interface{})
		if p.Path == "" {
			p.Path = "/"
		}
		wsOpts["v2ray-http-upgrade"] = true
		wsOpts["v2ray-http-upgrade-fast-open"] = true
		wsOpts["path"] = p.Path
		if p.Host != "" {
			headers := make(map[string]string)
			headers["Host"] = p.Host
			wsOpts["headers"] = headers
		}
		options["ws-opts"] = wsOpts
	case "http", "h2":
		h2Opts := make(map[string]interface{})
		if p.Path != "" {
			h2Opts["path"] = p.Path
		}
		if p.Host != "" {
			h2Opts["host"] = []string{p.Host}
		}
		options["h2-opts"] = h2Opts

	case "grpc":
		grpcOpts := make(map[string]interface{})
		if p.Path != "" {
			grpcOpts["grpc-service-name"] = p.Path
		}
		options["grpc-opts"] = grpcOpts
		if p.Host != "" {
			options["servername"] = p.Host
		}

	case "quic":
		quicOpts := make(map[string]interface{})
		if p.Host != "" {
			quicOpts["mode"] = p.Host
		}
		if p.Path != "" {
			quicOpts["key"] = p.Path
		}
		options["quic-opts"] = quicOpts
	}
	return options
}
