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

	"gopkg.in/yaml.v3"

	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/proxy"
	"github.com/metacubex/mihomo/adapter"
	C "github.com/metacubex/mihomo/config"
)

type CustomContent struct {
	Proxies  []proxy.ProxyInterface
	Groups   []config.ProxyGroupConfig
	RawRules []string
}

// ParseSubscription parses a subscription URL and returns proxy list
func ParseSubscription(subURL string, proxy string) (*CustomContent, error) {
	// Fetch subscription content
	content, err := fetchSubscription(subURL, proxy)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch subscription: %w", err)
	}

	// Try to detect subscription format and parse
	custom, err := parseContent(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse subscription: %w", err)
	}

	return custom, nil
}

// ParseSubscriptionFile parses a file url like "file://xxxx.yaml" returns proxy list
func ParseSubscriptionFile(subURL string) (*CustomContent, error) {
	// get file path
	filePath := strings.TrimPrefix(subURL, "file://")
	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read subscription file: %w", err)
	}

	// Try to detect subscription format and parse
	custom, err := parseContent(string(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse subscription: %w", err)
	}

	return custom, nil
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

func parseContent(content string) (*CustomContent, error) {
	// Try base64 decode first (standard subscription format)
	decoded, err := base64.StdEncoding.DecodeString(content)
	if err == nil {
		content = string(decoded)
	}

	var proxies []proxy.ProxyInterface
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse different proxy formats
		p, err := ParseProxyLine(line)
		if err != nil {
			// Try parsing as YAML/JSON (Clash format)
			if custom, err := ParseMihomoConfig(content); err == nil {
				return custom, nil
			}else{
				log.Printf("Failed to parse line as proxy or mihomo config: %s, %v", line, err)
				break
			}
		}

		proxies = append(proxies, p)
	}

	if len(proxies) == 0 {
		return nil, fmt.Errorf("no valid proxies found")
	}

	return &CustomContent{
		Proxies: proxies,
	}, nil
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
func ParseProxyLine(line string) (proxy.SubconverterProxy, error) {
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
	} else if strings.HasPrefix(line, "hysteria://") || strings.HasPrefix(line, "hysteria2://") || strings.HasPrefix(line, "hy2://") {
		return parseHysteria(line)
	} else if strings.HasPrefix(line, "tuic://") {
		return parseTUIC(line)
	}

	// Extract protocol prefix for error message
	idx := strings.Index(line, "://")
	if idx == -1 {
		return nil, fmt.Errorf("invalid proxy link format")
	}
	protocol := line[:idx]
	return nil, fmt.Errorf("unsupported proxy protocol: %s", protocol)
}

func parseShadowsocks(line string) (proxy.SubconverterProxy, error) {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "ss://") {
		return nil, fmt.Errorf("not a valid ss:// link")
	}

	line = line[5:]

	var remark, password, method, server, port, plugin, pluginOpts, group string

	if idx := strings.Index(line, "#"); idx != -1 {
		remark = urlDecode(line[idx+1:])
		line = line[:idx]
	}

	if idx := strings.Index(line, "?"); idx != -1 {
		queryStr := line[idx+1:]
		line = line[:idx]

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

	if strings.Contains(line, "@") {
		parts := strings.Split(line, "@")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid ss link format")
		}

		userInfo := urlSafeBase64Decode(parts[0])
		methodPass := strings.SplitN(userInfo, ":", 2)
		if len(methodPass) != 2 {
			return nil, fmt.Errorf("invalid userinfo format")
		}
		method = methodPass[0]
		password = methodPass[1]

		serverInfo := parts[1]

		if strings.HasPrefix(serverInfo, "[") {
			endIdx := strings.Index(serverInfo, "]")
			if endIdx == -1 {
				return nil, fmt.Errorf("invalid IPv6 format")
			}
			server = serverInfo[:endIdx+1]
			if endIdx+1 < len(serverInfo) && serverInfo[endIdx+1] == ':' {
				port = serverInfo[endIdx+2:]
			} else {
				return nil, fmt.Errorf("invalid server:port format")
			}
		} else {
			serverPort := strings.Split(serverInfo, ":")
			if len(serverPort) != 2 {
				return nil, fmt.Errorf("invalid server:port format")
			}
			server = serverPort[0]
			port = serverPort[1]
		}
	} else {
		decoded := urlSafeBase64Decode(line)

		atIdx := strings.Index(decoded, "@")
		if atIdx == -1 {
			return nil, fmt.Errorf("invalid ss link format")
		}

		userInfo := decoded[:atIdx]
		serverInfo := decoded[atIdx+1:]

		methodPass := strings.SplitN(userInfo, ":", 2)
		if len(methodPass) != 2 {
			return nil, fmt.Errorf("invalid userinfo format")
		}
		method = methodPass[0]
		password = methodPass[1]

		serverPort := strings.Split(serverInfo, ":")
		if len(serverPort) != 2 {
			return nil, fmt.Errorf("invalid server:port format")
		}
		server = serverPort[0]
		port = serverPort[1]
	}

	portNum, err := strconv.Atoi(port)
	if err != nil || portNum == 0 {
		return nil, fmt.Errorf("invalid port: %s", port)
	}

	if remark == "" {
		remark = server + ":" + port
	}

	if group == "" {
		group = "SS"
	}

	p := &proxy.ShadowsocksProxy{
		BaseProxy: proxy.BaseProxy{
			Type:   "ss",
			Remark: remark,
			Server: server,
			Port:   portNum,
			Group:  group,
		},
		Password:      password,
		EncryptMethod: method,
		Plugin:        plugin,
		PluginOpts:    pluginOpts,
	}

	mihomoProxy, err := adapter.ParseProxy(p.ProxyOptions())
	if err != nil {
		return p, nil
	} else {
		return &proxy.MihomoProxy{
			ProxyInterface: p,
			Clash:          mihomoProxy,
			Options:        p.ProxyOptions(),
		}, nil
	}

}

func parseShadowsocksR(line string) (proxy.SubconverterProxy, error) {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "ssr://") {
		return nil, fmt.Errorf("not a valid ssr:// link")
	}

	line = strings.ReplaceAll(line[6:], "\r", "")
	line = urlSafeBase64Decode(line)

	var remarks, group, server, port, method, password, protocol, protocolParam, obfs, obfsParam string

	if idx := strings.Index(line, "/?"); idx != -1 {
		queryStr := line[idx+2:]
		line = line[:idx]

		params, _ := url.ParseQuery(queryStr)
		group = urlSafeBase64Decode(params.Get("group"))
		remarks = urlSafeBase64Decode(params.Get("remarks"))
		obfsParam = strings.TrimSpace(urlSafeBase64Decode(params.Get("obfsparam")))
		protocolParam = strings.TrimSpace(urlSafeBase64Decode(params.Get("protoparam")))
	}

	re := regexp.MustCompile(`(\S+):(\d+?):(\S+?):(\S+?):(\S+?):(\S+)`)
	matches := re.FindStringSubmatch(line)
	if len(matches) != 7 {
		return nil, fmt.Errorf("invalid ssr link format")
	}

	server = matches[1]
	port = matches[2]
	protocol = matches[3]
	method = matches[4]
	obfs = matches[5]
	password = urlSafeBase64Decode(matches[6])

	portNum, err := strconv.Atoi(port)
	if err != nil || portNum == 0 {
		return nil, fmt.Errorf("invalid port: %s", port)
	}

	if group == "" {
		group = "SSR"
	}
	if remarks == "" {
		remarks = server + ":" + port
	}

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
	var p *proxy.ShadowsocksRProxy
	if isSS {
		p = &proxy.ShadowsocksRProxy{
			BaseProxy: proxy.BaseProxy{
				Type:   "ss",
				Remark: remarks,
				Server: server,
				Port:   portNum,
				Group:  group,
			},
			Password:      password,
			EncryptMethod: method,
		}
	} else {

	}

	p = &proxy.ShadowsocksRProxy{
		BaseProxy: proxy.BaseProxy{
			Type:   "ssr",
			Remark: remarks,
			Server: server,
			Port:   portNum,
			Group:  group,
		},
		Password:      password,
		EncryptMethod: method,
		Protocol:      protocol,
		ProtocolParam: protocolParam,
		Obfs:          obfs,
		ObfsParam:     obfsParam,
	}
	mihomoProxy, err := adapter.ParseProxy(p.ProxyOptions())
	if err != nil {
		log.Printf("mihomo proxy parse failed %v", err)
		return p, nil
	} else {
		return &proxy.MihomoProxy{
			ProxyInterface: p,
			Clash:          mihomoProxy,
			Options:        p.ProxyOptions(),
		}, nil
	}
}

func parseVMess(line string) (proxy.SubconverterProxy, error) {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "vmess://") {
		return nil, fmt.Errorf("not a valid vmess:// link")
	}

	line = line[8:]
	decoded := urlSafeBase64Decode(line)

	var vmessData map[string]interface{}
	if err := json.Unmarshal([]byte(decoded), &vmessData); err != nil {
		return nil, fmt.Errorf("failed to parse vmess JSON: %w", err)
	}

	ps := getStringField(vmessData, "ps")
	add := getStringField(vmessData, "add")
	port := getStringField(vmessData, "port")
	id := getStringField(vmessData, "id")
	aid := getStringField(vmessData, "aid")
	net := getStringField(vmessData, "net")
	host := getStringField(vmessData, "host")
	path := getStringField(vmessData, "path")
	tls := getStringField(vmessData, "tls")
	sni := getStringField(vmessData, "sni")

	if net == "" {
		net = "tcp"
	}
	if aid == "" {
		aid = "0"
	}

	portNum, err := strconv.Atoi(port)
	if err != nil || portNum == 0 {
		return nil, fmt.Errorf("invalid port: %s", port)
	}

	alterID, _ := strconv.Atoi(aid)

	if ps == "" {
		ps = add + ":" + port
	}

	version := getStringField(vmessData, "v")
	if version == "1" || version == "" {
		if host != "" && strings.Contains(host, ";") {
			parts := strings.SplitN(host, ";", 2)
			host = parts[0]
			if path == "" {
				path = parts[1]
			}
		}
	}

	p := &proxy.VMessProxy{
		BaseProxy: proxy.BaseProxy{
			Type:   "vmess",
			Remark: ps,
			Server: add,
			Port:   portNum,
			Group:  "VMess",
		},
		UUID:    id,
		AlterID: alterID,
		Network: net,
		Path:    path,
		Host:    host,
		TLS:     tls == "tls",
		SNI:     sni,
	}
	mihomoProxy, err := adapter.ParseProxy(p.ProxyOptions())
	if err != nil {
		log.Printf("mihomo proxy parse failed %v", err)
		return p, nil
	} else {
		return &proxy.MihomoProxy{
			ProxyInterface: p,
			Clash:          mihomoProxy,
			Options:        p.ProxyOptions(),
		}, nil
	}

}

func parseTrojan(line string) (proxy.SubconverterProxy, error) {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "trojan://") {
		return nil, fmt.Errorf("not a valid trojan:// link")
	}

	line = line[9:]

	var remark, password, server, port, sni, network, path, group string
	var allowInsecure bool

	if idx := strings.LastIndex(line, "#"); idx != -1 {
		remark = urlDecode(line[idx+1:])
		line = line[:idx]
	}

	if idx := strings.Index(line, "?"); idx != -1 {
		queryStr := line[idx+1:]
		line = line[:idx]

		params, _ := url.ParseQuery(queryStr)

		sni = params.Get("sni")
		if sni == "" {
			sni = params.Get("peer")
		}

		if params.Get("allowInsecure") == "1" || params.Get("allowInsecure") == "true" {
			allowInsecure = true
		}

		group = urlDecode(params.Get("group"))

		if params.Get("ws") == "1" {
			network = "ws"
			path = params.Get("wspath")
		} else if params.Get("type") == "ws" {
			network = "ws"
			path = params.Get("path")
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

	re := regexp.MustCompile(`(.*?)@(.*):(.*)`)
	matches := re.FindStringSubmatch(line)
	if len(matches) != 4 {
		return nil, fmt.Errorf("invalid trojan link format")
	}

	password = matches[1]
	server = matches[2]
	port = matches[3]

	portNum, err := strconv.Atoi(port)
	if err != nil || portNum == 0 {
		return nil, fmt.Errorf("invalid port: %s", port)
	}

	if remark == "" {
		remark = server + ":" + port
	}
	if group == "" {
		group = "Trojan"
	}

	p := &proxy.TrojanProxy{
		BaseProxy: proxy.BaseProxy{
			Type:   "trojan",
			Remark: remark,
			Server: server,
			Port:   portNum,
			Group:  group,
		},
		Password:      password,
		Network:       network,
		Path:          path,
		Host:          sni,
		TLS:           true, // Trojan always uses TLS
		AllowInsecure: allowInsecure,
	}
	mihomoProxy, err := adapter.ParseProxy(p.ProxyOptions())
	if err != nil {
		log.Printf("mihomo proxy parse failed %v", err)

		return p, nil
	} else {
		return &proxy.MihomoProxy{
			ProxyInterface: p,
			Clash:          mihomoProxy,
			Options:        p.ProxyOptions(),
		}, nil
	}
}

func parseVLESS(line string) (proxy.SubconverterProxy, error) {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "vless://") {
		return nil, fmt.Errorf("not a valid vless:// link")
	}

	line = line[8:]

	var remark, uuid, server, port, network, flow, security, sni, path, host, group string
	var allowInsecure bool

	if idx := strings.LastIndex(line, "#"); idx != -1 {
		remark = urlDecode(line[idx+1:])
		line = line[:idx]
	}

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
			host = params.Get("quicSecurity")
			path = params.Get("key")
		}
	}

	re := regexp.MustCompile(`(.*?)@(.*):(.*)`)
	matches := re.FindStringSubmatch(line)
	if len(matches) != 4 {
		return nil, fmt.Errorf("invalid vless link format")
	}

	uuid = matches[1]
	server = matches[2]
	port = matches[3]

	portNum, err := strconv.Atoi(port)
	if err != nil || portNum == 0 {
		return nil, fmt.Errorf("invalid port: %s", port)
	}

	if remark == "" {
		remark = server + ":" + port
	}
	if group == "" {
		group = "VLESS"
	}

	p := &proxy.VLESSProxy{
		BaseProxy: proxy.BaseProxy{
			Type:   "vless",
			Remark: remark,
			Server: server,
			Port:   portNum,
			Group:  group,
		},
		UUID:          uuid,
		Network:       network,
		Path:          path,
		Host:          host,
		TLS:           security == "tls" || security == "reality",
		AllowInsecure: allowInsecure,
		Flow:          flow,
		SNI:           sni,
	}
	mihomoProxy, err := adapter.ParseProxy(p.ProxyOptions())
	if err != nil {
		log.Printf("mihomo proxy parse failed %v", err)
		return p, nil
	} else {
		return &proxy.MihomoProxy{
			ProxyInterface: p,
			Clash:          mihomoProxy,
			Options:        p.ProxyOptions(),
		}, nil
	}

}

func parseHysteria(line string) (proxy.SubconverterProxy, error) {
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
		return nil, fmt.Errorf("not a valid hysteria link")
	}

	prefixLen := len(protocol) + 3
	line = line[prefixLen:]

	var remark, server, port, password, obfs string
	var insecure bool

	if idx := strings.LastIndex(line, "#"); idx != -1 {
		remark = urlDecode(line[idx+1:])
		line = line[:idx]
	}

	var params url.Values
	if idx := strings.Index(line, "?"); idx != -1 {
		queryStr := line[idx+1:]
		line = line[:idx]
		params, _ = url.ParseQuery(queryStr)

		if params.Get("insecure") == "1" || params.Get("insecure") == "true" {
			insecure = true
		}
		params.Del("insecure")
		obfs = params.Get("obfs")
		params.Del("obfs")
	}

	if strings.Contains(line, "@") {
		parts := strings.SplitN(line, "@", 2)
		password = parts[0]
		line = parts[1]
	}

	serverPort := strings.Split(line, ":")
	if len(serverPort) != 2 {
		return nil, fmt.Errorf("invalid server:port format")
	}
	server = serverPort[0]
	port = serverPort[1]

	portNum, err := strconv.Atoi(port)
	if err != nil || portNum == 0 {
		return nil, fmt.Errorf("invalid port: %s", port)
	}

	if remark == "" {
		remark = server + ":" + port
	}

	p := &proxy.HysteriaProxy{
		BaseProxy: proxy.BaseProxy{
			Type:   protocol,
			Remark: remark,
			Server: server,
			Port:   portNum,
			Group:  strings.ToUpper(protocol),
		},
		Password:      password,
		Obfs:          obfs,
		AllowInsecure: insecure,
		Params:        params,
	}
	mihomoProxy, err := adapter.ParseProxy(p.ProxyOptions())
	if err != nil {
		log.Printf("mihomo proxy parse failed %v", err)
		return p, nil
	} else {
		return &proxy.MihomoProxy{
			ProxyInterface: p,
			Clash:          mihomoProxy,
			Options:        p.ProxyOptions(),
		}, nil
	}
}

func parseTUIC(line string) (proxy.SubconverterProxy, error) {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "tuic://") {
		return nil, fmt.Errorf("not a valid tuic link")
	}

	line = line[7:]

	var remark, uuid, password, server, port string
	var insecure bool

	if idx := strings.LastIndex(line, "#"); idx != -1 {
		remark = urlDecode(line[idx+1:])
		line = line[:idx]
	}

	var params url.Values
	if idx := strings.Index(line, "?"); idx != -1 {
		queryStr := line[idx+1:]
		line = line[:idx]
		params, _ = url.ParseQuery(queryStr)

		if params.Get("allow_insecure") == "1" || params.Get("allow_insecure") == "true" {
			insecure = true
		}
		params.Del("allow_insecure")
	}

	if strings.Contains(line, "@") {
		parts := strings.SplitN(line, "@", 2)
		auth := parts[0]
		line = parts[1]

		authParts := strings.SplitN(auth, ":", 2)
		uuid = authParts[0]
		if len(authParts) == 2 {
			password = authParts[1]
		}
	}

	serverPort := strings.Split(line, ":")
	if len(serverPort) != 2 {
		return nil, fmt.Errorf("invalid server:port format")
	}
	server = serverPort[0]
	port = serverPort[1]

	portNum, err := strconv.Atoi(port)
	if err != nil || portNum == 0 {
		return nil, fmt.Errorf("invalid port: %s", port)
	}

	if remark == "" {
		remark = server + ":" + port
	}

	p := &proxy.TUICProxy{
		BaseProxy: proxy.BaseProxy{
			Type:   "tuic",
			Remark: remark,
			Server: server,
			Port:   portNum,
			Group:  "TUIC",
		},
		UUID:          uuid,
		Password:      password,
		AllowInsecure: insecure,
		Params:        params,
	}
	mihomoProxy, err := adapter.ParseProxy(p.ProxyOptions())
	if err != nil {
		log.Printf("mihomo proxy parse failed %v", err)
		return p, nil
	} else {
		return &proxy.MihomoProxy{
			ProxyInterface: p,
			Clash:          mihomoProxy,
			Options:        p.ProxyOptions(),
		}, nil
	}
}

// ParseMihomoConfig parses Clash YAML format and returns proxies
func ParseMihomoConfig(content string) (*CustomContent, error) {
	// Try parsing as YAML (Clash format)
	var clashConfig C.RawConfig

	if err := yaml.Unmarshal([]byte(content), &clashConfig); err != nil {
		return nil, fmt.Errorf("failed to parse clash format: %w", err)
	}

	if len(clashConfig.Proxy) == 0 {
		return nil, fmt.Errorf("no proxies found in clash format")
	}
	custom := &CustomContent{Proxies: make([]proxy.ProxyInterface, 0)}

	for _, proxyMap := range clashConfig.Proxy {
		mihomoProxy, err := adapter.ParseProxy(proxyMap)

		if err != nil {
			log.Printf("failed to parse proxy in clash format: %v", err)
			continue // Skip invalid proxies
		}
		addr := mihomoProxy.Addr()
		// parse addr to server and port
		hostPort := strings.Split(addr, ":")
		if len(hostPort) != 2 {
			log.Printf("invalid address format: %s", addr)
			continue
		}
		server := hostPort[0]
		port, err := strconv.Atoi(hostPort[1])
		if err != nil {
			log.Printf("invalid port in address: %s", addr)
			continue
		}
		custom.Proxies = append(custom.Proxies, &proxy.MihomoProxy{
			ProxyInterface: &proxy.BaseProxy{
				Type:   mihomoProxy.Type().String(),
				Remark: mihomoProxy.Name(),
				Server: server,
				Port:   port,
			},
			Clash:   mihomoProxy,
			Options: proxyMap,
		})
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
