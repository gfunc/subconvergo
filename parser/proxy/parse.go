package proxy

import (
	"fmt"
	"strings"

	"github.com/gfunc/subconvergo/proxy/core"
)

// ParseProxy parses a proxy line using explicit protocol detection
// This mimics the routing logic in subconverter's explode function
func ParseProxy(line string) (core.SubconverterProxy, error) {
	// ShadowsocksR
	if strings.HasPrefix(line, "ssr://") {
		return (&ShadowsocksRParser{}).Parse(line)
	}
	// VMess
	if strings.HasPrefix(line, "vmess://") || strings.HasPrefix(line, "vmess1://") {
		return (&VMessParser{}).Parse(line)
	}
	// Shadowsocks
	if strings.HasPrefix(line, "ss://") {
		return (&ShadowsocksParser{}).Parse(line)
	}
	// Socks5
	if strings.HasPrefix(line, "socks://") || strings.HasPrefix(line, "https://t.me/socks") || strings.HasPrefix(line, "tg://socks") {
		return (&Socks5Parser{}).Parse(line)
	}
	// HTTP
	if strings.HasPrefix(line, "https://t.me/http") || strings.HasPrefix(line, "tg://http") || strings.HasPrefix(line, "http://") || strings.HasPrefix(line, "https://") {
		return (&HttpParser{}).Parse(line)
	}
	// Trojan
	if strings.HasPrefix(line, "trojan://") {
		return (&TrojanParser{}).Parse(line)
	}
	// Hysteria2
	if strings.HasPrefix(line, "hysteria2://") || strings.HasPrefix(line, "hy2://") {
		return (&Hysteria2Parser{}).Parse(line)
	}
	// TUIC
	if strings.HasPrefix(line, "tuic://") {
		return (&TUICParser{}).Parse(line)
	}
	// AnyTLS
	if strings.HasPrefix(line, "anytls://") {
		return (&AnyTLSParser{}).Parse(line)
	}
	// VLESS
	if strings.HasPrefix(line, "vless://") {
		return (&VLESSParser{}).Parse(line)
	}
	// Hysteria (v1)
	if strings.HasPrefix(line, "hysteria://") {
		return (&HysteriaParser{}).Parse(line)
	}
	// Snell
	if strings.HasPrefix(line, "snell://") {
		return (&SnellParser{}).Parse(line)
	}
	// WireGuard
	if strings.HasPrefix(line, "wireguard://") || strings.HasPrefix(line, "wg://") {
		return (&WireGuardParser{}).Parse(line)
	}

	// Fallback for parsers that support LineMatcher but don't have a simple prefix (e.g. Surge format "Name = ss, ...")
	if (&ShadowsocksParser{}).CanParseLine(line) {
		return (&ShadowsocksParser{}).Parse(line)
	}
	if (&HttpParser{}).CanParseLine(line) {
		return (&HttpParser{}).Parse(line)
	}
	if (&Socks5Parser{}).CanParseLine(line) {
		return (&Socks5Parser{}).Parse(line)
	}
	if (&SnellParser{}).CanParseLine(line) {
		return (&SnellParser{}).Parse(line)
	}
	if (&TrojanParser{}).CanParseLine(line) {
		return (&TrojanParser{}).Parse(line)
	}
	if (&VMessParser{}).CanParseLine(line) {
		return (&VMessParser{}).Parse(line)
	}

	return nil, fmt.Errorf("invalid proxy format")
}
