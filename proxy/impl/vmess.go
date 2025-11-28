package impl

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/proxy/core"
)

// VMessProxy represents a VMess proxy
type VMessProxy struct {
	core.BaseProxy `yaml:",inline"`
	UUID           string `yaml:"uuid" json:"uuid"`
	AlterID        int    `yaml:"alter_id" json:"alter_id"`
	Cipher         string `yaml:"cipher" json:"cipher"`
	Network        string `yaml:"network" json:"network"`
	FakeType       string `yaml:"fake_type" json:"fake_type"` // Obfuscation type
	Path           string `yaml:"path" json:"path"`
	Host           string `yaml:"host" json:"host"`
	TLS            bool   `yaml:"tls" json:"tls"`
	SNI            string `yaml:"sni" json:"sni"`
}

func (p *VMessProxy) ToSingleConfig(ext *config.ProxySetting) (string, error) {
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

func (p *VMessProxy) ToClashConfig(ext *config.ProxySetting) (map[string]interface{}, error) {
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

	if ext != nil {
		if ext.UDP {
			options["udp"] = true
		}
		if ext.SCV {
			options["skip-cert-verify"] = true
		}
		if ext.TLS13 {
			options["tls13"] = true
		}
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
	return options, nil
}

func (p *VMessProxy) ToSurgeConfig(ext *config.ProxySetting) (string, error) {
	surgeVer := 3
	if ext != nil && ext.SurgeVer != 0 {
		surgeVer = ext.SurgeVer
	}
	if surgeVer < 4 {
		return "", fmt.Errorf("VMess not supported in Surge < 4")
	}

	parts := []string{"vmess", p.Server, fmt.Sprintf("%d", p.Port), fmt.Sprintf("username=%s", p.UUID)}

	if p.Network == "ws" {
		parts = append(parts, "ws=true")
		if p.Path != "" {
			parts = append(parts, fmt.Sprintf("ws-path=%s", p.Path))
		}
		if p.Host != "" {
			parts = append(parts, fmt.Sprintf("ws-headers=Host:%s", p.Host))
		}
	}

	if p.TLS {
		parts = append(parts, "tls=true")
		if p.SNI != "" {
			parts = append(parts, fmt.Sprintf("sni=%s", p.SNI))
		}
	}

	if ext != nil {
		if ext.TFO {
			parts = append(parts, "tfo=true")
		}
		if ext.UDP {
			parts = append(parts, "udp-relay=true")
		}
		if ext.TLS13 {
			parts = append(parts, "tls13=true")
		}
		if ext.SCV {
			parts = append(parts, "skip-cert-verify=true")
		}
	}

	return fmt.Sprintf("%s = %s", p.Remark, strings.Join(parts, ", ")), nil
}

func (p *VMessProxy) ToLoonConfig(ext *config.ProxySetting) (string, error) {
	if p.Network == "grpc" {
		return "", fmt.Errorf("VMess gRPC not supported in Loon")
	}
	parts := []string{"vmess", p.Server, fmt.Sprintf("%d", p.Port), fmt.Sprintf("username=%s", p.UUID)}

	if p.Network == "ws" {
		parts = append(parts, "transport=ws")
		if p.Path != "" {
			parts = append(parts, fmt.Sprintf("path=%s", p.Path))
		}
		if p.Host != "" {
			parts = append(parts, fmt.Sprintf("ws-headers=Host:%s", p.Host))
		}
	}

	if p.TLS {
		parts = append(parts, "tls=true")
		if p.SNI != "" {
			parts = append(parts, fmt.Sprintf("sni=%s", p.SNI))
		}
	}

	if ext != nil {
		if ext.TFO {
			parts = append(parts, "tfo=true")
		}
		if ext.UDP {
			parts = append(parts, "udp-relay=true")
		}
	}

	return fmt.Sprintf("%s = %s", p.Remark, strings.Join(parts, ", ")), nil
}

func (p *VMessProxy) ToQuantumultXConfig(ext *config.ProxySetting) (string, error) {
	parts := []string{fmt.Sprintf("vmess=%s:%d", p.Server, p.Port), "method=aes-128-gcm", fmt.Sprintf("password=%s", p.UUID)}

	if p.Network == "ws" {
		parts = append(parts, "obfs=ws")
		if p.Path != "" {
			parts = append(parts, fmt.Sprintf("obfs-uri=%s", p.Path))
		}
		if p.Host != "" {
			parts = append(parts, fmt.Sprintf("obfs-host=%s", p.Host))
		}
	} else if p.TLS {
		parts = append(parts, "obfs=over-tls")
		if p.SNI != "" {
			parts = append(parts, fmt.Sprintf("obfs-host=%s", p.SNI))
		}
	}

	if ext != nil {
		if ext.TFO {
			parts = append(parts, "fast-open=true")
		}
		if ext.UDP {
			parts = append(parts, "udp-relay=true")
		}
	}

	parts = append(parts, fmt.Sprintf("tag=%s", p.Remark))
	return strings.Join(parts, ", "), nil
}

func (p *VMessProxy) ToSingboxConfig(ext *config.ProxySetting) (map[string]interface{}, error) {
	outbound := map[string]interface{}{
		"type":        "vmess",
		"tag":         p.Remark,
		"server":      p.Server,
		"server_port": p.Port,
		"uuid":        p.UUID,
		"alter_id":    p.AlterID,
		"security":    "auto",
	}

	if p.TLS {
		tls := map[string]interface{}{
			"enabled": true,
		}
		if p.SNI != "" {
			tls["server_name"] = p.SNI
		}
		if ext.SCV {
			tls["insecure"] = true
		}
		outbound["tls"] = tls
	}

	if p.Network == "ws" {
		transport := map[string]interface{}{
			"type": "ws",
		}
		if p.Path != "" {
			transport["path"] = p.Path
		}
		if p.Host != "" {
			transport["headers"] = map[string]string{
				"Host": p.Host,
			}
		}
		outbound["transport"] = transport
	} else if p.Network == "grpc" {
		transport := map[string]interface{}{
			"type": "grpc",
		}
		if p.Path != "" {
			transport["service_name"] = p.Path
		}
		outbound["transport"] = transport
	} else if p.Network == "http" || p.Network == "h2" {
		transport := map[string]interface{}{
			"type": "http",
		}
		if p.Path != "" {
			transport["path"] = p.Path
		}
		if p.Host != "" {
			transport["host"] = []string{p.Host}
		}
		outbound["transport"] = transport
	}

	return outbound, nil
}
