package impl

import (
	"encoding/base64"
	"fmt"
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
	link := fmt.Sprintf("ss://%s@%s:%d#%s", encoded, p.Server, p.Port, core.UrlEncode(p.Remark))
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

// ShadowsocksRProxy represents a ShadowsocksR proxy
type ShadowsocksRProxy struct {
	core.BaseProxy `yaml:",inline"`
	Password       string `yaml:"password" json:"password"`
	EncryptMethod  string `yaml:"encrypt_method" json:"encrypt_method"`
	Protocol       string `yaml:"protocol" json:"protocol"`
	ProtocolParam  string `yaml:"protocol_param" json:"protocol_param"`
	Obfs           string `yaml:"obfs" json:"obfs"`
	ObfsParam      string `yaml:"obfs_param" json:"obfs_param"`
}

func (p *ShadowsocksRProxy) ToShareLink(ext *config.ProxySetting) (string, error) {
	// Format: ssr://base64(server:port:protocol:method:obfs:base64(password)/?...)
	passEncoded := base64.URLEncoding.EncodeToString([]byte(p.Password))
	mainPart := fmt.Sprintf("%s:%d:%s:%s:%s:%s",
		p.Server,
		p.Port,
		p.Protocol,
		p.EncryptMethod,
		p.Obfs,
		passEncoded,
	)

	// Add query parameters
	params := []string{}
	if p.ObfsParam != "" {
		params = append(params, fmt.Sprintf("obfsparam=%s", base64.URLEncoding.EncodeToString([]byte(p.ObfsParam))))
	}
	if p.ProtocolParam != "" {
		params = append(params, fmt.Sprintf("protoparam=%s", base64.URLEncoding.EncodeToString([]byte(p.ProtocolParam))))
	}
	if p.Remark != "" {
		params = append(params, fmt.Sprintf("remarks=%s", base64.URLEncoding.EncodeToString([]byte(p.Remark))))
	}

	if len(params) > 0 {
		mainPart += "/?" + strings.Join(params, "&")
	}

	encoded := base64.URLEncoding.EncodeToString([]byte(mainPart))
	return "ssr://" + encoded, nil
}

func (p *ShadowsocksRProxy) ToClashConfig(ext *config.ProxySetting) (options map[string]interface{}) {
	if p.Type == "ss" {
		options = map[string]interface{}{
			"type":     "ss",
			"name":     p.Remark,
			"server":   p.Server,
			"port":     p.Port,
			"cipher":   p.EncryptMethod,
			"password": p.Password,
		}
	} else {
		options = map[string]interface{}{
			"type":           "ssr",
			"name":           p.Remark,
			"server":         p.Server,
			"port":           p.Port,
			"cipher":         p.EncryptMethod,
			"password":       p.Password,
			"protocol":       p.Protocol,
			"obfs":           p.Obfs,
			"protocol-param": p.ProtocolParam,
			"obfs-param":     p.ObfsParam,
		}
	}
	return options
}
