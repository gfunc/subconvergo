package impl

import (
	"fmt"
	"net/url"
	"strings"

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

func (p *SnellProxy) ToSingleConfig(ext *config.ProxySetting) (string, error) {
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

func (p *SnellProxy) ToClashConfig(ext *config.ProxySetting) (map[string]interface{}, error) {
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
	return options, nil
}

func (p *SnellProxy) ToSurgeConfig(ext *config.ProxySetting) (string, error) {
	parts := []string{"snell", p.Server, fmt.Sprintf("%d", p.Port), fmt.Sprintf("psk=%s", p.Psk)}
	if p.Obfs != "" {
		parts = append(parts, fmt.Sprintf("obfs=%s", p.Obfs))
		if p.ObfsParam != "" {
			parts = append(parts, fmt.Sprintf("obfs-host=%s", p.ObfsParam))
		}
	}
	if p.Version > 0 {
		parts = append(parts, fmt.Sprintf("version=%d", p.Version))
	}
	if ext.TFO {
		parts = append(parts, "tfo=true")
	}
	return fmt.Sprintf("%s = %s", p.Remark, strings.Join(parts, ", ")), nil
}

func (p *SnellProxy) ToLoonConfig(ext *config.ProxySetting) (string, error) {
	return "", fmt.Errorf("snell not supported in Loon")
}

func (p *SnellProxy) ToQuantumultXConfig(ext *config.ProxySetting) (string, error) {
	return "", fmt.Errorf("snell not supported in Quantumult X")
}

func (p *SnellProxy) ToSingboxConfig(ext *config.ProxySetting) (map[string]interface{}, error) {
	return nil, fmt.Errorf("snell not supported in sing-box")
}
