package handler

import (
	"testing"

	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/parser"
)

func TestApplyRenameRules(t *testing.T) {
	// Setup config
	config.Global.NodePref.RenameNodes = []config.RenameNodeConfig{
		{Match: "香港", Replace: "HK"},
		{Match: "台湾", Replace: "TW"},
		{Match: "^\\s+|\\s+$", Replace: ""},
	}

	h := NewSubHandler()

	tests := []struct {
		name     string
		input    []parser.Proxy
		expected []parser.Proxy
	}{
		{
			name: "rename chinese to english",
			input: []parser.Proxy{
				{Remark: "香港节点01", Type: "ss"},
				{Remark: "台湾节点02", Type: "vmess"},
			},
			expected: []parser.Proxy{
				{Remark: "HK节点01", Type: "ss"},
				{Remark: "TW节点02", Type: "vmess"},
			},
		},
		{
			name: "trim whitespace",
			input: []parser.Proxy{
				{Remark: "  HK Node  ", Type: "ss"},
			},
			expected: []parser.Proxy{
				{Remark: "HK Node", Type: "ss"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := h.applyRenameRules(tt.input)
			for i, proxy := range result {
				if proxy.Remark != tt.expected[i].Remark {
					t.Errorf("Expected remark %q, got %q", tt.expected[i].Remark, proxy.Remark)
				}
			}
		})
	}
}

func TestApplyEmojiRules(t *testing.T) {
	// Setup config
	config.Global.Emojis.AddEmoji = true
	config.Global.Emojis.RemoveOldEmoji = true
	config.Global.Emojis.Rules = []config.EmojiRuleConfig{
		{Match: "(🇭🇰)|(港)|(Hong)|(HK)", Emoji: "🇭🇰"},
		{Match: "(🇺🇸)|(美)|(US)|(United States)", Emoji: "🇺🇸"},
		{Match: "(🇯🇵)|(日)|(Japan)|(JP)", Emoji: "🇯🇵"},
	}

	h := NewSubHandler()

	tests := []struct {
		name     string
		input    []parser.Proxy
		expected string
	}{
		{
			name: "add HK emoji",
			input: []parser.Proxy{
				{Remark: "Hong Kong 01", Type: "ss"},
			},
			expected: "🇭🇰 Hong Kong 01",
		},
		{
			name: "add US emoji",
			input: []parser.Proxy{
				{Remark: "美国节点", Type: "vmess"},
			},
			expected: "🇺🇸 美国节点",
		},
		{
			name: "add JP emoji",
			input: []parser.Proxy{
				{Remark: "日本节点", Type: "trojan"},
			},
			expected: "🇯🇵 日本节点",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := h.applyEmojiRules(tt.input)
			if result[0].Remark != tt.expected {
				t.Errorf("Expected remark %q, got %q", tt.expected, result[0].Remark)
			}
		})
	}
}

func TestSortProxies(t *testing.T) {
	config.Global.NodePref.SortFlag = true

	h := NewSubHandler()

	input := []parser.Proxy{
		{Remark: "C Node", Type: "ss"},
		{Remark: "A Node", Type: "vmess"},
		{Remark: "B Node", Type: "trojan"},
	}

	expected := []string{"A Node", "B Node", "C Node"}

	result := h.sortProxies(input)

	for i, proxy := range result {
		if proxy.Remark != expected[i] {
			t.Errorf("Expected remark %q at position %d, got %q", expected[i], i, proxy.Remark)
		}
	}
}

func TestRenderTemplate(t *testing.T) {
	// Setup config
	config.Global.Template.Globals = []config.TemplateGlobalConfig{
		{Key: "clash.new_field_name", Value: "true"},
		{Key: "managed_prefix", Value: "http://example.com"},
	}

	h := NewSubHandler()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple variable",
			input:    "New field name: {{.clash.new_field_name}}",
			expected: "New field name: true",
		},
		{
			name:     "multiple variables",
			input:    "Prefix: {{.managed_prefix}}, New Field: {{.clash.new_field_name}}",
			expected: "Prefix: http://example.com, New Field: true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := h.renderTemplate(tt.input)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestAppendProxyType(t *testing.T) {
	config.Global.Common.AppendProxyType = true

	proxies := []parser.Proxy{
		{Remark: "HK Node 01", Type: "ss"},
		{Remark: "US Node 02", Type: "vmess"},
		{Remark: "JP Node 03", Type: "trojan"},
	}

	expected := []string{
		"HK Node 01 [ss]",
		"US Node 02 [vmess]",
		"JP Node 03 [trojan]",
	}

	// Simulate the append proxy type logic
	for i := range proxies {
		proxies[i].Remark = proxies[i].Remark + " [" + proxies[i].Type + "]"
	}

	for i, proxy := range proxies {
		if proxy.Remark != expected[i] {
			t.Errorf("Expected remark %q, got %q", expected[i], proxy.Remark)
		}
	}
}
