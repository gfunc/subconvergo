package sub

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gfunc/subconvergo/parser/core"
	"github.com/gfunc/subconvergo/parser/proxy"
	"github.com/gfunc/subconvergo/parser/utils"
	proxyCore "github.com/gfunc/subconvergo/proxy/core"
)

type NetchSubscriptionParser struct{}

func (p *NetchSubscriptionParser) Name() string {
	return "Netch"
}

func (p *NetchSubscriptionParser) CanParse(content string) bool {
	return strings.Contains(content, "\"ModeFileNameType\"")
}

func (p *NetchSubscriptionParser) Parse(content string) (*core.SubContent, error) {
	var js map[string]interface{}
	if err := json.Unmarshal([]byte(content), &js); err != nil {
		return nil, fmt.Errorf("failed to parse Netch JSON: %w", err)
	}

	servers, ok := js["Server"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("Server is not a list")
	}

	var proxies []proxyCore.ProxyInterface

	for _, item := range servers {
		cfg, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		typeVal := utils.ToInt(cfg["Type"])

		switch typeVal {
		case 1: // SS
			parser := &proxy.ShadowsocksParser{}
			if ss, err := parser.ParseNetch(cfg); err == nil {
				proxies = append(proxies, ss)
			}
		case 2: // SSR
			parser := &proxy.ShadowsocksRParser{}
			if ssr, err := parser.ParseNetch(cfg); err == nil {
				proxies = append(proxies, ssr)
			}
		case 3: // VMess
			parser := &proxy.VMessParser{}
			if vmess, err := parser.ParseNetch(cfg); err == nil {
				proxies = append(proxies, vmess)
			}
		case 4: // Socks5
			parser := &proxy.Socks5Parser{}
			if socks, err := parser.ParseNetch(cfg); err == nil {
				proxies = append(proxies, socks)
			}
		}
	}

	return &core.SubContent{
		Proxies: proxies,
	}, nil
}
