// Package parser provides subscription parsing functionality for various proxy protocols.
//
// This package implements parsing for common proxy share link formats:
//   - Shadowsocks (ss://)
//   - ShadowsocksR (ssr://)
//   - VMess (vmess://)
//   - Trojan (trojan://)
//   - VLESS (vless://)
//   - Hysteria (hysteria://, hysteria2://, hy2://)
//   - TUIC (tuic://)
//   - Clash YAML format
//
// All parsers validate proxies using the mihomo adapter to ensure compatibility
// and correctness. The parsing behavior matches the subconverter C++ implementation.
//
// For protocols not explicitly implemented, the package attempts to use mihomo's
// built-in parsers as a fallback, allowing support for additional protocols that
// mihomo supports.
package parser

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gfunc/subconvergo/cache"
	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/parser/core"
	"github.com/gfunc/subconvergo/parser/proxy"
	"github.com/gfunc/subconvergo/parser/sub"
	proxyCore "github.com/gfunc/subconvergo/proxy/core"
)

type SubParser struct {
	Index     int
	URL       string
	Proxy     string
	Tag       string
	Group     string
	UserAgent string
}

func (sp *SubParser) Parse() (*core.SubContent, error) {
	// Handle tag: prefix (tag:group_name,link)
	if strings.HasPrefix(sp.URL, "tag:") {
		if idx := strings.Index(sp.URL, ","); idx != -1 {
			sp.Group = sp.URL[4:idx]
			sp.URL = sp.URL[idx+1:]
		}
	}

	// Handle nullnode
	if sp.URL == "nullnode" {
		// Return empty content for now, effectively skipping it or we could add a dummy node
		return &core.SubContent{
			Proxies:  make([]proxyCore.ProxyInterface, 0),
			Groups:   make([]config.ProxyGroupConfig, 0),
			RawRules: make([]string, 0),
		}, nil
	}

	sp.parseURL()
	sc := &core.SubContent{
		Proxies:  make([]proxyCore.ProxyInterface, 0),
		Groups:   make([]config.ProxyGroupConfig, 0),
		RawRules: make([]string, 0),
	}

	// Check if file exists locally
	isFile := false
	if strings.HasPrefix(sp.URL, "file://") {
		isFile = true
	} else {
		if _, err := os.Stat(sp.URL); err == nil {
			isFile = true
		}
	}

	if strings.HasPrefix(sp.URL, "https://t.me/") || strings.HasPrefix(sp.URL, "tg://") {
		// Telegram links are treated as single proxies
		parserProxy, err := ParseProxy(sp.URL)
		if err != nil {
			return nil, fmt.Errorf("failed to parse telegram proxy line: %w", err)
		}
		if sp.Tag != "" {
			parserProxy.SetRemark(sp.Tag)
		}
		parserProxy.SetGroupId(sp.Index)
		if sp.Group != "" {
			parserProxy.SetGroup(sp.Group)
		} else {
			parserProxy.SetGroup(parserProxy.GetRemark())
		}
		sc.Proxies = append(sc.Proxies, parserProxy)
	} else if strings.HasPrefix(sp.URL, "http://") || strings.HasPrefix(sp.URL, "https://") || strings.HasPrefix(sp.URL, "Netch://") {
		// Parse subscription
		custom, err := sp.ParseSubscription()
		if err != nil {
			return nil, fmt.Errorf("failed to parse subscription: %w", err)
		}
		for _, p := range custom.Proxies {
			p.SetGroupId(sp.Index)
			if sp.Group != "" {
				p.SetGroup(sp.Group)
			}
		}
		sc.Proxies = append(sc.Proxies, custom.Proxies...)
		if custom.Proxies != nil {
			sc.Groups = append(sc.Groups, custom.Groups...)
		}
		if custom.RawRules != nil {
			sc.RawRules = append(sc.RawRules, custom.RawRules...)
		}
	} else if isFile {
		custom, err := sp.ParseSubscriptionFile()
		if err != nil {
			return nil, fmt.Errorf("failed to parse subscription: %w", err)
		}
		for _, p := range custom.Proxies {
			p.SetGroupId(sp.Index)
			if sp.Group != "" {
				p.SetGroup(sp.Group)
			}
		}
		sc.Proxies = append(sc.Proxies, custom.Proxies...)
		if custom.Proxies != nil {
			sc.Groups = append(sc.Groups, custom.Groups...)
		}
		if custom.RawRules != nil {
			sc.RawRules = append(sc.RawRules, custom.RawRules...)
		}
	} else {
		parserProxy, err := ParseProxy(sp.URL)
		if err != nil {
			return nil, fmt.Errorf("failed to parse proxy line: %w", err)
		}
		if sp.Tag != "" {
			parserProxy.SetRemark(sp.Tag)
		}
		parserProxy.SetGroupId(sp.Index)
		if sp.Group != "" {
			parserProxy.SetGroup(sp.Group)
		}
		sc.Proxies = append(sc.Proxies, parserProxy)
	}
	log.Printf("[parser.SubParser.Parse] index=%d url=%s proxies=%d", sp.Index, sp.URL, len(sc.Proxies))
	return sc, nil
}

func (sp *SubParser) parseURL() {
	// get tag from url after #
	u, err := url.Parse(sp.URL)
	if err != nil {
		return
	}
	sp.Tag = u.Fragment

	// remove tag from sp.URL
	if idx := strings.Index(sp.URL, "#"); idx != -1 {
		sp.URL = sp.URL[:idx]
	}
}

// ParseSubscription parses a subscription URL and returns proxy list
func (sp *SubParser) ParseSubscription() (*core.SubContent, error) {
	// Fetch subscription content
	content, err := sp.fetchSubscription()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch subscription: %w", err)
	}

	// Try to detect subscription format and parse
	custom, err := sub.ParseSubscription(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse subscription: %w", err)
	}
	log.Printf("[parser.ParseSubscription] index=%d url=%s parsed=%d proxies", sp.Index, sp.URL, len(custom.Proxies))
	return custom, nil
}

// ParseSubscriptionFile parses a file url like "file://xxxx.yaml" returns proxy list
func (sp *SubParser) ParseSubscriptionFile() (*core.SubContent, error) {
	// get file path
	filePath := strings.TrimPrefix(sp.URL, "file://")
	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read subscription file: %w", err)
	}
	log.Printf("[parser.ParseSubscriptionFile] index=%d path=%s size=%d", sp.Index, filePath, len(content))

	// Try to detect subscription format and parse
	custom, err := sub.ParseSubscription(string(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse subscription: %w", err)
	}
	return custom, nil
}

func (sp *SubParser) fetchSubscription() (string, error) {
	// Check cache first
	cacheKey := ""
	if config.Global.Advanced.EnableCache {
		cacheKey = cache.GlobalManager.GetKey(sp.URL)
		if data, ok := cache.GlobalManager.Get(cacheKey, config.Global.Advanced.CacheSubscription); ok {
			log.Printf("[parser.fetchSubscription] index=%d url=%s served from cache", sp.Index, sp.URL)
			return string(data), nil
		}
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Configure proxy if specified
	if sp.Proxy != "" && sp.Proxy != "NONE" {
		proxyURL, err := url.Parse(sp.Proxy)
		if err == nil {
			client.Transport = &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
			}
		}
	}

	targetURL := sp.URL
	if strings.HasPrefix(targetURL, "Netch://") {
		targetURL = "http://" + strings.TrimPrefix(targetURL, "Netch://")
	}

	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		log.Printf("[parser.fetchSubscription] index=%d url=%s create request error=%v", sp.Index, sp.URL, err)
		return "", err
	}

	if sp.UserAgent != "" {
		req.Header.Set("User-Agent", sp.UserAgent)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[parser.fetchSubscription] index=%d url=%s error=%v", sp.Index, sp.URL, err)
		// Try stale cache if enabled
		if config.Global.Advanced.EnableCache && config.Global.Advanced.ServeCacheOnFetchFail {
			if data, ok := cache.GlobalManager.GetStale(cacheKey); ok {
				log.Printf("[parser.fetchSubscription] index=%d url=%s served from stale cache", sp.Index, sp.URL)
				return string(data), nil
			}
		}
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		statusErr := fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
		log.Printf("[parser.fetchSubscription] index=%d url=%s status=%d statusText=%s", sp.Index, sp.URL, resp.StatusCode, resp.Status)
		// Try stale cache if enabled
		if config.Global.Advanced.EnableCache && config.Global.Advanced.ServeCacheOnFetchFail {
			if data, ok := cache.GlobalManager.GetStale(cacheKey); ok {
				log.Printf("[parser.fetchSubscription] index=%d url=%s served from stale cache", sp.Index, sp.URL)
				return string(data), nil
			}
		}
		return "", statusErr
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[parser.fetchSubscription] index=%d url=%s read error=%v", sp.Index, sp.URL, err)
		return "", err
	}

	// Save to cache
	if config.Global.Advanced.EnableCache {
		if err := cache.GlobalManager.Set(cacheKey, body); err != nil {
			log.Printf("[parser.fetchSubscription] failed to save cache: %v", err)
		}
	}

	log.Printf("[parser.fetchSubscription] index=%d url=%s size=%d", sp.Index, sp.URL, len(body))
	return string(body), nil
}

// ProcessRemark ensures unique remarks by appending _N suffix for duplicates
func ProcessRemark(remark string, existingRemarks map[string]int) string {
	if count, exists := existingRemarks[remark]; exists {
		newRemark := fmt.Sprintf("%s_%d", remark, count+1)
		existingRemarks[remark] = count + 1
		return newRemark
	}
	existingRemarks[remark] = 1
	return remark
}

// ParseProxy parses a proxy line and returns a ProxyInterface
func ParseProxy(line string) (proxyCore.ParsableProxy, error) {
	// Use the explicit routing logic
	p, err := proxy.ParseProxy(line)
	if err == nil {
		return p, nil
	}

	// Fallback to Mihomo generic parser if no specific parser found
	// Extract protocol prefix for error message
	idx := strings.Index(line, "://")
	if idx == -1 {
		return nil, fmt.Errorf("invalid proxy link format")
	}
	protocol := line[:idx]
	log.Printf("[parser.ParseProxy] unsupported native proxy protocol=%s line=%s", protocol, line)
	return sub.ParseMihomoProxy(protocol, line[idx:])
}
