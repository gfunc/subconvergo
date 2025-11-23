package impl

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/proxy/core"
	"github.com/gfunc/subconvergo/proxy/utils"
)

// TrojanProxy represents a Trojan proxy
type TrojanProxy struct {
	core.BaseProxy `yaml:",inline"`
	Password       string `yaml:"password" json:"password"`
	Network        string `yaml:"network" json:"network"`
	Path           string `yaml:"path" json:"path"`
	Host           string `yaml:"host" json:"host"`
	TLS            bool   `yaml:"tls" json:"tls"`
	AllowInsecure  bool   `yaml:"allow_insecure" json:"allow_insecure"`
}

func (p *TrojanProxy) ToSingleConfig(ext *config.ProxySetting) (string, error) {
	// Format: trojan://password@server:port?params#remark
	link := fmt.Sprintf("trojan://%s@%s:%d", p.Password, p.Server, p.Port)

	params := []string{}
	if p.Host != "" {
		params = append(params, fmt.Sprintf("sni=%s", p.Host))
	}
	if p.Network == "ws" {
		params = append(params, "type=ws")
		if p.Path != "" {
			params = append(params, fmt.Sprintf("path=%s", url.QueryEscape(p.Path)))
		}
	}
	if p.AllowInsecure {
		params = append(params, "allowInsecure=1")
	}

	if len(params) > 0 {
		link += "?" + strings.Join(params, "&")
	}

	if p.Remark != "" {
		link += "#" + utils.UrlEncode(p.Remark)
	}

	return link, nil
}

func (p *TrojanProxy) ToClashConfig(ext *config.ProxySetting) (map[string]interface{}, error) {
	options := map[string]interface{}{
		"type":     "trojan",
		"name":     p.Remark,
		"server":   p.Server,
		"port":     p.Port,
		"password": p.Password,
	}
	if p.Host != "" {
		options["sni"] = p.Host
	}

	if p.AllowInsecure {
		options["skip-cert-verify"] = true
	}

	if p.Network != "" {
		options["network"] = p.Network

		switch p.Network {
		case "ws":
			wsOpts := make(map[string]interface{})
			if p.Path != "" {
				wsOpts["path"] = p.Path
			}
			options["ws-opts"] = wsOpts

		case "grpc":
			grpcOpts := make(map[string]interface{})
			if p.Path != "" {
				grpcOpts["grpc-service-name"] = p.Path
			}
			options["grpc-opts"] = grpcOpts
		}
	}
	return options, nil
}

func (p *TrojanProxy) ToSurgeConfig(ext *config.ProxySetting) (string, error) {
	parts := []string{"trojan", fmt.Sprintf("%s:%d", p.Server, p.Port), fmt.Sprintf("password=%s", p.Password)}

	if p.Network == "ws" {
		parts = append(parts, "ws=true")
		if p.Path != "" {
			parts = append(parts, fmt.Sprintf("ws-path=%s", p.Path))
		}
		if p.Host != "" {
			parts = append(parts, fmt.Sprintf("ws-headers=Host:%s", p.Host))
		}
	}

	if p.Host != "" {
		parts = append(parts, fmt.Sprintf("sni=%s", p.Host))
	}
	if p.AllowInsecure {
		parts = append(parts, "skip-cert-verify=true")
	}

	if ext.TFO {
		parts = append(parts, "tfo=true")
	}
	return fmt.Sprintf("%s = %s", p.Remark, strings.Join(parts, ", ")), nil
}

func (p *TrojanProxy) ToLoonConfig(ext *config.ProxySetting) (string, error) {
	parts := []string{"trojan", fmt.Sprintf("%s:%d", p.Server, p.Port), fmt.Sprintf("password=%s", p.Password)}

	if p.Network == "ws" {
		parts = append(parts, "ws=true")
		if p.Path != "" {
			parts = append(parts, fmt.Sprintf("ws-path=%s", p.Path))
		}
		if p.Host != "" {
			parts = append(parts, fmt.Sprintf("ws-headers=Host:%s", p.Host))
		}
	}

	if p.Host != "" {
		parts = append(parts, fmt.Sprintf("sni=%s", p.Host))
	}
	if p.AllowInsecure {
		parts = append(parts, "skip-cert-verify=true")
	}

	return fmt.Sprintf("%s = %s", p.Remark, strings.Join(parts, ", ")), nil
}

func (p *TrojanProxy) ToQuantumultXConfig(ext *config.ProxySetting) (string, error) {
	// Format: trojan=server:port, password=password, over-tls=true, tls-host=host, fast-open=true/false, udp-relay=true/false, tag=tag
	parts := []string{fmt.Sprintf("trojan=%s:%d", p.Server, p.Port)}
	parts = append(parts, fmt.Sprintf("password=%s", p.Password))
	// parts = append(parts, "over-tls=true") // Removed to match original generator
	if p.Host != "" {
		parts = append(parts, fmt.Sprintf("tls-host=%s", p.Host))
	}
	if p.AllowInsecure {
		parts = append(parts, "tls-verification=false")
	}
	// if p.Network == "ws" {
	// 	parts = append(parts, "obfs=wss")
	// 	if p.Path != "" {
	// 		parts = append(parts, fmt.Sprintf("obfs-uri=%s", p.Path))
	// 	}
	// 	if p.Host != "" {
	// 		parts = append(parts, fmt.Sprintf("obfs-host=%s", p.Host))
	// 	}
	// }
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

func (p *TrojanProxy) ToSingboxConfig(ext *config.ProxySetting) (map[string]interface{}, error) {
	outbound := map[string]interface{}{
		"type":        "trojan",
		"tag":         p.Remark,
		"server":      p.Server,
		"server_port": p.Port,
		"password":    p.Password,
	}

	tls := map[string]interface{}{
		"enabled": true,
	}
	if p.Host != "" {
		tls["server_name"] = p.Host
	}
	if p.AllowInsecure {
		tls["insecure"] = true
	}
	outbound["tls"] = tls

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
	}

	return outbound, nil
}
