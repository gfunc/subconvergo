package impl

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/proxy/core"
)

// AnyTLSProxy represents an AnyTLS proxy
type AnyTLSProxy struct {
	core.BaseProxy           `yaml:",inline"`
	Password                 string   `yaml:"password" json:"password"`
	SNI                      string   `yaml:"sni" json:"sni"`
	Alpn                     []string `yaml:"alpn" json:"alpn"`
	Fingerprint              string   `yaml:"fingerprint" json:"fingerprint"`
	IdleSessionCheckInterval int      `yaml:"idle_session_check_interval" json:"idle_session_check_interval"`
	IdleSessionTimeout       int      `yaml:"idle_session_timeout" json:"idle_session_timeout"`
	MinIdleSession           int      `yaml:"min_idle_session" json:"min_idle_session"`
	TFO                      bool     `yaml:"tfo" json:"tfo"`
	AllowInsecure            bool     `yaml:"allow_insecure" json:"allow_insecure"`
}

func (p *AnyTLSProxy) ToShareLink(ext *config.ProxySetting) (string, error) {
	// Format: anytls://password@server:port?peer=sni&alpn=alpn&hpkp=fingerprint&tfo=1&insecure=1#remark
	link := fmt.Sprintf("anytls://%s@%s:%d", p.Password, p.Server, p.Port)

	values := url.Values{}
	if p.SNI != "" {
		values.Add("peer", p.SNI)
	}
	if len(p.Alpn) > 0 {
		values.Add("alpn", strings.Join(p.Alpn, ","))
	}
	if p.Fingerprint != "" {
		values.Add("hpkp", p.Fingerprint)
	}
	if p.TFO {
		values.Add("tfo", "1")
	}
	if p.AllowInsecure {
		values.Add("insecure", "1")
	}
	if p.IdleSessionCheckInterval != 0 {
		values.Add("idle_session_check_interval", fmt.Sprintf("%d", p.IdleSessionCheckInterval))
	}
	if p.IdleSessionTimeout != 0 {
		values.Add("idle_session_timeout", fmt.Sprintf("%d", p.IdleSessionTimeout))
	}
	if p.MinIdleSession != 0 {
		values.Add("min_idle_session", fmt.Sprintf("%d", p.MinIdleSession))
	}

	if len(values) > 0 {
		link += "?" + values.Encode()
	}

	if p.Remark != "" {
		link += "#" + core.UrlEncode(p.Remark)
	}

	return link, nil
}

func (p *AnyTLSProxy) ToClashConfig(ext *config.ProxySetting) map[string]interface{} {
	options := map[string]interface{}{
		"type":   "anytls",
		"name":   p.Remark,
		"server": p.Server,
		"port":   p.Port,
	}

	if p.Password != "" {
		options["password"] = p.Password
	}
	if p.SNI != "" {
		options["sni"] = p.SNI
	}
	if len(p.Alpn) > 0 {
		options["alpn"] = p.Alpn
	}
	if p.Fingerprint != "" {
		options["fingerprint"] = p.Fingerprint
	}
	if p.IdleSessionCheckInterval != 0 {
		options["idle-session-check-interval"] = p.IdleSessionCheckInterval
	}
	if p.IdleSessionTimeout != 0 {
		options["idle-session-timeout"] = p.IdleSessionTimeout
	}
	if p.MinIdleSession != 0 {
		options["min-idle-session"] = p.MinIdleSession
	}
	if p.AllowInsecure {
		options["skip-cert-verify"] = true
	}
	if p.TFO {
		options["tfo"] = true
	}

	return options
}
