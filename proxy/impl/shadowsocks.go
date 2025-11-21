package impl

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/proxy/core"
)

// ShadowsocksProxy represents a Shadowsocks proxy
type ShadowsocksProxy struct {
	core.BaseProxy `yaml:",inline"`
	Password       string                 `yaml:"password" json:"password"`
	EncryptMethod  string                 `yaml:"encrypt_method" json:"encrypt_method"`
	Plugin         string                 `yaml:"plugin" json:"plugin"`
	PluginOpts     map[string]interface{} `yaml:"plugin_opts" json:"plugin_opts"`
}

func (p *ShadowsocksProxy) ToShareLink(ext *config.ProxySetting) (string, error) {
	// Format: ss://base64(method:password)@server:port#remark
	userInfo := fmt.Sprintf("%s:%s", p.EncryptMethod, p.Password)
	encoded := base64.URLEncoding.EncodeToString([]byte(userInfo))
	link := fmt.Sprintf("ss://%s@%s:%d", encoded, p.Server, p.Port)

	if p.Plugin != "" {
		pluginStr := p.Plugin
		if len(p.PluginOpts) > 0 {
			var opts []string
			for k, v := range p.PluginOpts {
				opts = append(opts, fmt.Sprintf("%s=%v", k, v))
			}
			sort.Strings(opts)
			pluginStr += ";" + strings.Join(opts, ";")
		}
		link += fmt.Sprintf("/?plugin=%s", url.QueryEscape(pluginStr))
	}

	link += fmt.Sprintf("#%s", core.UrlEncode(p.Remark))
	return link, nil
}

func (p *ShadowsocksProxy) ToClashConfig(ext *config.ProxySetting) map[string]interface{} {
	options := map[string]interface{}{
		"type":     "ss",
		"name":     p.Remark,
		"server":   p.Server,
		"port":     p.Port,
		"cipher":   p.EncryptMethod,
		"password": p.Password,
	}

	if ext != nil && ext.UDP {
		options["udp"] = true
	}

	if p.Plugin != "" {
		opts := make(map[string]interface{})
		if len(p.PluginOpts) != 0 {

			switch p.Plugin {
			case "simple-obfs":
			case "obfs-local":
				options["plugin"] = "obfs"
				opts["mode"] = p.PluginOpts["obfs"]
				opts["host"] = p.PluginOpts["obfs-host"]
			case "v2ray-plugin":
				options["plugin"] = "v2ray-plugin"
				opts["mode"] = p.PluginOpts["mode"]
				opts["host"] = p.PluginOpts["host"]
				opts["path"] = p.PluginOpts["path"]
				opts["tls"] = p.PluginOpts["tls"]
				opts["mux"] = p.PluginOpts["mux"]
				opts["skip-cert-verify"] = ext.SCV
			}
			options["plugin-opts"] = opts
		}
	}

	return options
}
