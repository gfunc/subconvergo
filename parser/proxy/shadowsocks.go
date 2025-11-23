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

func (p *ShadowsocksParser) Parse(line string) (core.SubconverterProxy, error) {
	line = strings.TrimSpace(line)
	if isSurgeSS(line) {
		return p.ParseSurge(line)
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

// ParseSurge parses a Surge format line
func (p *ShadowsocksParser) ParseSurge(line string) (core.SubconverterProxy, error) {
	parts := strings.SplitN(line, "=", 2)
	remark := strings.TrimSpace(parts[0])
	params := strings.Split(strings.TrimSpace(parts[1]), ",")

	if len(params) < 5 {
		return nil, fmt.Errorf("invalid surge ss format")
	}

	// params[0] is "ss"
	server := strings.TrimSpace(params[1])
	portStr := strings.TrimSpace(params[2])
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid port: %s", portStr)
	}
	encrypt := strings.TrimSpace(params[3])
	password := strings.TrimSpace(params[4])

	ss := &impl.ShadowsocksProxy{
		BaseProxy: core.BaseProxy{
			Type:   "ss",
			Server: server,
			Port:   port,
			Remark: remark,
		},
		Password:      password,
		EncryptMethod: encrypt,
	}

	// Parse optional args
	for i := 5; i < len(params); i++ {
		arg := strings.TrimSpace(params[i])
		if strings.HasPrefix(arg, "obfs=") {
			ss.Plugin = "obfs" // Surge uses "obfs" param, but in SS it maps to simple-obfs or similar?
			// subconverter maps obfs=http to simple-obfs?
			// Let's check subconverter logic if needed.
			// For now, just store it.
			if ss.PluginOpts == nil {
				ss.PluginOpts = make(map[string]interface{})
			}
			ss.PluginOpts["obfs"] = strings.TrimPrefix(arg, "obfs=")
		} else if strings.HasPrefix(arg, "obfs-host=") {
			if ss.PluginOpts == nil {
				ss.PluginOpts = make(map[string]interface{})
			}
			ss.PluginOpts["obfs-host"] = strings.TrimPrefix(arg, "obfs-host=")
		}
		// Handle tfo, udp-relay, etc.
	}

	return ss, nil
}

// ParseClash parses a Clash config map
func (p *ShadowsocksParser) ParseClash(config map[string]interface{}) (core.SubconverterProxy, error) {
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

	return ss, nil
}

// ParseNetch parses a Netch config map
func (p *ShadowsocksParser) ParseNetch(config map[string]interface{}) (core.SubconverterProxy, error) {
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
		// Netch plugin opts parsing?
		// Assuming simple string for now or implement parsing if needed
	}

	return ss, nil
}

// ParseSSTap parses a SSTap config map
func (p *ShadowsocksParser) ParseSSTap(config map[string]interface{}) (core.SubconverterProxy, error) {
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
		// Parse plugin opts
	}

	return ss, nil
}

// ParseSSD parses a SSD config map (resolved)
func (p *ShadowsocksParser) ParseSSD(config map[string]interface{}) (core.SubconverterProxy, error) {
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

	return ss, nil
}

func (p *ShadowsocksParser) ParseSSAndroid(config map[string]interface{}) (*impl.ShadowsocksProxy, error) {
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

	return ss, nil
}