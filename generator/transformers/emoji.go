package transformers

import (
	"regexp"
	"strings"

	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/generator/utils"
	proxyCore "github.com/gfunc/subconvergo/proxy/core"
)

type EmojiTransformer struct {
	Config config.EmojiConfig
}

func NewEmojiTransformer(cfg config.EmojiConfig) *EmojiTransformer {
	return &EmojiTransformer{
		Config: cfg,
	}
}

func (t *EmojiTransformer) Transform(proxies []proxyCore.ProxyInterface, global *config.Settings) ([]proxyCore.ProxyInterface, error) {
	if !t.Config.AddEmoji {
		return proxies, nil
	}

	if len(t.Config.Rules) == 0 {
		return proxies, nil
	}

	for i := range proxies {
		// Remove old emoji first if configured
		if t.Config.RemoveOldEmoji {
			proxies[i].SetRemark(removeEmoji(proxies[i].GetRemark()))
		}

		// Add new emoji based on rules
		for _, rule := range t.Config.Rules {
			if (rule.Match == "" && rule.Script == "") || rule.Emoji == "" {
				continue
			}

			// TODO: Implement script support if rule.Script is provided

			// Apply matcher-based filtering (supports !!TYPE=, !!GROUP=, etc.)
			if rule.Match != "" {
				matched, realRule := utils.ApplyMatcher(rule.Match, proxies[i])
				if !matched {
					continue
				}

				// If there's a real regex rule after the matcher, check if remark matches
				if realRule != "" {
					matched, err := regexp.MatchString(realRule, proxies[i].GetRemark())
					if err != nil || !matched {
						continue
					}
				}

				// Add emoji and break (only first matching rule)
				proxies[i].SetRemark(rule.Emoji + " " + proxies[i].GetRemark())
				break
			}
		}
	}

	return proxies, nil
}

// removeEmoji removes emoji characters from a string
func removeEmoji(s string) string {
	// Simple implementation - remove common emoji patterns
	// This regex removes most emoji and flag sequences
	re := regexp.MustCompile(`[\x{1F600}-\x{1F64F}\x{1F300}-\x{1F5FF}\x{1F680}-\x{1F6FF}\x{2600}-\x{26FF}\x{2700}-\x{27BF}\x{1F900}-\x{1F9FF}\x{1F1E0}-\x{1F1FF}]`)
	s = re.ReplaceAllString(s, "")

	// Remove leading/trailing spaces
	return strings.TrimSpace(s)
}
