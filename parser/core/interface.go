package core

import (
	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/proxy/core"
)

// SubContent represents the parsed content of a subscription
type SubContent struct {
	Proxies  []core.ProxyInterface
	Groups   []config.ProxyGroupConfig
	RawRules []string
}

// ProxyParser defines how to parse a single proxy configuration
type ProxyParser interface {
	SingeSourceParserMixin
	// Name returns the protocol name (e.g., "Shadowsocks")
	Name() string
}

// LineMatcher indicates the parser can handle single line configurations (e.g. ss://...)
type LineMatcher interface {
	CanParseLine(line string) bool
}

// SubscriptionParser defines how to parse a full subscription/config file
type SubscriptionParser interface {
	// Name returns the parser name
	Name() string
	// CanParse checks if the content can be parsed by this parser
	CanParse(content string) bool
	// Parse converts the content into SubContent
	Parse(content string) (*SubContent, error)
}

type SingeSourceParserMixin interface {
	// ParseSingle converts the content into a ProxyInterface
	ParseSingle(content string) (core.SubconverterProxy, error)
}

type ClashSourceParserMixin interface {
	ParseClash(config map[string]interface{}) (core.SubconverterProxy, error)
}

type V2RaySourceParserMixin interface {
	ParseV2Ray(config map[string]interface{}) (core.SubconverterProxy, error)
}

type NetchSourceParserMixin interface {
	ParseNetch(config map[string]interface{}) (core.SubconverterProxy, error)
}

type SSSourceParserMixin interface {
	ParseSS(config map[string]interface{}) (core.SubconverterProxy, error)
}

type SSTapSourceParserMixin interface {
	ParseSSTap(config map[string]interface{}) (core.SubconverterProxy, error)
}
type SSAndroidSourceParserMixin interface {
	ParseSSAndroid(config map[string]interface{}) (core.SubconverterProxy, error)
}

// SurgeSourceParserMixin defines the interface for proxies that can be parsed from Surge config lines
type SurgeSourceParserMixin interface {
	ParseSurge(content string) (core.SubconverterProxy, error)
}
