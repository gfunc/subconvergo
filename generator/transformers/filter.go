package transformers

import (
	"log"
	"regexp"
	"strings"

	"github.com/gfunc/subconvergo/config"
	proxyCore "github.com/gfunc/subconvergo/proxy/core"
)

type FilterTransformer struct {
	Include []string
	Exclude []string
}

func NewFilterTransformer(include, exclude []string) *FilterTransformer {
	return &FilterTransformer{
		Include: include,
		Exclude: exclude,
	}
}

func (t *FilterTransformer) Transform(proxies []proxyCore.ProxyInterface, global *config.Settings) ([]proxyCore.ProxyInterface, error) {
	// Apply exclude filter
	if len(t.Exclude) > 0 {
		proxies = filterProxies(proxies, t.Exclude, false)
	}

	// Apply include filter
	if len(t.Include) > 0 {
		proxies = filterProxies(proxies, t.Include, true)
	}

	return proxies, nil
}

func filterProxies(proxies []proxyCore.ProxyInterface, patterns []string, include bool) []proxyCore.ProxyInterface {
	if len(patterns) == 0 {
		return proxies
	}

	log.Printf("[FilterTransformer] Filtering %d proxies with patterns %v (include=%v)", len(proxies), patterns, include)

	// Pre-compile regexes for all patterns
	type compiledPat struct {
		raw string
		re  *regexp.Regexp
	}
	compiled := make([]compiledPat, len(patterns))
	for i, p := range patterns {
		compiled[i].raw = p
		if p == "" {
			continue
		}
		var expr string
		if strings.HasPrefix(p, "/") && strings.HasSuffix(p, "/") && len(p) > 2 {
			expr = p[1 : len(p)-1]
		} else {
			expr = regexp.QuoteMeta(p)
		}
		if re, err := regexp.Compile(expr); err == nil {
			compiled[i].re = re
		}
	}

	var result []proxyCore.ProxyInterface
	for _, pr := range proxies {
		matched := false
		for i, p := range patterns {
			if p == "" {
				continue
			}
			if compiled[i].re != nil {
				if compiled[i].re.MatchString(pr.GetRemark()) {
					matched = true
					break
				}
			} else {
				// regex failed to compile; fallback to substring contains
				if strings.Contains(pr.GetRemark(), p) {
					matched = true
					break
				}
			}
		}
		log.Printf("[FilterTransformer] Proxy '%s' matched=%v (include=%v) -> keep=%v", pr.GetRemark(), matched, include, include == matched)
		if include == matched {
			result = append(result, pr)
		}
	}

	return result
}
