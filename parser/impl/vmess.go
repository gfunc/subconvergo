package impl

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/gfunc/subconvergo/proxy/core"
	"github.com/gfunc/subconvergo/proxy/impl"
	"github.com/metacubex/mihomo/adapter"
)

type VMessParser struct{}

func (p *VMessParser) Name() string {
	return "VMess"
}

func (p *VMessParser) CanParse(line string) bool {
	return strings.HasPrefix(line, "vmess://")
}

func (p *VMessParser) Parse(line string) (core.SubconverterProxy, error) {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "vmess://") {
		return nil, fmt.Errorf("not a valid vmess:// link")
	}

	line = line[8:]

	// Check for standard VMess format (uuid@host:port)
	if strings.Contains(line, "@") && !strings.Contains(line, "://") { // Avoid matching other protocols if any
		return p.parseStdVMess(line)
	}

	decoded := urlSafeBase64Decode(line)

	// Check if it's JSON
	if strings.HasPrefix(strings.TrimSpace(decoded), "{") {
		return p.parseJSONVMess(decoded)
	}

	// Check if it's base64 encoded standard VMess
	if strings.Contains(decoded, "@") {
		return p.parseStdVMess(decoded)
	}

	return nil, fmt.Errorf("unknown vmess format")
}

func (p *VMessParser) parseStdVMess(line string) (core.SubconverterProxy, error) {
	parts := strings.SplitN(line, "@", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid standard vmess format")
	}
	uuid := parts[0]
	rest := parts[1]

	var remark string
	var params url.Values

	if idx := strings.Index(rest, "#"); idx != -1 {
		remark = urlDecode(rest[idx+1:])
		rest = rest[:idx]
	}

	if idx := strings.Index(rest, "?"); idx != -1 {
		queryStr := rest[idx+1:]
		rest = rest[:idx]
		params, _ = url.ParseQuery(queryStr)
	}

	serverPort := strings.Split(rest, ":")
	if len(serverPort) != 2 {
		return nil, fmt.Errorf("invalid server:port")
	}
	server := serverPort[0]
	portStr := serverPort[1]

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid port: %s", portStr)
	}

	if remark == "" {
		remark = params.Get("remark")
		if remark == "" {
			remark = server + ":" + portStr
		}
	}

	net := params.Get("network")
	if net == "" {
		net = "tcp"
	}

	aidStr := params.Get("aid")
	aid := 0
	if aidStr != "" {
		aid, _ = strconv.Atoi(aidStr)
	}

	tls := params.Get("tls") == "1" || params.Get("tls") == "true" || params.Get("tls") == "tls"
	sni := params.Get("sni")
	if sni == "" {
		sni = params.Get("peer")
	}
	path := params.Get("path")
	host := params.Get("host")
	if host == "" {
		host = params.Get("obfsParam")
	}

	pObj := &impl.VMessProxy{
		BaseProxy: core.BaseProxy{
			Type:   "vmess",
			Remark: remark,
			Server: server,
			Port:   port,
			Group:  "VMess",
		},
		UUID:    uuid,
		AlterID: aid,
		Network: net,
		Path:    path,
		Host:    host,
		TLS:     tls,
		SNI:     sni,
	}
	return p.toMihomoProxy(pObj)
}

func (p *VMessParser) parseJSONVMess(decoded string) (core.SubconverterProxy, error) {
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

	pObj := &impl.VMessProxy{
		BaseProxy: core.BaseProxy{
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
	return p.toMihomoProxy(pObj)
}

func (p *VMessParser) toMihomoProxy(pObj *impl.VMessProxy) (core.SubconverterProxy, error) {
	option := pObj.ToClashConfig(nil)
	mihomoProxy, err := adapter.ParseProxy(option)
	if err != nil {
		return pObj, nil
	} else {
		return &impl.MihomoProxy{
			ProxyInterface: pObj,
			Clash:          mihomoProxy,
			Options:        option,
		}, nil
	}
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
