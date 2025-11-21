package impl

import (
	"encoding/json"
	"fmt"
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
	mihomoProxy, err := adapter.ParseProxy(pObj.ToClashConfig(nil))
	if err != nil {
		return pObj, nil
	} else {
		return &impl.MihomoProxy{
			ProxyInterface: pObj,
			Clash:          mihomoProxy,
			Options:        pObj.ToClashConfig(nil),
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
