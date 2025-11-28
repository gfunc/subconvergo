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

type ShadowsocksParser struct{}

func (p *ShadowsocksParser) Name() string {
	return "Shadowsocks"
}

func (p *ShadowsocksParser) CanParseLine(line string) bool {
	return strings.HasPrefix(line, "ss://") || isSurgeSS(line)
}

func isSurgeSS(line string) bool {
	parts := strings.SplitN(line, "=", 2)
	if len(parts) != 2 {
		return false
	}
	val := strings.TrimSpace(parts[1])
	return strings.HasPrefix(val, "ss,")
}

func (p *ShadowsocksParser) ParseSingle(line string) (core.ParsableProxy, error) {
	line = strings.TrimSpace(line)
	if isSurgeSS(line) {
		parts := strings.SplitN(line, "=", 2)
		remark := strings.TrimSpace(parts[0])
		content := strings.TrimSpace(parts[1])

		proxy, err := p.ParseSurge(content)
		if err != nil {
			return nil, err
		}
		proxy.SetRemark(remark)
		return proxy, nil
	}

	if !strings.HasPrefix(line, "ss://") {
		return nil, fmt.Errorf("not a valid ss:// link")
	}

	line = line[5:]

	var remark, password, method, server, port, plugin, pluginOpts string
	group := core.SS_DEFAULT_GROUP

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

// ParseSurge parses a Surge config string
func (p *ShadowsocksParser) ParseSurge(content string) (core.ParsableProxy, error) {
	params := strings.Split(content, ",")
	if len(params) < 3 {
		return nil, fmt.Errorf("invalid surge ss config: %s", content)
	}

	proxyType := strings.TrimSpace(params[0])
	server := strings.TrimSpace(params[1])
	portStr := strings.TrimSpace(params[2])
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid port: %s", portStr)
	}

	ss := &impl.ShadowsocksProxy{
		BaseProxy: core.BaseProxy{
			Type:   "ss",
			Server: server,
			Port:   port,
		},
	}

	startIdx := 3
	// Check for positional method/password
	// Surge 2 "custom" has method, password at 3, 4
	if strings.ToLower(proxyType) == "custom" {
		if len(params) >= 5 {
			ss.EncryptMethod = strings.TrimSpace(params[3])
			ss.Password = strings.TrimSpace(params[4])
			startIdx = 5
		}
	}

	for i := startIdx; i < len(params); i++ {
		kv := strings.SplitN(strings.TrimSpace(params[i]), "=", 2)
		if len(kv) == 2 {
			k := strings.TrimSpace(kv[0])
			v := strings.TrimSpace(kv[1])
			switch k {
			case "encrypt-method":
				ss.EncryptMethod = v
			case "password":
				ss.Password = v
			case "obfs":
				ss.Plugin = "obfs"
				if ss.PluginOpts == nil {
					ss.PluginOpts = make(map[string]interface{})
				}
				ss.PluginOpts["obfs"] = v
			case "obfs-host":
				if ss.PluginOpts == nil {
					ss.PluginOpts = make(map[string]interface{})
				}
				ss.PluginOpts["obfs-host"] = v
			}
		}
	}

	return utils.ToMihomoProxy(ss)
}

// ParseClash parses a Clash config map
func (p *ShadowsocksParser) ParseClash(config map[string]interface{}) (core.ParsableProxy, error) {
	server := utils.GetStringField(config, "server")
	port := utils.GetIntField(config, "port")
	cipher := utils.GetStringField(config, "cipher")
	password := utils.GetStringField(config, "password")
	name := utils.GetStringField(config, "name")
	plugin := utils.GetStringField(config, "plugin")

	ss := &impl.ShadowsocksProxy{
		BaseProxy: core.BaseProxy{
			Type:   "ss",
			Server: server,
			Port:   port,
			Remark: name,
		},
		Password:      password,
		EncryptMethod: cipher,
		Plugin:        plugin,
	}

	if opts, ok := config["plugin-opts"].(map[string]interface{}); ok {
		ss.PluginOpts = opts
	} else if opts, ok := config["plugin-opts"].(map[interface{}]interface{}); ok {
		// Handle map[interface{}]interface{} which might come from yaml unmarshal
		ss.PluginOpts = make(map[string]interface{})
		for k, v := range opts {
			if ks, ok := k.(string); ok {
				ss.PluginOpts[ks] = v
			}
		}
	}

	return utils.ToMihomoProxy(ss)
}

// ParseNetch parses a Netch config map
func (p *ShadowsocksParser) ParseNetch(config map[string]interface{}) (core.ParsableProxy, error) {
	remark := utils.GetStringField(config, "Remark")
	hostname := utils.GetStringField(config, "Hostname")
	port := utils.GetIntField(config, "Port")
	password := utils.GetStringField(config, "Password")
	method := utils.GetStringField(config, "EncryptMethod")
	plugin := utils.GetStringField(config, "Plugin")
	pluginOptsStr := utils.GetStringField(config, "PluginOpts")

	ss := &impl.ShadowsocksProxy{
		BaseProxy: core.BaseProxy{
			Type:   "ss",
			Server: hostname,
			Port:   port,
			Remark: remark,
		},
		Password:      password,
		EncryptMethod: method,
		Plugin:        plugin,
	}

	if pluginOptsStr != "" {
		ss.PluginOpts = utils.ParsePluginOpts(pluginOptsStr)
	}

	return utils.ToMihomoProxy(ss)
}

// ParseSSTap parses a SSTap config map
func (p *ShadowsocksParser) ParseSSTap(config map[string]interface{}) (core.ParsableProxy, error) {
	name := utils.GetStringField(config, "name")
	server := utils.GetStringField(config, "server")
	port := utils.GetIntField(config, "port")
	password := utils.GetStringField(config, "password")
	method := utils.GetStringField(config, "method")
	plugin := utils.GetStringField(config, "plugin")
	pluginOptsStr := utils.GetStringField(config, "plugin_opts")

	ss := &impl.ShadowsocksProxy{
		BaseProxy: core.BaseProxy{
			Type:   "ss",
			Server: server,
			Port:   port,
			Remark: name,
		},
		Password:      password,
		EncryptMethod: method,
		Plugin:        plugin,
	}

	if pluginOptsStr != "" {
		ss.PluginOpts = utils.ParsePluginOpts(pluginOptsStr)
	}

	return utils.ToMihomoProxy(ss)
}

// ParseSSD parses a SSD config map (resolved)
func (p *ShadowsocksParser) ParseSSD(config map[string]interface{}) (core.ParsableProxy, error) {
	// Expects keys: server, port, encryption, password, plugin, plugin_options, remarks, airport
	server := utils.GetStringField(config, "server")
	port := utils.GetIntField(config, "port")
	encrypt := utils.GetStringField(config, "encryption")
	pass := utils.GetStringField(config, "password")
	plugin := utils.GetStringField(config, "plugin")
	pluginO := utils.GetStringField(config, "plugin_options")
	remarks := utils.GetStringField(config, "remarks")
	group := utils.GetStringField(config, "airport")

	ss := &impl.ShadowsocksProxy{
		BaseProxy: core.BaseProxy{
			Type:   "ss",
			Server: server,
			Port:   port,
			Remark: remarks,
			Group:  group,
		},
		Password:      pass,
		EncryptMethod: encrypt,
		Plugin:        plugin,
	}

	if pluginO != "" {
		ss.PluginOpts = make(map[string]interface{})
		parts := strings.Split(pluginO, ";")
		for _, part := range parts {
			kv := strings.SplitN(part, "=", 2)
			if len(kv) == 2 {
				ss.PluginOpts[kv[0]] = kv[1]
			}
		}
	}

	return utils.ToMihomoProxy(ss)
}

func (p *ShadowsocksParser) ParseSSAndroid(config map[string]interface{}) (core.ParsableProxy, error) {
	server := utils.GetStringField(config, "server")
	port := utils.GetIntField(config, "server_port")
	remarks := utils.GetStringField(config, "remarks")
	password := utils.GetStringField(config, "password")
	method := utils.GetStringField(config, "method")
	plugin := utils.GetStringField(config, "plugin")
	pluginOpts := utils.GetStringField(config, "plugin_opts")
	group := utils.GetStringField(config, "group")

	if remarks == "" {
		remarks = fmt.Sprintf("%s:%d", server, port)
	}

	ss := &impl.ShadowsocksProxy{
		BaseProxy: core.BaseProxy{
			Type:   "ss",
			Server: server,
			Port:   port,
			Remark: remarks,
			Group:  group,
		},
		Password:      password,
		EncryptMethod: method,
		Plugin:        plugin,
	}

	if pluginOpts != "" {
		ss.PluginOpts = utils.ParsePluginOpts(pluginOpts)
	}

	return utils.ToMihomoProxy(ss)
}

func (p *ShadowsocksParser) ParseSS(cfg map[string]interface{}) (core.ParsableProxy, error) {
	server := utils.GetStringField(cfg, "server")
	port := utils.GetIntField(cfg, "server_port")
	if port == 0 {
		return nil, fmt.Errorf("invalid port")
	}

	remark := utils.GetStringField(cfg, "remarks")
	if remark == "" {
		remark = fmt.Sprintf("%s:%d", server, port)
	}

	password := utils.GetStringField(cfg, "password")
	method := utils.GetStringField(cfg, "method")
	plugin := utils.GetStringField(cfg, "plugin")
	pluginOpts := utils.GetStringField(cfg, "plugin_opts")
	group := utils.GetStringField(cfg, "group")
	if group == "" {
		group = core.SS_DEFAULT_GROUP
	}

	ss := &impl.ShadowsocksProxy{
		BaseProxy: core.BaseProxy{
			Type:   "ss",
			Remark: remark,
			Server: server,
			Port:   port,
			Group:  group,
		},
		Password:      password,
		EncryptMethod: method,
		Plugin:        plugin,
	}

	if pluginOpts != "" {
		ss.PluginOpts = utils.ParsePluginOpts(pluginOpts)
	}

	return utils.ToMihomoProxy(ss)
}
