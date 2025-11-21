package impl

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/gfunc/subconvergo/proxy/core"
	"github.com/gfunc/subconvergo/proxy/impl"
	"github.com/metacubex/mihomo/adapter"
)

type HysteriaParser struct{}

func (p *HysteriaParser) Name() string {
	return "Hysteria"
}

func (p *HysteriaParser) CanParse(line string) bool {
	return strings.HasPrefix(line, "hysteria://") || strings.HasPrefix(line, "hysteria2://") || strings.HasPrefix(line, "hy2://")
}

func (p *HysteriaParser) Parse(line string) (core.SubconverterProxy, error) {
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

	pObj := &impl.HysteriaProxy{
		BaseProxy: core.BaseProxy{
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
