package proxy

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/gfunc/subconvergo/parser/utils"
	"github.com/gfunc/subconvergo/proxy/core"
	"github.com/gfunc/subconvergo/proxy/impl"
)

type VMessParser struct{}

func (p *VMessParser) Name() string {
	return "VMess"
}

func (p *VMessParser) CanParseLine(line string) bool {
	return strings.HasPrefix(line, "vmess://") || isSurgeVMess(line)
}

func isSurgeVMess(line string) bool {
	parts := strings.SplitN(line, "=", 2)
	if len(parts) != 2 {
		return false
	}
	val := strings.TrimSpace(parts[1])
	return strings.HasPrefix(val, "vmess,")
}

func (p *VMessParser) ParseSingle(line string) (core.SubconverterProxy, error) {
	line = strings.TrimSpace(line)
	if isSurgeVMess(line) {
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

	if !strings.HasPrefix(line, "vmess://") {
		return nil, fmt.Errorf("not a valid vmess:// link")
	}

	line = line[8:]

	// Check for standard VMess format (uuid@host:port)
	if strings.Contains(line, "@") && !strings.Contains(line, "://") { // Avoid matching other protocols if any
		return p.parseStdVMess(line)
	}

	decoded := utils.UrlSafeBase64Decode(line)

	// Check if it's JSON
	if strings.HasPrefix(strings.TrimSpace(decoded), "{") {
		return p.parseJSONVMess(decoded)
	}

	// Check if it's base64 encoded standard VMess
	if strings.Contains(decoded, "@") {
		return p.parseStdVMess(decoded)
	}

	return nil, fmt.Errorf("unknown vmess format")
}

func (p *VMessParser) parseStdVMess(line string) (core.SubconverterProxy, error) {
	parts := strings.SplitN(line, "@", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid standard vmess format")
	}
	uuid := parts[0]
	rest := parts[1]

	var remark string
	var params url.Values

	if idx := strings.Index(rest, "#"); idx != -1 {
		remark = utils.UrlDecode(rest[idx+1:])
		rest = rest[:idx]
	}

	if idx := strings.Index(rest, "?"); idx != -1 {
		queryStr := rest[idx+1:]
		rest = rest[:idx]
		params, _ = url.ParseQuery(queryStr)
	}

	serverPort := strings.Split(rest, ":")
	if len(serverPort) != 2 {
		return nil, fmt.Errorf("invalid server:port")
	}
	server := serverPort[0]
	portStr := serverPort[1]

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid port: %s", portStr)
	}

	if remark == "" {
		remark = params.Get("remark")
		if remark == "" {
			remark = server + ":" + portStr
		}
	}

	net := params.Get("network")
	if net == "" {
		net = "tcp"
	}

	aidStr := params.Get("aid")
	aid := 0
	if aidStr != "" {
		aid, _ = strconv.Atoi(aidStr)
	}

	tls := params.Get("tls") == "1" || params.Get("tls") == "true" || params.Get("tls") == "tls"
	sni := params.Get("sni")
	if sni == "" {
		sni = params.Get("peer")
	}
	path := params.Get("path")
	host := params.Get("host")
	if host == "" {
		host = params.Get("obfsParam")
	}

	pObj := &impl.VMessProxy{
		BaseProxy: core.BaseProxy{
			Type:   "vmess",
			Remark: remark,
			Server: server,
			Port:   port,
			Group:  core.V2RAY_DEFAULT_GROUP,
		},
		UUID:    uuid,
		AlterID: aid,
		Network: net,
		Path:    path,
		Host:    host,
		TLS:     tls,
		SNI:     sni,
	}
	return utils.ToMihomoProxy(pObj)
}

func (p *VMessParser) parseJSONVMess(decoded string) (core.SubconverterProxy, error) {
	var vmessData map[string]interface{}
	if err := json.Unmarshal([]byte(decoded), &vmessData); err != nil {
		return nil, fmt.Errorf("failed to parse vmess JSON: %w", err)
	}

	ps := utils.GetStringField(vmessData, "ps")
	add := utils.GetStringField(vmessData, "add")
	port := utils.GetStringField(vmessData, "port")
	id := utils.GetStringField(vmessData, "id")
	aid := utils.GetStringField(vmessData, "aid")
	net := utils.GetStringField(vmessData, "net")
	host := utils.GetStringField(vmessData, "host")
	path := utils.GetStringField(vmessData, "path")
	tls := utils.GetStringField(vmessData, "tls")
	sni := utils.GetStringField(vmessData, "sni")

	if net == "" {
		net = "tcp"
	}
	if aid == "" {
		aid = "0"
	}

	portNum, err := strconv.Atoi(port)
	if err != nil || portNum == 0 {
		return nil, fmt.Errorf("invalid port: %s", port)
	}

	alterID, _ := strconv.Atoi(aid)

	if ps == "" {
		ps = add + ":" + port
	}

	version := utils.GetStringField(vmessData, "v")
	if version == "1" || version == "" {
		if host != "" && strings.Contains(host, ";") {
			parts := strings.SplitN(host, ";", 2)
			host = parts[0]
			if path == "" {
				path = parts[1]
			}
		}
	}

	pObj := &impl.VMessProxy{
		BaseProxy: core.BaseProxy{
			Type:   "vmess",
			Remark: ps,
			Server: add,
			Port:   portNum,
			Group:  core.V2RAY_DEFAULT_GROUP,
		},
		UUID:    id,
		AlterID: alterID,
		Network: net,
		Path:    path,
		Host:    host,
		TLS:     tls == "tls",
		SNI:     sni,
	}
	return utils.ToMihomoProxy(pObj)
}

// ParseSurge parses a Surge config string
func (p *VMessParser) ParseSurge(content string) (core.SubconverterProxy, error) {
	params := strings.Split(content, ",")
	if len(params) < 3 {
		return nil, fmt.Errorf("invalid surge vmess config: %s", content)
	}

	server := strings.TrimSpace(params[1])
	portStr := strings.TrimSpace(params[2])
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid port: %s", portStr)
	}

	vmess := &impl.VMessProxy{
		BaseProxy: core.BaseProxy{
			Type:   "vmess",
			Server: server,
			Port:   port,
		},
		Network: "tcp", // Default
	}

	for i := 3; i < len(params); i++ {
		kv := strings.SplitN(strings.TrimSpace(params[i]), "=", 2)
		if len(kv) == 2 {
			k := strings.TrimSpace(kv[0])
			v := strings.TrimSpace(kv[1])
			switch k {
			case "username":
				vmess.UUID = v
			case "ws":
				if v == "true" {
					vmess.Network = "ws"
				}
			case "ws-path":
				vmess.Path = v
			case "tls":
				vmess.TLS = v == "true"
			case "sni":
				vmess.SNI = v
			}
		}
	}

	return utils.ToMihomoProxy(vmess)
}

// ParseClash parses a Clash config map
func (p *VMessParser) ParseClash(config map[string]interface{}) (core.SubconverterProxy, error) {
	server := utils.GetStringField(config, "server")
	port := utils.GetIntField(config, "port")
	uuid := utils.GetStringField(config, "uuid")
	alterId := utils.GetIntField(config, "alterId")
	cipher := utils.GetStringField(config, "cipher")
	name := utils.GetStringField(config, "name")
	network := utils.GetStringField(config, "network")
	tls := utils.GetStringField(config, "tls") == "true" || config["tls"] == true
	sni := utils.GetStringField(config, "servername")
	if sni == "" {
		sni = utils.GetStringField(config, "sni")
	}

	// ws-opts, http-opts, etc.
	var path, host string
	if wsOpts, ok := config["ws-opts"].(map[string]interface{}); ok {
		path = utils.GetStringField(wsOpts, "path")
		if headers, ok := wsOpts["headers"].(map[string]interface{}); ok {
			host = utils.GetStringField(headers, "Host")
		}
	}
	// ... handle other transports

	vmess := &impl.VMessProxy{
		BaseProxy: core.BaseProxy{
			Type:   "vmess",
			Server: server,
			Port:   port,
			Remark: name,
		},
		UUID:    uuid,
		AlterID: alterId,
		Cipher:  cipher,
		Network: network,
		TLS:     tls,
		SNI:     sni,
		Path:    path,
		Host:    host,
	}
	return utils.ToMihomoProxy(vmess)
}

// ParseNetch parses a Netch config map
func (p *VMessParser) ParseNetch(config map[string]interface{}) (core.SubconverterProxy, error) {
	remark := utils.GetStringField(config, "Remark")
	hostname := utils.GetStringField(config, "Hostname")
	port := utils.GetIntField(config, "Port")
	uuid := utils.GetStringField(config, "UserID")
	alterId := utils.GetIntField(config, "AlterID")
	cipher := utils.GetStringField(config, "EncryptMethod")
	network := utils.GetStringField(config, "TransferProtocol")
	tls := utils.GetStringField(config, "TLS") == "tls"
	host := utils.GetStringField(config, "Host")
	path := utils.GetStringField(config, "Path")
	fakeType := utils.GetStringField(config, "FakeType")

	vmess := &impl.VMessProxy{
		BaseProxy: core.BaseProxy{
			Type:   "vmess",
			Server: hostname,
			Port:   port,
			Remark: remark,
		},
		UUID:     uuid,
		AlterID:  alterId,
		Cipher:   cipher,
		Network:  network,
		TLS:      tls,
		Host:     host,
		Path:     path,
		FakeType: fakeType,
	}
	return utils.ToMihomoProxy(vmess)
}

// ParseV2Ray parses a V2Ray config map (resolved)
func (p *VMessParser) ParseV2Ray(config map[string]interface{}) (core.SubconverterProxy, error) {
	// Expects resolved fields
	server := utils.GetStringField(config, "address")
	port := utils.GetIntField(config, "port")
	id := utils.GetStringField(config, "id")
	aid := utils.GetIntField(config, "alterId")
	security := utils.GetStringField(config, "security")
	network := utils.GetStringField(config, "network")
	typeStr := utils.GetStringField(config, "type")
	host := utils.GetStringField(config, "host")
	path := utils.GetStringField(config, "path")
	tls := utils.GetStringField(config, "tls") == "true"
	sni := utils.GetStringField(config, "sni")
	remark := utils.GetStringField(config, "remark")

	vmess := &impl.VMessProxy{
		BaseProxy: core.BaseProxy{
			Type:   "vmess",
			Server: server,
			Port:   port,
			Remark: remark,
		},
		UUID:     id,
		AlterID:  aid,
		Cipher:   security,
		Network:  network,
		FakeType: typeStr,
		Host:     host,
		Path:     path,
		TLS:      tls,
		SNI:      sni,
	}
	return utils.ToMihomoProxy(vmess)
}
