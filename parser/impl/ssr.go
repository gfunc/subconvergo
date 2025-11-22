package impl

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/gfunc/subconvergo/parser/utils"
	"github.com/gfunc/subconvergo/proxy/core"
	"github.com/gfunc/subconvergo/proxy/impl"
)

type ShadowsocksRParser struct{}

func (p *ShadowsocksRParser) Name() string {
	return "ShadowsocksR"
}

func (p *ShadowsocksRParser) CanParse(line string) bool {
	return strings.HasPrefix(line, "ssr://")
}

func (p *ShadowsocksRParser) Parse(line string) (core.SubconverterProxy, error) {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "ssr://") {
		return nil, fmt.Errorf("not a valid ssr:// link")
	}

	line = strings.ReplaceAll(line[6:], "\r", "")
	line = utils.UrlSafeBase64Decode(line)

	var remarks, group, server, port, method, password, protocol, protocolParam, obfs, obfsParam string

	if idx := strings.Index(line, "/?"); idx != -1 {
		queryStr := line[idx+2:]
		line = line[:idx]

		params, _ := url.ParseQuery(queryStr)
		group = utils.UrlSafeBase64Decode(params.Get("group"))
		remarks = utils.UrlSafeBase64Decode(params.Get("remarks"))
		obfsParam = strings.TrimSpace(utils.UrlSafeBase64Decode(params.Get("obfsparam")))
		protocolParam = strings.TrimSpace(utils.UrlSafeBase64Decode(params.Get("protoparam")))
	}

	re := regexp.MustCompile(`(\S+):(\d+?):(\S+?):(\S+?):(\S+?):(\S+)`)
	matches := re.FindStringSubmatch(line)
	if len(matches) != 7 {
		return nil, fmt.Errorf("invalid ssr link format")
	}

	server = matches[1]
	port = matches[2]
	protocol = matches[3]
	method = matches[4]
	obfs = matches[5]
	password = utils.UrlSafeBase64Decode(matches[6])

	portNum, err := strconv.Atoi(port)
	if err != nil || portNum == 0 {
		return nil, fmt.Errorf("invalid port: %s", port)
	}

	if group == "" {
		group = "SSR"
	}
	if remarks == "" {
		remarks = server + ":" + port
	}

	ssCiphers := []string{"rc4-md5", "aes-128-gcm", "aes-192-gcm", "aes-256-gcm", "aes-128-cfb",
		"aes-192-cfb", "aes-256-cfb", "aes-128-ctr", "aes-192-ctr", "aes-256-ctr", "chacha20-ietf-poly1305",
		"xchacha20-ietf-poly1305", "2022-blake3-aes-128-gcm", "2022-blake3-aes-256-gcm"}

	isSS := false
	for _, cipher := range ssCiphers {
		if cipher == method && (obfs == "" || obfs == "plain") && (protocol == "" || protocol == "origin") {
			isSS = true
			break
		}
	}
	var pObj *impl.ShadowsocksRProxy
	if isSS {
		pObj = &impl.ShadowsocksRProxy{
			BaseProxy: core.BaseProxy{
				Type:   "ss",
				Remark: remarks,
				Server: server,
				Port:   portNum,
				Group:  group,
			},
			Password:      password,
			EncryptMethod: method,
		}
	} else {
		pObj = &impl.ShadowsocksRProxy{
			BaseProxy: core.BaseProxy{
				Type:   "ssr",
				Remark: remarks,
				Server: server,
				Port:   portNum,
				Group:  group,
			},
			Password:      password,
			EncryptMethod: method,
			Protocol:      protocol,
			ProtocolParam: protocolParam,
			Obfs:          obfs,
			ObfsParam:     obfsParam,
		}
	}

	return utils.ToMihomoProxy(pObj)
}
