package parser

import (
	"github.com/gfunc/subconvergo/parser/core"
	"github.com/gfunc/subconvergo/parser/impl"
)

func init() {
	core.RegisterParser(&impl.ShadowsocksParser{})
	core.RegisterParser(&impl.ShadowsocksRParser{})
	core.RegisterParser(&impl.VMessParser{})
	core.RegisterParser(&impl.TrojanParser{})
	core.RegisterParser(&impl.VLESSParser{})
	core.RegisterParser(&impl.HysteriaParser{})
	core.RegisterParser(&impl.TUICParser{})
	core.RegisterParser(&impl.AnyTLSParser{})
}
