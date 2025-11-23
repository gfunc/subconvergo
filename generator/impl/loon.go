package impl

import (
	"fmt"
	"strings"

	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/generator/core"
	"github.com/gfunc/subconvergo/generator/utils"
	pc "github.com/gfunc/subconvergo/proxy/core"
	pimpl "github.com/gfunc/subconvergo/proxy/impl"
)

// LoonGenerator implements the Generator interface for Loon
type LoonGenerator struct{}

func init() {
	core.RegisterGenerator(&LoonGenerator{})
}

// Name returns the generator name
func (g *LoonGenerator) Name() string {
	return "loon"
}

// Generate produces the Loon configuration
func (g *LoonGenerator) Generate(proxies []pc.ProxyInterface, groups []config.ProxyGroupConfig, rules []string, global *config.Settings, opts core.GeneratorOptions) (string, error) {
	// Loon format is very similar to Surge
	var output strings.Builder

	output.WriteString(opts.Base)
	output.WriteString("\n\n[Proxy]\n")

	// Generate proxy section (same format as Surge)
	for _, proxy := range proxies {
		line := convertToLoon(proxy, opts)
		if line != "" {
			output.WriteString(line)
			output.WriteString("\n")
		}
	}

	// Generate proxy groups
	if len(groups) > 0 {
		output.WriteString("\n[Proxy Group]\n")

		for _, group := range groups {
			line := convertGroupToLoon(group, proxies)
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
			output.WriteString(fmt.Sprintf("RULE-SET,%s,%s\n", ruleset.Ruleset, ruleset.Group))
		}
		output.WriteString("FINAL,DIRECT\n")
	}

	return output.String(), nil
}

func convertGroupToLoon(group config.ProxyGroupConfig, proxies []pc.ProxyInterface) string {
	var sb strings.Builder
	sb.WriteString(group.Name)
	sb.WriteString(" = ")
	sb.WriteString(strings.ToLower(group.Type))

	filtered := utils.FilterProxiesByRules(proxies, group.Rule)
	if len(filtered) == 0 {
		filtered = []string{"DIRECT"}
	}

	for _, p := range filtered {
		sb.WriteString(", ")
		sb.WriteString(p)
	}

	sb.WriteString(", img-url=https://raw.githubusercontent.com/Koolson/Qure/master/IconSet/Proxy.png")

	return sb.String()
}

// convertToLoon converts a single proxy to Loon format
func convertToLoon(proxy pc.ProxyInterface, opts core.GeneratorOptions) string {
	switch t := proxy.(type) {
	case *pimpl.HttpProxy:
		// Format: http,server,port,username,"password"
		// Format: https,server,port,username,"password"
		scheme := "http"
		if t.Tls {
			scheme = "https"
		}
		part := fmt.Sprintf("%s,%s,%d,%s,\"%s\"", scheme, t.Server, t.Port, t.Username, t.Password)
		if t.Tls && opts.SCV {
			part += ",skip-cert-verify=true"
		}
		return fmt.Sprintf("%s = %s", t.Remark, part)

	case *pimpl.WireGuardProxy:
		// Format: wireguard, interface-ip=..., private-key=..., peers=[{...}]
		parts := []string{"wireguard"}
		if t.Ip != "" {
			parts = append(parts, fmt.Sprintf("interface-ip=%s", t.Ip))
		}
		if t.Ipv6 != "" {
			parts = append(parts, fmt.Sprintf("interface-ipv6=%s", t.Ipv6))
		}
		parts = append(parts, fmt.Sprintf("private-key=%s", t.PrivateKey))
		for _, dns := range t.Dns {
			if strings.Contains(dns, ":") {
				parts = append(parts, fmt.Sprintf("dnsv6=%s", dns))
			} else {
				parts = append(parts, fmt.Sprintf("dns=%s", dns))
			}
		}
		if t.Mtu > 0 {
			parts = append(parts, fmt.Sprintf("mtu=%d", t.Mtu))
		}

		peer := fmt.Sprintf("public-key=%s,endpoint=%s:%d", t.PublicKey, t.Server, t.Port)
		if t.PreSharedKey != "" {
			peer += fmt.Sprintf(",preshared-key=%s", t.PreSharedKey)
		}
		parts = append(parts, fmt.Sprintf("peers=[{%s}]", peer))

		return fmt.Sprintf("%s = %s", t.Remark, strings.Join(parts, ","))

	case *pimpl.Hysteria2Proxy:
		// Format: hysteria2,server,port,"password"
		parts := []string{"hysteria2", t.Server, fmt.Sprintf("%d", t.Port), fmt.Sprintf("\"%s\"", t.Password)}
		if t.Sni != "" {
			parts = append(parts, fmt.Sprintf("sni=%s", t.Sni))
		}
		if t.SkipCertVerify {
			parts = append(parts, "skip-cert-verify=true")
		}
		return fmt.Sprintf("%s = %s", t.Remark, strings.Join(parts, ","))

	default:
		return convertToSurge(proxy, opts)
	}
}
