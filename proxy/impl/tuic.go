package impl

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/proxy/core"
	"github.com/gfunc/subconvergo/proxy/utils"
)

// TUICProxy represents a TUIC proxy
type TUICProxy struct {
	core.BaseProxy `yaml:",inline"`
	UUID           string     `yaml:"uuid" json:"uuid"`
	Password       string     `yaml:"password" json:"password"`
	AllowInsecure  bool       `yaml:"allow_insecure" json:"allow_insecure"`
	Params         url.Values `yaml:"-" json:"params"`
}

func (p *TUICProxy) ToSingleConfig(ext *config.ProxySetting) (string, error) {
	link := fmt.Sprintf("tuic://%s", p.UUID)
	if p.Password != "" {
		link += ":" + p.Password
	}
	link += fmt.Sprintf("@%s:%d", p.Server, p.Port)

	if p.AllowInsecure {
		p.Params.Add("allow_insecure", "1")
	}

	if len(p.Params) > 0 {
		link += "?" + p.Params.Encode()
	}

	if p.Remark != "" {
		link += "#" + utils.UrlEncode(p.Remark)
	}

	return link, nil
}

func (p *TUICProxy) ToClashConfig(ext *config.ProxySetting) (map[string]interface{}, error) {
	options := map[string]interface{}{
		"type":   "tuic",
		"name":   p.Remark,
		"server": p.Server,
		"port":   p.Port,
		"uuid":   p.UUID,
	}

	if p.Password != "" {
		options["password"] = p.Password
	}

	if p.AllowInsecure {
		options["skip-cert-verify"] = true
	}
	if p.Params != nil {
		if sni := p.Params.Get("sni"); sni != "" {
			options["sni"] = sni
		}
		if alpn := p.Params.Get("alpn"); alpn != "" {
			options["alpn"] = strings.Split(alpn, ",")
		}
		if congestion := p.Params.Get("congestion_control"); congestion != "" {
			options["congestion-controller"] = congestion
		}
		if udpRelay := p.Params.Get("udp_relay_mode"); udpRelay != "" {
			options["udp-relay-mode"] = udpRelay
		}
	}
	return options, nil
}

func (p *TUICProxy) ToSurgeConfig(ext *config.ProxySetting) (string, error) {
	surgeVer := 3
	if ext != nil && ext.SurgeVer != 0 {
		surgeVer = ext.SurgeVer
	}
	if surgeVer < 4 {
		return "", fmt.Errorf("TUIC not supported in Surge < 4")
	}

	parts := []string{"tuic", p.Server, fmt.Sprintf("%d", p.Port)}
	if p.UUID != "" {
		parts = append(parts, fmt.Sprintf("uuid=%s", p.UUID))
	}
	if p.Password != "" {
		parts = append(parts, fmt.Sprintf("password=%s", p.Password))
	}

	if p.Params != nil {
		if sni := p.Params.Get("sni"); sni != "" {
			parts = append(parts, fmt.Sprintf("sni=%s", sni))
		}
		if alpn := p.Params.Get("alpn"); alpn != "" {
			parts = append(parts, fmt.Sprintf("alpn=%s", alpn))
		}
		if congestion := p.Params.Get("congestion_control"); congestion != "" {
			parts = append(parts, fmt.Sprintf("congestion-controller=%s", congestion))
		}
	}

	if p.AllowInsecure {
		parts = append(parts, "skip-cert-verify=true")
	}
	if ext.TFO {
		parts = append(parts, "tfo=true")
	}
	return fmt.Sprintf("%s = %s", p.Remark, strings.Join(parts, ", ")), nil
}

func (p *TUICProxy) ToLoonConfig(ext *config.ProxySetting) (string, error) {
	return "", fmt.Errorf("ToLoonConfig not supported for proxy type tuic")
}

func (p *TUICProxy) ToQuantumultXConfig(ext *config.ProxySetting) (string, error) {
	return "", fmt.Errorf("TUIC not supported in Quantumult X")
}

func (p *TUICProxy) ToSingboxConfig(ext *config.ProxySetting) (map[string]interface{}, error) {
	outbound := map[string]interface{}{
		"type":        "tuic",
		"tag":         p.Remark,
		"server":      p.Server,
		"server_port": p.Port,
		"uuid":        p.UUID,
	}
	if p.Password != "" {
		outbound["password"] = p.Password
	}

	if p.Params != nil {
		if sni := p.Params.Get("sni"); sni != "" {
			outbound["sni"] = sni
		}
		if congestion := p.Params.Get("congestion_control"); congestion != "" {
			outbound["congestion_controller"] = congestion
		}
		if udpRelay := p.Params.Get("udp_relay_mode"); udpRelay != "" {
			outbound["udp_relay_mode"] = udpRelay
		}
	}

	tls := map[string]interface{}{
		"enabled": true,
	}
	if p.AllowInsecure || ext.SCV {
		tls["insecure"] = true
	}
	if p.Params != nil {
		if alpn := p.Params.Get("alpn"); alpn != "" {
			tls["alpn"] = strings.Split(alpn, ",")
		}
	}
	outbound["tls"] = tls

	return outbound, nil
}
