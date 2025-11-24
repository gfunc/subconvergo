package sub

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/gfunc/subconvergo/parser/core"
	"github.com/gfunc/subconvergo/parser/utils"
	proxyCore "github.com/gfunc/subconvergo/proxy/core"
	"github.com/gfunc/subconvergo/proxy/impl"
)

type SSRSubscriptionParser struct{}

func (p *SSRSubscriptionParser) Name() string {
	return "SSR"
}

func (p *SSRSubscriptionParser) CanParse(content string) bool {
	return strings.Contains(content, "\"serverSubscribes\"") || (strings.Contains(content, "\"local_address\"") && strings.Contains(content, "\"local_port\""))
}

func (p *SSRSubscriptionParser) Parse(content string) (*core.SubContent, error) {
	var js map[string]interface{}
	if err := json.Unmarshal([]byte(content), &js); err != nil {
		return nil, fmt.Errorf("failed to parse SSR JSON: %w", err)
	}

	var proxies []proxyCore.ProxyInterface

	// Single libev config
	if _, ok := js["local_port"]; ok {
		if _, ok := js["local_address"]; ok {
			p := parseSSRNode(js)
			if p != nil {
				proxies = append(proxies, p)
			}
			return &core.SubContent{Proxies: proxies}, nil
		}
	}

	// Configs list
	if configs, ok := js["configs"].([]interface{}); ok {
		for _, item := range configs {
			if cfg, ok := item.(map[string]interface{}); ok {
				p := parseSSRNode(cfg)
				if p != nil {
					proxies = append(proxies, p)
				}
			}
		}
	}

	return &core.SubContent{Proxies: proxies}, nil
}

func parseSSRNode(cfg map[string]interface{}) proxyCore.ProxyInterface {
	server := utils.ToString(cfg["server"])
	portStr := utils.ToString(cfg["server_port"])
	if server == "" || portStr == "0" || portStr == "" {
		return nil
	}
	port, _ := strconv.Atoi(portStr)

	remarks := utils.ToString(cfg["remarks"])
	if remarks == "" {
		remarks = fmt.Sprintf("%s:%s", server, portStr)
	}
	group := utils.ToString(cfg["group"])
	if group == "" {
		group = proxyCore.SSR_DEFAULT_GROUP
	}

	method := utils.ToString(cfg["method"])
	password := utils.ToString(cfg["password"])
	protocol := utils.ToString(cfg["protocol"])
	protocolParam := utils.ToString(cfg["protocolparam"])
	if protocolParam == "" {
		protocolParam = utils.ToString(cfg["protocol_param"])
	}
	obfs := utils.ToString(cfg["obfs"])
	obfsParam := utils.ToString(cfg["obfsparam"])
	if obfsParam == "" {
		obfsParam = utils.ToString(cfg["obfs_param"])
	}

	// Check if it's actually SS (subconverter logic)
	// if(find(ss_ciphers.begin(), ss_ciphers.end(), method) != ss_ciphers.end() && (obfs.empty() || obfs == "plain") && (protocol.empty() || protocol == "origin"))
	// We skip this check for now and assume SSR if protocol/obfs are present, or SS if not?
	// But we are constructing SSRProxy.
	// If it's SS, we should construct SSProxy.
	// Let's just construct SSRProxy for now, as SSR is superset?
	// Or better, check if protocol/obfs are empty/plain/origin.

	isSS := false
	if (obfs == "" || obfs == "plain") && (protocol == "" || protocol == "origin") {
		isSS = true
	}

	if isSS {
		plugin := utils.ToString(cfg["plugin"])
		pluginOptsStr := utils.ToString(cfg["plugin_opts"])
		pluginOpts := make(map[string]interface{})
		if pluginOptsStr != "" {
			pluginOpts = utils.ParsePluginOpts(pluginOptsStr)
		}

		return &impl.ShadowsocksProxy{
			BaseProxy: proxyCore.BaseProxy{
				Type:   "ss",
				Remark: remarks,
				Server: server,
				Port:   port,
				Group:  proxyCore.SS_DEFAULT_GROUP, // Use SS default group if it's SS
			},
			Password:      password,
			EncryptMethod: method,
			Plugin:        plugin,
			PluginOpts:    pluginOpts,
		}
	}

	return &impl.ShadowsocksRProxy{
		BaseProxy: proxyCore.BaseProxy{
			Type:   "ssr",
			Remark: remarks,
			Server: server,
			Port:   port,
			Group:  group,
		},
		Password:      password,
		EncryptMethod: method,
		Protocol:      protocol,
		ProtocolParam: protocolParam,
		Obfs:          obfs,
		ObfsParam:     obfsParam,
	}
}
