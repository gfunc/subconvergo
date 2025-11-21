package impl

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/proxy/core"
)

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
