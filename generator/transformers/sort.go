package transformers

import (
	"sort"

	"github.com/gfunc/subconvergo/config"
	proxyCore "github.com/gfunc/subconvergo/proxy/core"
)

type SortTransformer struct {
	Enabled bool
}

func NewSortTransformer(enabled bool) *SortTransformer {
	return &SortTransformer{
		Enabled: enabled,
	}
}

func (t *SortTransformer) Transform(proxies []proxyCore.ProxyInterface, global *config.Settings) ([]proxyCore.ProxyInterface, error) {
	if !t.Enabled {
		return proxies, nil
	}

	// Simple alphabetical sort by remark
	// TODO: Implement script-based sorting if sortScript is provided
	sort.Slice(proxies, func(i, j int) bool {
		return proxies[i].GetRemark() < proxies[j].GetRemark()
	})

	return proxies, nil
}
