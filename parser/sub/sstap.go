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

type SSTapSubscriptionParser struct{}

func (p *SSTapSubscriptionParser) Name() string {
	return "SSTap"
}

func (p *SSTapSubscriptionParser) CanParse(content string) bool {
	return strings.Contains(content, "\"idInUse\"")
}

func (p *SSTapSubscriptionParser) Parse(content string) (*core.SubContent, error) {
	var js map[string]interface{}
	if err := json.Unmarshal([]byte(content), &js); err != nil {
		return nil, fmt.Errorf("failed to parse SSTap JSON: %w", err)
	}

	proxiesList, ok := js["proxies"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("proxies is not a list")
	}

	var proxies []proxyCore.ProxyInterface

	for _, item := range proxiesList {
		cfg, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		mode := utils.ToString(cfg["mode"]) // type

		// SSTap types: ss, ssr, socks5, http
		switch mode {
		case "ss":
			parser := &proxy.ShadowsocksParser{}
			if ss, err := parser.ParseSSTap(cfg); err == nil {
				proxies = append(proxies, ss)
			}
		case "ssr":
			parser := &proxy.ShadowsocksRParser{}
			if ssr, err := parser.ParseSSTap(cfg); err == nil {
				proxies = append(proxies, ssr)
			}
		case "socks5":
			parser := &proxy.Socks5Parser{}
			if socks, err := parser.ParseSSTap(cfg); err == nil {
				proxies = append(proxies, socks)
			}
		case "http":
			parser := &proxy.HttpParser{}
			if http, err := parser.ParseSSTap(cfg); err == nil {
				proxies = append(proxies, http)
			}
		}
	}

	return &core.SubContent{
		Proxies: proxies,
	}, nil
}
