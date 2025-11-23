package proxy

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/gfunc/subconvergo/parser/utils"
	"github.com/gfunc/subconvergo/proxy/core"
	"github.com/gfunc/subconvergo/proxy/impl"
)

type HysteriaParser struct{}

func (p *HysteriaParser) Name() string {
	return "Hysteria"
}

func (p *HysteriaParser) CanParseLine(line string) bool {
	return strings.HasPrefix(line, "hysteria://")
}

func (p *HysteriaParser) Parse(line string) (core.SubconverterProxy, error) {
	line = strings.TrimSpace(line)

	if !strings.HasPrefix(line, "hysteria://") {
		return nil, fmt.Errorf("not a valid hysteria link")
	}

	protocol := "hysteria"
	line = line[11:] // len("hysteria://")

	var remark, server, port, password, obfs string
	var insecure bool

	if idx := strings.LastIndex(line, "#"); idx != -1 {
		remark = utils.UrlDecode(line[idx+1:])
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
	} else {
		if pass := params.Get("password"); pass != "" {
			password = pass
			params.Del("password")
		} else if pass := params.Get("auth"); pass != "" {
			password = pass
			params.Del("auth")
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

	pObj := &impl.HysteriaProxy{
		BaseProxy: core.BaseProxy{
			Type:   protocol,
			Remark: remark,
			Server: server,
			Port:   portNum,
			Group:  core.HYSTERIA_DEFAULT_GROUP,
		},
		Password:      password,
		Obfs:          obfs,
		AllowInsecure: insecure,
		Params:        params,
	}
	return utils.ToMihomoProxy(pObj)
}
