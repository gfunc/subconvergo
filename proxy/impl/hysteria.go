package impl

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/proxy/core"
	"github.com/gfunc/subconvergo/proxy/utils"
)

// HysteriaProxy represents a Hysteria or Hysteria2 proxy
type HysteriaProxy struct {
	core.BaseProxy `yaml:",inline"`
	Password       string     `yaml:"password" json:"password"`
	Obfs           string     `yaml:"obfs" json:"obfs"`
	AllowInsecure  bool       `yaml:"allow_insecure" json:"allow_insecure"`
	Params         url.Values `yaml:"-" json:"params"`
}

func (p *HysteriaProxy) ToSingleConfig(ext *config.ProxySetting) (string, error) {
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
		link += "#" + utils.UrlEncode(p.Remark)
	}

	return link, nil
}

func (p *HysteriaProxy) ToClashConfig(ext *config.ProxySetting) (map[string]interface{}, error) {
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
	return options, nil
}

func (p *HysteriaProxy) ToSurgeConfig(ext *config.ProxySetting) (string, error) {
	parts := []string{p.Type, p.Server, fmt.Sprintf("%d", p.Port)}
	if p.Type == "hysteria2" {
		parts = append(parts, fmt.Sprintf("password=%s", p.Password))
	} else {
		parts = append(parts, fmt.Sprintf("auth=%s", p.Password))
	}

	if p.Params != nil {
		if sni := p.Params.Get("sni"); sni != "" {
			parts = append(parts, fmt.Sprintf("sni=%s", sni))
		} else if peer := p.Params.Get("peer"); peer != "" {
			parts = append(parts, fmt.Sprintf("sni=%s", peer))
		}
		if alpn := p.Params.Get("alpn"); alpn != "" {
			parts = append(parts, fmt.Sprintf("alpn=%s", alpn))
		}
		if p.Type == "hysteria" {
			up := p.Params.Get("upmbps")
			if up == "" {
				up = p.Params.Get("up")
			}
			down := p.Params.Get("downmbps")
			if down == "" {
				down = p.Params.Get("down")
			}
			if up != "" {
				parts = append(parts, fmt.Sprintf("up=%s", up))
			}
			if down != "" {
				parts = append(parts, fmt.Sprintf("down=%s", down))
			}
		}
		if p.Type == "hysteria2" {
			if obfsPassword := p.Params.Get("obfs-password"); obfsPassword != "" {
				parts = append(parts, fmt.Sprintf("obfs-password=%s", obfsPassword))
			}
		}
	}

	if p.AllowInsecure {
		parts = append(parts, "skip-cert-verify=true")
	}
	if p.Obfs != "" {
		parts = append(parts, fmt.Sprintf("obfs=%s", p.Obfs))
	}
	if ext.TFO {
		parts = append(parts, "tfo=true")
	}
	return fmt.Sprintf("%s = %s", p.Remark, strings.Join(parts, ", ")), nil
}

func (p *HysteriaProxy) ToLoonConfig(ext *config.ProxySetting) (string, error) {
	// Format: hysteria,server,port,auth_str=...,...
	parts := []string{p.Type, p.Server, fmt.Sprintf("%d", p.Port)}
	if p.Type == "hysteria2" {
		parts = append(parts, fmt.Sprintf("password=\"%s\"", p.Password))
	} else {
		parts = append(parts, fmt.Sprintf("auth_str=\"%s\"", p.Password))
	}

	if p.Params != nil {
		if sni := p.Params.Get("sni"); sni != "" {
			parts = append(parts, fmt.Sprintf("sni=%s", sni))
		} else if peer := p.Params.Get("peer"); peer != "" {
			parts = append(parts, fmt.Sprintf("sni=%s", peer))
		}
		if alpn := p.Params.Get("alpn"); alpn != "" {
			parts = append(parts, fmt.Sprintf("alpn=\"%s\"", alpn))
		}
		if p.Type == "hysteria" {
			up := p.Params.Get("upmbps")
			if up == "" {
				up = p.Params.Get("up")
			}
			down := p.Params.Get("downmbps")
			if down == "" {
				down = p.Params.Get("down")
			}
			if up != "" {
				parts = append(parts, fmt.Sprintf("up=%s", up))
			}
			if down != "" {
				parts = append(parts, fmt.Sprintf("down=%s", down))
			}
		}
	}

	if p.AllowInsecure {
		parts = append(parts, "skip-cert-verify=true")
	}
	return fmt.Sprintf("%s = %s", p.Remark, strings.Join(parts, ",")), nil
}

func (p *HysteriaProxy) ToQuantumultXConfig(ext *config.ProxySetting) (string, error) {
	return "", fmt.Errorf("hysteria not supported in Quantumult X")
}

func (p *HysteriaProxy) ToSingboxConfig(ext *config.ProxySetting) (map[string]interface{}, error) {
	outbound := map[string]interface{}{
		"type":        p.Type,
		"tag":         p.Remark,
		"server":      p.Server,
		"server_port": p.Port,
	}

	if p.Type == "hysteria2" {
		outbound["password"] = p.Password
	} else {
		outbound["auth_str"] = p.Password
	}

	tls := map[string]interface{}{
		"enabled": true,
	}

	if p.Params != nil {
		if sni := p.Params.Get("sni"); sni != "" {
			tls["server_name"] = sni
		} else if peer := p.Params.Get("peer"); peer != "" {
			tls["server_name"] = peer
		}
		if alpn := p.Params.Get("alpn"); alpn != "" {
			tls["alpn"] = strings.Split(alpn, ",")
		}
		if p.Type == "hysteria" {
			up := p.Params.Get("upmbps")
			if up == "" {
				up = p.Params.Get("up")
			}
			down := p.Params.Get("downmbps")
			if down == "" {
				down = p.Params.Get("down")
			}
			if up != "" {
				if upNum, err := strconv.Atoi(up); err == nil {
					outbound["up_mbps"] = upNum
				}
			}
			if down != "" {
				if downNum, err := strconv.Atoi(down); err == nil {
					outbound["down_mbps"] = downNum
				}
			}
		}
		if p.Type == "hysteria2" {
			if obfsPassword := p.Params.Get("obfs-password"); obfsPassword != "" {
				outbound["obfs"] = map[string]interface{}{
					"type":     p.Obfs,
					"password": obfsPassword,
				}
			}
		}
	}

	if p.AllowInsecure {
		tls["insecure"] = true
	}
	outbound["tls"] = tls

	if p.Obfs != "" && p.Type == "hysteria" {
		outbound["obfs"] = p.Obfs
	}

	return outbound, nil
}
