package sub

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/parser/core"
	"github.com/gfunc/subconvergo/parser/proxy"
	proxyCore "github.com/gfunc/subconvergo/proxy/core"
	"github.com/gfunc/subconvergo/proxy/impl"
	"github.com/metacubex/mihomo/adapter"
	C "github.com/metacubex/mihomo/config"
	"gopkg.in/yaml.v3"
)

type ClashSubscriptionParser struct{}

func (p *ClashSubscriptionParser) Name() string {
	return "Clash"
}

func (p *ClashSubscriptionParser) CanParse(content string) bool {
	return strings.Contains(content, "proxies:") || strings.Contains(content, "Proxy:")
}

func (p *ClashSubscriptionParser) Parse(content string) (*core.SubContent, error) {
	return ParseMihomoConfig(content)
}

// ParseMihomoConfig parses Clash YAML format and returns proxies
func ParseMihomoConfig(content string) (*core.SubContent, error) {
	// Try parsing as YAML (Clash format)
	var clashConfig C.RawConfig

	if err := yaml.Unmarshal([]byte(content), &clashConfig); err != nil {
		return nil, fmt.Errorf("failed to parse clash format: %w", err)
	}

	if len(clashConfig.Proxy) == 0 {
		return nil, fmt.Errorf("no proxies found in clash format")
	}
	custom := &core.SubContent{Proxies: make([]proxyCore.ProxyInterface, 0)}

	for _, proxyMap := range clashConfig.Proxy {
		proxyType, ok := proxyMap["type"].(string)
		if !ok {
			continue
		}

		var p proxyCore.ProxyInterface
		var err error

		switch proxyType {
		case "ss":
			parser := &proxy.ShadowsocksParser{}
			p, err = parser.ParseClash(proxyMap)
		case "ssr":
			parser := &proxy.ShadowsocksRParser{}
			p, err = parser.ParseClash(proxyMap)
		case "vmess":
			parser := &proxy.VMessParser{}
			p, err = parser.ParseClash(proxyMap)
		case "trojan":
			parser := &proxy.TrojanParser{}
			p, err = parser.ParseClash(proxyMap)
		case "snell":
			parser := &proxy.SnellParser{}
			p, err = parser.ParseClash(proxyMap)
		case "http":
			parser := &proxy.HttpParser{}
			p, err = parser.ParseClash(proxyMap)
		case "socks5":
			parser := &proxy.Socks5Parser{}
			p, err = parser.ParseClash(proxyMap)
		default:
			p, err = parseMihomoProxy(proxyMap)
		}

		if err != nil {
			log.Printf("failed to parse proxy in clash format: %v", err)
			continue
		}
		custom.Proxies = append(custom.Proxies, p)
	}
	custom.Groups = make([]config.ProxyGroupConfig, 0)
	for _, groupMap := range clashConfig.ProxyGroup {
		groupBytes, err := json.Marshal(groupMap)
		if err != nil {
			log.Printf("failed to marshal proxy group: %v", err)
			continue
		}
		var proxyGroup config.ProxyGroupConfig
		if err := json.Unmarshal(groupBytes, &proxyGroup); err != nil {
			log.Printf("failed to unmarshal proxy group: %v", err)
			continue
		}
		if len(proxyGroup.Proxies) > 0 {
			proxyGroup.Rule = make([]string, len(proxyGroup.Proxies))
			for i, p := range proxyGroup.Proxies {
				proxyGroup.Rule[i] = fmt.Sprintf("[]%s", p)
			}
		}
		custom.Groups = append(custom.Groups, proxyGroup)
	}

	custom.RawRules = clashConfig.Rule
	return custom, nil
}

func ParseMihomoProxy(protocol, content string) (*impl.MihomoProxy, error) {
	// Try base64 decode first (standard subscription format)
	decoded, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		decoded = []byte(content)
	}
	options := make(map[string]any)
	if err := yaml.Unmarshal(decoded, &options); err != nil {
		return nil, fmt.Errorf("failed to parse clash proxy: %s://%s, %w", protocol, content, err)
	}
	if _, ok := options["type"]; !ok {
		options["type"] = protocol
	}
	return parseMihomoProxy(options)
}

func parseMihomoProxy(options map[string]any) (*impl.MihomoProxy, error) {
	mihomoProxy, err := adapter.ParseProxy(options)
	if err != nil {
		return nil, err
	}
	addr := mihomoProxy.Addr()
	// parse addr to server and port
	server, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, fmt.Errorf("invalid address format: %s, %w", addr, err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid port in address: %s", addr)
	}
	return &impl.MihomoProxy{
		ProxyInterface: &proxyCore.BaseProxy{
			Type:   options["type"].(string),
			Remark: mihomoProxy.Name(),
			Server: server,
			Port:   port,
		},
		Clash:   mihomoProxy,
		Options: options,
	}, nil
}
