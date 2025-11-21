package transformers

import (
	"strings"
	"testing"

	"github.com/gfunc/subconvergo/config"
	pc "github.com/gfunc/subconvergo/proxy/core"
)

func TestRenameTransformer(t *testing.T) {
	rules := []config.RenameNodeConfig{
		{Match: "HK", Replace: "Hong Kong"},
		{Match: "US", Replace: "United States"},
		{Match: "!!TYPE=SS!!test", Replace: "Shadowsocks"},
	}

	transformer := NewRenameTransformer(rules)

	proxies := []pc.ProxyInterface{
		&pc.BaseProxy{Remark: "HK Node"},
		&pc.BaseProxy{Remark: "US Server"},
		&pc.BaseProxy{Remark: "test", Type: "ss"},
	}

	result, err := transformer.Transform(proxies, config.Global)
	if err != nil {
		t.Fatalf("Transform failed: %v", err)
	}

	if len(result) != 3 {
		t.Errorf("Expected 3 proxies, got %d", len(result))
	}

	if !strings.Contains(result[0].GetRemark(), "Hong Kong") {
		t.Errorf("Expected 'Hong Kong' in remark, got '%s'", result[0].GetRemark())
	}
	if !strings.Contains(result[1].GetRemark(), "United States") {
		t.Errorf("Expected 'United States' in remark, got '%s'", result[1].GetRemark())
	}
	if result[2].GetRemark() != "Shadowsocks" {
		t.Errorf("Expected 'Shadowsocks' in remark, got '%s'", result[2].GetRemark())
	}
}

func TestEmojiTransformer(t *testing.T) {
	cfg := config.EmojiConfig{
		AddEmoji:       true,
		RemoveOldEmoji: true,
		Rules: []config.EmojiRuleConfig{
			{Match: "US|America", Emoji: "🇺🇸"},
			{Match: "HK|Hong", Emoji: "🇭🇰"},
		},
	}

	transformer := NewEmojiTransformer(cfg)

	proxies := []pc.ProxyInterface{
		&pc.BaseProxy{Remark: "US Node"},
		&pc.BaseProxy{Remark: "HK Server"},
		&pc.BaseProxy{Remark: "🇺🇸 Existing Emoji"},
	}

	result, err := transformer.Transform(proxies, config.Global)
	if err != nil {
		t.Fatalf("Transform failed: %v", err)
	}

	if len(result) != 3 {
		t.Errorf("Expected 3 proxies, got %d", len(result))
	}

	if !strings.Contains(result[0].GetRemark(), "🇺🇸") {
		t.Errorf("Expected emoji 🇺🇸 in remark, got '%s'", result[0].GetRemark())
	}
	if !strings.Contains(result[1].GetRemark(), "🇭🇰") {
		t.Errorf("Expected emoji 🇭🇰 in remark, got '%s'", result[1].GetRemark())
	}
	// Check if old emoji was removed and new one added (or kept if rule matches)
	// "🇺🇸 Existing Emoji" -> remove -> "Existing Emoji" -> match? No match in rules for "Existing Emoji"
	// So it should just have emoji removed.
	if strings.Contains(result[2].GetRemark(), "🇺🇸") {
		t.Errorf("Expected old emoji removed, got '%s'", result[2].GetRemark())
	}
}

func TestSortTransformer(t *testing.T) {
	transformer := NewSortTransformer(true)

	proxies := []pc.ProxyInterface{
		&pc.BaseProxy{Remark: "Z"},
		&pc.BaseProxy{Remark: "A"},
		&pc.BaseProxy{Remark: "M"},
	}

	result, err := transformer.Transform(proxies, config.Global)
	if err != nil {
		t.Fatalf("Transform failed: %v", err)
	}

	if result[0].GetRemark() != "A" {
		t.Errorf("Expected first proxy to be A, got %s", result[0].GetRemark())
	}
	if result[1].GetRemark() != "M" {
		t.Errorf("Expected second proxy to be M, got %s", result[1].GetRemark())
	}
	if result[2].GetRemark() != "Z" {
		t.Errorf("Expected third proxy to be Z, got %s", result[2].GetRemark())
	}
}

func TestFilterTransformer(t *testing.T) {
	proxies := []pc.ProxyInterface{
		&pc.BaseProxy{Remark: "HK Node"},
		&pc.BaseProxy{Remark: "US Node"},
	}
	transformer := NewFilterTransformer([]string{"HK"}, nil)
	result, err := transformer.Transform(proxies, config.Global)
	if err != nil {
		t.Fatalf("Transform failed: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("Expected 1 proxy, got %d", len(result))
	}
}

func TestFilterTransformerRegexInclude(t *testing.T) {
	proxies := []pc.ProxyInterface{
		&pc.BaseProxy{Remark: "HK Node"},
		&pc.BaseProxy{Remark: "US Node"},
		&pc.BaseProxy{Remark: "HK Server"},
	}
	// Include remarks that start with HK using regex
	transformer := NewFilterTransformer([]string{"/^HK/"}, nil)
	result, err := transformer.Transform(proxies, config.Global)
	if err != nil {
		t.Fatalf("Transform failed: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 proxies starting with HK, got %d", len(result))
	}
	for _, p := range result {
		if !strings.HasPrefix(p.GetRemark(), "HK") {
			t.Fatalf("unexpected remark in regex include: %s", p.GetRemark())
		}
	}
}

func TestFilterTransformerRegexExclude(t *testing.T) {
	proxies := []pc.ProxyInterface{
		&pc.BaseProxy{Remark: "HK Node"},
		&pc.BaseProxy{Remark: "US Node"},
		&pc.BaseProxy{Remark: "JP Node"},
	}
	// Exclude remarks matching US or JP
	transformer := NewFilterTransformer(nil, []string{"/(US|JP)/"})
	result, err := transformer.Transform(proxies, config.Global)
	if err != nil {
		t.Fatalf("Transform failed: %v", err)
	}
	if len(result) != 1 || !strings.Contains(result[0].GetRemark(), "HK") {
		t.Fatalf("expected only HK to remain, got %v", result)
	}
}

func TestFilterTransformerWithIncludeExclude(t *testing.T) {
	proxies := []pc.ProxyInterface{
		&pc.BaseProxy{Remark: "HK Node"},
		&pc.BaseProxy{Remark: "US Node"},
		&pc.BaseProxy{Remark: "JP Node"},
	}

	transformer := NewFilterTransformer([]string{"HK", "US"}, nil)
	result, err := transformer.Transform(proxies, config.Global)
	if err != nil {
		t.Fatalf("Transform failed: %v", err)
	}
	if len(result) == 0 {
		t.Error("Filter should not empty all proxies with includes")
	}

	transformer = NewFilterTransformer(nil, []string{"HK"})
	result, err = transformer.Transform(proxies, config.Global)
	if err != nil {
		t.Fatalf("Transform failed: %v", err)
	}
	if len(result) == 0 {
		t.Error("Exclude should keep some proxies")
	}
}
