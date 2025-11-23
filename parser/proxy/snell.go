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

func (p *SnellParser) Parse(line string) (core.SubconverterProxy, error) {
	line = strings.TrimSpace(line)
	if isSurgeSnell(line) {
		return p.ParseSurge(line)
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

	return proxy, nil
}

// ParseSurge parses a Surge format line
func (p *SnellParser) ParseSurge(line string) (core.SubconverterProxy, error) {
	parts := strings.SplitN(line, "=", 2)
	remark := strings.TrimSpace(parts[0])
	params := strings.Split(strings.TrimSpace(parts[1]), ",")

	if len(params) < 3 {
		return nil, fmt.Errorf("invalid surge snell format")
	}

	// params[0] is "snell"
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
			Remark: remark,
			Group:  core.SNELL_DEFAULT_GROUP,
		},
	}

	for i := 3; i < len(params); i++ {
		arg := strings.TrimSpace(params[i])
		if strings.HasPrefix(arg, "psk=") {
			snell.Psk = strings.TrimPrefix(arg, "psk=")
		} else if strings.HasPrefix(arg, "obfs=") {
			snell.Obfs = strings.TrimPrefix(arg, "obfs=")
		} else if strings.HasPrefix(arg, "obfs-host=") {
			snell.ObfsParam = strings.TrimPrefix(arg, "obfs-host=")
		} else if strings.HasPrefix(arg, "version=") {
			v, _ := strconv.Atoi(strings.TrimPrefix(arg, "version="))
			snell.Version = v
		}
	}

	return snell, nil
}

// ParseClash parses a Clash config map
func (p *SnellParser) ParseClash(config map[string]interface{}) (core.SubconverterProxy, error) {
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
	return snell, nil
}
