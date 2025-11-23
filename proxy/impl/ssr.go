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

func (p *ShadowsocksRProxy) ToSingleConfig(ext *config.ProxySetting) (string, error) {
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

func (p *ShadowsocksRProxy) ToClashConfig(ext *config.ProxySetting) (map[string]interface{}, error) {
	var options map[string]interface{}
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
	return options, nil
}

func (p *ShadowsocksRProxy) ToSurgeConfig(ext *config.ProxySetting) (string, error) {
	return "", fmt.Errorf("SSR not supported in Surge")
}

func (p *ShadowsocksRProxy) ToLoonConfig(ext *config.ProxySetting) (string, error) {
	// Format: ssr,server,port,encrypt_method,"password",protocol,protocol_param,obfs,obfs_param
	parts := []string{"ssr", p.Server, fmt.Sprintf("%d", p.Port), p.EncryptMethod, fmt.Sprintf("\"%s\"", p.Password), p.Protocol}

	// Loon SSR format can be tricky with optional params, usually:
	// protocol-param, obfs, obfs-param
	// Assuming standard order if supported.
	// Actually Loon supports SSR.

	// protocol_param
	if p.ProtocolParam != "" {
		parts = append(parts, fmt.Sprintf("protocol-param=\"%s\"", p.ProtocolParam))
	}
	parts = append(parts, p.Obfs)
	if p.ObfsParam != "" {
		parts = append(parts, fmt.Sprintf("obfs-param=\"%s\"", p.ObfsParam))
	}

	return fmt.Sprintf("%s = %s", p.Remark, strings.Join(parts, ",")), nil
}

func (p *ShadowsocksRProxy) ToQuantumultXConfig(ext *config.ProxySetting) (string, error) {
	// Format: shadowsock=server:port, method=method, password=password, ssr-protocol=protocol, ssr-protocol-param=param, obfs=obfs, obfs-host=param, tag=tag
	parts := []string{fmt.Sprintf("shadowsocks=%s:%d", p.Server, p.Port)}
	parts = append(parts, fmt.Sprintf("method=%s", p.EncryptMethod))
	parts = append(parts, fmt.Sprintf("password=%s", p.Password))
	parts = append(parts, fmt.Sprintf("ssr-protocol=%s", p.Protocol))
	if p.ProtocolParam != "" {
		parts = append(parts, fmt.Sprintf("ssr-protocol-param=%s", p.ProtocolParam))
	}
	parts = append(parts, fmt.Sprintf("obfs=%s", p.Obfs))
	if p.ObfsParam != "" {
		parts = append(parts, fmt.Sprintf("obfs-host=%s", p.ObfsParam))
	}

	if ext.TFO {
		parts = append(parts, "fast-open=true")
	}
	if ext.UDP {
		parts = append(parts, "udp-relay=true")
	}
	parts = append(parts, fmt.Sprintf("tag=%s", p.Remark))
	return strings.Join(parts, ", "), nil
}

func (p *ShadowsocksRProxy) ToSingboxConfig(ext *config.ProxySetting) (map[string]interface{}, error) {
	return nil, fmt.Errorf("SSR not supported in sing-box")
}
