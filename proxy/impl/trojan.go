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

func (p *TrojanProxy) ToShareLink(ext *config.ProxySetting) (string, error) {
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

func (p *TrojanProxy) ToClashConfig(ext *config.ProxySetting) map[string]interface{} {
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
	return options
}
