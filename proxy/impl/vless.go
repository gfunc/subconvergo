package impl

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/proxy/core"
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

func (p *VLESSProxy) ToShareLink(ext *config.ProxySetting) (string, error) {
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
		link += "#" + core.UrlEncode(p.Remark)
	}

	return link, nil
}

func (p *VLESSProxy) ToClashConfig(ext *config.ProxySetting) map[string]interface{} {
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

	return options
}
