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

func (p *HttpParser) ParseSingle(line string) (core.ParsableProxy, error) {
	line = strings.TrimSpace(line)
	if isSurgeHttp(line) {
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

	return utils.ToMihomoProxy(proxy)
}

// ParseSurge parses a Surge config string
func (p *HttpParser) ParseSurge(content string) (core.ParsableProxy, error) {
	params := strings.Split(content, ",")
	if len(params) < 3 {
		return nil, fmt.Errorf("invalid surge http config: %s", content)
	}

	proxyType := strings.TrimSpace(params[0])
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
		},
		Tls: proxyType == "https",
	}

	startIdx := 3
	if len(params) >= 5 && !strings.Contains(params[3], "=") {
		http.Username = strings.TrimSpace(params[3])
		http.Password = strings.TrimSpace(params[4])
		startIdx = 5
	}

	for i := startIdx; i < len(params); i++ {
		kv := strings.SplitN(strings.TrimSpace(params[i]), "=", 2)
		if len(kv) == 2 {
			k := strings.TrimSpace(kv[0])
			v := strings.TrimSpace(kv[1])
			switch k {
			case "username":
				http.Username = v
			case "password":
				http.Password = v
			case "skip-cert-verify":
				http.SkipCertVerify = v == "true"
			}
		}
	}

	return utils.ToMihomoProxy(http)
}

// ParseClash parses a Clash config map
func (p *HttpParser) ParseClash(config map[string]interface{}) (core.ParsableProxy, error) {
	server := utils.GetStringField(config, "server")
	port := utils.GetIntField(config, "port")
	name := utils.GetStringField(config, "name")
	username := utils.GetStringField(config, "username")
	password := utils.GetStringField(config, "password")
	tls := utils.GetStringField(config, "tls") == "true" || config["tls"] == true
	skipCertVerify := utils.GetStringField(config, "skip-cert-verify") == "true" || config["skip-cert-verify"] == true

	http := &impl.HttpProxy{
		BaseProxy: core.BaseProxy{
			Type:   "http",
			Server: server,
			Port:   port,
			Remark: name,
		},
		Username:       username,
		Password:       password,
		Tls:            tls,
		SkipCertVerify: skipCertVerify,
	}
	return utils.ToMihomoProxy(http)
}

// ParseSSTap parses a SSTap config map
func (p *HttpParser) ParseSSTap(config map[string]interface{}) (core.ParsableProxy, error) {
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
	return utils.ToMihomoProxy(http)
}

func (p *HttpParser) ParseTelegram(line string) (core.ParsableProxy, error) {
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

	return utils.ToMihomoProxy(proxy)
}
