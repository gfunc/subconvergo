package generator

import (
	"testing"

	"github.com/gfunc/subconvergo/generator/core"
	_ "github.com/gfunc/subconvergo/generator/impl"
	"github.com/gfunc/subconvergo/generator/utils"
	pc "github.com/gfunc/subconvergo/proxy/core"
	"github.com/gfunc/subconvergo/proxy/impl"
)

func TestApplyMatcher(t *testing.T) {
	tests := []struct {
		name     string
		rule     string
		proxy    pc.ProxyInterface
		expected bool
		realRule string
	}{
		{
			name: "GROUP matcher - match",
			rule: "!!GROUP=US!!.*",
			proxy: &pc.BaseProxy{
				Remark: "US Node 1",
				Group:  "US Premium",
			},
			expected: true,
			realRule: ".*",
		},
		{
			name: "GROUP matcher - no match",
			rule: "!!GROUP=HK!!.*",
			proxy: &pc.BaseProxy{
				Remark: "US Node 1",
				Group:  "US Premium",
			},
			expected: false,
			realRule: ".*",
		},
		{
			name: "TYPE matcher - shadowsocks",
			rule: "!!TYPE=SS|VMess!!.*",
			proxy: &pc.BaseProxy{
				Remark: "Test Node",
				Type:   "ss",
			},
			expected: true,
			realRule: ".*",
		},
		{
			name: "PORT matcher - range",
			rule: "!!PORT=443!!.*",
			proxy: &pc.BaseProxy{
				Remark: "Test Node",
				Port:   443,
			},
			expected: true,
			realRule: ".*",
		},
		{
			name: "SERVER matcher",
			rule: "!!SERVER=example\\.com!!.*",
			proxy: &pc.BaseProxy{
				Remark: "Test Node",
				Server: "example.com",
			},
			expected: true,
			realRule: ".*",
		},
		{
			name: "No matcher - pass through",
			rule: "US.*",
			proxy: &pc.BaseProxy{
				Remark: "US Node 1",
			},
			expected: true,
			realRule: "US.*",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched, realRule := utils.ApplyMatcher(tt.rule, tt.proxy)
			if matched != tt.expected {
				t.Errorf("ApplyMatcher() matched = %v, want %v", matched, tt.expected)
			}
			if realRule != tt.realRule {
				t.Errorf("ApplyMatcher() realRule = %v, want %v", realRule, tt.realRule)
			}
		})
	}
}

func TestMatchRange(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		value    int
		expected bool
	}{
		{"single value match", "443", 443, true},
		{"single value no match", "443", 8080, false},
		{"range match", "8000-9000", 8388, true},
		{"range no match", "8000-9000", 443, false},
		{"comma separated match", "443,8080,8388", 8080, true},
		{"comma separated no match", "443,8080,8388", 9000, false},
		{"complex range", "1-100,443,8000-9000", 8388, true},
		{"empty pattern", "", 1234, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.MatchRange(tt.pattern, tt.value)
			if result != tt.expected {
				t.Errorf("MatchRange(%s, %d) = %v, want %v", tt.pattern, tt.value, result, tt.expected)
			}
		})
	}
}

func TestFilterProxiesByRules(t *testing.T) {
	proxies := []pc.ProxyInterface{
		&pc.BaseProxy{Type: "ss", Remark: "US Node 1", Server: "us1.example.com", Port: 443, Group: "US"},
		&pc.BaseProxy{Type: "ss", Remark: "US Node 2", Server: "us2.example.com", Port: 8080, Group: "US"},
		&pc.BaseProxy{Type: "vmess", Remark: "HK Node 1", Server: "hk1.example.com", Port: 443, Group: "HK"},
		&pc.BaseProxy{Type: "trojan", Remark: "JP Node 1", Server: "jp1.example.com", Port: 443, Group: "JP"},
		&pc.BaseProxy{Type: "trojan", Remark: "SG Node 1", Server: "sg1.example.com", Port: 8388, Group: "SG"},
	}

	tests := []struct {
		name     string
		rules    []string
		expected []string
	}{
		{
			name:     "Filter by type SS",
			rules:    []string{"!!TYPE=SS"},
			expected: []string{"US Node 1", "US Node 2"},
		},
		{
			name:     "Filter by type VMess or Trojan",
			rules:    []string{"!!TYPE=VMESS|TROJAN"},
			expected: []string{"HK Node 1", "JP Node 1", "SG Node 1"},
		},
		{
			name:     "Filter by port 443",
			rules:    []string{"!!PORT=443"},
			expected: []string{"US Node 1", "HK Node 1", "JP Node 1"},
		},
		{
			name:     "Filter by group US",
			rules:    []string{"!!GROUP=US"},
			expected: []string{"US Node 1", "US Node 2"},
		},
		{
			name:     "Filter by regex pattern",
			rules:    []string{"HK.*"},
			expected: []string{"HK Node 1"},
		},
		{
			name:     "Filter with TYPE and regex",
			rules:    []string{"!!TYPE=SS!!US.*"},
			expected: []string{"US Node 1", "US Node 2"},
		},
		{
			name:     "Direct inclusion",
			rules:    []string{"[]DIRECT", "[]REJECT"},
			expected: []string{"DIRECT", "REJECT"},
		},
		{
			name:     "Multiple rules",
			rules:    []string{"!!TYPE=SS", "!!TYPE=VMESS"},
			expected: []string{"US Node 1", "US Node 2", "HK Node 1"},
		},
		{
			name:     "Server pattern",
			rules:    []string{"!!SERVER=.*\\.example\\.com"},
			expected: []string{"US Node 1", "US Node 2", "HK Node 1", "JP Node 1", "SG Node 1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.FilterProxiesByRules(proxies, tt.rules)
			if len(result) != len(tt.expected) {
				t.Errorf("FilterProxiesByRules() returned %d proxies, want %d", len(result), len(tt.expected))
				t.Logf("Got: %v", result)
				t.Logf("Want: %v", tt.expected)
				return
			}
			for i, name := range result {
				if name != tt.expected[i] {
					t.Errorf("FilterProxiesByRules()[%d] = %v, want %v", i, name, tt.expected[i])
				}
			}
		})
	}
}

func TestGenerate(t *testing.T) {
	proxies := []pc.ProxyInterface{
		&impl.ShadowsocksProxy{BaseProxy: pc.BaseProxy{Type: "ss", Remark: "SS1", Server: "ss.com", Port: 443}, EncryptMethod: "aes-256-gcm", Password: "pass"},
		&impl.VMessProxy{BaseProxy: pc.BaseProxy{Type: "vmess", Remark: "VM1", Server: "vm.com", Port: 443}, UUID: "uuid"},
		&impl.TrojanProxy{BaseProxy: pc.BaseProxy{Type: "trojan", Remark: "TJ1", Server: "tj.com", Port: 443}, Password: "pass"},
	}

	tests := []struct {
		name   string
		target string
		base   string
	}{
		{"Clash", "clash", "proxies: []\nrules: []"},
		{"Surge", "surge", "[General]\n"},
		{"Loon", "loon", "[General]\n"},
		{"QuantumultX", "quanx", "[general]\n"},
		{"SingBox", "singbox", `{"outbounds":[],"route":{}}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := core.GeneratorOptions{Target: tt.target}
			_, err := Generate(proxies, opts, tt.base)
			if err != nil {
				t.Errorf("Generate(%s) failed: %v", tt.target, err)
			}
		})
	}
}
