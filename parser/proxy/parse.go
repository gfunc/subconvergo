package proxy

import (
	"fmt"
	"strings"

	"github.com/gfunc/subconvergo/proxy/core"
)

// ParseProxy parses a proxy line using explicit protocol detection
// This mimics the routing logic in subconverter's explode function
func ParseProxy(line string) (core.ParsableProxy, error) {
	// ShadowsocksR
	if strings.HasPrefix(line, "ssr://") {
		return (&ShadowsocksRParser{}).ParseSingle(line)
	}
	// VMess
	if strings.HasPrefix(line, "vmess://") || strings.HasPrefix(line, "vmess1://") {
		return (&VMessParser{}).ParseSingle(line)
	}
	// Shadowsocks
	if strings.HasPrefix(line, "ss://") {
		return (&ShadowsocksParser{}).ParseSingle(line)
	}
	// Socks5
	if strings.HasPrefix(line, "socks://") || strings.HasPrefix(line, "https://t.me/socks") || strings.HasPrefix(line, "tg://socks") {
		return (&Socks5Parser{}).ParseSingle(line)
	}
	// HTTP
	if strings.HasPrefix(line, "https://t.me/http") || strings.HasPrefix(line, "tg://http") || strings.HasPrefix(line, "http://") || strings.HasPrefix(line, "https://") {
		return (&HttpParser{}).ParseSingle(line)
	}
	// Trojan
	if strings.HasPrefix(line, "trojan://") {
		return (&TrojanParser{}).ParseSingle(line)
	}
	// Hysteria2
	if strings.HasPrefix(line, "hysteria2://") || strings.HasPrefix(line, "hy2://") {
		return (&Hysteria2Parser{}).ParseSingle(line)
	}
	// TUIC
	if strings.HasPrefix(line, "tuic://") {
		return (&TUICParser{}).ParseSingle(line)
	}
	// AnyTLS
	if strings.HasPrefix(line, "anytls://") {
		return (&AnyTLSParser{}).ParseSingle(line)
	}
	// VLESS
	if strings.HasPrefix(line, "vless://") {
		return (&VLESSParser{}).ParseSingle(line)
	}
	// Hysteria (v1)
	if strings.HasPrefix(line, "hysteria://") {
		return (&HysteriaParser{}).ParseSingle(line)
	}
	// Snell
	if strings.HasPrefix(line, "snell://") {
		return (&SnellParser{}).ParseSingle(line)
	}
	// WireGuard
	if strings.HasPrefix(line, "wireguard://") || strings.HasPrefix(line, "wg://") {
		return (&WireGuardParser{}).ParseSingle(line)
	}

	// Fallback for parsers that support LineMatcher but don't have a simple prefix (e.g. Surge format "Name = ss, ...")
	if (&ShadowsocksParser{}).CanParseLine(line) {
		return (&ShadowsocksParser{}).ParseSingle(line)
	}
	if (&HttpParser{}).CanParseLine(line) {
		return (&HttpParser{}).ParseSingle(line)
	}
	if (&Socks5Parser{}).CanParseLine(line) {
		return (&Socks5Parser{}).ParseSingle(line)
	}
	if (&SnellParser{}).CanParseLine(line) {
		return (&SnellParser{}).ParseSingle(line)
	}
	if (&TrojanParser{}).CanParseLine(line) {
		return (&TrojanParser{}).ParseSingle(line)
	}
	if (&VMessParser{}).CanParseLine(line) {
		return (&VMessParser{}).ParseSingle(line)
	}

	return nil, fmt.Errorf("invalid proxy format: %s", line)
}
