package proxy

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

func (p *ShadowsocksRParser) CanParseLine(line string) bool {
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
		group = core.SSR_DEFAULT_GROUP
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

// ParseClash parses a Clash config map
func (p *ShadowsocksRParser) ParseClash(config map[string]interface{}) (core.SubconverterProxy, error) {
	server := utils.GetStringField(config, "server")
	port := utils.GetIntField(config, "port")
	name := utils.GetStringField(config, "name")
	password := utils.GetStringField(config, "password")
	cipher := utils.GetStringField(config, "cipher")
	protocol := utils.GetStringField(config, "protocol")
	protocolParam := utils.GetStringField(config, "protocol-param")
	obfs := utils.GetStringField(config, "obfs")
	obfsParam := utils.GetStringField(config, "obfs-param")

	ssr := &impl.ShadowsocksRProxy{
		BaseProxy: core.BaseProxy{
			Type:   "ssr",
			Server: server,
			Port:   port,
			Remark: name,
		},
		Password:      password,
		EncryptMethod: cipher,
		Protocol:      protocol,
		ProtocolParam: protocolParam,
		Obfs:          obfs,
		ObfsParam:     obfsParam,
	}
	return ssr, nil
}

// ParseNetch parses a Netch config map
func (p *ShadowsocksRParser) ParseNetch(config map[string]interface{}) (core.SubconverterProxy, error) {
	remark := utils.GetStringField(config, "Remark")
	hostname := utils.GetStringField(config, "Hostname")
	port := utils.GetIntField(config, "Port")
	password := utils.GetStringField(config, "Password")
	method := utils.GetStringField(config, "EncryptMethod")
	protocol := utils.GetStringField(config, "Protocol")
	protocolParam := utils.GetStringField(config, "ProtocolParam")
	obfs := utils.GetStringField(config, "OBFS")
	obfsParam := utils.GetStringField(config, "OBFSParam")

	ssr := &impl.ShadowsocksRProxy{
		BaseProxy: core.BaseProxy{
			Type:   "ssr",
			Server: hostname,
			Port:   port,
			Remark: remark,
		},
		Password:      password,
		EncryptMethod: method,
		Protocol:      protocol,
		ProtocolParam: protocolParam,
		Obfs:          obfs,
		ObfsParam:     obfsParam,
	}
	return ssr, nil
}

// ParseSSTap parses a SSTap config map
func (p *ShadowsocksRParser) ParseSSTap(config map[string]interface{}) (core.SubconverterProxy, error) {
	name := utils.GetStringField(config, "name")
	server := utils.GetStringField(config, "server")
	port := utils.GetIntField(config, "port")
	password := utils.GetStringField(config, "password")
	method := utils.GetStringField(config, "method")
	protocol := utils.GetStringField(config, "protocol")
	protocolParam := utils.GetStringField(config, "protocol_param")
	obfs := utils.GetStringField(config, "obfs")
	obfsParam := utils.GetStringField(config, "obfs_param")

	ssr := &impl.ShadowsocksRProxy{
		BaseProxy: core.BaseProxy{
			Type:   "ssr",
			Server: server,
			Port:   port,
			Remark: name,
		},
		Password:      password,
		EncryptMethod: method,
		Protocol:      protocol,
		ProtocolParam: protocolParam,
		Obfs:          obfs,
		ObfsParam:     obfsParam,
	}
	return ssr, nil
}
