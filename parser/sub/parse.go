package sub

import (
	"strings"

	"github.com/gfunc/subconvergo/parser/core"
)

// ParseSubscription mimics subconverter's explodeConfContent logic to route parsing
func ParseSubscription(content string) (*core.SubContent, error) {
	// 1. Explicit Format Detection (explodeConfContent)
	if strings.Contains(content, "\"version\"") {
		return (&SSSubscriptionParser{}).Parse(content)
	}
	if strings.Contains(content, "\"serverSubscribes\"") {
		return (&SSRSubscriptionParser{}).Parse(content)
	}
	if strings.Contains(content, "\"uiItem\"") || strings.Contains(content, "vnext") {
		return (&V2RaySubscriptionParser{}).Parse(content)
	}
	if strings.Contains(content, "\"proxy_apps\"") || (strings.Contains(content, "\"server\"") && strings.Contains(content, "\"server_port\"") && strings.Contains(content, "\"method\"")) {
		return (&SSAndroidSubscriptionParser{}).Parse(content)
	}
	if strings.Contains(content, "\"idInUse\"") {
		return (&SSTapSubscriptionParser{}).Parse(content)
	}
	if strings.Contains(content, "\"local_address\"") && strings.Contains(content, "\"local_port\"") {
		return (&SSRSubscriptionParser{}).Parse(content)
	}
	if strings.Contains(content, "\"ModeFileNameType\"") {
		return (&NetchSubscriptionParser{}).Parse(content)
	}

	// 2. Fallback (explodeSub)
	// SSD
	if strings.HasPrefix(content, "ssd://") {
		return (&SSDSubscriptionParser{}).Parse(content)
	}

	// Clash
	clashParser := &ClashSubscriptionParser{}
	if clashParser.CanParse(content) {
		return clashParser.Parse(content)
	}

	// Surge
	surgeParser := &SurgeSubscriptionParser{}
	if surgeParser.CanParse(content) {
		return surgeParser.Parse(content)
	}
	// General
	return (&SingleSubscriptionParser{}).Parse(content)
}
