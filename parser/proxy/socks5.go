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

type Socks5Parser struct{}

func (p *Socks5Parser) Name() string {
	return "Socks5"
}

func (p *Socks5Parser) CanParseLine(line string) bool {
	return strings.HasPrefix(line, "socks5://") ||
		strings.HasPrefix(line, "tg://socks") ||
		strings.HasPrefix(line, "https://t.me/socks") ||
		isSurgeSocks5(line)
}

func isSurgeSocks5(line string) bool {
	parts := strings.SplitN(line, "=", 2)
	if len(parts) != 2 {
		return false
	}
	val := strings.TrimSpace(parts[1])
	return strings.HasPrefix(val, "socks5,") || strings.HasPrefix(val, "socks5-tls,")
}

func (p *Socks5Parser) ParseSingle(line string) (core.ParsableProxy, error) {
	line = strings.TrimSpace(line)
	if isSurgeSocks5(line) {
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

	if strings.HasPrefix(line, "tg://socks") || strings.HasPrefix(line, "https://t.me/socks") {
		return p.ParseTelegram(line)
	}

	if !strings.HasPrefix(line, "socks5://") {
		return nil, fmt.Errorf("not a valid socks5:// link")
	}

	u, err := url.Parse(line)
	if err != nil {
		return nil, fmt.Errorf("invalid socks5 link: %w", err)
	}

	proxy := &impl.Socks5Proxy{}
	proxy.Type = "socks5"
	proxy.Server = u.Hostname()
	port := u.Port()
	if port != "" {
		p, _ := strconv.Atoi(port)
		proxy.Port = p
	} else {
		proxy.Port = 1080
	}

	if u.User != nil {
		proxy.Username = u.User.Username()
		proxy.Password, _ = u.User.Password()
	}

	proxy.Remark = utils.UrlDecode(u.Fragment)
	if proxy.Remark == "" {
		proxy.Remark = proxy.Server
	}
	proxy.Group = core.SOCKS_DEFAULT_GROUP

	return utils.ToMihomoProxy(proxy)
}

// ParseSurge parses a Surge config string
func (p *Socks5Parser) ParseSurge(content string) (core.ParsableProxy, error) {
	params := strings.Split(content, ",")
	if len(params) < 3 {
		return nil, fmt.Errorf("invalid surge socks5 config: %s", content)
	}

	proxyType := strings.TrimSpace(params[0])
	server := strings.TrimSpace(params[1])
	portStr := strings.TrimSpace(params[2])
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid port: %s", portStr)
	}

	socks := &impl.Socks5Proxy{
		BaseProxy: core.BaseProxy{
			Type:   "socks5",
			Server: server,
			Port:   port,
		},
	}

	if proxyType == "socks5-tls" {
		socks.TLS = true
	}

	startIdx := 3
	if len(params) >= 5 && !strings.Contains(params[3], "=") {
		socks.Username = strings.TrimSpace(params[3])
		socks.Password = strings.TrimSpace(params[4])
		startIdx = 5
	}

	for i := startIdx; i < len(params); i++ {
		kv := strings.SplitN(strings.TrimSpace(params[i]), "=", 2)
		if len(kv) == 2 {
			k := strings.TrimSpace(kv[0])
			v := strings.TrimSpace(kv[1])
			switch k {
			case "username":
				socks.Username = v
			case "password":
				socks.Password = v
			case "tls":
				socks.TLS = v == "true"
			}
		}
	}

	return utils.ToMihomoProxy(socks)
}

// ParseClash parses a Clash config map
func (p *Socks5Parser) ParseClash(config map[string]interface{}) (core.ParsableProxy, error) {
	server := utils.GetStringField(config, "server")
	port := utils.GetIntField(config, "port")
	name := utils.GetStringField(config, "name")
	username := utils.GetStringField(config, "username")
	password := utils.GetStringField(config, "password")
	tls := utils.GetStringField(config, "tls") == "true" || config["tls"] == true

	socks := &impl.Socks5Proxy{
		BaseProxy: core.BaseProxy{
			Type:   "socks5",
			Server: server,
			Port:   port,
			Remark: name,
		},
		Username: username,
		Password: password,
		TLS:      tls,
	}
	return utils.ToMihomoProxy(socks)
}

// ParseNetch parses a Netch config map
func (p *Socks5Parser) ParseNetch(config map[string]interface{}) (core.ParsableProxy, error) {
	remark := utils.GetStringField(config, "Remark")
	hostname := utils.GetStringField(config, "Hostname")
	port := utils.GetIntField(config, "Port")
	username := utils.GetStringField(config, "Username")
	password := utils.GetStringField(config, "Password")

	socks := &impl.Socks5Proxy{
		BaseProxy: core.BaseProxy{
			Type:   "socks5",
			Server: hostname,
			Port:   port,
			Remark: remark,
		},
		Username: username,
		Password: password,
	}
	return utils.ToMihomoProxy(socks)
}

// ParseSSTap parses a SSTap config map
func (p *Socks5Parser) ParseSSTap(config map[string]interface{}) (core.ParsableProxy, error) {
	name := utils.GetStringField(config, "name")
	server := utils.GetStringField(config, "server")
	port := utils.GetIntField(config, "port")
	username := utils.GetStringField(config, "username")
	password := utils.GetStringField(config, "password")

	socks := &impl.Socks5Proxy{
		BaseProxy: core.BaseProxy{
			Type:   "socks5",
			Server: server,
			Port:   port,
			Remark: name,
		},
		Username: username,
		Password: password,
	}
	return utils.ToMihomoProxy(socks)
}

func (p *Socks5Parser) ParseTelegram(line string) (core.ParsableProxy, error) {
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

	proxy := &impl.Socks5Proxy{}
	proxy.Type = "socks5"
	proxy.Server = server
	proxy.Port = port
	proxy.Username = user
	proxy.Password = pass
	proxy.Remark = remark

	return utils.ToMihomoProxy(proxy)
}
