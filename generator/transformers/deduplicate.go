package transformers

import (
	"fmt"
	"strings"

	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/proxy/core"
)

// DeduplicateTransformer ensures all proxies have unique names
type DeduplicateTransformer struct{}

// NewDeduplicateTransformer creates a new DeduplicateTransformer
func NewDeduplicateTransformer() *DeduplicateTransformer {
	return &DeduplicateTransformer{}
}

// Transform implements the Transformer interface
func (t *DeduplicateTransformer) Transform(proxies []core.ProxyInterface, global *config.Settings) ([]core.ProxyInterface, error) {
	finalProxies := make([]core.ProxyInterface, 0, len(proxies))
	usedNames := make(map[string]bool)

	for _, p := range proxies {
		originalName := p.GetRemark()

		// Replace '=' with '-' to avoid parse errors in some clients (e.g. Surge)
		originalName = strings.ReplaceAll(originalName, "=", "-")

		// Quote if contains comma
		if strings.Contains(originalName, ",") {
			if !strings.HasPrefix(originalName, "\"") {
				originalName = "\"" + originalName + "\""
			}
		}

		name := originalName
		count := 2
		for usedNames[name] {
			name = fmt.Sprintf("%s %d", originalName, count)
			count++
		}

		// Always update remark if it changed (either by sanitization or deduplication)
		if name != p.GetRemark() {
			p.SetRemark(name)
		}

		usedNames[name] = true
		finalProxies = append(finalProxies, p)
	}
	return finalProxies, nil
}
