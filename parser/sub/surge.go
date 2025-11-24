package sub

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/gfunc/subconvergo/parser/core"
	"github.com/gfunc/subconvergo/parser/proxy"
	"github.com/gfunc/subconvergo/parser/utils"
	proxyCore "github.com/gfunc/subconvergo/proxy/core"
	"gopkg.in/ini.v1"
)

type SurgeSubscriptionParser struct{}

func (p *SurgeSubscriptionParser) Name() string {
	return "Surge"
}

func (p *SurgeSubscriptionParser) CanParse(content string) bool {
	content = utils.UrlSafeBase64Decode(content)
	surgeRegex := regexp.MustCompile(`.+\s*?=\s*?(vmess|shadowsocks|http|trojan|ss|ssr|snell|socks5)\s*?,`)
	
	cfg, err := ini.Load([]byte(content))
	if err != nil {
		return false
	}

	section := cfg.Section("Proxy")
	if section == nil {
		return false
	}
	return surgeRegex.MatchString(content)
}

func (p *SurgeSubscriptionParser) Parse(content string) (*core.SubContent, error) {
	if !strings.Contains(content, "[Proxy]") {
		content = "[Proxy]\n" + content
	}
	// Preprocess content to handle [Proxy] section correctly if needed
	// subconverter does: surge = regReplace(surge, R"(^[\S\s]*?\[)", "[", false);
	// This removes everything before the first [
	if idx := strings.Index(content, "["); idx != -1 {
		content = content[idx:]
	}

	cfg, err := ini.Load([]byte(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse Surge INI: %w", err)
	}

	section := cfg.Section("Proxy")
	if section == nil {
		return nil, fmt.Errorf("no [Proxy] section found")
	}

	var proxies []proxyCore.ProxyInterface

	// Regex to split "Remark = Type, ..."
	// subconverter: "(.*?)\\s*=\\s*(.*)"
	// INI parser handles the key=value split.
	// Key is Remark, Value is the rest.

	for _, key := range section.Keys() {
		remark := key.Name()
		value := key.Value()

		// Split value by comma
		parts := strings.Split(value, ",")
		if len(parts) < 1 {
			continue
		}

		proxyType := strings.TrimSpace(parts[0])

		// Skip built-in types
		switch strings.ToLower(proxyType) {
		case "direct", "reject", "reject-tinygif":
			continue
		}

		var p proxyCore.SubconverterProxy
		var parseErr error

		switch strings.ToLower(proxyType) {
		case "custom": // Surge 2 SS
			p, parseErr = (&proxy.ShadowsocksParser{}).ParseSurge(value)

		case "ss": // Surge 3 SS
			p, parseErr = (&proxy.ShadowsocksParser{}).ParseSurge(value)

		case "socks5", "socks5-tls":
			p, parseErr = (&proxy.Socks5Parser{}).ParseSurge(value)

		case "vmess":
			p, parseErr = (&proxy.VMessParser{}).ParseSurge(value)

		case "http", "https":
			p, parseErr = (&proxy.HttpParser{}).ParseSurge(value)

		case "trojan":
			p, parseErr = (&proxy.TrojanParser{}).ParseSurge(value)

		case "snell":
			p, parseErr = (&proxy.SnellParser{}).ParseSurge(value)

		case "wireguard":
			// WireGuard in Surge is complex, often referencing another section.
			// subconverter: section-name=... -> read [WireGuard section-name]
			// We need to find section-name in the value string
			var sectionName string
			for _, part := range parts {
				kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
				if len(kv) == 2 && strings.TrimSpace(kv[0]) == "section-name" {
					sectionName = strings.TrimSpace(kv[1])
					break
				}
			}

			wgConfig := value
			if sectionName != "" {
				wgSection := cfg.Section("WireGuard " + sectionName)
				if wgSection != nil {
					for _, k := range wgSection.Keys() {
						wgConfig += fmt.Sprintf(", %s=%s", k.Name(), k.Value())
					}
				}
			}
			p, parseErr = (&proxy.WireGuardParser{}).ParseSurge(wgConfig)
		}

		if parseErr == nil && p != nil {
			p.SetRemark(remark)
			proxies = append(proxies, p)
		}
	}

	return &core.SubContent{
		Proxies: proxies,
	}, nil
}
