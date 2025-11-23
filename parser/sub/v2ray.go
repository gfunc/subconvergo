package sub

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/gfunc/subconvergo/parser/core"
	"github.com/gfunc/subconvergo/parser/proxy"
	proxyCore "github.com/gfunc/subconvergo/proxy/core"
)

type V2RaySubscriptionParser struct{}

func (p *V2RaySubscriptionParser) Name() string {
	return "V2Ray"
}

func (p *V2RaySubscriptionParser) CanParse(content string) bool {
	return strings.Contains(content, "\"uiItem\"") || strings.Contains(content, "vnext")
}

func (p *V2RaySubscriptionParser) Parse(content string) (*core.SubContent, error) {
	var js map[string]interface{}
	if err := json.Unmarshal([]byte(content), &js); err != nil {
		return nil, fmt.Errorf("failed to parse V2Ray JSON: %w", err)
	}

	outbounds, ok := js["outbounds"].([]interface{})
	if !ok || len(outbounds) == 0 {
		return nil, fmt.Errorf("no outbounds found")
	}

	// subconverter only parses the first outbound
	outbound, ok := outbounds[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid outbound format")
	}

	settings, ok := outbound["settings"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("no settings in outbound")
	}

	vnext, ok := settings["vnext"].([]interface{})
	if !ok || len(vnext) == 0 {
		return nil, fmt.Errorf("no vnext in settings")
	}

	serverObj, ok := vnext[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid vnext format")
	}

	address := toString(serverObj["address"])
	port := toString(serverObj["port"])

	users, ok := serverObj["users"].([]interface{})
	if !ok || len(users) == 0 {
		return nil, fmt.Errorf("no users in vnext")
	}
	user, ok := users[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid user format")
	}

	id := toString(user["id"])
	aid := toString(user["alterId"])
	security := toString(user["security"])

	streamSettings, _ := outbound["streamSettings"].(map[string]interface{})
	network := toString(streamSettings["network"])
	securityType := toString(streamSettings["security"])
	tls := securityType == "tls"

	var path, host, sni, typeStr string
	var net = network
	if net == "" {
		net = "tcp"
	}

	if net == "ws" {
		wsSettings, _ := streamSettings["wsSettings"].(map[string]interface{})
		path = toString(wsSettings["path"])
		headers, _ := wsSettings["headers"].(map[string]interface{})
		host = toString(headers["Host"])
	} else if net == "tcp" {
		tcpSettings, _ := streamSettings["tcpSettings"].(map[string]interface{})
		header, _ := tcpSettings["header"].(map[string]interface{})
		typeStr = toString(header["type"])
		// ...
	}
	// ... more parsing logic for other transports ...

	if tls {
		tlsSettings, _ := streamSettings["tlsSettings"].(map[string]interface{})
		sni = toString(tlsSettings["serverName"])
	}

	config := map[string]interface{}{
		"address":  address,
		"port":     port,
		"id":       id,
		"alterId":  aid,
		"security": security,
		"network":  net,
		"type":     typeStr,
		"host":     host,
		"path":     path,
		"tls":      strconv.FormatBool(tls),
		"sni":      sni,
		"remark":   "V2Ray Config",
	}

	parser := &proxy.VMessParser{}
	prx, err := parser.ParseV2Ray(config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse V2Ray proxy: %w", err)
	}

	return &core.SubContent{
		Proxies: []proxyCore.ProxyInterface{prx},
	}, nil
}
