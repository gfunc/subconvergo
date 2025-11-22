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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/parser/core"
	proxyCore "github.com/gfunc/subconvergo/proxy/core"
	"github.com/gfunc/subconvergo/proxy/impl"
	"github.com/metacubex/mihomo/adapter"
	C "github.com/metacubex/mihomo/config"
	"gopkg.in/yaml.v3"
)

type SubContent struct {
	Proxies  []proxyCore.ProxyInterface
	Groups   []config.ProxyGroupConfig
	RawRules []string
}

type SubParser struct {
	Index int
	URL   string
	Proxy string
	Tag   string
}

func (sp *SubParser) Parse() (*SubContent, error) {
	sp.parseURL()
	sc := &SubContent{
		Proxies:  make([]proxyCore.ProxyInterface, 0),
		Groups:   make([]config.ProxyGroupConfig, 0),
		RawRules: make([]string, 0),
	}

	if strings.HasPrefix(sp.URL, "http://") || strings.HasPrefix(sp.URL, "https://") {
		// Parse subscription
		custom, err := sp.ParseSubscription()
		if err != nil {
			return nil, fmt.Errorf("failed to parse subscription: %w", err)
		}
		sc.Proxies = append(sc.Proxies, custom.Proxies...)
		if custom.Proxies != nil {
			sc.Groups = append(sc.Groups, custom.Groups...)
		}
		if custom.RawRules != nil {
			sc.RawRules = append(sc.RawRules, custom.RawRules...)
		}
	} else if strings.HasPrefix(sp.URL, "file://") {
		custom, err := sp.ParseSubscriptionFile()
		if err != nil {
			return nil, fmt.Errorf("failed to parse subscription: %w", err)
		}
		sc.Proxies = append(sc.Proxies, custom.Proxies...)
		if custom.Proxies != nil {
			sc.Groups = append(sc.Groups, custom.Groups...)
		}
		if custom.RawRules != nil {
			sc.RawRules = append(sc.RawRules, custom.RawRules...)
		}
	} else {
		parserProxy, err := ParseProxyLine(sp.URL)
		if err != nil {
			return nil, fmt.Errorf("failed to parse proxy line: %w", err)
		}
		if sp.Tag != "" {
			parserProxy.SetRemark(sp.Tag)
		}
		parserProxy.SetGroupId(sp.Index)
		parserProxy.SetGroup(parserProxy.GetRemark())
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
func (sp *SubParser) ParseSubscription() (*SubContent, error) {
	// Fetch subscription content
	content, err := sp.fetchSubscription()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch subscription: %w", err)
	}

	// Try to detect subscription format and parse
	custom, err := sp.parseContent(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse subscription: %w", err)
	}
	for _, p := range custom.Proxies {
		p.SetGroupId(sp.Index)
		p.SetGroup(sp.Tag)
	}
	log.Printf("[parser.ParseSubscription] index=%d url=%s parsed=%d proxies", sp.Index, sp.URL, len(custom.Proxies))
	return custom, nil
}

// ParseSubscriptionFile parses a file url like "file://xxxx.yaml" returns proxy list
func (sp *SubParser) ParseSubscriptionFile() (*SubContent, error) {
	// get file path
	filePath := strings.TrimPrefix(sp.URL, "file://")
	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read subscription file: %w", err)
	}
	log.Printf("[parser.ParseSubscriptionFile] index=%d path=%s size=%d", sp.Index, filePath, len(content))

	// Try to detect subscription format and parse
	custom, err := sp.parseContent(string(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse subscription: %w", err)
	}
	for _, p := range custom.Proxies {
		p.SetGroupId(sp.Index)
		p.SetGroup(sp.Tag)
	}
	return custom, nil
}

func (sp *SubParser) fetchSubscription() (string, error) {
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

	resp, err := client.Get(sp.URL)
	if err != nil {
		log.Printf("[parser.fetchSubscription] index=%d url=%s error=%v", sp.Index, sp.URL, err)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		statusErr := fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
		log.Printf("[parser.fetchSubscription] index=%d url=%s status=%d statusText=%s", sp.Index, sp.URL, resp.StatusCode, resp.Status)
		return "", statusErr
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[parser.fetchSubscription] index=%d url=%s read error=%v", sp.Index, sp.URL, err)
		return "", err
	}

	log.Printf("[parser.fetchSubscription] index=%d url=%s size=%d", sp.Index, sp.URL, len(body))
	return string(body), nil
}

func (sp *SubParser) parseContent(content string) (*SubContent, error) {
	// Try base64 decode first (standard subscription format)
	decoded, err := base64.StdEncoding.DecodeString(content)
	if err == nil {
		content = string(decoded)
	}

	var proxies []proxyCore.ProxyInterface
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Skip comment lines
		if strings.HasPrefix(line, "#") {
			continue
		}

		// Parse different proxy formats
		p, err := ParseProxyLine(line)
		if err != nil {
			// Try parsing as YAML/JSON (Clash format)
			if custom, err := ParseMihomoConfig(content); err == nil {
				return custom, nil
			} else {
				log.Printf("[parser.parseContent] url=%s line=%s err=%v", sp.URL, line, err)
				break
			}
		}

		proxies = append(proxies, p)
	}

	if len(proxies) == 0 {
		log.Printf("[parser.parseContent] url=%s no valid proxies found", sp.URL)
		return nil, fmt.Errorf("no valid proxies found")
	}

	return &SubContent{
		Proxies: proxies,
	}, nil
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

// ParseProxyLine parses a proxy line and returns a ProxyInterface
func ParseProxyLine(line string) (proxyCore.SubconverterProxy, error) {
	// Use the registry to find a matching parser
	p, err := core.ParseLine(line)
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
	log.Printf("[parser.ParseProxyLine] unsupported native proxy protocol=%s line=%s", protocol, line)
	return ParseMihomoProxy(protocol, line[idx:])
}

// ParseMihomoConfig parses Clash YAML format and returns proxies
func ParseMihomoConfig(content string) (*SubContent, error) {
	// Try parsing as YAML (Clash format)
	var clashConfig C.RawConfig

	if err := yaml.Unmarshal([]byte(content), &clashConfig); err != nil {
		return nil, fmt.Errorf("failed to parse clash format: %w", err)
	}

	if len(clashConfig.Proxy) == 0 {
		return nil, fmt.Errorf("no proxies found in clash format")
	}
	custom := &SubContent{Proxies: make([]proxyCore.ProxyInterface, 0)}

	for _, proxyMap := range clashConfig.Proxy {
		mihomoProxy, err := parseMihomoProxy(proxyMap)
		if err != nil {
			log.Printf("failed to parse proxy in clash format: %v", err)
			continue
		}
		custom.Proxies = append(custom.Proxies, mihomoProxy)
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

	if len(custom.Proxies) == 0 {
		return nil, fmt.Errorf("no valid proxies in clash format")
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
	hostPort := strings.Split(addr, ":")
	if len(hostPort) != 2 {
		return nil, fmt.Errorf("invalid address format: %s", addr)
	}
	server := hostPort[0]
	port, err := strconv.Atoi(hostPort[1])
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
