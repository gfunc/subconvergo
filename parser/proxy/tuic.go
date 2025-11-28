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

type TUICParser struct{}

func (p *TUICParser) Name() string {
	return "TUIC"
}

func (p *TUICParser) CanParseLine(line string) bool {
	return strings.HasPrefix(line, "tuic://")
}

func (p *TUICParser) ParseSingle(line string) (core.ParsableProxy, error) {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "tuic://") {
		return nil, fmt.Errorf("not a valid tuic link")
	}

	line = line[7:]

	var remark, uuid, password, server, port string
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

	pObj := &impl.TUICProxy{
		BaseProxy: core.BaseProxy{
			Type:   "tuic",
			Remark: remark,
			Server: server,
			Port:   portNum,
			Group:  core.TUIC_DEFAULT_GROUP,
		},
		UUID:          uuid,
		Password:      password,
		AllowInsecure: insecure,
		Params:        params,
	}
	return utils.ToMihomoProxy(pObj)
}

func (p *TUICParser) ParseClash(option map[string]interface{}) (core.ParsableProxy, error) {
	server := utils.GetStringField(option, "server")
	port := utils.GetIntField(option, "port")
	name := utils.GetStringField(option, "name")
	uuid := utils.GetStringField(option, "uuid")
	password := utils.GetStringField(option, "password")

	if server == "" || port == 0 || uuid == "" {
		return nil, fmt.Errorf("missing required fields for tuic")
	}

	params := url.Values{}
	if cc := utils.GetStringField(option, "congestion-controller"); cc != "" {
		params.Set("congestion_control", cc)
	}
	if udp := utils.GetStringField(option, "udp-relay-mode"); udp != "" {
		params.Set("udp_relay_mode", udp)
	}
	if sni := utils.GetStringField(option, "sni"); sni != "" {
		params.Set("sni", sni)
	}
	if alpn := utils.GetStringField(option, "alpn"); alpn != "" {
		// alpn in clash is list, but we store as comma separated string in params for now?
		// Or maybe TUICProxy should handle it better.
		// utils.GetStringField might not work for list.
		// But let's assume it's a string or handle it.
		// Clash alpn is []string.
		if alpnList, ok := option["alpn"].([]interface{}); ok {
			var alpns []string
			for _, a := range alpnList {
				if s, ok := a.(string); ok {
					alpns = append(alpns, s)
				}
			}
			if len(alpns) > 0 {
				params.Set("alpn", strings.Join(alpns, ","))
			}
		}
	}

	pObj := &impl.TUICProxy{
		BaseProxy: core.BaseProxy{
			Type:   "tuic",
			Remark: name,
			Server: server,
			Port:   port,
			Group:  core.TUIC_DEFAULT_GROUP,
		},
		UUID:          uuid,
		Password:      password,
		AllowInsecure: utils.GetBoolField(option, "skip-cert-verify"),
		Params:        params,
	}
	return utils.ToMihomoProxy(pObj)
}
