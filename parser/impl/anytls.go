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

type AnyTLSParser struct{}

func (p *AnyTLSParser) Name() string {
	return "AnyTLS"
}

func (p *AnyTLSParser) CanParse(line string) bool {
	return strings.HasPrefix(line, "anytls://")
}

func (p *AnyTLSParser) Parse(line string) (core.SubconverterProxy, error) {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "anytls://") {
		return nil, fmt.Errorf("not a valid anytls:// link")
	}

	line = line[9:] // Remove anytls://

	var remark, password, server, portStr string
	var sni, fingerprint string
	var alpn []string
	var idleSessionCheckInterval, idleSessionTimeout, minIdleSession int
	var tfo, allowInsecure bool

	// Handle remark
	if idx := strings.LastIndex(line, "#"); idx != -1 {
		remark = utils.UrlDecode(line[idx+1:])
		line = line[:idx]
	}

	// Handle query params
	if idx := strings.Index(line, "?"); idx != -1 {
		queryStr := line[idx+1:]
		line = line[:idx]

		params, _ := url.ParseQuery(queryStr)

		password = params.Get("password")
		sni = params.Get("peer")
		if val := params.Get("alpn"); val != "" {
			alpn = strings.Split(val, ",")
		}
		fingerprint = params.Get("hpkp")

		if params.Get("tfo") == "1" || params.Get("tfo") == "true" {
			tfo = true
		}
		if params.Get("insecure") == "1" || params.Get("insecure") == "true" {
			allowInsecure = true
		}

		idleSessionCheckInterval, _ = strconv.Atoi(params.Get("idle_session_check_interval"))
		idleSessionTimeout, _ = strconv.Atoi(params.Get("idle_session_timeout"))
		minIdleSession, _ = strconv.Atoi(params.Get("min_idle_session"))
	}

	// Handle user info (password@server:port)
	if idx := strings.Index(line, "@"); idx != -1 {
		// password@server:port
		password = line[:idx]
		line = line[idx+1:]
	}

	// Handle server:port
	if idx := strings.LastIndex(line, ":"); idx != -1 {
		server = line[:idx]
		portStr = line[idx+1:]
	} else {
		return nil, fmt.Errorf("invalid server:port format")
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid port: %v", err)
	}

	if remark == "" {
		remark = fmt.Sprintf("%s:%d", server, port)
	}

	pObj := &impl.AnyTLSProxy{
		BaseProxy: core.BaseProxy{
			Type:   "anytls",
			Remark: remark,
			Server: server,
			Port:   port,
		},
		Password:                 password,
		SNI:                      sni,
		Alpn:                     alpn,
		Fingerprint:              fingerprint,
		IdleSessionCheckInterval: idleSessionCheckInterval,
		IdleSessionTimeout:       idleSessionTimeout,
		MinIdleSession:           minIdleSession,
		TFO:                      tfo,
		AllowInsecure:            allowInsecure,
	}
	return utils.ToMihomoProxy(pObj)
}
