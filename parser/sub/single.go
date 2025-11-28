package sub

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/gfunc/subconvergo/parser/core"
	"github.com/gfunc/subconvergo/parser/proxy"
	proxyCore "github.com/gfunc/subconvergo/proxy/core"
)

// SingleSubscriptionParser parses base64 encoded or plain text subscriptions
type SingleSubscriptionParser struct{}

func (p *SingleSubscriptionParser) Name() string {
	return "General"
}

func (p *SingleSubscriptionParser) CanParse(content string) bool {
	// Always try to parse as general subscription if other parsers failed
	return true
}

func (p *SingleSubscriptionParser) Parse(content string) (*core.SubContent, error) {
	// Try to decode base64 first
	decoded, err := decodeBase64(content)
	if err == nil && len(decoded) > 0 {
		content = decoded
	}

	var proxies []proxyCore.ProxyInterface
	// Split by newline or space (subconverter logic handles \r, \n)
	// subconverter: while(getline(strstream, strLink, delimiter))
	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		p, err := proxy.ParseProxy(line)
		if err != nil {
			// Fallback to Mihomo
			idx := strings.Index(line, "://")
			if idx != -1 {
				protocol := line[:idx]
				p, err = ParseMihomoProxy(protocol, line[idx:])
			}
		}

		if err == nil && p != nil {
			proxies = append(proxies, p)
		}
	}

	if len(proxies) == 0 {
		return nil, fmt.Errorf("no proxies found in subscription")
	}

	return &core.SubContent{
		Proxies: proxies,
	}, nil
}

func decodeBase64(s string) (string, error) {
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return "", fmt.Errorf("empty string")
	}
	// Pad with =
	if m := len(s) % 4; m != 0 {
		s += strings.Repeat("=", 4-m)
	}
	// Try standard
	b, err := base64.StdEncoding.DecodeString(s)
	if err == nil {
		return string(b), nil
	}
	// Try URL safe
	b, err = base64.URLEncoding.DecodeString(s)
	if err == nil {
		return string(b), nil
	}
	return "", err
}
