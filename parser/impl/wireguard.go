package impl

import (
	"fmt"

	"github.com/gfunc/subconvergo/proxy/core"
)

type WireGuardParser struct{}

func (p *WireGuardParser) Name() string {
	return "WireGuard"
}

func (p *WireGuardParser) CanParse(line string) bool {
	// WireGuard usually doesn't have a standard link format.
	// But we can support a custom one if needed, or just return false.
	// For now, let's assume no standard link format support.
	return false
}

func (p *WireGuardParser) Parse(line string) (core.SubconverterProxy, error) {
	return nil, fmt.Errorf("wireguard link parsing not supported")
}
