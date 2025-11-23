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

type HttpParser struct{}

func (p *HttpParser) Name() string {
	return "HTTP"
}

func (p *HttpParser) CanParse(line string) bool {
	return strings.HasPrefix(line, "http://") || strings.HasPrefix(line, "https://")
}

func (p *HttpParser) Parse(line string) (core.SubconverterProxy, error) {
	u, err := url.Parse(line)
	if err != nil {
		return nil, fmt.Errorf("invalid http/https link: %w", err)
	}

	proxy := &impl.HttpProxy{}
	proxy.Type = "http"
	proxy.Tls = u.Scheme == "https"
	proxy.Server = u.Hostname()
	port := u.Port()
	if port == "" {
		if proxy.Tls {
			proxy.Port = 443
		} else {
			proxy.Port = 80
		}
	} else {
		p, _ := strconv.Atoi(port)
		proxy.Port = p
	}

	if u.User != nil {
		proxy.Username = u.User.Username()
		proxy.Password, _ = u.User.Password()
	}

	proxy.Remark = utils.UrlDecode(u.Fragment)
	if proxy.Remark == "" {
		proxy.Remark = proxy.Server
	}

	return proxy, nil
}
