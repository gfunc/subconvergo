package impl

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/proxy/core"
	"github.com/gfunc/subconvergo/proxy/utils"
)

// VLESSProxy represents a VLESS proxy
type VLESSProxy struct {
	core.BaseProxy `yaml:",inline"`
	UUID           string `yaml:"uuid" json:"uuid"`
	Network        string `yaml:"network" json:"network"`
	Path           string `yaml:"path" json:"path"`
	Host           string `yaml:"host" json:"host"`
	TLS            bool   `yaml:"tls" json:"tls"`
	AllowInsecure  bool   `yaml:"allow_insecure" json:"allow_insecure"`
	Flow           string `yaml:"flow" json:"flow"`
	SNI            string `yaml:"sni" json:"sni"`
}

func (p *VLESSProxy) ToSingleConfig(ext *config.ProxySetting) (string, error) {
	// Format: vless://uuid@server:port?params#remark
	link := fmt.Sprintf("vless://%s@%s:%d", p.UUID, p.Server, p.Port)

	params := []string{fmt.Sprintf("type=%s", p.Network)}

	if p.TLS {
		params = append(params, "security=tls")
		if p.Host != "" {
			params = append(params, fmt.Sprintf("sni=%s", p.Host))
		}
	}

	if p.Network == "ws" && p.Path != "" {
		params = append(params, fmt.Sprintf("path=%s", url.QueryEscape(p.Path)))
	}
	if p.Host != "" && p.Network == "ws" {
		params = append(params, fmt.Sprintf("host=%s", p.Host))
	}

	link += "?" + strings.Join(params, "&")

	if p.Remark != "" {
		link += "#" + utils.UrlEncode(p.Remark)
	}

	return link, nil
}

func (p *VLESSProxy) ToClashConfig(ext *config.ProxySetting) (map[string]interface{}, error) {
	options := map[string]interface{}{
		"type":    "vless",
		"name":    p.Remark,
		"server":  p.Server,
		"port":    p.Port,
		"uuid":    p.UUID,
		"network": p.Network,
	}

	if p.Flow != "" {
		options["flow"] = p.Flow
	}

	if p.TLS {
		options["tls"] = true
		if p.SNI != "" {
			options["servername"] = p.SNI
		}
	}

	if p.AllowInsecure {
		options["skip-cert-verify"] = true
	}

	switch p.Network {
	case "ws":
		wsOpts := make(map[string]interface{})
		if p.Path != "" {
			wsOpts["path"] = p.Path
		}
		if p.Host != "" {
			headers := make(map[string]string)
			headers["Host"] = p.Host
			wsOpts["headers"] = headers
		}
		options["ws-opts"] = wsOpts

	case "grpc":
		grpcOpts := make(map[string]interface{})
		if p.Path != "" {
			grpcOpts["grpc-service-name"] = p.Path
		}
		options["grpc-opts"] = grpcOpts

	case "http", "h2":
		h2Opts := make(map[string]interface{})
		if p.Path != "" {
			h2Opts["path"] = p.Path
		}
		if p.Host != "" {
			h2Opts["host"] = []string{p.Host}
		}
		options["h2-opts"] = h2Opts
	}

	return options, nil
}

func (p *VLESSProxy) ToSurgeConfig(ext *config.ProxySetting) (string, error) {
	return "", fmt.Errorf("VLESS not supported in Surge")
}

func (p *VLESSProxy) ToLoonConfig(ext *config.ProxySetting) (string, error) {
	// Format: vless,server,port,uuid,transport,args
	parts := []string{"vless", p.Server, fmt.Sprintf("%d", p.Port), p.UUID}
	if p.Network == "ws" {
		parts = append(parts, "transport=ws")
		if p.Path != "" {
			parts = append(parts, fmt.Sprintf("path=\"%s\"", p.Path))
		}
		if p.Host != "" {
			parts = append(parts, fmt.Sprintf("host=\"%s\"", p.Host))
		}
	}
	if p.TLS {
		parts = append(parts, "over-tls=true")
		if p.SNI != "" {
			parts = append(parts, fmt.Sprintf("u-tls-name=%s", p.SNI))
		}
	}
	if p.AllowInsecure {
		parts = append(parts, "skip-cert-verify=true")
	}
	return fmt.Sprintf("%s = %s", p.Remark, strings.Join(parts, ",")), nil
}

func (p *VLESSProxy) ToQuantumultXConfig(ext *config.ProxySetting) (string, error) {
	return "", fmt.Errorf("VLESS not supported in Quantumult X")
}

func (p *VLESSProxy) ToSingboxConfig(ext *config.ProxySetting) (map[string]interface{}, error) {
	outbound := map[string]interface{}{
		"type":        "vless",
		"tag":         p.Remark,
		"server":      p.Server,
		"server_port": p.Port,
		"uuid":        p.UUID,
	}

	if p.Flow != "" {
		outbound["flow"] = p.Flow
	}

	if p.TLS {
		tls := map[string]interface{}{
			"enabled": true,
		}
		if p.SNI != "" {
			tls["server_name"] = p.SNI
		}
		if p.AllowInsecure {
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
