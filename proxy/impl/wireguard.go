package impl

import (
	"fmt"
	"strings"

	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/proxy/core"
)

type WireGuardProxy struct {
	core.BaseProxy `yaml:",inline"`
	Ip             string   `yaml:"ip" json:"ip"`
	Ipv6           string   `yaml:"ipv6" json:"ipv6"`
	PrivateKey     string   `yaml:"private-key" json:"private-key"`
	PublicKey      string   `yaml:"public-key" json:"public-key"`
	PreSharedKey   string   `yaml:"pre-shared-key" json:"pre-shared-key"`
	Dns            []string `yaml:"dns" json:"dns"`
	Mtu            int      `yaml:"mtu" json:"mtu"`
	Udp            bool     `yaml:"udp" json:"udp"`
}

func (p *WireGuardProxy) ToSingleConfig(ext *config.ProxySetting) (string, error) {
	// WireGuard doesn't have a standard share link format like others, usually config file.
	// But maybe generic format?
	return "", fmt.Errorf("wireguard share link not supported")
}

func (p *WireGuardProxy) ToClashConfig(ext *config.ProxySetting) (map[string]interface{}, error) {
	options := map[string]interface{}{
		"type":        "wireguard",
		"name":        p.Remark,
		"server":      p.Server,
		"port":        p.Port,
		"ip":          p.Ip,
		"private-key": p.PrivateKey,
		"public-key":  p.PublicKey,
	}
	if p.Ipv6 != "" {
		options["ipv6"] = p.Ipv6
	}
	if p.PreSharedKey != "" {
		options["pre-shared-key"] = p.PreSharedKey
	}
	if len(p.Dns) > 0 {
		options["dns"] = p.Dns
	}
	if p.Mtu > 0 {
		options["mtu"] = p.Mtu
	}
	if p.Udp {
		options["udp"] = true
	}
	return options, nil
}

func (p *WireGuardProxy) ToSurgeConfig(ext *config.ProxySetting) (string, error) {
	return "", fmt.Errorf("wireguard not supported in Surge")
}

func (p *WireGuardProxy) ToLoonConfig(ext *config.ProxySetting) (string, error) {
	// Format: wireguard,server,port,ip,private-key,peer-public-key
	parts := []string{"wireguard", p.Server, fmt.Sprintf("%d", p.Port), p.Ip, p.PrivateKey, p.PublicKey}
	if p.PreSharedKey != "" {
		parts = append(parts, fmt.Sprintf("pre-shared-key=%s", p.PreSharedKey))
	}
	if len(p.Dns) > 0 {
		parts = append(parts, fmt.Sprintf("dns=%s", strings.Join(p.Dns, ",")))
	}
	if p.Mtu > 0 {
		parts = append(parts, fmt.Sprintf("mtu=%d", p.Mtu))
	}
	return fmt.Sprintf("%s = %s", p.Remark, strings.Join(parts, ",")), nil
}

func (p *WireGuardProxy) ToQuantumultXConfig(ext *config.ProxySetting) (string, error) {
	return "", fmt.Errorf("wireguard not supported in Quantumult X")
}

func (p *WireGuardProxy) ToSingboxConfig(ext *config.ProxySetting) (map[string]interface{}, error) {
	outbound := map[string]interface{}{
		"type":            "wireguard",
		"tag":             p.Remark,
		"server":          p.Server,
		"server_port":     p.Port,
		"local_address":   []string{p.Ip},
		"private_key":     p.PrivateKey,
		"peer_public_key": p.PublicKey,
	}
	if p.Ipv6 != "" {
		outbound["local_address"] = append(outbound["local_address"].([]string), p.Ipv6)
	}
	if p.PreSharedKey != "" {
		outbound["pre_shared_key"] = p.PreSharedKey
	}
	if p.Mtu > 0 {
		outbound["mtu"] = p.Mtu
	}
	if len(p.Dns) > 0 {
		outbound["dns_servers"] = p.Dns
	}
	return outbound, nil
}
