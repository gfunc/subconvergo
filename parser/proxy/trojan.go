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

type TrojanParser struct{}

func (p *TrojanParser) Name() string {
	return "Trojan"
}

func (p *TrojanParser) CanParseLine(line string) bool {
	return strings.HasPrefix(line, "trojan://") || isSurgeTrojan(line)
}

func isSurgeTrojan(line string) bool {
	parts := strings.SplitN(line, "=", 2)
	if len(parts) != 2 {
		return false
	}
	val := strings.TrimSpace(parts[1])
	return strings.HasPrefix(val, "trojan,")
}

func (p *TrojanParser) Parse(line string) (core.SubconverterProxy, error) {
	line = strings.TrimSpace(line)
	if isSurgeTrojan(line) {
		return p.ParseSurge(line)
	}

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
		group = core.TROJAN_DEFAULT_GROUP
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

// ParseSurge parses a Surge format line
func (p *TrojanParser) ParseSurge(line string) (core.SubconverterProxy, error) {
	parts := strings.SplitN(line, "=", 2)
	remark := strings.TrimSpace(parts[0])
	params := strings.Split(strings.TrimSpace(parts[1]), ",")

	if len(params) < 4 {
		return nil, fmt.Errorf("invalid surge trojan format")
	}

	// params[0] is "trojan"
	server := strings.TrimSpace(params[1])
	portStr := strings.TrimSpace(params[2])
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid port: %s", portStr)
	}

	trojan := &impl.TrojanProxy{
		BaseProxy: core.BaseProxy{
			Type:   "trojan",
			Server: server,
			Port:   port,
			Remark: remark,
		},
	}

	for i := 3; i < len(params); i++ {
		arg := strings.TrimSpace(params[i])
		if strings.HasPrefix(arg, "password=") {
			trojan.Password = strings.TrimPrefix(arg, "password=")
		} else if strings.HasPrefix(arg, "sni=") {
			trojan.Host = strings.TrimPrefix(arg, "sni=")
		} else if strings.HasPrefix(arg, "skip-cert-verify=") {
			trojan.AllowInsecure = strings.TrimPrefix(arg, "skip-cert-verify=") == "true"
		}
		// Handle other args
	}

	return trojan, nil
}

// ParseClash parses a Clash config map
func (p *TrojanParser) ParseClash(config map[string]interface{}) (core.SubconverterProxy, error) {
	server := utils.GetStringField(config, "server")
	port := utils.GetIntField(config, "port")
	name := utils.GetStringField(config, "name")
	password := utils.GetStringField(config, "password")
	sni := utils.GetStringField(config, "sni")
	skipCertVerify := utils.GetStringField(config, "skip-cert-verify") == "true" || config["skip-cert-verify"] == true
	network := utils.GetStringField(config, "network")

	// ws-opts, grpc-opts, etc.
	var path, host string
	if wsOpts, ok := config["ws-opts"].(map[string]interface{}); ok {
		path = utils.GetStringField(wsOpts, "path")
		if headers, ok := wsOpts["headers"].(map[string]interface{}); ok {
			host = utils.GetStringField(headers, "Host")
		}
	}

	trojan := &impl.TrojanProxy{
		BaseProxy: core.BaseProxy{
			Type:   "trojan",
			Server: server,
			Port:   port,
			Remark: name,
		},
		Password:      password,
		Host:          sni, // SNI is usually Host in Trojan
		AllowInsecure: skipCertVerify,
		Network:       network,
		Path:          path,
	}
	if host != "" {
		// If host is specified in headers, it might override SNI or be used for WS
		// TrojanProxy struct has Host field which is usually SNI/Peer
	}
	return trojan, nil
}
