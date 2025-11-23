package impl

import (
	"fmt"

	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/proxy/core"
	"github.com/gfunc/subconvergo/proxy/utils"
)

type HttpProxy struct {
	core.BaseProxy `yaml:",inline"`
	Username       string `yaml:"username" json:"username"`
	Password       string `yaml:"password" json:"password"`
	Tls            bool   `yaml:"tls" json:"tls"`
}

func (p *HttpProxy) ToShareLink(ext *config.ProxySetting) (string, error) {
	// http://user:pass@server:port#remark
	// https://user:pass@server:port#remark
	scheme := "http"
	if p.Tls {
		scheme = "https"
	}

	userInfo := ""
	if p.Username != "" || p.Password != "" {
		userInfo = fmt.Sprintf("%s:%s@", p.Username, p.Password)
	}

	link := fmt.Sprintf("%s://%s%s:%d", scheme, userInfo, p.Server, p.Port)
	link += fmt.Sprintf("#%s", utils.UrlEncode(p.Remark))
	return link, nil
}

func (p *HttpProxy) ToClashConfig(ext *config.ProxySetting) map[string]interface{} {
	options := map[string]interface{}{
		"type":   "http",
		"name":   p.Remark,
		"server": p.Server,
		"port":   p.Port,
	}
	if p.Username != "" {
		options["username"] = p.Username
	}
	if p.Password != "" {
		options["password"] = p.Password
	}
	if p.Tls {
		options["tls"] = true
	}
	return options
}
