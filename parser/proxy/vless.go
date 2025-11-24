package proxy

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/gfunc/subconvergo/parser/utils"
	"github.com/gfunc/subconvergo/proxy/core"
	"github.com/gfunc/subconvergo/proxy/impl"
)

type VLESSParser struct{}

func (p *VLESSParser) Name() string {
	return "VLESS"
}

func (p *VLESSParser) CanParseLine(line string) bool {
	return strings.HasPrefix(line, "vless://")
}

func (p *VLESSParser) ParseSingle(line string) (core.SubconverterProxy, error) {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "vless://") {
		return nil, fmt.Errorf("not a valid vless:// link")
	}

	line = line[8:]

	var remark, uuid, server, port, network, flow, security, sni, path, host, group string
	var allowInsecure bool

	if idx := strings.LastIndex(line, "#"); idx != -1 {
		remark = utils.UrlDecode(line[idx+1:])
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

		group = utils.UrlDecode(params.Get("group"))

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
		group = core.VLESS_DEFAULT_GROUP
	}

	pObj := &impl.VLESSProxy{
		BaseProxy: core.BaseProxy{
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
	return utils.ToMihomoProxy(pObj)
}

// ParseClash parses a Clash config map
func (p *VLESSParser) ParseClash(config map[string]interface{}) (core.SubconverterProxy, error) {
	server := utils.GetStringField(config, "server")
	port := utils.GetIntField(config, "port")
	name := utils.GetStringField(config, "name")
	uuid := utils.GetStringField(config, "uuid")
	network := utils.GetStringField(config, "network")
	tls := config["tls"] == true
	flow := utils.GetStringField(config, "flow")
	servername := utils.GetStringField(config, "servername")

	v := &impl.VLESSProxy{
		BaseProxy: core.BaseProxy{
			Type:   "vless",
			Server: server,
			Port:   port,
			Remark: name,
		},
		UUID:    uuid,
		Network: network,
		TLS:     tls,
		Flow:    flow,
		SNI:     servername,
	}

	if config["skip-cert-verify"] == true {
		v.AllowInsecure = true
	}

	if wsOpts, ok := config["ws-opts"].(map[string]interface{}); ok {
		v.Path = utils.GetStringField(wsOpts, "path")
		if headers, ok := wsOpts["headers"].(map[string]interface{}); ok {
			v.Host = utils.GetStringField(headers, "Host")
		}
	}
	if grpcOpts, ok := config["grpc-opts"].(map[string]interface{}); ok {
		v.Path = utils.GetStringField(grpcOpts, "grpc-service-name")
	}

	return utils.ToMihomoProxy(v)
}
