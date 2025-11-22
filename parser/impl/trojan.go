package impl

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

type TrojanParser struct{}

func (p *TrojanParser) Name() string {
	return "Trojan"
}

func (p *TrojanParser) CanParse(line string) bool {
	return strings.HasPrefix(line, "trojan://")
}

func (p *TrojanParser) Parse(line string) (core.SubconverterProxy, error) {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "trojan://") {
		return nil, fmt.Errorf("not a valid trojan:// link")
	}

	line = line[9:]

	var remark, password, server, port, sni, network, path, group string
	var allowInsecure bool

	if idx := strings.LastIndex(line, "#"); idx != -1 {
		remark = utils.UrlDecode(line[idx+1:])
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

		group = utils.UrlDecode(params.Get("group"))

		if params.Get("ws") == "1" {
			network = "ws"
			path = params.Get("wspath")
		} else if params.Get("type") == "ws" {
			network = "ws"
			path = params.Get("path")
			if strings.HasPrefix(path, "%2F") {
				path = utils.UrlDecode(path)
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

	pObj := &impl.TrojanProxy{
		BaseProxy: core.BaseProxy{
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
	return utils.ToMihomoProxy(pObj)
}
