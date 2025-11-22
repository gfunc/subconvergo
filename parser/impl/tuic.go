package impl

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

func (p *TUICParser) CanParse(line string) bool {
	return strings.HasPrefix(line, "tuic://")
}

func (p *TUICParser) Parse(line string) (core.SubconverterProxy, error) {
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
			Group:  "TUIC",
		},
		UUID:          uuid,
		Password:      password,
		AllowInsecure: insecure,
		Params:        params,
	}
	return utils.ToMihomoProxy(pObj)
}
