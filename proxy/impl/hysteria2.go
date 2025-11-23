package impl

import (
	"fmt"
	"net/url"

	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/proxy/core"
	"github.com/gfunc/subconvergo/proxy/utils"
)

type Hysteria2Proxy struct {
	core.BaseProxy `yaml:",inline"`
	Password       string `yaml:"password" json:"password"`
	Sni            string `yaml:"sni" json:"sni"`
	SkipCertVerify bool   `yaml:"skip-cert-verify" json:"skip-cert-verify"`
	Obfs           string `yaml:"obfs" json:"obfs"`
	ObfsPassword   string `yaml:"obfs-password" json:"obfs-password"`
}

func (p *Hysteria2Proxy) ToShareLink(ext *config.ProxySetting) (string, error) {
	// hysteria2://password@server:port?sni=...&obfs=...&obfs-password=...
	link := fmt.Sprintf("hysteria2://%s@%s:%d", p.Password, p.Server, p.Port)
	params := url.Values{}
	if p.Sni != "" {
		params.Add("sni", p.Sni)
	}
	if p.SkipCertVerify {
		params.Add("insecure", "1")
	}
	if p.Obfs != "" {
		params.Add("obfs", p.Obfs)
		if p.ObfsPassword != "" {
			params.Add("obfs-password", p.ObfsPassword)
		}
	}

	if len(params) > 0 {
		link += "?" + params.Encode()
	}
	link += fmt.Sprintf("#%s", utils.UrlEncode(p.Remark))
	return link, nil
}

func (p *Hysteria2Proxy) ToClashConfig(ext *config.ProxySetting) map[string]interface{} {
	options := map[string]interface{}{
		"type":     "hysteria2",
		"name":     p.Remark,
		"server":   p.Server,
		"port":     p.Port,
		"password": p.Password,
	}
	if p.Sni != "" {
		options["sni"] = p.Sni
	}
	if p.SkipCertVerify {
		options["skip-cert-verify"] = true
	}
	if p.Obfs != "" {
		options["obfs"] = p.Obfs
		if p.ObfsPassword != "" {
			options["obfs-password"] = p.ObfsPassword
		}
	}
	return options
}
