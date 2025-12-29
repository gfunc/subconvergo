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

// ParseClash parses a Clash config map
func (p *Hysteria2Parser) ParseClash(config map[string]interface{}) (core.ParsableProxy, error) {
	server := utils.GetStringField(config, "server")
	port := utils.GetIntField(config, "port")
	name := utils.GetStringField(config, "name")
	password := utils.GetStringField(config, "password")
	if password == "" {
		password = utils.GetStringField(config, "auth")
	}
	sni := utils.GetStringField(config, "sni")
	obfs := utils.GetStringField(config, "obfs")
	obfsPassword := utils.GetStringField(config, "obfs-password")
	skipCertVerify := config["skip-cert-verify"] == true

	udp := utils.GetBoolPtrField(config, "udp")
	tfo := utils.GetBoolPtrField(config, "tfo")
	scv := utils.GetBoolPtrField(config, "skip-cert-verify")
	tls13 := utils.GetBoolPtrField(config, "tls13")

	h := &impl.Hysteria2Proxy{
		BaseProxy: core.BaseProxy{
			Type:   "hysteria2",
			Server: server,
			Port:   port,
			Remark: name,
			UDP:    udp,
			TFO:    tfo,
			SCV:    scv,
			TLS13:  tls13,
		},
		Password:       password,
		Sni:            sni,
		SkipCertVerify: skipCertVerify,
		Obfs:           obfs,
		ObfsPassword:   obfsPassword,
	}

	h.Mport = utils.GetStringField(config, "mport")
	h.InitialStreamReceiveWindow = uint64(utils.GetIntField(config, "initial-stream-receive-window"))
	h.MaxStreamReceiveWindow = uint64(utils.GetIntField(config, "max-stream-receive-window"))
	h.InitialConnectionReceiveWindow = uint64(utils.GetIntField(config, "initial-connection-receive-window"))
	h.MaxConnectionReceiveWindow = uint64(utils.GetIntField(config, "max-connection-receive-window"))
	h.UdpMTU = utils.GetIntField(config, "udp-mtu")
	h.IpVersion = utils.GetStringField(config, "ip-version")
	h.ClientFingerprint = utils.GetStringField(config, "client-fingerprint")

	if echOpts, ok := config["ech-opts"].(map[string]interface{}); ok {
		h.EchEnable = utils.GetBoolPtrField(echOpts, "enable")
		h.EchConfig = utils.GetStringField(echOpts, "config")
	}

	h.Certificate = utils.GetStringField(config, "certificate")
	h.PrivateKeyPem = utils.GetStringField(config, "private-key")

	return utils.ToMihomoProxy(h)
}
