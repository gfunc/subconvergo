package impl

import (
	"fmt"

	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/proxy/core"
	"github.com/gfunc/subconvergo/proxy/utils"
)

// Socks5Proxy represents a Socks5 proxy
type Socks5Proxy struct {
	core.BaseProxy `yaml:",inline"`
	Username       string `yaml:"username" json:"username"`
	Password       string `yaml:"password" json:"password"`
	TLS            bool   `yaml:"tls" json:"tls"`
}

func (p *Socks5Proxy) ToSingleConfig(ext *config.ProxySetting) (string, error) {
	// Format: socks5://user:pass@server:port#remark
	var link string
	if p.Username != "" {
		link = fmt.Sprintf("socks5://%s:%s@%s:%d", p.Username, p.Password, p.Server, p.Port)
	} else {
		link = fmt.Sprintf("socks5://%s:%d", p.Server, p.Port)
	}

	link += fmt.Sprintf("#%s", utils.UrlEncode(p.Remark))
	return link, nil
}

func (p *Socks5Proxy) ToClashConfig(ext *config.ProxySetting) (map[string]interface{}, error) {
	options := map[string]interface{}{
		"type":     "socks5",
		"name":     p.Remark,
		"server":   p.Server,
		"port":     p.Port,
		"username": p.Username,
		"password": p.Password,
		"tls":      p.TLS,
	}
	return options, nil
}

func (p *Socks5Proxy) ToSurgeConfig(ext *config.ProxySetting) (string, error) {
	// Format: Name = socks5, Server, Port, username=..., password=...
	line := fmt.Sprintf("%s = socks5, %s, %d", p.Remark, p.Server, p.Port)
	if p.TLS {
		line = fmt.Sprintf("%s = socks5-tls, %s, %d", p.Remark, p.Server, p.Port)
	}
	if p.Username != "" {
		line += fmt.Sprintf(", username=%s", p.Username)
	}
	if p.Password != "" {
		line += fmt.Sprintf(", password=%s", p.Password)
	}
	return line, nil
}

func (p *Socks5Proxy) ToQuanXConfig(ext *config.ProxySetting) (string, error) {
	// Format: socks5=Server:Port, username=..., password=..., tag=Remark
	line := fmt.Sprintf("socks5=%s:%d", p.Server, p.Port)
	if p.Username != "" {
		line += fmt.Sprintf(", username=%s", p.Username)
	}
	if p.Password != "" {
		line += fmt.Sprintf(", password=%s", p.Password)
	}
	line += fmt.Sprintf(", tag=%s", p.Remark)
	return line, nil
}

func (p *Socks5Proxy) ToLoonConfig(ext *config.ProxySetting) (string, error) {
	return p.ToSurgeConfig(ext)
}

func (p *Socks5Proxy) ToSingBoxConfig(ext *config.ProxySetting) (map[string]interface{}, error) {
	outbound := map[string]interface{}{
		"type":        "socks",
		"tag":         p.Remark,
		"server":      p.Server,
		"server_port": p.Port,
		"username":    p.Username,
		"password":    p.Password,
	}
	return outbound, nil
}
