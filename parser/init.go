package parser

import (
	"github.com/gfunc/subconvergo/parser/core"
	"github.com/gfunc/subconvergo/parser/proxy"
	"github.com/gfunc/subconvergo/parser/sub"
)

func init() {
	core.RegisterParser(&proxy.ShadowsocksParser{})
	core.RegisterParser(&proxy.ShadowsocksRParser{})
	core.RegisterParser(&proxy.VMessParser{})
	core.RegisterParser(&proxy.TrojanParser{})
	core.RegisterParser(&proxy.VLESSParser{})
	core.RegisterParser(&proxy.HysteriaParser{})
	core.RegisterParser(&proxy.Hysteria2Parser{})
	core.RegisterParser(&proxy.HttpParser{})
	core.RegisterParser(&proxy.Socks5Parser{})
	core.RegisterParser(&proxy.SnellParser{})
	core.RegisterParser(&proxy.WireGuardParser{})
	core.RegisterParser(&proxy.TUICParser{})
	core.RegisterParser(&proxy.AnyTLSParser{})

	core.RegisterSubscriptionParser(&sub.ClashSubscriptionParser{})
	core.RegisterSubscriptionParser(&sub.SSDSubscriptionParser{})
	core.RegisterSubscriptionParser(&sub.SSSubscriptionParser{})
	core.RegisterSubscriptionParser(&sub.V2RaySubscriptionParser{})
	core.RegisterSubscriptionParser(&sub.SSTapSubscriptionParser{})
	core.RegisterSubscriptionParser(&sub.NetchSubscriptionParser{})
	core.RegisterSubscriptionParser(&sub.SSAndroidSubscriptionParser{})
	core.RegisterSubscriptionParser(&sub.SSRSubscriptionParser{})
	core.RegisterSubscriptionParser(&sub.Base64SubscriptionParser{})
	core.RegisterSubscriptionParser(&sub.PlainSubscriptionParser{})
}
