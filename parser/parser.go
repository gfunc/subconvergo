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
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/metacubex/mihomo/adapter"
	mihomoConfig "github.com/metacubex/mihomo/config"
	mihomoAdapter "github.com/metacubex/mihomo/adapter"
	"github.com/metacubex/mihomo/constant"
)

// Proxy represents a parsed proxy node
type Proxy struct {
	Type          string
	Remark        string
	Server        string
	Port          int
	Password      string
	Username      string
	EncryptMethod string
	Protocol      string
	ProtocolParam string
	Obfs          string
	ObfsParam     string
	Plugin        string
	PluginOpts    string
	UUID          string
	AlterID       int
	TLS           bool
	Network       string
	Path          string
	Host          string
	UDP           bool
	TCPFastOpen   bool
	AllowInsecure bool
	TLS13         bool
	Group         string

	// Raw mihomo proxy for validation
	MihomoProxy constant.Proxy
}


// ParseSubscription parses a subscription URL and returns proxy list
func ParseSubscription(subURL string, proxy string) ([]Proxy, error) {
	// Fetch subscription content
	content, err := fetchSubscription(subURL, proxy)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch subscription: %w", err)
	}

	// Try to detect subscription format and parse
	proxies, err := parseContent(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse subscription: %w", err)
	}

	return proxies, nil
}

// ParseSubscriptionFile parses a file url like "file://xxxx.yaml" returns proxy list
func ParseSubscriptionFile(subURL string, proxy string) ([]Proxy, error) {
	// get file path
	filePath := strings.TrimPrefix(subURL, "file://")
	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read subscription file: %w", err)
	}

	// Try to detect subscription format and parse
	proxies, err := parseContent(string(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse subscription: %w", err)
	}

	return proxies, nil
}

func fetchSubscription(subURL string, proxy string) (string, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Configure proxy if specified
	if proxy != "" && proxy != "NONE" {
		proxyURL, err := url.Parse(proxy)
		if err == nil {
			client.Transport = &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
			}
		}
	}

	resp, err := client.Get(subURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func parseContent(content string) ([]Proxy, error) {
	// Try base64 decode first (standard subscription format)
	decoded, err := base64.StdEncoding.DecodeString(content)
	if err == nil {
		content = string(decoded)
	}

	var proxies []Proxy
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse different proxy formats
		proxy, err := ParseProxyLine(line)
		if err != nil {
			// Try parsing as YAML/JSON (Clash format)
			if clashProxies, err := parseClashFormat(content); err == nil {
				return clashProxies, nil
			}
			continue
		}

		proxies = append(proxies, proxy)
	}

	if len(proxies) == 0 {
		return nil, fmt.Errorf("no valid proxies found")
	}

	return proxies, nil
}

// Helper functions for URL encoding/decoding
func urlDecode(s string) string {
	decoded, err := url.QueryUnescape(s)
	if err != nil {
		return s
	}
	return decoded
}

func urlSafeBase64Decode(s string) string {
	// Replace URL-safe characters
	s = strings.ReplaceAll(s, "-", "+")
	s = strings.ReplaceAll(s, "_", "/")

	// Add padding if necessary
	if m := len(s) % 4; m != 0 {
		s += strings.Repeat("=", 4-m)
	}

	decoded, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		// Try without padding
		decoded, err = base64.RawStdEncoding.DecodeString(s)
		if err != nil {
			return s
		}
	}
	return string(decoded)
}

func parsePluginOpts(opts string) map[string]interface{} {
	result := make(map[string]interface{})
	pairs := strings.Split(opts, ";")
	for _, pair := range pairs {
		if pair == "" {
			continue
		}
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) == 2 {
			result[kv[0]] = kv[1]
		} else {
			result[kv[0]] = "true"
		}
	}
	return result
}

func ParseProxyLine(line string) (Proxy, error) {
	// Parse proxy share links: ss://, ssr://, vmess://, trojan://, etc.

	if strings.HasPrefix(line, "ss://") {
		return parseShadowsocks(line)
	} else if strings.HasPrefix(line, "ssr://") {
		return parseShadowsocksR(line)
	} else if strings.HasPrefix(line, "vmess://") {
		return parseVMess(line)
	} else if strings.HasPrefix(line, "trojan://") {
		return parseTrojan(line)
	} else if strings.HasPrefix(line, "vless://") {
		return parseVLESS(line)
	}

	// Try using mihomo's parser as fallback for other protocols
	return parseFallbackProtocol(line)
}

// parseFallbackProtocol attempts to parse unsupported protocols using mihomo
func parseFallbackProtocol(line string) (Proxy, error) {
	// Extract protocol prefix
	idx := strings.Index(line, "://")
	if idx == -1 {
		return Proxy{}, fmt.Errorf("invalid proxy link format")
	}

	protocol := line[:idx]

	// Try to parse using mihomo's built-in parsers
	// mihomo supports: hysteria, hysteria2, tuic, wireguard, etc.
	switch protocol {
	case "hysteria", "hysteria2", "hy2":
		return parseHysteria(line)
	case "tuic":
		return parseTUIC(line)
	default:
		return Proxy{}, fmt.Errorf("unsupported proxy protocol: %s", protocol)
	}
}

// parseHysteria parses hysteria:// and hysteria2:// links using mihomo
func parseHysteria(line string) (Proxy, error) {
	line = strings.TrimSpace(line)

	var protocol string
	if strings.HasPrefix(line, "hysteria2://") || strings.HasPrefix(line, "hy2://") {
		protocol = "hysteria2"
		if strings.HasPrefix(line, "hy2://") {
			line = "hysteria2://" + line[6:]
		}
	} else if strings.HasPrefix(line, "hysteria://") {
		protocol = "hysteria"
	} else {
		return Proxy{}, fmt.Errorf("not a valid hysteria link")
	}

	// Remove protocol prefix
	prefixLen := len(protocol) + 3 // "protocol://"
	line = line[prefixLen:]

	var remark, server, port, password, obfs string
	var insecure bool

	// Extract remark (after #)
	if idx := strings.LastIndex(line, "#"); idx != -1 {
		remark = urlDecode(line[idx+1:])
		line = line[:idx]
	}

	// Extract query parameters (after ?)
	var params url.Values
	if idx := strings.Index(line, "?"); idx != -1 {
		queryStr := line[idx+1:]
		line = line[:idx]
		params, _ = url.ParseQuery(queryStr)

		if params.Get("insecure") == "1" || params.Get("insecure") == "true" {
			insecure = true
		}
		obfs = params.Get("obfs")
	}

	// Parse password@server:port or server:port
	var auth string
	if strings.Contains(line, "@") {
		parts := strings.SplitN(line, "@", 2)
		auth = parts[0]
		line = parts[1]
		password = auth
	}

	// Parse server:port
	serverPort := strings.Split(line, ":")
	if len(serverPort) != 2 {
		return Proxy{}, fmt.Errorf("invalid server:port format")
	}
	server = serverPort[0]
	port = serverPort[1]

	portNum, err := strconv.Atoi(port)
	if err != nil || portNum == 0 {
		return Proxy{}, fmt.Errorf("invalid port: %s", port)
	}

	if remark == "" {
		remark = server + ":" + port
	}

	// Create mihomo config
	mihomoConfig := map[string]interface{}{
		"type":   protocol,
		"name":   remark,
		"server": server,
		"port":   portNum,
	}

	if password != "" {
		if protocol == "hysteria2" {
			mihomoConfig["password"] = password
		} else {
			mihomoConfig["auth-str"] = password
		}
	}

	if insecure {
		mihomoConfig["skip-cert-verify"] = true
	}

	if obfs != "" {
		mihomoConfig["obfs"] = obfs
	}

	// Add additional params
	if params != nil {
		if sni := params.Get("sni"); sni != "" {
			mihomoConfig["sni"] = sni
		}
		if peer := params.Get("peer"); peer != "" {
			mihomoConfig["sni"] = peer
		}
		if alpn := params.Get("alpn"); alpn != "" {
			mihomoConfig["alpn"] = strings.Split(alpn, ",")
		}

		// Hysteria v1-specific (required fields)
		if protocol == "hysteria" {
			// Extract auth from query params if not in URL
			if password == "" {
				if auth := params.Get("auth"); auth != "" {
					password = auth
					mihomoConfig["auth-str"] = auth
				}
			}

			// up and down are required for Hysteria v1
			up := params.Get("upmbps")
			down := params.Get("downmbps")
			if up == "" {
				up = "10" // Default to 10 Mbps if not specified
			}
			if down == "" {
				down = "50" // Default to 50 Mbps if not specified
			}
			if upNum, err := strconv.Atoi(up); err == nil {
				mihomoConfig["up"] = upNum
			}
			if downNum, err := strconv.Atoi(down); err == nil {
				mihomoConfig["down"] = downNum
			}
		}

		// Hysteria2-specific
		if protocol == "hysteria2" {
			if obfsPassword := params.Get("obfs-password"); obfsPassword != "" {
				mihomoConfig["obfs-password"] = obfsPassword
			}
		}
	} else if protocol == "hysteria" {
		// No params but hysteria v1 requires up/down
		mihomoConfig["up"] = 10
		mihomoConfig["down"] = 50
	}

	// Validate using mihomo
	mihomoProxy, err := adapter.ParseProxy(mihomoConfig)
	if err != nil {
		return Proxy{}, fmt.Errorf("mihomo validation failed: %w", err)
	}

	return Proxy{
		Type:          protocol,
		Remark:        remark,
		Server:        server,
		Port:          portNum,
		Password:      password,
		Obfs:          obfs,
		AllowInsecure: insecure,
		Group:         strings.ToUpper(protocol),
		MihomoProxy:   mihomoProxy,
	}, nil
}

// parseTUIC parses tuic:// links using mihomo
func parseTUIC(line string) (Proxy, error) {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "tuic://") {
		return Proxy{}, fmt.Errorf("not a valid tuic link")
	}

	// Remove tuic:// prefix
	line = line[7:]

	var remark, uuid, password, server, port string
	var insecure bool

	// Extract remark (after #)
	if idx := strings.LastIndex(line, "#"); idx != -1 {
		remark = urlDecode(line[idx+1:])
		line = line[:idx]
	}

	// Extract query parameters (after ?)
	var params url.Values
	if idx := strings.Index(line, "?"); idx != -1 {
		queryStr := line[idx+1:]
		line = line[:idx]
		params, _ = url.ParseQuery(queryStr)

		if params.Get("allow_insecure") == "1" || params.Get("allow_insecure") == "true" {
			insecure = true
		}
	}

	// Parse uuid:password@server:port or uuid@server:port
	if strings.Contains(line, "@") {
		parts := strings.SplitN(line, "@", 2)
		auth := parts[0]
		line = parts[1]

		// Auth can be uuid:password or just uuid
		authParts := strings.SplitN(auth, ":", 2)
		uuid = authParts[0]
		if len(authParts) == 2 {
			password = authParts[1]
		}
	}

	// Parse server:port
	serverPort := strings.Split(line, ":")
	if len(serverPort) != 2 {
		return Proxy{}, fmt.Errorf("invalid server:port format")
	}
	server = serverPort[0]
	port = serverPort[1]

	portNum, err := strconv.Atoi(port)
	if err != nil || portNum == 0 {
		return Proxy{}, fmt.Errorf("invalid port: %s", port)
	}

	if remark == "" {
		remark = server + ":" + port
	}

	// Create mihomo config
	mihomoConfig := map[string]interface{}{
		"type":   "tuic",
		"name":   remark,
		"server": server,
		"port":   portNum,
		"uuid":   uuid,
	}

	if password != "" {
		mihomoConfig["password"] = password
	}

	if insecure {
		mihomoConfig["skip-cert-verify"] = true
	}

	// Add additional params
	if params != nil {
		if sni := params.Get("sni"); sni != "" {
			mihomoConfig["sni"] = sni
		}
		if alpn := params.Get("alpn"); alpn != "" {
			mihomoConfig["alpn"] = strings.Split(alpn, ",")
		}
		if congestion := params.Get("congestion_control"); congestion != "" {
			mihomoConfig["congestion-controller"] = congestion
		}
		if udpRelay := params.Get("udp_relay_mode"); udpRelay != "" {
			mihomoConfig["udp-relay-mode"] = udpRelay
		}
	}

	// Validate using mihomo
	mihomoProxy, err := adapter.ParseProxy(mihomoConfig)
	if err != nil {
		return Proxy{}, fmt.Errorf("mihomo validation failed: %w", err)
	}

	return Proxy{
		Type:          "tuic",
		Remark:        remark,
		Server:        server,
		Port:          portNum,
		UUID:          uuid,
		Password:      password,
		AllowInsecure: insecure,
		Group:         "TUIC",
		MihomoProxy:   mihomoProxy,
	}, nil
}

// Use mihomo to parse and validate Clash format
func parseClashFormat(content string) ([]Proxy, error) {
	// Parse using mihomo's config parser
	rawCfg, err := mihomoConfig.UnmarshalRawConfig([]byte(content))
	if err != nil {
		return nil, err
	}

	var proxies []Proxy
	for _, rawProxy := range rawCfg.Proxy {
		clashProxy, err := mihomoAdapter.ParseProxy(rawProxy)
		if err == nil {
			proxies = append(proxies, Proxy{
				Type: clashProxy.Type().String(),
				Remark: clashProxy.Name(),
				MihomoProxy: clashProxy,
			})
			continue
		}
		// Convert mihomo proxy to our Proxy struct
		proxy, err := convertMihomoProxy(rawProxy)
		if err != nil {
			continue
		}
		proxies = append(proxies, proxy)
	}

	return proxies, nil
}

func convertMihomoProxy(rawProxy map[string]interface{}) (Proxy, error) {
	// Use mihomo's adapter to parse and validate proxy
	mihomoProxy, err := adapter.ParseProxy(rawProxy)
	if err != nil {
		return Proxy{}, err
	}

	proxy := Proxy{
		Remark:      rawProxy["name"].(string),
		Server:      rawProxy["server"].(string),
		MihomoProxy: mihomoProxy,
	}

	// Extract common fields
	if port, ok := rawProxy["port"].(int); ok {
		proxy.Port = port
	}

	if proxyType, ok := rawProxy["type"].(string); ok {
		proxy.Type = proxyType
	}

	// Type-specific parsing
	switch proxy.Type {
	case "ss":
		if cipher, ok := rawProxy["cipher"].(string); ok {
			proxy.EncryptMethod = cipher
		}
		if password, ok := rawProxy["password"].(string); ok {
			proxy.Password = password
		}
	case "vmess":
		if uuid, ok := rawProxy["uuid"].(string); ok {
			proxy.UUID = uuid
		}
		if alterId, ok := rawProxy["alterId"].(int); ok {
			proxy.AlterID = alterId
		}
	case "trojan":
		if password, ok := rawProxy["password"].(string); ok {
			proxy.Password = password
		}
	}

	return proxy, nil
}

// parseShadowsocks parses ss:// share links
// Format: ss://base64(method:password)@server:port#remark
// or: ss://base64(method:password@server:port)#remark
func parseShadowsocks(line string) (Proxy, error) {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "ss://") {
		return Proxy{}, fmt.Errorf("not a valid ss:// link")
	}

	// Remove ss:// prefix
	line = line[5:]

	var remark, password, method, server, port, plugin, pluginOpts, group string

	// Extract remark (after #)
	if idx := strings.Index(line, "#"); idx != -1 {
		remark = urlDecode(line[idx+1:])
		line = line[:idx]
	}

	// Extract plugin options (after ?)
	if idx := strings.Index(line, "?"); idx != -1 {
		queryStr := line[idx+1:]
		line = line[:idx]

		// Parse plugin parameter
		params, _ := url.ParseQuery(queryStr)
		if pluginStr := params.Get("plugin"); pluginStr != "" {
			pluginStr = urlDecode(pluginStr)
			if idx := strings.Index(pluginStr, ";"); idx != -1 {
				plugin = pluginStr[:idx]
				pluginOpts = pluginStr[idx+1:]
			} else {
				plugin = pluginStr
			}
		}

		if groupStr := params.Get("group"); groupStr != "" {
			group = urlSafeBase64Decode(groupStr)
		}
	}

	// Parse server info
	if strings.Contains(line, "@") {
		// New format: method:password@server:port (base64 encoded userinfo)
		parts := strings.Split(line, "@")
		if len(parts) != 2 {
			return Proxy{}, fmt.Errorf("invalid ss link format")
		}

		// Decode userinfo
		userInfo := urlSafeBase64Decode(parts[0])
		methodPass := strings.SplitN(userInfo, ":", 2)
		if len(methodPass) != 2 {
			return Proxy{}, fmt.Errorf("invalid userinfo format")
		}
		method = methodPass[0]
		password = methodPass[1]

		// Parse server:port
		serverInfo := parts[1]

		// Handle IPv6 addresses in brackets
		if strings.HasPrefix(serverInfo, "[") {
			// IPv6: [host]:port
			endIdx := strings.Index(serverInfo, "]")
			if endIdx == -1 {
				return Proxy{}, fmt.Errorf("invalid IPv6 format")
			}
			server = serverInfo[:endIdx+1] // Include brackets
			if endIdx+1 < len(serverInfo) && serverInfo[endIdx+1] == ':' {
				port = serverInfo[endIdx+2:]
			} else {
				return Proxy{}, fmt.Errorf("invalid server:port format")
			}
		} else {
			// IPv4 or hostname: host:port
			serverPort := strings.Split(serverInfo, ":")
			if len(serverPort) != 2 {
				return Proxy{}, fmt.Errorf("invalid server:port format")
			}
			server = serverPort[0]
			port = serverPort[1]
		}
	} else {
		// Old format: entire string is base64 encoded
		decoded := urlSafeBase64Decode(line)

		// Parse: method:password@server:port
		atIdx := strings.Index(decoded, "@")
		if atIdx == -1 {
			return Proxy{}, fmt.Errorf("invalid ss link format")
		}

		userInfo := decoded[:atIdx]
		serverInfo := decoded[atIdx+1:]

		methodPass := strings.SplitN(userInfo, ":", 2)
		if len(methodPass) != 2 {
			return Proxy{}, fmt.Errorf("invalid userinfo format")
		}
		method = methodPass[0]
		password = methodPass[1]

		serverPort := strings.Split(serverInfo, ":")
		if len(serverPort) != 2 {
			return Proxy{}, fmt.Errorf("invalid server:port format")
		}
		server = serverPort[0]
		port = serverPort[1]
	}

	portNum, err := strconv.Atoi(port)
	if err != nil || portNum == 0 {
		return Proxy{}, fmt.Errorf("invalid port: %s", port)
	}

	if remark == "" {
		remark = server + ":" + port
	}

	if group == "" {
		group = "SS"
	}

	// Create mihomo proxy config for validation
	mihomoConfig := map[string]interface{}{
		"type":     "ss",
		"name":     remark,
		"server":   server,
		"port":     portNum,
		"cipher":   method,
		"password": password,
	}



	if plugin != "" {
		mihomoConfig["plugin"] = plugin
		if pluginOpts != "" {
			mihomoConfig["plugin-opts"] = parsePluginOpts(pluginOpts)
		}
	}

	// Validate using mihomo
	mihomoProxy, err := adapter.ParseProxy(mihomoConfig)
	if err != nil {
		return Proxy{}, fmt.Errorf("mihomo validation failed: %w", err)
	}

	return Proxy{
		Type:          "ss",
		Remark:        remark,
		Server:        server,
		Port:          portNum,
		Password:      password,
		EncryptMethod: method,
		Plugin:        plugin,
		PluginOpts:    pluginOpts,
		Group:         group,
		MihomoProxy:   mihomoProxy,
	}, nil
}

func parseShadowsocksR(line string) (Proxy, error) {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "ssr://") {
		return Proxy{}, fmt.Errorf("not a valid ssr:// link")
	}

	// Remove ssr:// prefix and decode
	line = strings.ReplaceAll(line[6:], "\r", "")
	line = urlSafeBase64Decode(line)

	var remarks, group, server, port, method, password, protocol, protocolParam, obfs, obfsParam string

	// Extract query parameters (after /?)
	if idx := strings.Index(line, "/?"); idx != -1 {
		queryStr := line[idx+2:]
		line = line[:idx]

		params, _ := url.ParseQuery(queryStr)
		group = urlSafeBase64Decode(params.Get("group"))
		remarks = urlSafeBase64Decode(params.Get("remarks"))
		obfsParam = strings.TrimSpace(urlSafeBase64Decode(params.Get("obfsparam")))
		protocolParam = strings.TrimSpace(urlSafeBase64Decode(params.Get("protoparam")))
	}

	// Parse: server:port:protocol:method:obfs:password_base64
	re := regexp.MustCompile(`(\S+):(\d+?):(\S+?):(\S+?):(\S+?):(\S+)`)
	matches := re.FindStringSubmatch(line)
	if len(matches) != 7 {
		return Proxy{}, fmt.Errorf("invalid ssr link format")
	}

	server = matches[1]
	port = matches[2]
	protocol = matches[3]
	method = matches[4]
	obfs = matches[5]
	password = urlSafeBase64Decode(matches[6])

	portNum, err := strconv.Atoi(port)
	if err != nil || portNum == 0 {
		return Proxy{}, fmt.Errorf("invalid port: %s", port)
	}

	if group == "" {
		group = "SSR"
	}
	if remarks == "" {
		remarks = server + ":" + port
	}

	// Check if this is actually a plain SS (when obfs is plain and protocol is origin)
	ssCiphers := []string{"rc4-md5", "aes-128-gcm", "aes-192-gcm", "aes-256-gcm", "aes-128-cfb",
		"aes-192-cfb", "aes-256-cfb", "aes-128-ctr", "aes-192-ctr", "aes-256-ctr", "chacha20-ietf-poly1305",
		"xchacha20-ietf-poly1305", "2022-blake3-aes-128-gcm", "2022-blake3-aes-256-gcm"}

	isSS := false
	for _, cipher := range ssCiphers {
		if cipher == method && (obfs == "" || obfs == "plain") && (protocol == "" || protocol == "origin") {
			isSS = true
			break
		}
	}

	if isSS {
		// Convert to SS
		mihomoConfig := map[string]interface{}{
			"type":     "ss",
			"name":     remarks,
			"server":   server,
			"port":     portNum,
			"cipher":   method,
			"password": password,
		}

		mihomoProxy, err := adapter.ParseProxy(mihomoConfig)
		if err != nil {
			return Proxy{}, fmt.Errorf("mihomo validation failed: %w", err)
		}

		return Proxy{
			Type:          "ss",
			Remark:        remarks,
			Server:        server,
			Port:          portNum,
			Password:      password,
			EncryptMethod: method,
			Group:         group,
			MihomoProxy:   mihomoProxy,
		}, nil
	}

	// Create SSR proxy
	mihomoConfig := map[string]interface{}{
		"type":     "ssr",
		"name":     remarks,
		"server":   server,
		"port":     portNum,
		"cipher":   method,
		"password": password,
		"protocol": protocol,
		"obfs":     obfs,
	}

	if protocolParam != "" {
		mihomoConfig["protocol-param"] = protocolParam
	}
	if obfsParam != "" {
		mihomoConfig["obfs-param"] = obfsParam
	}

	mihomoProxy, err := adapter.ParseProxy(mihomoConfig)
	if err != nil {
		return Proxy{}, fmt.Errorf("mihomo validation failed: %w", err)
	}

	return Proxy{
		Type:          "ssr",
		Remark:        remarks,
		Server:        server,
		Port:          portNum,
		Password:      password,
		EncryptMethod: method,
		Protocol:      protocol,
		ProtocolParam: protocolParam,
		Obfs:          obfs,
		ObfsParam:     obfsParam,
		Group:         group,
		MihomoProxy:   mihomoProxy,
	}, nil
}

func parseVMess(line string) (Proxy, error) {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "vmess://") {
		return Proxy{}, fmt.Errorf("not a valid vmess:// link")
	}

	// Remove vmess:// prefix
	line = line[8:]

	// Decode base64
	decoded := urlSafeBase64Decode(line)

	// Parse JSON
	var vmessData map[string]interface{}
	if err := json.Unmarshal([]byte(decoded), &vmessData); err != nil {
		return Proxy{}, fmt.Errorf("failed to parse vmess JSON: %w", err)
	}

	// Extract fields
	ps := getStringField(vmessData, "ps")
	add := getStringField(vmessData, "add")
	port := getStringField(vmessData, "port")
	id := getStringField(vmessData, "id")
	aid := getStringField(vmessData, "aid")
	net := getStringField(vmessData, "net")
	_ = getStringField(vmessData, "type") // headerType, not used in mihomo config
	host := getStringField(vmessData, "host")
	path := getStringField(vmessData, "path")
	tls := getStringField(vmessData, "tls")
	sni := getStringField(vmessData, "sni")

	// Default values
	if net == "" {
		net = "tcp"
	}
	if aid == "" {
		aid = "0"
	}

	portNum, err := strconv.Atoi(port)
	if err != nil || portNum == 0 {
		return Proxy{}, fmt.Errorf("invalid port: %s", port)
	}

	alterID, _ := strconv.Atoi(aid)

	if ps == "" {
		ps = add + ":" + port
	}

	// Handle version-specific path parsing
	version := getStringField(vmessData, "v")
	if version == "1" || version == "" {
		// Version 1: host can contain host;path
		if host != "" && strings.Contains(host, ";") {
			parts := strings.SplitN(host, ";", 2)
			host = parts[0]
			if path == "" {
				path = parts[1]
			}
		}
	}

	// Create mihomo config
	mihomoConfig := map[string]interface{}{
		"type":    "vmess",
		"name":    ps,
		"server":  add,
		"port":    portNum,
		"uuid":    id,
		"alterId": alterID,
		"cipher":  "auto",
		"network": net,
	}

	if tls == "tls" {
		mihomoConfig["tls"] = true
		if sni != "" {
			mihomoConfig["servername"] = sni
		}
	}

	// Network-specific options
	switch net {
	case "ws", "httpupgrade":
		wsOpts := make(map[string]interface{})
		if path == "" {
			path = "/"
		}
		wsOpts["path"] = path
		if host != "" {
			headers := make(map[string]string)
			headers["Host"] = host
			wsOpts["headers"] = headers
		}
		mihomoConfig["ws-opts"] = wsOpts

	case "http", "h2":
		h2Opts := make(map[string]interface{})
		if path != "" {
			h2Opts["path"] = path // mihomo expects string
		}
		if host != "" {
			h2Opts["host"] = []string{host}
		}
		mihomoConfig["h2-opts"] = h2Opts

	case "grpc":
		grpcOpts := make(map[string]interface{})
		if path != "" {
			grpcOpts["grpc-service-name"] = path
		}
		mihomoConfig["grpc-opts"] = grpcOpts
		if host != "" {
			mihomoConfig["servername"] = host
		}

	case "quic":
		quicOpts := make(map[string]interface{})
		if host != "" {
			quicOpts["mode"] = host
		}
		if path != "" {
			quicOpts["key"] = path
		}
		mihomoConfig["quic-opts"] = quicOpts
	}

	// Validate using mihomo
	mihomoProxy, err := adapter.ParseProxy(mihomoConfig)
	if err != nil {
		return Proxy{}, fmt.Errorf("mihomo validation failed: %w", err)
	}

	return Proxy{
		Type:        "vmess",
		Remark:      ps,
		Server:      add,
		Port:        portNum,
		UUID:        id,
		AlterID:     alterID,
		Network:     net,
		Path:        path,
		Host:        host,
		TLS:         tls == "tls",
		Group:       "VMess",
		MihomoProxy: mihomoProxy,
	}, nil
}

func getStringField(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case string:
			return val
		case float64:
			return strconv.FormatFloat(val, 'f', -1, 64)
		case int:
			return strconv.Itoa(val)
		}
	}
	return ""
}

func parseTrojan(line string) (Proxy, error) {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "trojan://") {
		return Proxy{}, fmt.Errorf("not a valid trojan:// link")
	}

	// Remove trojan:// prefix
	line = line[9:]

	var remark, password, server, port, sni, network, path, group string
	var allowInsecure bool

	// Extract remark (after #)
	if idx := strings.LastIndex(line, "#"); idx != -1 {
		remark = urlDecode(line[idx+1:])
		line = line[:idx]
	}

	// Extract query parameters (after ?)
	if idx := strings.Index(line, "?"); idx != -1 {
		queryStr := line[idx+1:]
		line = line[:idx]

		params, _ := url.ParseQuery(queryStr)

		// SNI can be in 'sni' or 'peer' parameter
		sni = params.Get("sni")
		if sni == "" {
			sni = params.Get("peer")
		}

		if params.Get("allowInsecure") == "1" || params.Get("allowInsecure") == "true" {
			allowInsecure = true
		}

		group = urlDecode(params.Get("group"))

		// Check for WebSocket
		if params.Get("ws") == "1" {
			network = "ws"
			path = params.Get("wspath")
		} else if params.Get("type") == "ws" {
			network = "ws"
			path = params.Get("path")
			// Path might be URL encoded
			if strings.HasPrefix(path, "%2F") {
				path = urlDecode(path)
			}
		} else if params.Get("type") == "grpc" {
			network = "grpc"
			path = params.Get("serviceName")
			if path == "" {
				path = params.Get("path")
			}
		}
	}

	// Parse password@server:port
	re := regexp.MustCompile(`(.*?)@(.*):(.*)`)
	matches := re.FindStringSubmatch(line)
	if len(matches) != 4 {
		return Proxy{}, fmt.Errorf("invalid trojan link format")
	}

	password = matches[1]
	server = matches[2]
	port = matches[3]

	portNum, err := strconv.Atoi(port)
	if err != nil || portNum == 0 {
		return Proxy{}, fmt.Errorf("invalid port: %s", port)
	}

	if remark == "" {
		remark = server + ":" + port
	}
	if group == "" {
		group = "Trojan"
	}

	// Create mihomo config
	mihomoConfig := map[string]interface{}{
		"type":     "trojan",
		"name":     remark,
		"server":   server,
		"port":     portNum,
		"password": password,
	}

	if sni != "" {
		mihomoConfig["sni"] = sni
	}

	if allowInsecure {
		mihomoConfig["skip-cert-verify"] = true
	}

	if network != "" {
		mihomoConfig["network"] = network

		switch network {
		case "ws":
			wsOpts := make(map[string]interface{})
			if path != "" {
				wsOpts["path"] = path
			}
			mihomoConfig["ws-opts"] = wsOpts

		case "grpc":
			grpcOpts := make(map[string]interface{})
			if path != "" {
				grpcOpts["grpc-service-name"] = path
			}
			mihomoConfig["grpc-opts"] = grpcOpts
		}
	}

	// Validate using mihomo
	mihomoProxy, err := adapter.ParseProxy(mihomoConfig)
	if err != nil {
		return Proxy{}, fmt.Errorf("mihomo validation failed: %w", err)
	}

	return Proxy{
		Type:          "trojan",
		Remark:        remark,
		Server:        server,
		Port:          portNum,
		Password:      password,
		Network:       network,
		Path:          path,
		Host:          sni,
		TLS:           true, // Trojan always uses TLS
		AllowInsecure: allowInsecure,
		Group:         group,
		MihomoProxy:   mihomoProxy,
	}, nil
}

func parseVLESS(line string) (Proxy, error) {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "vless://") {
		return Proxy{}, fmt.Errorf("not a valid vless:// link")
	}

	// Remove vless:// prefix
	line = line[8:]

	var remark, uuid, server, port, network, flow, security, sni, path, host, group string
	var allowInsecure bool

	// Extract remark (after #)
	if idx := strings.LastIndex(line, "#"); idx != -1 {
		remark = urlDecode(line[idx+1:])
		line = line[:idx]
	}

	// Extract query parameters (after ?)
	if idx := strings.Index(line, "?"); idx != -1 {
		queryStr := line[idx+1:]
		line = line[:idx]

		params, _ := url.ParseQuery(queryStr)

		network = params.Get("type")
		if network == "" {
			network = "tcp"
		}

		security = params.Get("security")
		flow = params.Get("flow")
		sni = params.Get("sni")

		if params.Get("allowInsecure") == "1" || params.Get("allowInsecure") == "true" {
			allowInsecure = true
		}

		group = urlDecode(params.Get("group"))

		// Network-specific parameters
		switch network {
		case "ws":
			path = params.Get("path")
			host = params.Get("host")
		case "grpc":
			path = params.Get("serviceName")
			if path == "" {
				path = params.Get("path")
			}
		case "http", "h2":
			path = params.Get("path")
			host = params.Get("host")
		case "quic":
			// QUIC-specific params
			host = params.Get("quicSecurity")
			path = params.Get("key")
		}
	}

	// Parse uuid@server:port
	re := regexp.MustCompile(`(.*?)@(.*):(.*)`)
	matches := re.FindStringSubmatch(line)
	if len(matches) != 4 {
		return Proxy{}, fmt.Errorf("invalid vless link format")
	}

	uuid = matches[1]
	server = matches[2]
	port = matches[3]

	portNum, err := strconv.Atoi(port)
	if err != nil || portNum == 0 {
		return Proxy{}, fmt.Errorf("invalid port: %s", port)
	}

	if remark == "" {
		remark = server + ":" + port
	}
	if group == "" {
		group = "VLESS"
	}

	// Create mihomo config
	mihomoConfig := map[string]interface{}{
		"type":    "vless",
		"name":    remark,
		"server":  server,
		"port":    portNum,
		"uuid":    uuid,
		"network": network,
	}

	if flow != "" {
		mihomoConfig["flow"] = flow
	}

	if security == "tls" || security == "reality" {
		mihomoConfig["tls"] = true
		if sni != "" {
			mihomoConfig["servername"] = sni
		}
	}

	if allowInsecure {
		mihomoConfig["skip-cert-verify"] = true
	}

	// Network-specific options
	switch network {
	case "ws":
		wsOpts := make(map[string]interface{})
		if path != "" {
			wsOpts["path"] = path
		}
		if host != "" {
			headers := make(map[string]string)
			headers["Host"] = host
			wsOpts["headers"] = headers
		}
		mihomoConfig["ws-opts"] = wsOpts

	case "grpc":
		grpcOpts := make(map[string]interface{})
		if path != "" {
			grpcOpts["grpc-service-name"] = path
		}
		mihomoConfig["grpc-opts"] = grpcOpts

	case "http", "h2":
		h2Opts := make(map[string]interface{})
		if path != "" {
			h2Opts["path"] = path // mihomo expects string
		}
		if host != "" {
			h2Opts["host"] = []string{host}
		}
		mihomoConfig["h2-opts"] = h2Opts
	}

	// Validate using mihomo
	mihomoProxy, err := adapter.ParseProxy(mihomoConfig)
	if err != nil {
		return Proxy{}, fmt.Errorf("mihomo validation failed: %w", err)
	}

	return Proxy{
		Type:          "vless",
		Remark:        remark,
		Server:        server,
		Port:          portNum,
		UUID:          uuid,
		Network:       network,
		Path:          path,
		Host:          host,
		TLS:           security == "tls" || security == "reality",
		AllowInsecure: allowInsecure,
		Group:         group,
		MihomoProxy:   mihomoProxy,
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
