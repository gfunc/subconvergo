package sub

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gfunc/subconvergo/parser/core"
	"github.com/gfunc/subconvergo/parser/proxy"
	proxyCore "github.com/gfunc/subconvergo/proxy/core"
)

type SSAndroidSubscriptionParser struct{}

func (p *SSAndroidSubscriptionParser) Name() string {
	return "SSAndroid"
}

func (p *SSAndroidSubscriptionParser) CanParse(content string) bool {
	return strings.Contains(content, "\"proxy_apps\"")
}

func (p *SSAndroidSubscriptionParser) Parse(content string) (*core.SubContent, error) {
	// Wrap content in {"nodes": ...} as per subconverter logic
	wrappedContent := fmt.Sprintf(`{"nodes":%s}`, content)

	var js map[string]interface{}
	if err := json.Unmarshal([]byte(wrappedContent), &js); err != nil {
		return nil, fmt.Errorf("failed to parse SSAndroid JSON: %w", err)
	}

	nodesList, ok := js["nodes"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("nodes is not a list")
	}

	var proxies []proxyCore.ProxyInterface
	group := proxyCore.SS_DEFAULT_GROUP

	for _, item := range nodesList {
		cfg, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		parser := &proxy.ShadowsocksParser{}
		p, err := parser.ParseSSAndroid(cfg)
		if err == nil {
			p.SetGroup(group)
			proxies = append(proxies, p)
		}
	}

	return &core.SubContent{
		Proxies: proxies,
	}, nil
}
