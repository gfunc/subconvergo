package impl

import (
	"fmt"

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

func (p *WireGuardProxy) ToShareLink(ext *config.ProxySetting) (string, error) {
	// WireGuard doesn't have a standard share link format like others, usually config file.
	// But maybe generic format?
	return "", fmt.Errorf("wireguard share link not supported")
}

func (p *WireGuardProxy) ToClashConfig(ext *config.ProxySetting) map[string]interface{} {
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
	return options
}
