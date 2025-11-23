package impl

import (
	"fmt"
	"strings"

	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/generator/core"
	"github.com/gfunc/subconvergo/generator/utils"
	pc "github.com/gfunc/subconvergo/proxy/core"
	"github.com/gfunc/subconvergo/proxy/impl"
)

// SurgeGenerator implements the Generator interface for Surge
type SurgeGenerator struct{}

func init() {
	core.RegisterGenerator(&SurgeGenerator{})
}

// Name returns the generator name
func (g *SurgeGenerator) Name() string {
	return "surge"
}

// Generate produces the Surge configuration
func (g *SurgeGenerator) Generate(proxies []pc.ProxyInterface, groups []config.ProxyGroupConfig, rules []string, global *config.Settings, opts core.GeneratorOptions) (string, error) {
	var output strings.Builder

	// Add base configuration
	output.WriteString(opts.Base)
	output.WriteString("\n\n[Proxy]\n")

	// Generate proxy section
	for _, proxy := range proxies {
		line := convertToSurge(proxy, opts)
		if line != "" {
			output.WriteString(line)
			output.WriteString("\n")
		}
	}

	// Generate proxy groups
	if len(groups) > 0 {
		output.WriteString("\n[Proxy Group]\n")

		for _, group := range groups {
			line := convertGroupToSurge(group, proxies)
			if line != "" {
				output.WriteString(line)
				output.WriteString("\n")
			}
		}
	}

	// Generate rules
	if opts.Rule && len(opts.Rulesets) > 0 {
		output.WriteString("\n[Rule]\n")
		for _, ruleset := range opts.Rulesets {
			// Surge rules format: RULE-SET,ruleset-name,policy
			output.WriteString(fmt.Sprintf("RULE-SET,%s,%s\n", ruleset.Ruleset, ruleset.Group))
		}
		output.WriteString("FINAL,DIRECT\n")
	}

	return output.String(), nil
}

func convertToSurge(p pc.ProxyInterface, opts core.GeneratorOptions) string {
	var parts []string

	switch t := p.(type) {
	case *impl.ShadowsocksProxy:
		parts = append(parts, "ss")
		parts = append(parts, fmt.Sprintf("%s:%d", t.Server, t.Port))
		parts = append(parts, fmt.Sprintf("encrypt-method=%s", t.EncryptMethod))
		parts = append(parts, fmt.Sprintf("password=%s", t.Password))
		if opts.UDP {
			parts = append(parts, "udp-relay=true")
		}
		if opts.TFO {
			parts = append(parts, "tfo=true")
		}
		if t.Plugin == "obfs-local" || t.Plugin == "simple-obfs" {
			if mode, ok := t.PluginOpts["obfs"]; ok {
				parts = append(parts, fmt.Sprintf("obfs=%s", mode))
			}
			if host, ok := t.PluginOpts["obfs-host"]; ok {
				parts = append(parts, fmt.Sprintf("obfs-host=%s", host))
			}
		}

	case *impl.HttpProxy:
		if t.Tls {
			parts = append(parts, "https")
		} else {
			parts = append(parts, "http")
		}
		parts = append(parts, fmt.Sprintf("%s:%d", t.Server, t.Port))
		parts = append(parts, fmt.Sprintf("username=%s", t.Username))
		parts = append(parts, fmt.Sprintf("password=%s", t.Password))
		if opts.TFO {
			parts = append(parts, "tfo=true")
		}

	case *impl.SnellProxy:
		parts = append(parts, "snell")
		parts = append(parts, fmt.Sprintf("%s:%d", t.Server, t.Port))
		parts = append(parts, fmt.Sprintf("psk=%s", t.Psk))
		if t.Obfs != "" {
			parts = append(parts, fmt.Sprintf("obfs=%s", t.Obfs))
			if t.ObfsParam != "" {
				parts = append(parts, fmt.Sprintf("obfs-host=%s", t.ObfsParam))
			}
		}
		if t.Version > 0 {
			parts = append(parts, fmt.Sprintf("version=%d", t.Version))
		}
		if opts.TFO {
			parts = append(parts, "tfo=true")
		}

	case *impl.WireGuardProxy:
		parts = append(parts, "wireguard")
		parts = append(parts, fmt.Sprintf("%s:%d", t.Server, t.Port))
		parts = append(parts, fmt.Sprintf("private-key=%s", t.PrivateKey))
		if t.Ip != "" {
			parts = append(parts, fmt.Sprintf("self-ip=%s", t.Ip))
		}
		if t.Ipv6 != "" {
			parts = append(parts, fmt.Sprintf("self-ip-v6=%s", t.Ipv6))
		}
		if len(t.Dns) > 0 {
			parts = append(parts, fmt.Sprintf("dns-server=%s", strings.Join(t.Dns, ", ")))
		}
		if t.Mtu > 0 {
			parts = append(parts, fmt.Sprintf("mtu=%d", t.Mtu))
		}
		if t.PreSharedKey != "" {
			parts = append(parts, fmt.Sprintf("psk=%s", t.PreSharedKey))
		}

	case *impl.Hysteria2Proxy:
		parts = append(parts, "hysteria2")
		parts = append(parts, fmt.Sprintf("%s:%d", t.Server, t.Port))
		parts = append(parts, fmt.Sprintf("password=%s", t.Password))
		if t.Sni != "" {
			parts = append(parts, fmt.Sprintf("sni=%s", t.Sni))
		}
		if t.SkipCertVerify {
			parts = append(parts, "skip-cert-verify=true")
		}
		if t.Obfs != "" {
			parts = append(parts, fmt.Sprintf("obfs=%s", t.Obfs))
			if t.ObfsPassword != "" {
				parts = append(parts, fmt.Sprintf("obfs-password=%s", t.ObfsPassword))
			}
		}
		if opts.TFO {
			parts = append(parts, "tfo=true")
		}

	case *impl.VMessProxy:
		parts = append(parts, "vmess")
		parts = append(parts, fmt.Sprintf("%s:%d", t.Server, t.Port))
		parts = append(parts, fmt.Sprintf("username=%s", t.UUID))
		if t.Network == "ws" {
			parts = append(parts, "ws=true")
			if t.Path != "" {
				parts = append(parts, fmt.Sprintf("ws-path=%s", t.Path))
			}
			if t.Host != "" {
				parts = append(parts, fmt.Sprintf("ws-headers=Host:%s", t.Host))
			}
		}
		if t.TLS {
			parts = append(parts, "tls=true")
			if opts.SCV {
				parts = append(parts, "skip-cert-verify=true")
			}
			if t.SNI != "" {
				parts = append(parts, fmt.Sprintf("sni=%s", t.SNI))
			}
		}

	case *impl.TrojanProxy:
		parts = append(parts, "trojan")
		parts = append(parts, fmt.Sprintf("%s:%d", t.Server, t.Port))
		parts = append(parts, fmt.Sprintf("password=%s", t.Password))
		if opts.SCV || t.AllowInsecure {
			parts = append(parts, "skip-cert-verify=true")
		}
		if t.Network == "ws" {
			parts = append(parts, "ws=true")
			if t.Path != "" {
				parts = append(parts, fmt.Sprintf("ws-path=%s", t.Path))
			}
			if t.Host != "" {
				parts = append(parts, fmt.Sprintf("ws-headers=Host:%s", t.Host))
			}
		}
		if t.Host != "" {
			parts = append(parts, fmt.Sprintf("sni=%s", t.Host))
		}

	default:
		return ""
	}

	return fmt.Sprintf("%s = %s", p.GetRemark(), strings.Join(parts, ", "))
}

func convertGroupToSurge(group config.ProxyGroupConfig, proxies []pc.ProxyInterface) string {
	var parts []string

	// Map group types to Surge types
	surgeType := strings.ToLower(group.Type)
	if surgeType == "selector" {
		surgeType = "select"
	}
	parts = append(parts, surgeType)

	// Filter proxies based on group rules using advanced filtering
	filtered := utils.FilterProxiesByRules(proxies, group.Rule)
	if len(filtered) == 0 {
		filtered = []string{"DIRECT"}
	}
	parts = append(parts, filtered...)

	// Add type-specific options
	switch surgeType {
	case "url-test", "fallback", "load-balance":
		if group.URL != "" {
			parts = append(parts, fmt.Sprintf("url=%s", group.URL))
		}
		if group.Interval > 0 {
			parts = append(parts, fmt.Sprintf("interval=%d", group.Interval))
		}
		if group.Tolerance > 0 {
			parts = append(parts, fmt.Sprintf("tolerance=%d", group.Tolerance))
		}
	}

	return fmt.Sprintf("%s = %s", group.Name, strings.Join(parts, ", "))
}
