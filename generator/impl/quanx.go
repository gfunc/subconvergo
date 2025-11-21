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

// QuantumultXGenerator implements the Generator interface for Quantumult X
type QuantumultXGenerator struct{}

func init() {
	core.RegisterGenerator(&QuantumultXGenerator{})
}

// Name returns the generator name
func (g *QuantumultXGenerator) Name() string {
	return "quanx"
}

// Generate produces the Quantumult X configuration
func (g *QuantumultXGenerator) Generate(proxies []pc.ProxyInterface, groups []config.ProxyGroupConfig, rules []string, global *config.Settings, opts core.GeneratorOptions) (string, error) {
	var output strings.Builder

	// Add base configuration
	output.WriteString(opts.Base)
	output.WriteString("\n\n[server_local]\n")

	// Generate proxy section
	for _, proxy := range proxies {
		line := convertToQuantumultX(proxy, opts)
		if line != "" {
			output.WriteString(line)
			output.WriteString("\n")
		}
	}

	// Generate policy section
	if len(groups) > 0 {
		output.WriteString("\n[policy]\n")

		for _, group := range groups {
			line := convertGroupToQuantumultX(group, proxies)
			if line != "" {
				output.WriteString(line)
				output.WriteString("\n")
			}
		}
	}

	// Generate filter rules
	if opts.Rule && len(opts.Rulesets) > 0 {
		output.WriteString("\n[filter_local]\n")
		for _, ruleset := range opts.Rulesets {
			output.WriteString(fmt.Sprintf("# %s rules\n", ruleset.Ruleset))
		}
		output.WriteString(fmt.Sprintf("FINAL,%s\n", "DIRECT"))
	}

	return output.String(), nil
}

func convertToQuantumultX(p pc.ProxyInterface, opts core.GeneratorOptions) string {
	var parts []string

	switch t := p.(type) {
	case *impl.ShadowsocksProxy:
		// Format: shadowsocks=server:port, method=encrypt-method, password=password, tag=name
		parts = append(parts, "shadowsocks="+fmt.Sprintf("%s:%d", t.Server, t.Port))
		parts = append(parts, fmt.Sprintf("method=%s", t.EncryptMethod))
		parts = append(parts, fmt.Sprintf("password=%s", t.Password))
		if opts.UDP {
			parts = append(parts, "udp-relay=true")
		}
		if opts.TFO {
			parts = append(parts, "fast-open=true")
		}
		if t.Plugin == "obfs-local" || t.Plugin == "simple-obfs" {
			if mode, ok := t.PluginOpts["obfs"]; ok {
				parts = append(parts, fmt.Sprintf("obfs=%s", mode))
			}
			if host, ok := t.PluginOpts["obfs-host"]; ok {
				parts = append(parts, fmt.Sprintf("obfs-host=%s", host))
			}
		}
		parts = append(parts, fmt.Sprintf("tag=%s", t.Remark))

	case *impl.VMessProxy:
		// Format: vmess=server:port, method=encrypt-method, password=uuid, tag=name
		parts = append(parts, "vmess="+fmt.Sprintf("%s:%d", t.Server, t.Port))
		parts = append(parts, "method=aes-128-gcm")
		parts = append(parts, fmt.Sprintf("password=%s", t.UUID))
		if t.Network == "ws" {
			parts = append(parts, "obfs=ws")
			if t.Path != "" {
				parts = append(parts, fmt.Sprintf("obfs-uri=%s", t.Path))
			}
			if t.Host != "" {
				parts = append(parts, fmt.Sprintf("obfs-host=%s", t.Host))
			}
		} else if t.TLS {
			parts = append(parts, "obfs=over-tls")
			if t.SNI != "" {
				parts = append(parts, fmt.Sprintf("obfs-host=%s", t.SNI))
			}
		}
		if opts.SCV {
			parts = append(parts, "tls-verification=false")
		}
		parts = append(parts, fmt.Sprintf("tag=%s", t.Remark))

	case *impl.TrojanProxy:
		// Format: trojan=server:port, password=password, tag=name
		parts = append(parts, "trojan="+fmt.Sprintf("%s:%d", t.Server, t.Port))
		parts = append(parts, fmt.Sprintf("password=%s", t.Password))
		if opts.SCV || t.AllowInsecure {
			parts = append(parts, "tls-verification=false")
		}
		if t.Host != "" {
			parts = append(parts, fmt.Sprintf("tls-host=%s", t.Host))
		}
		parts = append(parts, fmt.Sprintf("tag=%s", t.Remark))

	default:
		return ""
	}

	return strings.Join(parts, ", ")
}

func convertGroupToQuantumultX(group config.ProxyGroupConfig, proxies []pc.ProxyInterface) string {
	groupType := strings.ToLower(group.Type)
	if groupType == "select" {
		groupType = "static"
	} else if groupType == "url-test" {
		groupType = "available"
	}

	var parts []string
	parts = append(parts, fmt.Sprintf("%s=%s", groupType, group.Name))

	// Filter proxies using advanced filtering
	filtered := utils.FilterProxiesByRules(proxies, group.Rule)
	if len(filtered) == 0 {
		filtered = []string{"direct"}
	}
	parts = append(parts, filtered...)

	// Add img-url if needed
	parts = append(parts, "img-url=https://raw.githubusercontent.com/Koolson/Qure/master/IconSet/Proxy.png")

	return strings.Join(parts, ", ")
}
