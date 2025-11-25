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

type Hysteria2Parser struct{}

func (p *Hysteria2Parser) Name() string {
	return "Hysteria2"
}

func (p *Hysteria2Parser) CanParseLine(line string) bool {
	return strings.HasPrefix(line, "hysteria2://") || strings.HasPrefix(line, "hy2://")
}

func (p *Hysteria2Parser) ParseSingle(line string) (core.ParsableProxy, error) {
	u, err := url.Parse(line)
	if err != nil {
		return nil, fmt.Errorf("invalid hysteria2 link: %w", err)
	}

	proxy := &impl.Hysteria2Proxy{}
	proxy.Type = "hysteria2"
	proxy.Server = u.Hostname()
	port := u.Port()
	if port == "" {
		return nil, fmt.Errorf("missing port")
	}
	portNum, err := strconv.Atoi(port)
	if err != nil {
		return nil, fmt.Errorf("invalid port: %w", err)
	}
	proxy.Port = portNum

	if u.User != nil {
		proxy.Password = u.User.Username() // hy2 uses username part as password usually, or just user info
		if p, has := u.User.Password(); has {
			proxy.Password = fmt.Sprintf("%s:%s", proxy.Password, p)
		}
	}

	q := u.Query()
	proxy.Sni = q.Get("sni")
	if q.Get("insecure") == "1" {
		proxy.SkipCertVerify = true
	}
	proxy.Obfs = q.Get("obfs")
	proxy.ObfsPassword = q.Get("obfs-password")

	proxy.Remark = utils.UrlDecode(u.Fragment)
	if proxy.Remark == "" {
		proxy.Remark = proxy.Server
	}
	proxy.Group = core.HYSTERIA2_DEFAULT_GROUP

	return proxy, nil
}
