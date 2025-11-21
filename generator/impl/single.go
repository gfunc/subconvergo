package impl

import (
	"encoding/base64"
	"log"
	"strings"

	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/generator/core"
	pc "github.com/gfunc/subconvergo/proxy/core"
)

// SingleGenerator implements the Generator interface for single links (ss, ssr, v2ray, trojan)
type SingleGenerator struct {
	Target string
}

func init() {
	core.RegisterGenerator(&SingleGenerator{Target: "ss"})
	core.RegisterGenerator(&SingleGenerator{Target: "ssr"})
	core.RegisterGenerator(&SingleGenerator{Target: "v2ray"})
	core.RegisterGenerator(&SingleGenerator{Target: "trojan"})
}

// Name returns the generator name
func (g *SingleGenerator) Name() string {
	return g.Target
}

// Generate produces the single link subscription
func (g *SingleGenerator) Generate(proxies []pc.ProxyInterface, groups []config.ProxyGroupConfig, rules []string, global *config.Settings, opts core.GeneratorOptions) (string, error) {
	// Generate simple subscription (base64 encoded links)
	var lines []string
	for _, p := range proxies {
		// if SubconverterProxy
		if mixin, ok := p.(pc.SubconverterProxy); ok {

			// Only include proxies matching the requested format
			if g.Target == "v2ray" && (p.GetType() == "vmess" || p.GetType() == "vless") {
				link, err := mixin.ToShareLink(&opts.ProxySetting)
				if err != nil {
					log.Printf("Failed to generate link for proxy: %v", err)
					continue
				}
				lines = append(lines, link)
			} else if g.Target == p.GetType() {
				link, err := mixin.ToShareLink(&opts.ProxySetting)
				if err != nil {
					log.Printf("Failed to generate link for proxy: %v", err)
					continue
				}
				lines = append(lines, link)
			}
		}
	}

	// Base64 encode the entire subscription
	subscription := strings.Join(lines, "\n")
	encoded := base64.StdEncoding.EncodeToString([]byte(subscription))
	return encoded, nil
}
