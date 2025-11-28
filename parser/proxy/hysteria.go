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

func (p *HysteriaParser) ParseSingle(line string) (core.ParsableProxy, error) {
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

// ParseClash parses a Clash config map
func (p *HysteriaParser) ParseClash(config map[string]interface{}) (core.ParsableProxy, error) {
	server := utils.GetStringField(config, "server")
	port := utils.GetIntField(config, "port")
	name := utils.GetStringField(config, "name")

	proxyType := utils.GetStringField(config, "type")

	password := utils.GetStringField(config, "auth-str")
	if password == "" {
		password = utils.GetStringField(config, "password")
	}

	obfs := utils.GetStringField(config, "obfs")

	params := url.Values{}
	if up := utils.GetStringField(config, "up"); up != "" {
		params.Set("up", up)
	}
	if down := utils.GetStringField(config, "down"); down != "" {
		params.Set("down", down)
	}
	if sni := utils.GetStringField(config, "sni"); sni != "" {
		params.Set("sni", sni)
	}
	if skipCertVerify := config["skip-cert-verify"]; skipCertVerify == true {
		params.Set("insecure", "1")
	}

	// Handle ALPN
	if alpn, ok := config["alpn"].([]interface{}); ok {
		var alpnStrs []string
		for _, a := range alpn {
			if s, ok := a.(string); ok {
				alpnStrs = append(alpnStrs, s)
			}
		}
		if len(alpnStrs) > 0 {
			params.Set("alpn", strings.Join(alpnStrs, ","))
		}
	}

	h := &impl.HysteriaProxy{
		BaseProxy: core.BaseProxy{
			Type:   proxyType,
			Server: server,
			Port:   port,
			Remark: name,
		},
		Password:      password,
		Obfs:          obfs,
		AllowInsecure: params.Get("insecure") == "1",
		Params:        params,
	}
	return utils.ToMihomoProxy(h)
}
