package impl

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/proxy/core"
	"github.com/gfunc/subconvergo/proxy/utils"
)

// ShadowsocksProxy represents a Shadowsocks proxy
type ShadowsocksProxy struct {
	core.BaseProxy `yaml:",inline"`
	Password       string                 `yaml:"password" json:"password"`
	EncryptMethod  string                 `yaml:"encrypt_method" json:"encrypt_method"`
	Plugin         string                 `yaml:"plugin" json:"plugin"`
	PluginOpts     map[string]interface{} `yaml:"plugin_opts" json:"plugin_opts"`
}

func (p *ShadowsocksProxy) ToSingleConfig(ext *config.ProxySetting) (string, error) {
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

	link += fmt.Sprintf("#%s", utils.UrlEncode(p.Remark))
	return link, nil
}

func (p *ShadowsocksProxy) ToClashConfig(ext *config.ProxySetting) (map[string]interface{}, error) {
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

	return options, nil
}

func (p *ShadowsocksProxy) ToSurgeConfig(ext *config.ProxySetting) (string, error) {
	if p.Plugin == "v2ray-plugin" {
		return "", fmt.Errorf("v2ray-plugin not supported in Surge")
	}
	parts := []string{"ss", p.Server, fmt.Sprintf("%d", p.Port)}
	parts = append(parts, fmt.Sprintf("encrypt-method=%s", p.EncryptMethod))
	parts = append(parts, fmt.Sprintf("password=%s", p.Password))
	if ext.UDP {
		parts = append(parts, "udp-relay=true")
	}
	if ext.TFO {
		parts = append(parts, "tfo=true")
	}
	if p.Plugin == "obfs-local" || p.Plugin == "simple-obfs" {
		if mode, ok := p.PluginOpts["obfs"]; ok {
			parts = append(parts, fmt.Sprintf("obfs=%s", mode))
		}
		if host, ok := p.PluginOpts["obfs-host"]; ok {
			parts = append(parts, fmt.Sprintf("obfs-host=%s", host))
		}
	}
	return fmt.Sprintf("%s = %s", p.Remark, strings.Join(parts, ", ")), nil
}

func (p *ShadowsocksProxy) ToLoonConfig(ext *config.ProxySetting) (string, error) {
	// Format: Name = Shadowsocks,server,port,method,"password"
	parts := []string{"Shadowsocks", p.Server, fmt.Sprintf("%d", p.Port), p.EncryptMethod, fmt.Sprintf("\"%s\"", p.Password)}

	if p.Plugin == "simple-obfs" || p.Plugin == "obfs-local" || p.Plugin == "obfs" {
		if mode, ok := p.PluginOpts["obfs"]; ok {
			parts = append(parts, fmt.Sprintf("%v", mode))
		} else if mode, ok := p.PluginOpts["mode"]; ok {
			parts = append(parts, fmt.Sprintf("%v", mode))
		}
		if host, ok := p.PluginOpts["obfs-host"]; ok {
			parts = append(parts, fmt.Sprintf("%v", host))
		} else if host, ok := p.PluginOpts["host"]; ok {
			parts = append(parts, fmt.Sprintf("%v", host))
		}
	} else if p.Plugin != "" {
		return "", fmt.Errorf("plugin %s not supported in Loon", p.Plugin)
	}

	return fmt.Sprintf("%s = %s", p.Remark, strings.Join(parts, ",")), nil
}

func (p *ShadowsocksProxy) ToQuantumultXConfig(ext *config.ProxySetting) (string, error) {
	var parts []string
	parts = append(parts, "shadowsocks="+fmt.Sprintf("%s:%d", p.Server, p.Port))
	parts = append(parts, fmt.Sprintf("method=%s", p.EncryptMethod))
	parts = append(parts, fmt.Sprintf("password=%s", p.Password))
	if ext.UDP {
		parts = append(parts, "udp-relay=true")
	}
	if ext.TFO {
		parts = append(parts, "fast-open=true")
	}
	if p.Plugin == "obfs-local" || p.Plugin == "simple-obfs" {
		if mode, ok := p.PluginOpts["obfs"]; ok {
			parts = append(parts, fmt.Sprintf("obfs=%s", mode))
		}
		if host, ok := p.PluginOpts["obfs-host"]; ok {
			parts = append(parts, fmt.Sprintf("obfs-host=%s", host))
		}
	}
	parts = append(parts, fmt.Sprintf("tag=%s", p.Remark))
	return strings.Join(parts, ", "), nil
}

func (p *ShadowsocksProxy) ToSingboxConfig(ext *config.ProxySetting) (map[string]interface{}, error) {
	outbound := map[string]interface{}{
		"type":        "shadowsocks",
		"tag":         p.Remark,
		"server":      p.Server,
		"server_port": p.Port,
		"method":      p.EncryptMethod,
		"password":    p.Password,
	}
	if p.Plugin != "" {
		outbound["plugin"] = p.Plugin
		outbound["plugin_opts"] = p.PluginOpts
	}
	return outbound, nil
}
