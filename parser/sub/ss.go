package sub

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gfunc/subconvergo/parser/core"
	"github.com/gfunc/subconvergo/parser/utils"
	pc "github.com/gfunc/subconvergo/proxy/core"
	"github.com/gfunc/subconvergo/proxy/impl"
)

type SSSubscriptionParser struct{}

func (p *SSSubscriptionParser) Name() string {
	return "SS"
}

func (p *SSSubscriptionParser) CanParse(content string) bool {
	return strings.Contains(content, "\"version\"")
}

func (p *SSSubscriptionParser) Parse(content string) (*core.SubContent, error) {
	var js map[string]interface{}
	if err := json.Unmarshal([]byte(content), &js); err != nil {
		return nil, fmt.Errorf("failed to parse SS JSON: %w", err)
	}

	section := "configs"
	if _, ok := js["servers"]; ok {
		if _, ok := js["version"]; ok {
			section = "servers"
		}
	}
	// subconverter logic: const char *section = json.HasMember("version") && json.HasMember("servers") ? "servers" : "configs";
	// But if it has "servers" but no "version", it defaults to "configs"?
	// Wait, if it has "servers" but no "version", subconverter uses "configs".
	// But if "configs" doesn't exist, it returns.
	// So if it has "servers" only, it might fail in subconverter unless "configs" also exists?
	// Let's follow subconverter logic exactly.
	// But wait, if I have a file with ONLY "servers", subconverter might fail?
	// Let's be more flexible.
	if _, ok := js["servers"]; ok {
		section = "servers"
	}
	if _, ok := js["configs"]; ok {
		section = "configs"
	}
	// subconverter prefers "servers" if "version" is present.
	if _, hasVersion := js["version"]; hasVersion {
		if _, hasServers := js["servers"]; hasServers {
			section = "servers"
		}
	}

	val, ok := js[section]
	if !ok {
		return nil, fmt.Errorf("no configs or servers found")
	}

	// It should be a list
	list, ok := val.([]interface{})
	if !ok {
		return nil, fmt.Errorf("section %s is not a list", section)
	}

	var proxies []pc.ProxyInterface
	group := toString(js["remarks"])
	if group == "" {
		group = "SS-Provider"
	}

	for _, item := range list {
		cfg, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		cfg["group"] = group
		
		if ss, err := p.parseSSConf(cfg); err == nil {
			proxies = append(proxies, ss)
		}
	}

	return &core.SubContent{
		Proxies: proxies,
	}, nil
}


func (p *SSSubscriptionParser) parseSSConf(config map[string]interface{}) (*impl.ShadowsocksProxy, error) {
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
		BaseProxy: pc.BaseProxy{
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
		// Parse plugin opts
	}

	return ss, nil
}