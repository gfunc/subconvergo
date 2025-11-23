package sub

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/gfunc/subconvergo/parser/core"
	proxyCore "github.com/gfunc/subconvergo/proxy/core"
)

// Base64SubscriptionParser parses base64 encoded subscriptions
type Base64SubscriptionParser struct{}

func (p *Base64SubscriptionParser) Name() string {
	return "Base64"
}

func (p *Base64SubscriptionParser) CanParse(content string) bool {
	// We can try to decode a small part to check if it's base64
	// But for now, let's assume it's a fallback or we check if it's NOT clash/etc.
	// Or we can just return true and let Parse fail.
    // However, ParseSubscription iterates and returns the first one that CanParse.
    // So we should be careful.
    // Let's try to decode the first line or the whole content if short.
    trimmed := strings.TrimSpace(content)
    if len(trimmed) == 0 {
        return false
    }
    // Check if it contains spaces (base64 usually doesn't, unless it's multiple lines of base64?)
    // Standard subscription is one big base64 string.
    if strings.ContainsAny(trimmed, " \t") {
        return false
    }
    // Try decoding a chunk
    chunk := trimmed
    if len(chunk) > 100 {
        chunk = chunk[:100]
    }
    _, err := base64.StdEncoding.DecodeString(chunk)
    return err == nil
}

func (p *Base64SubscriptionParser) Parse(content string) (*core.SubContent, error) {
	decoded, err := base64.StdEncoding.DecodeString(content)
	if err == nil {
		content = string(decoded)
	}

	var proxies []proxyCore.ProxyInterface
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if proxy, err := core.ParseProxy(line); err == nil {
			proxies = append(proxies, proxy)
		}
	}

    if len(proxies) == 0 {
        return nil, fmt.Errorf("no proxies found in base64 subscription")
    }

	return &core.SubContent{
		Proxies: proxies,
	}, nil
}

// PlainSubscriptionParser parses plain text subscriptions (line by line)
type PlainSubscriptionParser struct{}

func (p *PlainSubscriptionParser) Name() string {
	return "Plain"
}

func (p *PlainSubscriptionParser) CanParse(content string) bool {
	// Check if it contains common protocol prefixes
	return strings.Contains(content, "://")
}

func (p *PlainSubscriptionParser) Parse(content string) (*core.SubContent, error) {
	var proxies []proxyCore.ProxyInterface
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if proxy, err := core.ParseProxy(line); err == nil {
			proxies = append(proxies, proxy)
		}
	}

	if len(proxies) == 0 {
		return nil, fmt.Errorf("no proxies found in plain subscription")
	}

	return &core.SubContent{
		Proxies: proxies,
	}, nil
}
