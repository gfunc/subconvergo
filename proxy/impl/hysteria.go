package impl

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/proxy/core"
)

// HysteriaProxy represents a Hysteria or Hysteria2 proxy
type HysteriaProxy struct {
	core.BaseProxy `yaml:",inline"`
	Password       string     `yaml:"password" json:"password"`
	Obfs           string     `yaml:"obfs" json:"obfs"`
	AllowInsecure  bool       `yaml:"allow_insecure" json:"allow_insecure"`
	Params         url.Values `yaml:"-" json:"params"`
}

func (p *HysteriaProxy) ToShareLink(ext *config.ProxySetting) (string, error) {
	protocol := p.Type
	link := fmt.Sprintf("%s://%s@%s:%d", protocol, p.Password, p.Server, p.Port)

	if p.AllowInsecure {
		p.Params.Add("insecure", "1")
	}
	if p.Obfs != "" {
		p.Params.Add("obfs", p.Obfs)
	}

	if len(p.Params) > 0 {
		link += "?" + p.Params.Encode()
	}

	if p.Remark != "" {
		link += "#" + core.UrlEncode(p.Remark)
	}

	return link, nil
}

func (p *HysteriaProxy) ToClashConfig(ext *config.ProxySetting) map[string]interface{} {
	options := map[string]interface{}{
		"type":   p.Type,
		"name":   p.Remark,
		"server": p.Server,
		"port":   p.Port,
	}

	if p.Password != "" {
		if p.Type == "hysteria2" {
			options["password"] = p.Password
		} else {
			options["auth-str"] = p.Password
		}
	}

	if p.AllowInsecure {
		options["skip-cert-verify"] = true
	}

	if p.Obfs != "" {
		options["obfs"] = p.Obfs
	}

	if p.Params != nil {
		if sni := p.Params.Get("sni"); sni != "" {
			options["sni"] = sni
		}
		if peer := p.Params.Get("peer"); peer != "" {
			options["sni"] = peer
		}
		if alpn := p.Params.Get("alpn"); alpn != "" {
			options["alpn"] = strings.Split(alpn, ",")
		}

		if p.Type == "hysteria" {
			if p.Password == "" {
				if auth := p.Params.Get("auth"); auth != "" {
					p.Password = auth
					options["auth-str"] = auth
				}
			}

			up := p.Params.Get("upmbps")
			if up == "" {
				up = p.Params.Get("up")
			}
			down := p.Params.Get("downmbps")
			if down == "" {
				down = p.Params.Get("down")
			}
			if up == "" {
				up = "10"
			}
			if down == "" {
				down = "50"
			}
			if upNum, err := strconv.Atoi(up); err == nil {
				options["up"] = upNum
			}
			if downNum, err := strconv.Atoi(down); err == nil {
				options["down"] = downNum
			}
		}

		if p.Type == "hysteria2" {
			if obfsPassword := p.Params.Get("obfs-password"); obfsPassword != "" {
				options["obfs-password"] = obfsPassword
			}
		}
	} else if p.Type == "hysteria" {
		options["up"] = 10
		options["down"] = 50
	}
	return options
}
