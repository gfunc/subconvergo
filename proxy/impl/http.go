package impl

import (
	"fmt"
	"strings"

	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/proxy/core"
	"github.com/gfunc/subconvergo/proxy/utils"
)

type HttpProxy struct {
	core.BaseProxy `yaml:",inline"`
	Username       string `yaml:"username" json:"username"`
	Password       string `yaml:"password" json:"password"`
	Tls            bool   `yaml:"tls" json:"tls"`
	SkipCertVerify bool   `yaml:"skip-cert-verify" json:"skip-cert-verify"`
}

func (p *HttpProxy) ToSingleConfig(ext *config.ProxySetting) (string, error) {
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

func (p *HttpProxy) ToClashConfig(ext *config.ProxySetting) (map[string]interface{}, error) {
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
		if p.SkipCertVerify || ext.SCV {
			options["skip-cert-verify"] = true
		}
	}
	return options, nil
}

func (p *HttpProxy) ToSurgeConfig(ext *config.ProxySetting) (string, error) {
	scheme := "http"
	parts := []string{scheme, p.Server, fmt.Sprintf("%d", p.Port)}
	parts = append(parts, fmt.Sprintf("username=%s", p.Username))
	parts = append(parts, fmt.Sprintf("password=%s", p.Password))
	if p.Tls {
		parts = append(parts, "tls=true")
	}
	if ext.TFO {
		parts = append(parts, "tfo=true")
	}
	if p.Tls && (p.SkipCertVerify || ext.SCV) {
		parts = append(parts, "skip-cert-verify=true")
	}
	return fmt.Sprintf("%s = %s", p.Remark, strings.Join(parts, ", ")), nil
}

func (p *HttpProxy) ToLoonConfig(ext *config.ProxySetting) (string, error) {
	scheme := "http"
	if p.Tls {
		scheme = "https"
	}
	// Format: http,server,port,username,"password"
	part := fmt.Sprintf("%s,%s,%d,%s,\"%s\"", scheme, p.Server, p.Port, p.Username, p.Password)
	if p.Tls && (p.SkipCertVerify || ext.SCV) {
		part += ",skip-cert-verify=true"
	}
	return fmt.Sprintf("%s = %s", p.Remark, part), nil
}

func (p *HttpProxy) ToQuantumultXConfig(ext *config.ProxySetting) (string, error) {
	parts := []string{"http=" + fmt.Sprintf("%s:%d", p.Server, p.Port)}
	if p.Username != "" {
		parts = append(parts, fmt.Sprintf("username=%s", p.Username))
	} else {
		parts = append(parts, "username=none")
	}
	if p.Password != "" {
		parts = append(parts, fmt.Sprintf("password=%s", p.Password))
	} else {
		parts = append(parts, "password=none")
	}
	if p.Tls {
		parts = append(parts, "over-tls=true")
	} else {
		parts = append(parts, "over-tls=false")
	}
	if ext.TFO {
		parts = append(parts, "fast-open=true")
	}
	parts = append(parts, fmt.Sprintf("tag=%s", p.Remark))
	return strings.Join(parts, ", "), nil
}

func (p *HttpProxy) ToSingboxConfig(ext *config.ProxySetting) (map[string]interface{}, error) {
	outbound := map[string]interface{}{
		"type":        "http",
		"tag":         p.Remark,
		"server":      p.Server,
		"server_port": p.Port,
		"username":    p.Username,
		"password":    p.Password,
	}
	if p.Tls {
		outbound["tls"] = map[string]interface{}{
			"enabled":  true,
			"insecure": ext.SCV,
		}
	}
	return outbound, nil
}
