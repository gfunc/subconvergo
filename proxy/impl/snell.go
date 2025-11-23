package impl

import (
	"fmt"
	"net/url"

	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/proxy/core"
	"github.com/gfunc/subconvergo/proxy/utils"
)

type SnellProxy struct {
	core.BaseProxy `yaml:",inline"`
	Psk            string `yaml:"psk" json:"psk"`
	Obfs           string `yaml:"obfs" json:"obfs"`
	ObfsParam      string `yaml:"obfs-opts" json:"obfs-opts"` // Clash uses obfs-opts
	Version        int    `yaml:"version" json:"version"`
}

func (p *SnellProxy) ToShareLink(ext *config.ProxySetting) (string, error) {
	// snell://server:port?psk=...&obfs=...
	params := url.Values{}
	if p.Psk != "" {
		params.Add("psk", p.Psk)
	}
	if p.Obfs != "" {
		params.Add("obfs", p.Obfs)
	}
	if p.ObfsParam != "" {
		params.Add("obfs-host", p.ObfsParam) // Usually obfs-host for http obfs
	}
	if p.Version > 0 {
		params.Add("version", fmt.Sprintf("%d", p.Version))
	}

	link := fmt.Sprintf("snell://%s:%d?%s", p.Server, p.Port, params.Encode())
	link += fmt.Sprintf("#%s", utils.UrlEncode(p.Remark))
	return link, nil
}

func (p *SnellProxy) ToClashConfig(ext *config.ProxySetting) map[string]interface{} {
	options := map[string]interface{}{
		"type":   "snell",
		"name":   p.Remark,
		"server": p.Server,
		"port":   p.Port,
		"psk":    p.Psk,
	}
	if p.Version > 0 {
		options["version"] = p.Version
	}
	if p.Obfs != "" {
		options["obfs-opts"] = map[string]interface{}{
			"mode": p.Obfs,
			"host": p.ObfsParam,
		}
	}
	return options
}
