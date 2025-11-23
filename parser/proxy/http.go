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

type HttpParser struct{}

func (p *HttpParser) Name() string {
	return "HTTP"
}

func (p *HttpParser) CanParseLine(line string) bool {
	return strings.HasPrefix(line, "http://") || strings.HasPrefix(line, "https://") ||
		strings.HasPrefix(line, "tg://http") || strings.HasPrefix(line, "https://t.me/http") ||
		isSurgeHttp(line)
}

func isSurgeHttp(line string) bool {
	parts := strings.SplitN(line, "=", 2)
	if len(parts) != 2 {
		return false
	}
	val := strings.TrimSpace(parts[1])
	return strings.HasPrefix(val, "http,") || strings.HasPrefix(val, "https,")
}

func (p *HttpParser) Parse(line string) (core.SubconverterProxy, error) {
	line = strings.TrimSpace(line)
	if isSurgeHttp(line) {
		return p.ParseSurge(line)
	}

	if strings.HasPrefix(line, "tg://http") || strings.HasPrefix(line, "https://t.me/http") {
		return p.ParseTelegram(line)
	}

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
	proxy.Group = core.HTTP_DEFAULT_GROUP

	return proxy, nil
}

// ParseSurge parses a Surge format line
func (p *HttpParser) ParseSurge(line string) (core.SubconverterProxy, error) {
	parts := strings.SplitN(line, "=", 2)
	remark := strings.TrimSpace(parts[0])
	params := strings.Split(strings.TrimSpace(parts[1]), ",")

	if len(params) < 3 {
		return nil, fmt.Errorf("invalid surge http format")
	}

	// params[0] is "http" or "https"
	typeStr := strings.TrimSpace(params[0])
	server := strings.TrimSpace(params[1])
	portStr := strings.TrimSpace(params[2])
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid port: %s", portStr)
	}

	http := &impl.HttpProxy{
		BaseProxy: core.BaseProxy{
			Type:   "http",
			Server: server,
			Port:   port,
			Remark: remark,
		},
		Tls: typeStr == "https",
	}

	for i := 3; i < len(params); i++ {
		arg := strings.TrimSpace(params[i])
		if strings.HasPrefix(arg, "username=") {
			http.Username = strings.TrimPrefix(arg, "username=")
		} else if strings.HasPrefix(arg, "password=") {
			http.Password = strings.TrimPrefix(arg, "password=")
		} else if strings.HasPrefix(arg, "tls=") {
			http.Tls = strings.TrimPrefix(arg, "tls=") == "true"
		}
	}

	return http, nil
}

// ParseClash parses a Clash config map
func (p *HttpParser) ParseClash(config map[string]interface{}) (core.SubconverterProxy, error) {
	server := utils.GetStringField(config, "server")
	port := utils.GetIntField(config, "port")
	name := utils.GetStringField(config, "name")
	username := utils.GetStringField(config, "username")
	password := utils.GetStringField(config, "password")
	tls := utils.GetStringField(config, "tls") == "true" || config["tls"] == true

	http := &impl.HttpProxy{
		BaseProxy: core.BaseProxy{
			Type:   "http",
			Server: server,
			Port:   port,
			Remark: name,
		},
		Username: username,
		Password: password,
		Tls:      tls,
	}
	return http, nil
}

// ParseSSTap parses a SSTap config map
func (p *HttpParser) ParseSSTap(config map[string]interface{}) (core.SubconverterProxy, error) {
	name := utils.GetStringField(config, "name")
	server := utils.GetStringField(config, "server")
	port := utils.GetIntField(config, "port")
	username := utils.GetStringField(config, "username")
	password := utils.GetStringField(config, "password")

	http := &impl.HttpProxy{
		BaseProxy: core.BaseProxy{
			Type:   "http",
			Server: server,
			Port:   port,
			Remark: name,
		},
		Username: username,
		Password: password,
	}
	return http, nil
}

func (p *HttpParser) ParseTelegram(line string) (core.SubconverterProxy, error) {
	u, err := url.Parse(line)
	if err != nil {
		return nil, fmt.Errorf("invalid telegram link: %w", err)
	}
	q := u.Query()

	server := q.Get("server")
	portStr := q.Get("port")
	user := q.Get("user")
	pass := q.Get("pass")
	remark := q.Get("remarks")
	if remark == "" {
		remark = q.Get("remark")
	}

	port, _ := strconv.Atoi(portStr)

	proxy := &impl.HttpProxy{}
	proxy.Type = "http"
	proxy.Server = server
	proxy.Port = port
	proxy.Username = user
	proxy.Password = pass
	proxy.Remark = remark

	return proxy, nil
}
