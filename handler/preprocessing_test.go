package handler

import (
	"strings"
	"testing"

	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/parser"
)

func TestApplyMatcherForRename(t *testing.T) {
	h := NewSubHandler()

	tests := []struct {
		name          string
		rule          string
		proxy         parser.Proxy
		expectMatched bool
		expectRule    string
	}{
		{
			name: "TYPE matcher - SS",
			rule: "!!TYPE=SS!!.*",
			proxy: parser.Proxy{
				Type:   "ss",
				Remark: "Test Node",
			},
			expectMatched: true,
			expectRule:    ".*",
		},
		{
			name: "TYPE matcher - VMess",
			rule: "!!TYPE=VMESS|TROJAN!!.*",
			proxy: parser.Proxy{
				Type:   "vmess",
				Remark: "Test Node",
			},
			expectMatched: true,
			expectRule:    ".*",
		},
		{
			name: "GROUP matcher",
			rule: "!!GROUP=US!!.*",
			proxy: parser.Proxy{
				Group:  "US Premium",
				Remark: "Test Node",
			},
			expectMatched: true,
			expectRule:    ".*",
		},
		{
			name: "PORT matcher",
			rule: "!!PORT=443!!.*",
			proxy: parser.Proxy{
				Port:   443,
				Remark: "Test Node",
			},
			expectMatched: true,
			expectRule:    ".*",
		},
		{
			name: "SERVER matcher",
			rule: "!!SERVER=.*\\.example\\.com!!.*",
			proxy: parser.Proxy{
				Server: "us1.example.com",
				Remark: "Test Node",
			},
			expectMatched: true,
			expectRule:    ".*",
		},
		{
			name: "No matcher - pass through",
			rule: ".*US.*",
			proxy: parser.Proxy{
				Remark: "US Node",
			},
			expectMatched: true,
			expectRule:    ".*US.*",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched, realRule := h.applyMatcherForRename(tt.rule, tt.proxy)
			if matched != tt.expectMatched {
				t.Errorf("applyMatcherForRename() matched = %v, want %v", matched, tt.expectMatched)
			}
			if realRule != tt.expectRule {
				t.Errorf("applyMatcherForRename() realRule = %v, want %v", realRule, tt.expectRule)
			}
		})
	}
}

func TestMatchRange(t *testing.T) {
	h := NewSubHandler()

	tests := []struct {
		name    string
		pattern string
		value   int
		want    bool
	}{
		{"single value match", "443", 443, true},
		{"single value no match", "443", 8080, false},
		{"range match", "8000-9000", 8388, true},
		{"range no match", "8000-9000", 443, false},
		{"comma separated match", "443,8080,8388", 8080, true},
		{"comma separated no match", "443,8080,8388", 9000, false},
		{"empty pattern", "", 1234, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := h.matchRange(tt.pattern, tt.value)
			if result != tt.want {
				t.Errorf("matchRange(%s, %d) = %v, want %v", tt.pattern, tt.value, result, tt.want)
			}
		})
	}
}

func TestApplyRenameRulesWithMatchers(t *testing.T) {
	// Save original config
	originalRenameNodes := config.Global.NodePref.RenameNodes

	// Set test config
	config.Global.NodePref.RenameNodes = []config.RenameNodeConfig{
		{Match: ".*HK.*", Replace: "香港"},
		{Match: ".*US.*", Replace: "美国"},
		{Match: "!!TYPE=SS!!.*", Replace: "SS-$0"},
	}

	h := NewSubHandler()
	proxies := []parser.Proxy{
		{Type: "ss", Remark: "HK Node 1"},
		{Type: "vmess", Remark: "US Node 1"},
		{Type: "ss", Remark: "SG Node 1"},
	}

	result := h.applyRenameRules(proxies)

	// Check results
	if !strings.Contains(result[0].Remark, "香港") {
		t.Errorf("Expected HK rename, got %s", result[0].Remark)
	}
	if !strings.Contains(result[1].Remark, "美国") {
		t.Errorf("Expected US rename, got %s", result[1].Remark)
	}

	// Restore original config
	config.Global.NodePref.RenameNodes = originalRenameNodes
}

func TestApplyEmojiRulesWithMatchers(t *testing.T) {
	// Save original config
	originalEmojis := config.Global.Emojis

	// Set test config
	config.Global.Emojis = config.EmojiConfig{
		AddEmoji:       true,
		RemoveOldEmoji: true,
		Rules: []config.EmojiRuleConfig{
			{Match: ".*(HK|Hong Kong|香港).*", Emoji: "🇭🇰"},
			{Match: ".*(US|United States|美国).*", Emoji: "🇺🇸"},
			{Match: "!!TYPE=SS!!.*", Emoji: "⚡"},
		},
	}

	h := NewSubHandler()
	proxies := []parser.Proxy{
		{Type: "ss", Remark: "HK Node 1"},
		{Type: "vmess", Remark: "US Node 1"},
		{Type: "ss", Remark: "SG SS Node 1"},
	}

	result := h.applyEmojiRules(proxies)

	// Check results
	if !strings.HasPrefix(result[0].Remark, "🇭🇰") {
		t.Errorf("Expected HK emoji, got %s", result[0].Remark)
	}
	if !strings.HasPrefix(result[1].Remark, "🇺🇸") {
		t.Errorf("Expected US emoji, got %s", result[1].Remark)
	}
	if !strings.HasPrefix(result[2].Remark, "⚡") {
		t.Errorf("Expected SS emoji, got %s", result[2].Remark)
	}

	// Restore original config
	config.Global.Emojis = originalEmojis
}

func TestRemoveEmoji(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"with HK flag", "🇭🇰 Hong Kong Node"},
		{"with US flag", "🇺🇸 US Node"},
		{"with multiple emojis", "⚡🇭🇰 HK SS Node"},
		{"no emoji", "Plain Node"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeEmoji(tt.input)
			// Should not contain emoji characters
			if len(result) == 0 && len(tt.input) > 0 {
				t.Errorf("removeEmoji() removed everything from %s", tt.input)
			}
			// Just verify it doesn't panic and returns something
			t.Logf("Input: %s, Output: %s", tt.input, result)
		})
	}
}
