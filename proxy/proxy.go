package proxy

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

type ProxyInterface interface {
	GetType() string
	GetRemark() string
	GetServer() string
	GetPort() int
	GetGroup() string
	SetRemark(remark string)
	SetGroup(group string)
}

type SubconverterProxy interface {
	ProxyInterface
	GenerateLink() (string, error)
	ProxyOptions() map[string]interface{}
}

// BaseProxy contains common fields shared by all proxy types
type BaseProxy struct {
	Type   string `yaml:"type" json:"type"`
	Remark string `yaml:"remark" json:"remark"`
	Server string `yaml:"server" json:"server"`
	Port   int    `yaml:"port" json:"port"`
	Group  string `yaml:"group" json:"group"`
}

func (p *BaseProxy) GetType() string {
	return p.Type
}

func (p *BaseProxy) GetRemark() string {
	return p.Remark
}

func (p *BaseProxy) SetRemark(remark string) {
	p.Remark = remark
}

func (p *BaseProxy) GetServer() string {
	return p.Server
}

func (p *BaseProxy) GetPort() int {
	return p.Port
}

func (p *BaseProxy) GetGroup() string {
	return p.Group
}

func (p *BaseProxy) SetGroup(group string) {
	p.Group = group
}

// ShadowsocksProxy represents a Shadowsocks proxy
type ShadowsocksProxy struct {
	BaseProxy     `yaml:",inline"`
	Password      string `yaml:"password" json:"password"`
	EncryptMethod string `yaml:"encrypt_method" json:"encrypt_method"`
	Plugin        string `yaml:"plugin" json:"plugin"`
	PluginOpts    string `yaml:"plugin_opts" json:"plugin_opts"`
}

func (p *ShadowsocksProxy) GenerateLink() (string, error) {
	// Format: ss://base64(method:password)@server:port#remark
	userInfo := fmt.Sprintf("%s:%s", p.EncryptMethod, p.Password)
	encoded := base64.URLEncoding.EncodeToString([]byte(userInfo))
	link := fmt.Sprintf("ss://%s@%s:%d#%s", encoded, p.Server, p.Port, urlEncode(p.Remark))
	return link, nil
}

func (p *ShadowsocksProxy) ProxyOptions() map[string]interface{} {
	options := map[string]interface{}{
		"type":     "ss",
		"name":     p.Remark,
		"server":   p.Server,
		"port":     p.Port,
		"cipher":   p.EncryptMethod,
		"password": p.Password,
	}

	if p.Plugin != "" {
		options["plugin"] = p.Plugin
		if p.PluginOpts != "" {
			options["plugin-opts"] = parsePluginOpts(p.PluginOpts)
		}
	}

	return options
}

// ShadowsocksRProxy represents a ShadowsocksR proxy
type ShadowsocksRProxy struct {
	BaseProxy     `yaml:",inline"`
	Password      string `yaml:"password" json:"password"`
	EncryptMethod string `yaml:"encrypt_method" json:"encrypt_method"`
	Protocol      string `yaml:"protocol" json:"protocol"`
	ProtocolParam string `yaml:"protocol_param" json:"protocol_param"`
	Obfs          string `yaml:"obfs" json:"obfs"`
	ObfsParam     string `yaml:"obfs_param" json:"obfs_param"`
}

func (p *ShadowsocksRProxy) GenerateLink() (string, error) {
	// Format: ssr://base64(server:port:protocol:method:obfs:base64(password)/?...)
	passEncoded := base64.URLEncoding.EncodeToString([]byte(p.Password))
	mainPart := fmt.Sprintf("%s:%d:%s:%s:%s:%s",
		p.Server,
		p.Port,
		p.Protocol,
		p.EncryptMethod,
		p.Obfs,
		passEncoded,
	)

	// Add query parameters
	params := []string{}
	if p.ObfsParam != "" {
		params = append(params, fmt.Sprintf("obfsparam=%s", base64.URLEncoding.EncodeToString([]byte(p.ObfsParam))))
	}
	if p.ProtocolParam != "" {
		params = append(params, fmt.Sprintf("protoparam=%s", base64.URLEncoding.EncodeToString([]byte(p.ProtocolParam))))
	}
	if p.Remark != "" {
		params = append(params, fmt.Sprintf("remarks=%s", base64.URLEncoding.EncodeToString([]byte(p.Remark))))
	}

	if len(params) > 0 {
		mainPart += "/?" + strings.Join(params, "&")
	}

	encoded := base64.URLEncoding.EncodeToString([]byte(mainPart))
	return "ssr://" + encoded, nil
}

func (p *ShadowsocksRProxy) ProxyOptions() (options map[string]interface{}) {
	if p.Type == "ss" {
		options = map[string]interface{}{
			"type":     "ss",
			"name":     p.Remark,
			"server":   p.Server,
			"port":     p.Port,
			"cipher":   p.EncryptMethod,
			"password": p.Password,
		}
	} else {
		options = map[string]interface{}{
			"type":           "ssr",
			"name":           p.Remark,
			"server":         p.Server,
			"port":           p.Port,
			"cipher":         p.EncryptMethod,
			"password":       p.Password,
			"protocol":       p.Protocol,
			"obfs":           p.Obfs,
			"protocol-param": p.ProtocolParam,
			"obfs-param":     p.ObfsParam,
		}
	}
	return options
}

// VMessProxy represents a VMess proxy
type VMessProxy struct {
	BaseProxy `yaml:",inline"`
	UUID      string `yaml:"uuid" json:"uuid"`
	AlterID   int    `yaml:"alter_id" json:"alter_id"`
	Network   string `yaml:"network" json:"network"`
	Path      string `yaml:"path" json:"path"`
	Host      string `yaml:"host" json:"host"`
	TLS       bool   `yaml:"tls" json:"tls"`
	SNI       string `yaml:"sni" json:"sni"`
}

func (p *VMessProxy) GenerateLink() (string, error) {
	// VMess JSON format
	vmessData := map[string]interface{}{
		"v":    "2",
		"ps":   p.Remark,
		"add":  p.Server,
		"port": fmt.Sprintf("%d", p.Port),
		"id":   p.UUID,
		"aid":  fmt.Sprintf("%d", p.AlterID),
		"net":  p.Network,
		"type": "none",
		"host": p.Host,
		"path": p.Path,
		"tls":  "",
	}

	if p.TLS {
		vmessData["tls"] = "tls"
	}

	jsonBytes, err := json.Marshal(vmessData)
	if err != nil {
		return "", err
	}

	encoded := base64.StdEncoding.EncodeToString(jsonBytes)
	return "vmess://" + encoded, nil
}

func (p *VMessProxy) ProxyOptions() map[string]interface{} {
	options := map[string]interface{}{
		"type":    "vmess",
		"name":    p.Remark,
		"server":  p.Server,
		"port":    p.Port,
		"uuid":    p.UUID,
		"alterId": p.AlterID,
		"cipher":  "auto",
		"network": p.Network,
	}

	if p.TLS {
		options["tls"] = true
		if p.SNI != "" {
			options["servername"] = p.SNI
		}
	}
	switch p.Network {
	case "ws", "httpupgrade":
		wsOpts := make(map[string]interface{})
		if p.Path == "" {
			p.Path = "/"
		}
		wsOpts["path"] = p.Path
		if p.Host != "" {
			headers := make(map[string]string)
			headers["Host"] = p.Host
			wsOpts["headers"] = headers
		}
		options["ws-opts"] = wsOpts

	case "http", "h2":
		h2Opts := make(map[string]interface{})
		if p.Path != "" {
			h2Opts["path"] = p.Path
		}
		if p.Host != "" {
			h2Opts["host"] = []string{p.Host}
		}
		options["h2-opts"] = h2Opts

	case "grpc":
		grpcOpts := make(map[string]interface{})
		if p.Path != "" {
			grpcOpts["grpc-service-name"] = p.Path
		}
		options["grpc-opts"] = grpcOpts
		if p.Host != "" {
			options["servername"] = p.Host
		}

	case "quic":
		quicOpts := make(map[string]interface{})
		if p.Host != "" {
			quicOpts["mode"] = p.Host
		}
		if p.Path != "" {
			quicOpts["key"] = p.Path
		}
		options["quic-opts"] = quicOpts
	}
	return options
}

// TrojanProxy represents a Trojan proxy
type TrojanProxy struct {
	BaseProxy     `yaml:",inline"`
	Password      string `yaml:"password" json:"password"`
	Network       string `yaml:"network" json:"network"`
	Path          string `yaml:"path" json:"path"`
	Host          string `yaml:"host" json:"host"`
	TLS           bool   `yaml:"tls" json:"tls"`
	AllowInsecure bool   `yaml:"allow_insecure" json:"allow_insecure"`
}

func (p *TrojanProxy) GenerateLink() (string, error) {
	// Format: trojan://password@server:port?params#remark
	link := fmt.Sprintf("trojan://%s@%s:%d", p.Password, p.Server, p.Port)

	params := []string{}
	if p.Host != "" {
		params = append(params, fmt.Sprintf("sni=%s", p.Host))
	}
	if p.Network == "ws" {
		params = append(params, "type=ws")
		if p.Path != "" {
			params = append(params, fmt.Sprintf("path=%s", urlEncode(p.Path)))
		}
	}
	if p.AllowInsecure {
		params = append(params, "allowInsecure=1")
	}

	if len(params) > 0 {
		link += "?" + strings.Join(params, "&")
	}

	if p.Remark != "" {
		link += "#" + urlEncode(p.Remark)
	}

	return link, nil
}

func (p *TrojanProxy) ProxyOptions() map[string]interface{} {
	options := map[string]interface{}{
		"type":     "trojan",
		"name":     p.Remark,
		"server":   p.Server,
		"port":     p.Port,
		"password": p.Password,
	}
	if p.Host != "" {
		options["sni"] = p.Host
	}

	if p.AllowInsecure {
		options["skip-cert-verify"] = true
	}

	if p.Network != "" {
		options["network"] = p.Network

		switch p.Network {
		case "ws":
			wsOpts := make(map[string]interface{})
			if p.Path != "" {
				wsOpts["path"] = p.Path
			}
			options["ws-opts"] = wsOpts

		case "grpc":
			grpcOpts := make(map[string]interface{})
			if p.Path != "" {
				grpcOpts["grpc-service-name"] = p.Path
			}
			options["grpc-opts"] = grpcOpts
		}
	}
	return options
}

// VLESSProxy represents a VLESS proxy
type VLESSProxy struct {
	BaseProxy     `yaml:",inline"`
	UUID          string `yaml:"uuid" json:"uuid"`
	Network       string `yaml:"network" json:"network"`
	Path          string `yaml:"path" json:"path"`
	Host          string `yaml:"host" json:"host"`
	TLS           bool   `yaml:"tls" json:"tls"`
	AllowInsecure bool   `yaml:"allow_insecure" json:"allow_insecure"`
	Flow          string `yaml:"flow" json:"flow"`
	SNI           string `yaml:"sni" json:"sni"`
}

func (p *VLESSProxy) GenerateLink() (string, error) {
	// Format: vless://uuid@server:port?params#remark
	link := fmt.Sprintf("vless://%s@%s:%d", p.UUID, p.Server, p.Port)

	params := []string{fmt.Sprintf("type=%s", p.Network)}

	if p.TLS {
		params = append(params, "security=tls")
		if p.Host != "" {
			params = append(params, fmt.Sprintf("sni=%s", p.Host))
		}
	}

	if p.Network == "ws" && p.Path != "" {
		params = append(params, fmt.Sprintf("path=%s", urlEncode(p.Path)))
	}
	if p.Host != "" && p.Network == "ws" {
		params = append(params, fmt.Sprintf("host=%s", p.Host))
	}

	link += "?" + strings.Join(params, "&")

	if p.Remark != "" {
		link += "#" + urlEncode(p.Remark)
	}

	return link, nil
}

func (p *VLESSProxy) ProxyOptions() map[string]interface{} {
	options := map[string]interface{}{
		"type":    "vless",
		"name":    p.Remark,
		"server":  p.Server,
		"port":    p.Port,
		"uuid":    p.UUID,
		"network": p.Network,
	}

	if p.Flow != "" {
		options["flow"] = p.Flow
	}

	if p.TLS {
		options["tls"] = true
		if p.SNI != "" {
			options["servername"] = p.SNI
		}
	}

	if p.AllowInsecure {
		options["skip-cert-verify"] = true
	}

	switch p.Network {
	case "ws":
		wsOpts := make(map[string]interface{})
		if p.Path != "" {
			wsOpts["path"] = p.Path
		}
		if p.Host != "" {
			headers := make(map[string]string)
			headers["Host"] = p.Host
			wsOpts["headers"] = headers
		}
		options["ws-opts"] = wsOpts

	case "grpc":
		grpcOpts := make(map[string]interface{})
		if p.Path != "" {
			grpcOpts["grpc-service-name"] = p.Path
		}
		options["grpc-opts"] = grpcOpts

	case "http", "h2":
		h2Opts := make(map[string]interface{})
		if p.Path != "" {
			h2Opts["path"] = p.Path
		}
		if p.Host != "" {
			h2Opts["host"] = []string{p.Host}
		}
		options["h2-opts"] = h2Opts
	}

	return options
}

// HysteriaProxy represents a Hysteria or Hysteria2 proxy
type HysteriaProxy struct {
	BaseProxy     `yaml:",inline"`
	Password      string     `yaml:"password" json:"password"`
	Obfs          string     `yaml:"obfs" json:"obfs"`
	AllowInsecure bool       `yaml:"allow_insecure" json:"allow_insecure"`
	Params        url.Values `yaml:"-" json:"params"`
}

func (p *HysteriaProxy) GenerateLink() (string, error) {
	protocol := p.Type
	link := fmt.Sprintf("%s://%s@%s:%d", protocol, p.Password, p.Server, p.Port)

	if p.AllowInsecure {
		p.Params.Add("insecure", "1")
	}
	if p.Obfs != "" {
		p.Params.Add("obfs", p.Obfs)
	}

	if len(p.Params) > 0 {
		link += "?" + p.Params.Encode()
	}

	if p.Remark != "" {
		link += "#" + urlEncode(p.Remark)
	}

	return link, nil
}

func (p *HysteriaProxy) ProxyOptions() map[string]interface{} {
	options := map[string]interface{}{
		"type":   p.Type,
		"name":   p.Remark,
		"server": p.Server,
		"port":   p.Port,
	}

	if p.Password != "" {
		if p.Type == "hysteria2" {
			options["password"] = p.Password
		} else {
			options["auth-str"] = p.Password
		}
	}

	if p.AllowInsecure {
		options["skip-cert-verify"] = true
	}

	if p.Obfs != "" {
		options["obfs"] = p.Obfs
	}

	if p.Params != nil {
		if sni := p.Params.Get("sni"); sni != "" {
			options["sni"] = sni
		}
		if peer := p.Params.Get("peer"); peer != "" {
			options["sni"] = peer
		}
		if alpn := p.Params.Get("alpn"); alpn != "" {
			options["alpn"] = strings.Split(alpn, ",")
		}

		if p.Type == "hysteria" {
			if p.Password == "" {
				if auth := p.Params.Get("auth"); auth != "" {
					p.Password = auth
					options["auth-str"] = auth
				}
			}

			up := p.Params.Get("upmbps")
			down := p.Params.Get("downmbps")
			if up == "" {
				up = "10"
			}
			if down == "" {
				down = "50"
			}
			if upNum, err := strconv.Atoi(up); err == nil {
				options["up"] = upNum
			}
			if downNum, err := strconv.Atoi(down); err == nil {
				options["down"] = downNum
			}
		}

		if p.Type == "hysteria2" {
			if obfsPassword := p.Params.Get("obfs-password"); obfsPassword != "" {
				options["obfs-password"] = obfsPassword
			}
		}
	} else if p.Type == "hysteria" {
		options["up"] = 10
		options["down"] = 50
	}
	return options
}

// TUICProxy represents a TUIC proxy
type TUICProxy struct {
	BaseProxy     `yaml:",inline"`
	UUID          string     `yaml:"uuid" json:"uuid"`
	Password      string     `yaml:"password" json:"password"`
	AllowInsecure bool       `yaml:"allow_insecure" json:"allow_insecure"`
	Params        url.Values `yaml:"-" json:"params"`
}

func (p *TUICProxy) GenerateLink() (string, error) {
	link := fmt.Sprintf("tuic://%s", p.UUID)
	if p.Password != "" {
		link += ":" + p.Password
	}
	link += fmt.Sprintf("@%s:%d", p.Server, p.Port)

	if p.AllowInsecure {
		p.Params.Add("allow_insecure", "1")
	}

	if len(p.Params) > 0 {
		link += "?" + p.Params.Encode()
	}

	if p.Remark != "" {
		link += "#" + urlEncode(p.Remark)
	}

	return link, nil
}

func (p *TUICProxy) ProxyOptions() map[string]interface{} {
	options := map[string]interface{}{
		"type":   "tuic",
		"name":   p.Remark,
		"server": p.Server,
		"port":   p.Port,
		"uuid":   p.UUID,
	}

	if p.Password != "" {
		options["password"] = p.Password
	}

	if p.AllowInsecure {
		options["skip-cert-verify"] = true
	}
	if p.Params != nil {
		if sni := p.Params.Get("sni"); sni != "" {
			options["sni"] = sni
		}
		if alpn := p.Params.Get("alpn"); alpn != "" {
			options["alpn"] = strings.Split(alpn, ",")
		}
		if congestion := p.Params.Get("congestion_control"); congestion != "" {
			options["congestion-controller"] = congestion
		}
		if udpRelay := p.Params.Get("udp_relay_mode"); udpRelay != "" {
			options["udp-relay-mode"] = udpRelay
		}
	}
	return options
}

// Helper functions

func urlEncode(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, " ", "%20"), "#", "%23")
}

func parsePluginOpts(opts string) map[string]interface{} {
	result := make(map[string]interface{})
	pairs := strings.Split(opts, ";")
	for _, pair := range pairs {
		if pair == "" {
			continue
		}
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) == 2 {
			result[kv[0]] = kv[1]
		} else {
			result[kv[0]] = "true"
		}
	}
	return result
}
