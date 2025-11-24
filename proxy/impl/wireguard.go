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
	if p.Server == "" || p.Port == 0 {
		return nil, fmt.Errorf("wireguard server or port missing")
	}
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
	if p.Server == "" || p.Port == 0 {
		return "", fmt.Errorf("wireguard server or port missing")
	}

	// Format: wireguard, interface-ip=..., private-key=..., peers=[{...}]
	var parts []string
	parts = append(parts, "wireguard")

	if p.Ip != "" {
		parts = append(parts, fmt.Sprintf("interface-ip=%s", p.Ip))
	}
	if p.Ipv6 != "" {
		parts = append(parts, fmt.Sprintf("interface-ipv6=%s", p.Ipv6))
	}

	parts = append(parts, fmt.Sprintf("private-key=%s", p.PrivateKey))

	for _, dns := range p.Dns {
		if strings.Contains(dns, ":") {
			parts = append(parts, fmt.Sprintf("dnsv6=%s", dns))
		} else {
			parts = append(parts, fmt.Sprintf("dns=%s", dns))
		}
	}

	if p.Mtu > 0 {
		parts = append(parts, fmt.Sprintf("mtu=%d", p.Mtu))
	}

	// Peer generation
	peerParts := []string{}
	peerParts = append(peerParts, fmt.Sprintf("public-key = %s", p.PublicKey))
	peerParts = append(peerParts, fmt.Sprintf("endpoint = %s:%d", p.Server, p.Port))

	if p.PreSharedKey != "" {
		peerParts = append(peerParts, fmt.Sprintf("preshared-key = %s", p.PreSharedKey))
	}
	// AllowedIPs is not in BaseProxy, assuming 0.0.0.0/0 if not present?
	// Mihomo WireGuard struct might not have AllowedIPs exposed easily or it's in a different field?
	// Checking WireGuardProxy struct: no AllowedIPs field.
	// But subconverter output has allowed-ips.
	// If input doesn't have it, maybe default?
	// subconverter generatePeer uses node.AllowedIPs.

	peerStr := strings.Join(peerParts, ", ")
	parts = append(parts, fmt.Sprintf("peers=[{%s}]", peerStr))

	return fmt.Sprintf("%s = %s", p.Remark, strings.Join(parts, ", ")), nil
}

func (p *WireGuardProxy) ToQuantumultXConfig(ext *config.ProxySetting) (string, error) {
	return "", fmt.Errorf("wireguard not supported in Quantumult X")
}

func (p *WireGuardProxy) ToSingboxConfig(ext *config.ProxySetting) (map[string]interface{}, error) {
	if p.Server == "" || p.Port == 0 {
		return nil, fmt.Errorf("wireguard server or port missing")
	}

	// Construct peer object
	peer := map[string]interface{}{
		"server":      p.Server,
		"server_port": p.Port,
		"public_key":  p.PublicKey,
	}
	if p.PreSharedKey != "" {
		peer["pre_shared_key"] = p.PreSharedKey
	}
	// AllowedIPs default to 0.0.0.0/0 and ::/0 if not present (Subconverter behavior)
	peer["allowed_ips"] = []string{"0.0.0.0/0", "::/0"}

	outbound := map[string]interface{}{
		"type":          "wireguard",
		"tag":           p.Remark,
		"local_address": []string{p.Ip},
		"private_key":   p.PrivateKey,
		"peers":         []interface{}{peer},
	}

	if p.Ipv6 != "" {
		outbound["local_address"] = append(outbound["local_address"].([]string), p.Ipv6)
	}
	if p.Mtu > 0 {
		outbound["mtu"] = p.Mtu
	}
	if len(p.Dns) > 0 {
		outbound["dns_servers"] = p.Dns
	}
	return outbound, nil
}
