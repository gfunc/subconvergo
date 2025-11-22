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

type ShadowsocksParser struct{}

func (p *ShadowsocksParser) Name() string {
	return "Shadowsocks"
}

func (p *ShadowsocksParser) CanParse(line string) bool {
	return strings.HasPrefix(line, "ss://")
}

func (p *ShadowsocksParser) Parse(line string) (core.SubconverterProxy, error) {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "ss://") {
		return nil, fmt.Errorf("not a valid ss:// link")
	}

	line = line[5:]

	var remark, password, method, server, port, plugin, pluginOpts, group string

	if idx := strings.Index(line, "#"); idx != -1 {
		remark = utils.UrlDecode(line[idx+1:])
		line = line[:idx]
	}

	if idx := strings.Index(line, "?"); idx != -1 {
		queryStr := line[idx+1:]
		line = line[:idx]

		params, _ := url.ParseQuery(queryStr)
		if pluginStr := params.Get("plugin"); pluginStr != "" {
			pluginStr = utils.UrlDecode(pluginStr)
			if idx := strings.Index(pluginStr, ";"); idx != -1 {
				plugin = pluginStr[:idx]
				pluginOpts = pluginStr[idx+1:]
			} else {
				plugin = pluginStr
			}
		}

		if groupStr := params.Get("group"); groupStr != "" {
			group = utils.UrlSafeBase64Decode(groupStr)
		}
	}

	line = strings.TrimSuffix(line, "/")

	if strings.Contains(line, "@") {
		parts := strings.Split(line, "@")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid ss link format")
		}

		userInfo := utils.UrlSafeBase64Decode(parts[0])
		methodPass := strings.SplitN(userInfo, ":", 2)
		if len(methodPass) != 2 {
			return nil, fmt.Errorf("invalid userinfo format")
		}
		method = methodPass[0]
		password = methodPass[1]

		serverInfo := parts[1]

		if strings.HasPrefix(serverInfo, "[") {
			endIdx := strings.Index(serverInfo, "]")
			if endIdx == -1 {
				return nil, fmt.Errorf("invalid IPv6 format")
			}
			server = serverInfo[:endIdx+1]
			if endIdx+1 < len(serverInfo) && serverInfo[endIdx+1] == ':' {
				port = serverInfo[endIdx+2:]
			} else {
				return nil, fmt.Errorf("invalid server:port format")
			}
		} else {
			serverPort := strings.Split(serverInfo, ":")
			if len(serverPort) != 2 {
				return nil, fmt.Errorf("invalid server:port format")
			}
			server = serverPort[0]
			port = serverPort[1]
		}
	} else {
		decoded := utils.UrlSafeBase64Decode(line)

		atIdx := strings.Index(decoded, "@")
		if atIdx == -1 {
			return nil, fmt.Errorf("invalid ss link format")
		}

		userInfo := decoded[:atIdx]
		serverInfo := decoded[atIdx+1:]

		methodPass := strings.SplitN(userInfo, ":", 2)
		if len(methodPass) != 2 {
			return nil, fmt.Errorf("invalid userinfo format")
		}
		method = methodPass[0]
		password = methodPass[1]

		serverPort := strings.Split(serverInfo, ":")
		if len(serverPort) != 2 {
			return nil, fmt.Errorf("invalid server:port format")
		}
		server = serverPort[0]
		port = serverPort[1]
	}

	portNum, err := strconv.Atoi(port)
	if err != nil || portNum == 0 {
		return nil, fmt.Errorf("invalid port: %s", port)
	}

	if remark == "" {
		remark = server + ":" + port
	}

	if group == "" {
		group = "SS"
	}

	pObj := &impl.ShadowsocksProxy{
		BaseProxy: core.BaseProxy{
			Type:   "ss",
			Remark: remark,
			Server: server,
			Port:   portNum,
			Group:  group,
		},
		Password:      password,
		EncryptMethod: method,
		Plugin:        plugin,
		PluginOpts:    utils.ParsePluginOpts(pluginOpts),
	}
	return utils.ToMihomoProxy(pObj)
}
