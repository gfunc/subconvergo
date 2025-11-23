package sub

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/gfunc/subconvergo/parser/core"
	"github.com/gfunc/subconvergo/parser/proxy"
	proxyCore "github.com/gfunc/subconvergo/proxy/core"
)

type SSDSubscriptionParser struct{}

func (p *SSDSubscriptionParser) Name() string {
	return "SSD"
}

func (p *SSDSubscriptionParser) CanParse(content string) bool {
	return strings.HasPrefix(content, "ssd://")
}

func (p *SSDSubscriptionParser) Parse(content string) (*core.SubContent, error) {
	base64Str := strings.TrimPrefix(content, "ssd://")
	decoded, err := base64.RawURLEncoding.DecodeString(base64Str)
	if err != nil {
		// Try standard encoding if raw fails
		decoded, err = base64.URLEncoding.DecodeString(base64Str)
		if err != nil {
			return nil, fmt.Errorf("failed to decode SSD subscription: %w", err)
		}
	}

	var ssdData struct {
		Airport string          `json:"airport"`
		Port    interface{}     `json:"port"`
		Encrypt string          `json:"encryption"`
		Pass    string          `json:"password"`
		Plugin  string          `json:"plugin"`
		PluginO string          `json:"plugin_options"`
		Servers json.RawMessage `json:"servers"`
	}

	if err := json.Unmarshal(decoded, &ssdData); err != nil {
		return nil, fmt.Errorf("failed to parse SSD JSON: %w", err)
	}

	var proxies []proxyCore.ProxyInterface
	var servers []map[string]interface{}

	// servers can be array or object (map)
	if err := json.Unmarshal(ssdData.Servers, &servers); err != nil {
		// Try parsing as object (legacy SSD?)
		// subconverter: listType = 1
		// But Go's json.Unmarshal won't automatically convert object to slice.
		// Let's try unmarshalling to map
		var serversMap map[string]map[string]interface{}
		if err2 := json.Unmarshal(ssdData.Servers, &serversMap); err2 == nil {
			// Convert map to slice, keys might be indices or names
			// subconverter iterates map and uses key as index? No, it just iterates.
			for _, s := range serversMap {
				servers = append(servers, s)
			}
		} else {
			return nil, fmt.Errorf("failed to parse SSD servers: %w", err)
		}
	}

	defaultPort := toString(ssdData.Port)
	defaultEncrypt := ssdData.Encrypt
	defaultPass := ssdData.Pass
	defaultPlugin := ssdData.Plugin
	defaultPluginO := ssdData.PluginO

	for _, s := range servers {
		// Resolve defaults
		if toString(s["port"]) == "" {
			s["port"] = defaultPort
		}
		if toString(s["encryption"]) == "" {
			s["encryption"] = defaultEncrypt
		}
		if toString(s["password"]) == "" {
			s["password"] = defaultPass
		}
		if toString(s["plugin"]) == "" {
			s["plugin"] = defaultPlugin
		}
		if toString(s["plugin_options"]) == "" {
			s["plugin_options"] = defaultPluginO
		}

		s["airport"] = ssdData.Airport

		parser := &proxy.ShadowsocksParser{}
		if ss, err := parser.ParseSSD(s); err == nil {
			proxies = append(proxies, ss)
		}
	}

	return &core.SubContent{
		Proxies: proxies,
	}, nil
}

func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	case int:
		return strconv.Itoa(val)
	default:
		return fmt.Sprintf("%v", val)
	}
}
