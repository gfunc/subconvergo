package proxy

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/gfunc/subconvergo/parser/utils"
	"github.com/gfunc/subconvergo/proxy/core"
	"github.com/gfunc/subconvergo/proxy/impl"
)

type WireGuardParser struct{}

func (p *WireGuardParser) Name() string {
	return "WireGuard"
}

func (p *WireGuardParser) CanParseLine(line string) bool {
	// WireGuard usually doesn't have a standard link format.
	// But we can support a custom one if needed, or just return false.
	// For now, let's assume no standard link format support.
	return false
}

func (p *WireGuardParser) ParseSingle(line string) (core.SubconverterProxy, error) {
	return nil, fmt.Errorf("wireguard link parsing not supported")
}

// ParseSurge parses a Surge config string
func (p *WireGuardParser) ParseSurge(content string) (core.SubconverterProxy, error) {
	params := strings.Split(content, ",")
	wg := &impl.WireGuardProxy{
		BaseProxy: core.BaseProxy{
			Type: "wireguard",
		},
	}

	for _, param := range params {
		kv := strings.SplitN(strings.TrimSpace(param), "=", 2)
		if len(kv) == 2 {
			k := strings.TrimSpace(kv[0])
			v := strings.TrimSpace(kv[1])
			switch k {
			case "server":
				wg.Server = v
			case "port":
				wg.Port, _ = strconv.Atoi(v)
			case "private-key":
				wg.PrivateKey = v
			case "self-ip":
				wg.Ip = v
			case "self-ip-v6":
				wg.Ipv6 = v
			case "dns-server":
				wg.Dns = strings.Split(v, ",")
			case "mtu":
				wg.Mtu, _ = strconv.Atoi(v)
			case "public-key", "peer-public-key":
				wg.PublicKey = v
			case "pre-shared-key":
				wg.PreSharedKey = v
			case "endpoint":
				if host, portStr, err := net.SplitHostPort(v); err == nil {
					wg.Server = host
					wg.Port, _ = strconv.Atoi(portStr)
				}
			}
		}
	}

	if wg.Server == "" || wg.Port == 0 {
		// If server/port are not found, maybe they are positional?
		// But usually WG uses keys.
		// Let's check if params[1] and params[2] are positional if they don't have '='
		if len(params) >= 3 {
			if !strings.Contains(params[1], "=") {
				wg.Server = strings.TrimSpace(params[1])
			}
			if !strings.Contains(params[2], "=") {
				wg.Port, _ = strconv.Atoi(strings.TrimSpace(params[2]))
			}
		}
	}

	return utils.ToMihomoProxy(wg)
}

// ParseClash parses a Clash config map
func (p *WireGuardParser) ParseClash(config map[string]interface{}) (core.SubconverterProxy, error) {
	server := utils.GetStringField(config, "server")
	port := utils.GetIntField(config, "port")
	name := utils.GetStringField(config, "name")

	wg := &impl.WireGuardProxy{
		BaseProxy: core.BaseProxy{
			Type:   "wireguard",
			Server: server,
			Port:   port,
			Remark: name,
		},
		PrivateKey:   utils.GetStringField(config, "private-key"),
		PublicKey:    utils.GetStringField(config, "public-key"),
		PreSharedKey: utils.GetStringField(config, "pre-shared-key"),
		Ip:           utils.GetStringField(config, "ip"),
		Ipv6:         utils.GetStringField(config, "ipv6"),
		Mtu:          utils.GetIntField(config, "mtu"),
		Udp:          utils.GetBoolField(config, "udp"),
	}

	if dns, ok := config["dns"].([]interface{}); ok {
		for _, d := range dns {
			if s, ok := d.(string); ok {
				wg.Dns = append(wg.Dns, s)
			}
		}
	}

	return utils.ToMihomoProxy(wg)
}
