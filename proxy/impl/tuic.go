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

func (p *TUICProxy) ToShareLink(ext *config.ProxySetting) (string, error) {
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

func (p *TUICProxy) ToClashConfig(ext *config.ProxySetting) map[string]interface{} {
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
	return options
}
