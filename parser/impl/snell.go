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

type SnellParser struct{}

func (p *SnellParser) Name() string {
	return "Snell"
}

func (p *SnellParser) CanParse(line string) bool {
	return strings.HasPrefix(line, "snell://")
}

func (p *SnellParser) Parse(line string) (core.SubconverterProxy, error) {
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

	return proxy, nil
}
