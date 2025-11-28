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

type SnellParser struct{}

func (p *SnellParser) Name() string {
	return "Snell"
}

func (p *SnellParser) CanParseLine(line string) bool {
	return strings.HasPrefix(line, "snell://") || isSurgeSnell(line)
}

func isSurgeSnell(line string) bool {
	parts := strings.SplitN(line, "=", 2)
	if len(parts) != 2 {
		return false
	}
	val := strings.TrimSpace(parts[1])
	return strings.HasPrefix(val, "snell,")
}

func (p *SnellParser) ParseSingle(line string) (core.ParsableProxy, error) {
	line = strings.TrimSpace(line)
	if isSurgeSnell(line) {
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

	u, err := url.Parse(line)
	if err != nil {
		return nil, fmt.Errorf("invalid snell link: %w", err)
	}

	proxy := &impl.SnellProxy{}
	proxy.Type = "snell"
	proxy.Server = u.Hostname()
	port := u.Port()
	if port != "" {
		p, _ := strconv.Atoi(port)
		proxy.Port = p
	}

	q := u.Query()
	proxy.Psk = q.Get("psk")
	proxy.Obfs = q.Get("obfs")
	proxy.ObfsParam = q.Get("obfs-host")
	if v := q.Get("version"); v != "" {
		proxy.Version, _ = strconv.Atoi(v)
	}

	proxy.Remark = utils.UrlDecode(u.Fragment)
	if proxy.Remark == "" {
		proxy.Remark = proxy.Server
	}
	proxy.Group = core.SNELL_DEFAULT_GROUP

	return utils.ToMihomoProxy(proxy)
}

// ParseSurge parses a Surge config string
func (p *SnellParser) ParseSurge(content string) (core.ParsableProxy, error) {
	params := strings.Split(content, ",")
	if len(params) < 3 {
		return nil, fmt.Errorf("invalid surge snell config: %s", content)
	}

	server := strings.TrimSpace(params[1])
	portStr := strings.TrimSpace(params[2])
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid port: %s", portStr)
	}

	snell := &impl.SnellProxy{
		BaseProxy: core.BaseProxy{
			Type:   "snell",
			Server: server,
			Port:   port,
			Group:  core.SNELL_DEFAULT_GROUP,
		},
	}

	for i := 3; i < len(params); i++ {
		kv := strings.SplitN(strings.TrimSpace(params[i]), "=", 2)
		if len(kv) == 2 {
			k := strings.TrimSpace(kv[0])
			v := strings.TrimSpace(kv[1])
			switch k {
			case "psk":
				snell.Psk = v
			case "obfs":
				snell.Obfs = v
			case "obfs-host":
				snell.ObfsParam = v
			case "version":
				if ver, err := strconv.Atoi(v); err == nil {
					snell.Version = ver
				}
			}
		}
	}

	return utils.ToMihomoProxy(snell)
}

// ParseClash parses a Clash config map
func (p *SnellParser) ParseClash(config map[string]interface{}) (core.ParsableProxy, error) {
	server := utils.GetStringField(config, "server")
	port := utils.GetIntField(config, "port")
	name := utils.GetStringField(config, "name")
	psk := utils.GetStringField(config, "psk")
	version := utils.GetIntField(config, "version")
	obfsOpts, _ := config["obfs-opts"].(map[string]interface{})
	obfs := utils.GetStringField(obfsOpts, "mode")
	obfsHost := utils.GetStringField(obfsOpts, "host")

	snell := &impl.SnellProxy{
		BaseProxy: core.BaseProxy{
			Type:   "snell",
			Server: server,
			Port:   port,
			Remark: name,
		},
		Psk:       psk,
		Version:   version,
		Obfs:      obfs,
		ObfsParam: obfsHost,
	}
	return utils.ToMihomoProxy(snell)
}
