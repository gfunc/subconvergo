package transformers

import (
	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/proxy/core"
)

// Transformer modifies the proxy list or configuration before generation
type Transformer interface {
	Transform(proxies []core.ProxyInterface, global *config.Settings) ([]core.ProxyInterface, error)
}
