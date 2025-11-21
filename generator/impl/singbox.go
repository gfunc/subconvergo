package impl

import (
	"encoding/json"
	"fmt"

	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/generator/core"
	pc "github.com/gfunc/subconvergo/proxy/core"
	"github.com/gfunc/subconvergo/proxy/impl"
)

// SingBoxGenerator implements the Generator interface for sing-box
type SingBoxGenerator struct{}

func init() {
	core.RegisterGenerator(&SingBoxGenerator{})
}

// Name returns the generator name
func (g *SingBoxGenerator) Name() string {
	return "singbox"
}

// Generate produces the sing-box configuration
func (g *SingBoxGenerator) Generate(proxies []pc.ProxyInterface, groups []config.ProxyGroupConfig, rules []string, global *config.Settings, opts core.GeneratorOptions) (string, error) {
	// Parse base configuration as JSON
	var base map[string]interface{}
	if err := json.Unmarshal([]byte(opts.Base), &base); err != nil {
		return "", fmt.Errorf("failed to parse base config: %w", err)
	}

	// Convert proxies to sing-box outbounds
	var outbounds []map[string]interface{}

	// Add DIRECT and REJECT outbounds
	outbounds = append(outbounds, map[string]interface{}{
		"type": "direct",
		"tag":  "DIRECT",
	})
	outbounds = append(outbounds, map[string]interface{}{
		"type": "block",
		"tag":  "REJECT",
	})

	// Add proxy outbounds
	for _, proxy := range proxies {
		outbound := convertToSingBox(proxy, opts)
		if outbound != nil {
			outbounds = append(outbounds, outbound)
		}
	}

	base["outbounds"] = outbounds

	// Generate routing rules if enabled
	if opts.Rule && len(opts.Rulesets) > 0 {
		rules := generateSingBoxRules(opts.Rulesets)
		if route, ok := base["route"].(map[string]interface{}); ok {
			route["rules"] = rules
		}
	}

	// Add clash_mode if enabled
	if opts.ProxySetting.SingBoxAddClashMode {
		if experimental, ok := base["experimental"].(map[string]interface{}); ok {
			if clashApi, ok := experimental["clash_api"].(map[string]interface{}); ok {
				clashApi["default_mode"] = "rule"
			}
		}
	}

	// Marshal back to JSON with proper indentation
	output, err := json.MarshalIndent(base, "", "    ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal output: %w", err)
	}

	return string(output), nil
}

func convertToSingBox(p pc.ProxyInterface, opts core.GeneratorOptions) map[string]interface{} {
	outbound := map[string]interface{}{
		"tag":         p.GetRemark(),
		"type":        p.GetType(),
		"server":      p.GetServer(),
		"server_port": p.GetPort(),
	}

	switch t := p.(type) {
	case *impl.ShadowsocksProxy:
		outbound["type"] = "shadowsocks"
		outbound["method"] = t.EncryptMethod
		outbound["password"] = t.Password
		if t.Plugin != "" {
			outbound["plugin"] = t.Plugin
			outbound["plugin_opts"] = t.PluginOpts
		}

	case *impl.VMessProxy:
		outbound["uuid"] = t.UUID
		outbound["alter_id"] = t.AlterID
		outbound["security"] = "auto"
		if t.TLS {
			tls := map[string]interface{}{
				"enabled": true,
			}
			if t.SNI != "" {
				tls["server_name"] = t.SNI
			}
			if opts.SCV {
				tls["insecure"] = true
			}
			outbound["tls"] = tls
		}
		if t.Network == "ws" {
			transport := map[string]interface{}{
				"type": "ws",
				"path": t.Path,
			}
			if t.Host != "" {
				transport["headers"] = map[string]string{
					"Host": t.Host,
				}
			}
			outbound["transport"] = transport
		}

	case *impl.TrojanProxy:
		outbound["password"] = t.Password
		tls := map[string]interface{}{
			"enabled": true,
		}
		if t.Host != "" {
			tls["server_name"] = t.Host
		}
		if opts.SCV || t.AllowInsecure {
			tls["insecure"] = true
		}
		outbound["tls"] = tls

		if t.Network == "ws" {
			transport := map[string]interface{}{
				"type": "ws",
				"path": t.Path,
			}
			if t.Host != "" {
				transport["headers"] = map[string]string{
					"Host": t.Host,
				}
			}
			outbound["transport"] = transport
		}
	}

	return outbound
}

func generateSingBoxRules(rulesets []config.RulesetConfig) []map[string]interface{} {
	var rules []map[string]interface{}

	for _, ruleset := range rulesets {
		rule := map[string]interface{}{
			"rule_set": []string{ruleset.Ruleset},
			"outbound": ruleset.Group,
		}
		rules = append(rules, rule)
	}

	// Add final rule
	rules = append(rules, map[string]interface{}{
		"outbound": "DIRECT",
	})

	return rules
}
