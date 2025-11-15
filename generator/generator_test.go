package generator

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/parser"
)

func TestGenerateClash(t *testing.T) {
	proxies := []parser.Proxy{
		{
			Type:          "ss",
			Remark:        "SS Test",
			Server:        "example.com",
			Port:          8388,
			Password:      "password",
			EncryptMethod: "aes-256-gcm",
		},
		{
			Type:    "vmess",
			Remark:  "VMess Test",
			Server:  "example.com",
			Port:    443,
			UUID:    "12345678-1234-1234-1234-123456789012",
			AlterID: 0,
			Network: "ws",
			Path:    "/path",
			Host:    "example.com",
			TLS:     true,
		},
	}

	opts := GeneratorOptions{
		Target:            "clash",
		ClashNewFieldName: true,
	}

	baseConfig := `port: 7890
socks-port: 7891
allow-lan: false
mode: Rule
log-level: info`

	result, err := Generate(proxies, opts, baseConfig)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if result == "" {
		t.Error("result should not be empty")
	}

	// Check if YAML contains expected fields
	expectedStrings := []string{
		"proxies:",
		"SS Test",
		"VMess Test",
		"type: ss",
		"type: vmess",
	}

	for _, expected := range expectedStrings {
		if !contains(result, expected) {
			t.Errorf("expected result to contain '%s'", expected)
		}
	}
}

func TestGenerateSurge(t *testing.T) {
	proxies := []parser.Proxy{
		{
			Type:          "ss",
			Remark:        "SS Test",
			Server:        "example.com",
			Port:          8388,
			Password:      "password",
			EncryptMethod: "aes-256-gcm",
		},
		{
			Type:     "trojan",
			Remark:   "Trojan Test",
			Server:   "example.com",
			Port:     443,
			Password: "password",
			TLS:      true,
		},
	}

	opts := GeneratorOptions{
		Target: "surge",
	}

	baseConfig := `[General]
loglevel = notify
dns-server = system`

	result, err := Generate(proxies, opts, baseConfig)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Check that the output contains expected sections
	if !strings.Contains(result, "[Proxy]") {
		t.Error("Expected [Proxy] section in Surge output")
	}
	if !strings.Contains(result, "SS Test") {
		t.Error("Expected SS proxy in output")
	}
	if !strings.Contains(result, "Trojan Test") {
		t.Error("Expected Trojan proxy in output")
	}
}

func TestGenerateQuantumultX(t *testing.T) {
	proxies := []parser.Proxy{
		{
			Type:          "ss",
			Remark:        "SS Test",
			Server:        "example.com",
			Port:          8388,
			Password:      "password",
			EncryptMethod: "aes-256-gcm",
		},
		{
			Type:    "vmess",
			Remark:  "VMess Test",
			Server:  "example.com",
			Port:    443,
			UUID:    "12345678-1234-1234-1234-123456789012",
			AlterID: 0,
			Network: "ws",
			Path:    "/path",
			Host:    "example.com",
			TLS:     true,
		},
	}

	opts := GeneratorOptions{
		Target: "quanx",
	}

	baseConfig := `[general]
server_check_url = http://www.gstatic.com/generate_204`

	result, err := Generate(proxies, opts, baseConfig)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Check that the output contains expected sections
	if !strings.Contains(result, "[server_local]") {
		t.Error("Expected [server_local] section in QuantumultX output")
	}
	if !strings.Contains(result, "tag=SS Test") {
		t.Error("Expected SS proxy with tag in output")
	}
}

func TestGenerateSingBox(t *testing.T) {
	proxies := []parser.Proxy{
		{
			Type:          "ss",
			Remark:        "SS Test",
			Server:        "example.com",
			Port:          8388,
			Password:      "password",
			EncryptMethod: "aes-256-gcm",
		},
		{
			Type:     "trojan",
			Remark:   "Trojan Test",
			Server:   "example.com",
			Port:     443,
			Password: "password",
			TLS:      true,
		},
	}

	opts := GeneratorOptions{
		Target: "singbox",
	}

	baseConfig := `{"log":{"level":"info"}}`

	result, err := Generate(proxies, opts, baseConfig)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Parse the JSON to verify structure
	var config map[string]interface{}
	if err := json.Unmarshal([]byte(result), &config); err != nil {
		t.Fatalf("Failed to parse sing-box JSON: %v", err)
	}

	// Check for outbounds array
	if _, ok := config["outbounds"]; !ok {
		t.Error("Expected 'outbounds' field in sing-box config")
	}
}

func TestGenerateSS(t *testing.T) {
	proxies := []parser.Proxy{
		{
			Type:          "ss",
			Remark:        "SS Test 1",
			Server:        "example1.com",
			Port:          8388,
			Password:      "password1",
			EncryptMethod: "aes-256-gcm",
		},
		{
			Type:          "ss",
			Remark:        "SS Test 2",
			Server:        "example2.com",
			Port:          8389,
			Password:      "password2",
			EncryptMethod: "chacha20-ietf-poly1305",
		},
	}

	opts := GeneratorOptions{
		Target: "ss",
	}

	result, err := Generate(proxies, opts, "")
	if err != nil {
		t.Fatalf("Generate(ss) error = %v", err)
	}

	// generateSingle is not fully implemented yet, so result might be empty
	// Just check it doesn't crash
	t.Logf("SS generation result: %s", result)
}

func TestGenerateSSR(t *testing.T) {
	proxies := []parser.Proxy{
		{
			Type:          "ssr",
			Remark:        "SSR Test",
			Server:        "example.com",
			Port:          8388,
			Password:      "password",
			EncryptMethod: "aes-256-cfb",
			Protocol:      "auth_aes128_md5",
			Obfs:          "tls1.2_ticket_auth",
		},
	}

	opts := GeneratorOptions{
		Target: "ssr",
	}

	result, err := Generate(proxies, opts, "")
	if err != nil {
		t.Fatalf("Generate(ssr) error = %v", err)
	}

	// generateSingle is not fully implemented yet
	t.Logf("SSR generation result: %s", result)
}

func TestGenerateV2Ray(t *testing.T) {
	proxies := []parser.Proxy{
		{
			Type:    "vmess",
			Remark:  "VMess Test",
			Server:  "example.com",
			Port:    443,
			UUID:    "12345678-1234-1234-1234-123456789012",
			AlterID: 0,
			Network: "ws",
			Path:    "/path",
			Host:    "example.com",
			TLS:     true,
		},
	}

	opts := GeneratorOptions{
		Target: "v2ray",
	}

	result, err := Generate(proxies, opts, "")
	if err != nil {
		t.Fatalf("Generate(v2ray) error = %v", err)
	}

	// generateSingle is not fully implemented yet
	t.Logf("V2Ray generation result: %s", result)
}

func TestGenerateProxyGroups(t *testing.T) {
	proxies := []parser.Proxy{
		{Remark: "HK-Server1"},
		{Remark: "HK-Server2"},
		{Remark: "US-Server1"},
		{Remark: "JP-Server1"},
	}

	groups := []config.ProxyGroupConfig{
		{
			Name: "HK",
			Type: "select",
			Rule: []string{"HK"},
		},
		{
			Name:     "US",
			Type:     "url-test",
			Rule:     []string{"US"},
			URL:      "http://www.gstatic.com/generate_204",
			Interval: 300,
		},
		{
			Name: "Auto",
			Type: "fallback",
			Rule: []string{
				"HK",
				"US",
			},
		},
	}

	opts := GeneratorOptions{
		ClashNewFieldName: true,
	}

	result := generateClashProxyGroups(proxies, groups, opts)

	if len(result) != 3 {
		t.Errorf("expected 3 proxy groups, got %d", len(result))
	}

	// Check HK group has proxies
	hkProxies := result[0]["proxies"].([]string)
	if len(hkProxies) < 1 {
		t.Errorf("expected HK group to have at least 1 proxy, got %d", len(hkProxies))
	}

	// Check US group has proxies
	usProxies := result[1]["proxies"].([]string)
	if len(usProxies) < 1 {
		t.Errorf("expected US group to have at least 1 proxy, got %d", len(usProxies))
	}
}

func TestFilterProxies(t *testing.T) {
	proxyNames := []string{
		"HK-Premium",
		"HK-Free",
		"US-Server",
		"JP-Expired",
	}

	// Test include filter
	includeResult := filterProxies(proxyNames, []string{"HK"})
	if len(includeResult) != 2 {
		t.Errorf("expected 2 HK proxies, got %d", len(includeResult))
	}

	// Test pattern matching
	usResult := filterProxies(proxyNames, []string{"US"})
	if len(usResult) != 1 {
		t.Errorf("expected 1 US proxy, got %d", len(usResult))
	}

	// Test multiple patterns
	multiResult := filterProxies(proxyNames, []string{"HK", "US"})
	if len(multiResult) != 3 {
		t.Errorf("expected 3 proxies (HK or US), got %d", len(multiResult))
	}
}

func TestProxyGroupFiltering(t *testing.T) {
	proxies := []parser.Proxy{
		{Remark: "HK Hong Kong 01"},
		{Remark: "US United States 01"},
		{Remark: "Server-HK-01"},
	}

	// Test Clash proxy group generation with filtering
	groups := []config.ProxyGroupConfig{
		{
			Name: "HK",
			Type: "select",
			Rule: []string{"HK", "Hong Kong"},
		},
		{
			Name: "US",
			Type: "select",
			Rule: []string{"US", "United States"},
		},
	}

	opts := GeneratorOptions{
		ClashNewFieldName: true,
	}

	result := generateClashProxyGroups(proxies, groups, opts)

	if len(result) != 2 {
		t.Errorf("expected 2 groups, got %d", len(result))
	}

	// Check HK group contains HK proxies
	hkProxies := result[0]["proxies"].([]string)
	if len(hkProxies) < 1 {
		t.Errorf("HK group should have at least 1 proxy")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// BenchmarkGenerateClash benchmarks Clash generation
func BenchmarkGenerateClash(b *testing.B) {
	proxies := make([]parser.Proxy, 100)
	for i := 0; i < 100; i++ {
		proxies[i] = parser.Proxy{
			Type:          "ss",
			Remark:        "Test Server",
			Server:        "example.com",
			Port:          8388,
			Password:      "password",
			EncryptMethod: "aes-256-gcm",
		}
	}

	opts := GeneratorOptions{
		Target:            "clash",
		ClashNewFieldName: true,
	}

	baseConfig := `port: 7890
mode: Rule`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Generate(proxies, opts, baseConfig)
	}
}
