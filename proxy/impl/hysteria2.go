package impl

import (
	"fmt"
	"net/url"
	"strings"

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

func (p *Hysteria2Proxy) ToSingleConfig(ext *config.ProxySetting) (string, error) {
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

func (p *Hysteria2Proxy) ToClashConfig(ext *config.ProxySetting) (map[string]interface{}, error) {
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
	return options, nil
}

func (p *Hysteria2Proxy) ToSurgeConfig(ext *config.ProxySetting) (string, error) {
	surgeVer := 3
	if ext != nil && ext.SurgeVer != 0 {
		surgeVer = ext.SurgeVer
	}
	if surgeVer < 4 {
		return "", fmt.Errorf("Hysteria2 not supported in Surge < 4")
	}

	parts := []string{"hysteria2", p.Server, fmt.Sprintf("%d", p.Port)}
	parts = append(parts, fmt.Sprintf("password=%s", p.Password))
	if p.Sni != "" {
		parts = append(parts, fmt.Sprintf("sni=%s", p.Sni))
	}
	if p.SkipCertVerify {
		parts = append(parts, "skip-cert-verify=true")
	}
	if p.Obfs != "" {
		parts = append(parts, fmt.Sprintf("obfs=%s", p.Obfs))
		if p.ObfsPassword != "" {
			parts = append(parts, fmt.Sprintf("obfs-password=%s", p.ObfsPassword))
		}
	}
	if ext.TFO {
		parts = append(parts, "tfo=true")
	}
	return fmt.Sprintf("%s = %s", p.Remark, strings.Join(parts, ", ")), nil
}

func (p *Hysteria2Proxy) ToLoonConfig(ext *config.ProxySetting) (string, error) {
	parts := []string{"hysteria2", p.Server, fmt.Sprintf("%d", p.Port)}
	if p.Password != "" {
		parts = append(parts, fmt.Sprintf("password=%s", p.Password))
	}
	if p.Sni != "" {
		parts = append(parts, fmt.Sprintf("sni=%s", p.Sni))
	}
	if p.SkipCertVerify || ext.SCV {
		parts = append(parts, "skip-cert-verify=true")
	}
	if p.Obfs != "" {
		parts = append(parts, fmt.Sprintf("obfs=%s", p.Obfs))
		if p.ObfsPassword != "" {
			parts = append(parts, fmt.Sprintf("obfs-password=%s", p.ObfsPassword))
		}
	}
	return fmt.Sprintf("%s = %s", p.Remark, strings.Join(parts, ",")), nil
}

func (p *Hysteria2Proxy) ToQuantumultXConfig(ext *config.ProxySetting) (string, error) {
	return "", fmt.Errorf("ToQuantumultXConfig not supported for proxy type hysteria2")
}

func (p *Hysteria2Proxy) ToSingboxConfig(ext *config.ProxySetting) (map[string]interface{}, error) {
	outbound := map[string]interface{}{
		"type":        "hysteria2",
		"tag":         p.Remark,
		"server":      p.Server,
		"server_port": p.Port,
		"password":    p.Password,
	}
	if p.Sni != "" {
		outbound["tls"] = map[string]interface{}{
			"enabled":     true,
			"server_name": p.Sni,
			"insecure":    p.SkipCertVerify,
		}
	}
	if p.Obfs != "" {
		outbound["obfs"] = map[string]interface{}{
			"type":     p.Obfs,
			"password": p.ObfsPassword,
		}
	}
	return outbound, nil
}
