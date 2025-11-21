package core

import "github.com/gfunc/subconvergo/proxy/core"

// LineParser defines how to parse a single proxy line
type LineParser interface {
	// Name returns the protocol name (e.g., "Shadowsocks")
	Name() string
	// CanParse checks if the line starts with the protocol prefix
	CanParse(line string) bool
	// Parse converts the line into a ProxyInterface
	Parse(line string) (core.SubconverterProxy, error)
}
