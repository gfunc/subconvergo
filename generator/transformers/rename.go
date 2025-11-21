package transformers

import (
	"regexp"

	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/generator/utils"
	proxyCore "github.com/gfunc/subconvergo/proxy/core"
)

type RenameTransformer struct {
	Rules []config.RenameNodeConfig
}

func NewRenameTransformer(rules []config.RenameNodeConfig) *RenameTransformer {
	return &RenameTransformer{
		Rules: rules,
	}
}

func (t *RenameTransformer) Transform(proxies []proxyCore.ProxyInterface, global *config.Settings) ([]proxyCore.ProxyInterface, error) {
	if len(t.Rules) == 0 {
		return proxies, nil
	}

	for i := range proxies {
		originalRemark := proxies[i].GetRemark()

		for _, rule := range t.Rules {
			// Skip if no match pattern or both match and replace are empty
			if rule.Match == "" && rule.Script == "" {
				continue
			}

			// TODO: Implement script support if rule.Script is provided

			// Apply matcher-based filtering (supports !!TYPE=, !!GROUP=, etc.)
			if rule.Match != "" {
				matched, realRule := utils.ApplyMatcher(rule.Match, proxies[i])
				if !matched {
					continue
				}

				// If there's a real regex rule after the matcher, apply it
				if realRule != "" {
					re, err := regexp.Compile(realRule)
					if err != nil {
						continue
					}
					proxies[i].SetRemark(re.ReplaceAllString(proxies[i].GetRemark(), rule.Replace))
				}
			}
		}

		// If remark is empty after processing, restore original
		if proxies[i].GetRemark() == "" {
			proxies[i].SetRemark(originalRemark)
		}
	}
	return proxies, nil
}
