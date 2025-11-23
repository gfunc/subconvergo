package core

import (
	"fmt"
	"sync"

	"github.com/gfunc/subconvergo/proxy/core"
)

var (
	proxyParsers []ProxyParser
	subParsers   []SubscriptionParser
	mu           sync.RWMutex
)

// RegisterParser adds a proxy parser to the registry
func RegisterParser(p ProxyParser) {
	mu.Lock()
	defer mu.Unlock()
	proxyParsers = append(proxyParsers, p)
}

// RegisterSubscriptionParser adds a subscription parser to the registry
func RegisterSubscriptionParser(p SubscriptionParser) {
	mu.Lock()
	defer mu.Unlock()
	subParsers = append(subParsers, p)
}

// ParseProxy tries to parse a proxy config using registered parsers
func ParseProxy(content string) (core.SubconverterProxy, error) {
	mu.RLock()
	defer mu.RUnlock()

	for _, p := range proxyParsers {
		if matcher, ok := p.(LineMatcher); ok {
			if matcher.CanParseLine(content) {
				return p.Parse(content)
			}
		}
	}
	return nil, fmt.Errorf("no parser found for content: %s", content)
}

// ParseSubscription tries to parse a subscription using registered parsers
func ParseSubscription(content string) (*SubContent, error) {
	mu.RLock()
	defer mu.RUnlock()

	for _, p := range subParsers {
		if p.CanParse(content) {
			return p.Parse(content)
		}
	}
	return nil, fmt.Errorf("no subscription parser found")
}
