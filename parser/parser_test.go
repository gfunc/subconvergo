package parser

import (
	"strings"
	"testing"

	P "github.com/gfunc/subconvergo/proxy/impl"
	"github.com/stretchr/testify/assert"
)

// TestProxyInterfaceGenerateLink tests that ToShareLink works correctly for all proxy types
func TestProxyInterfaceGenerateLink(t *testing.T) {
	tests := []struct {
		name         string
		inputLink    string
		expectedType string
	}{
		{
			name:         "Shadowsocks proxy",
			inputLink:    "ss://YWVzLTI1Ni1nY206cGFzc3dvcmQ=@example.com:8388#Test%20SS",
			expectedType: "ss",
		},
		{
			name:         "VMess proxy",
			inputLink:    "vmess://eyJ2IjoiMiIsInBzIjoidGVzdC12bWVzcyIsImFkZCI6ImV4YW1wbGUuY29tIiwicG9ydCI6IjQ0MyIsImlkIjoiMTIzNDU2NzgtMTIzNC0xMjM0LTEyMzQtMTIzNDU2Nzg5MDEyIiwiYWlkIjoiMCIsIm5ldCI6InRjcCIsInR5cGUiOiJub25lIiwiaG9zdCI6IiIsInBhdGgiOiIiLCJ0bHMiOiIifQ==",
			expectedType: "vmess",
		},
		{
			name:         "Trojan proxy",
			inputLink:    "trojan://password@example.com:443#Test%20Trojan",
			expectedType: "trojan",
		},
		{
			name:         "VLESS proxy",
			inputLink:    "vless://12345678-1234-1234-1234-123456789012@example.com:443?type=tcp#Test%20VLESS",
			expectedType: "vless",
		},
		{
			name:         "Hysteria2 proxy",
			inputLink:    "hysteria2://password@example.com:443#Test%20Hysteria2",
			expectedType: "hysteria2",
		},
		{
			name:         "TUIC proxy",
			inputLink:    "tuic://12345678-1234-1234-1234-123456789012:password@example.com:443#Test%20TUIC",
			expectedType: "tuic",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the link using the OOP interface
			proxyInterface, err := ParseProxyLine(tt.inputLink)
			if err != nil {
				t.Fatalf("ParseProxyLine() error = %v", err)
			}

			// Verify the type
			if proxyInterface.GetType() != tt.expectedType {
				t.Errorf("GetType() = %v, want %v", proxyInterface.GetType(), tt.expectedType)
			}

			// Generate a new link
			generatedLink, _ := proxyInterface.ToShareLink(nil)

			if generatedLink == "" {
				t.Errorf("ToShareLink() returned empty string")
			}

			// Parse the generated link back and verify it matches
			proxyInterface2, err := ParseProxyLine(generatedLink)
			if err != nil {
				t.Fatalf("ParseProxyLine() of generated link error = %v", err)
			}

			// Verify key fields match
			if proxyInterface2.GetType() != tt.expectedType {
				t.Errorf("Re-parsed type = %v, want %v", proxyInterface2.GetType(), tt.expectedType)
			}
			if proxyInterface2.GetServer() != proxyInterface.GetServer() {
				t.Errorf("Re-parsed server = %v, want %v", proxyInterface2.GetServer(), proxyInterface.GetServer())
			}
			if proxyInterface2.GetPort() != proxyInterface.GetPort() {
				t.Errorf("Re-parsed port = %v, want %v", proxyInterface2.GetPort(), proxyInterface.GetPort())
			}
		})
	}
}

// TestProxyInterfaceSetters tests setter methods
func TestProxyInterfaceSetters(t *testing.T) {
	inputLink := "ss://YWVzLTI1Ni1nY206cGFzc3dvcmQ=@example.com:8388#Test"

	proxy, err := ParseProxyLine(inputLink)
	if err != nil {
		t.Fatalf("ParseProxyLine() error = %v", err)
	}

	// Test SetRemark
	newRemark := "New Remark"
	proxy.SetRemark(newRemark)
	if proxy.GetRemark() != newRemark {
		t.Errorf("After SetRemark(), GetRemark() = %v, want %v", proxy.GetRemark(), newRemark)
	}

	// Test SetGroup
	newGroup := "New Group"
	proxy.SetGroup(newGroup)
	if proxy.GetGroup() != newGroup {
		t.Errorf("After SetGroup(), GetGroup() = %v, want %v", proxy.GetGroup(), newGroup)
	}
}

// TestShadowsocksProxyGenerateLink tests Shadowsocks link generation
func TestShadowsocksProxyGenerateLink(t *testing.T) {
	tests := []struct {
		name      string
		inputLink string
	}{
		{
			name:      "SS with plugin",
			inputLink: "ss://YWVzLTI1Ni1nY206cGFzc3dvcmQ=@example.com:8388?plugin=obfs-local%3Bobfs%3Dhttp%3Bobfs-host%3Dexample.com#Test",
		},
		{
			name:      "SS basic",
			inputLink: "ss://YWVzLTI1Ni1nY206cGFzc3dvcmQ=@example.com:8388#Test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proxy, err := ParseProxyLine(tt.inputLink)
			if err != nil {
				t.Fatalf("ParseProxyLine error = %v", err)
			}

			link, _ := proxy.ToShareLink(nil)

			if !strings.HasPrefix(link, "ss://") {
				t.Errorf("ToShareLink() should start with ss://, got %v", link)
			}

			// Verify it can be parsed back
			_, err = ParseProxyLine(link)
			if err != nil {
				t.Errorf("Generated link cannot be parsed back: %v", err)
			}
		})
	}
}

// TestVMessProxyGenerateLink tests VMess link generation
func TestVMessProxyGenerateLink(t *testing.T) {
	inputLink := "vmess://eyJ2IjoiMiIsInBzIjoidGVzdCIsImFkZCI6ImV4YW1wbGUuY29tIiwicG9ydCI6IjQ0MyIsImlkIjoiMTIzNDU2NzgtMTIzNC0xMjM0LTEyMzQtMTIzNDU2Nzg5MDEyIiwiYWlkIjoiMCIsIm5ldCI6IndzIiwicGF0aCI6Ii9wYXRoIiwiaG9zdCI6ImV4YW1wbGUuY29tIiwidGxzIjoidGxzIn0="

	proxy, err := ParseProxyLine(inputLink)
	if err != nil {
		t.Fatalf("ParseProxyLine error = %v", err)
	}

	link, _ := proxy.ToShareLink(nil)

	if !strings.HasPrefix(link, "vmess://") {
		t.Errorf("ToShareLink() should start with vmess://, got %v", link)
	}

	// Verify it can be parsed back
	proxy2, err := ParseProxyLine(link)
	if err != nil {
		t.Errorf("Generated link cannot be parsed back: %v", err)
	}

	// Type assertion to check fields
	proxyVMess, ok := proxy.(*P.VMessProxy)
	if !ok {
		// Try MihomoProxy if it wraps it
		if mp, ok := proxy.(*P.MihomoProxy); ok {
			proxyVMess, ok = mp.ProxyInterface.(*P.VMessProxy)
			if !ok {
				t.Fatalf("Failed to get VMess proxy from MihomoProxy: %v", err)
			}
		} else {
			t.Fatalf("proxy is not of type *VMessProxy")
		}
	}

	proxy2VMess, ok := proxy2.(*P.VMessProxy)
	if !ok {
		if mp, ok := proxy2.(*P.MihomoProxy); ok {
			proxy2VMess, ok = mp.ProxyInterface.(*P.VMessProxy)
			if !ok {
				t.Fatalf("Failed to get VMess proxy from MihomoProxy: %v", err)
			}
		} else {
			t.Fatalf("proxy2 is not of type *VMessProxy")
		}
	}

	// Verify critical fields match
	if proxy2VMess.UUID != proxyVMess.UUID {
		t.Errorf("Re-parsed UUID = %v, want %v", proxy2VMess.UUID, proxyVMess.UUID)
	}
	if proxy2VMess.Network != proxyVMess.Network {
		t.Errorf("Re-parsed Network = %v, want %v", proxy2VMess.Network, proxyVMess.Network)
	}
}

// TestTrojanProxyGenerateLink tests Trojan link generation
func TestTrojanProxyGenerateLink(t *testing.T) {
	tests := []struct {
		name      string
		inputLink string
	}{
		{
			name:      "Trojan basic",
			inputLink: "trojan://password@example.com:443#Test",
		},
		{
			name:      "Trojan with WebSocket",
			inputLink: "trojan://password@example.com:443?type=ws&path=%2Fpath#Test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proxy, err := ParseProxyLine(tt.inputLink)
			if err != nil {
				t.Fatalf("ParseProxyLine error = %v", err)
			}

			link, _ := proxy.ToShareLink(nil)

			if !strings.HasPrefix(link, "trojan://") {
				t.Errorf("ToShareLink() should start with trojan://, got %v", link)
			}

			// Verify it can be parsed back
			_, err = ParseProxyLine(link)
			if err != nil {
				t.Errorf("Generated link cannot be parsed back: %v", err)
			}
		})
	}
}

// TestVLESSProxyGenerateLink tests VLESS link generation
func TestVLESSProxyGenerateLink(t *testing.T) {
	tests := []struct {
		name      string
		inputLink string
	}{
		{
			name:      "VLESS basic",
			inputLink: "vless://12345678-1234-1234-1234-123456789012@example.com:443?type=tcp#Test",
		},
		{
			name:      "VLESS with WebSocket",
			inputLink: "vless://12345678-1234-1234-1234-123456789012@example.com:443?type=ws&path=%2Fpath&security=tls#Test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proxy, err := ParseProxyLine(tt.inputLink)
			if err != nil {
				t.Fatalf("ParseProxyLine error = %v", err)
			}

			link, _ := proxy.ToShareLink(nil)

			if !strings.HasPrefix(link, "vless://") {
				t.Errorf("ToShareLink() should start with vless://, got %v", link)
			}

			// Verify it can be parsed back
			_, err = ParseProxyLine(link)
			if err != nil {
				t.Errorf("Generated link cannot be parsed back: %v", err)
			}
		})
	}
}

// TestHysteriaProxyGenerateLink tests Hysteria link generation
func TestHysteriaProxyGenerateLink(t *testing.T) {
	tests := []struct {
		name      string
		inputLink string
		protocol  string
	}{
		{
			name:      "Hysteria v1",
			inputLink: "hysteria://password@example.com:443?upmbps=10&downmbps=50#Test",
			protocol:  "hysteria",
		},
		{
			name:      "Hysteria2",
			inputLink: "hysteria2://password@example.com:443#Test",
			protocol:  "hysteria2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proxy, err := ParseProxyLine(tt.inputLink)
			if err != nil {
				t.Fatalf("ParseProxyLine error = %v", err)
			}

			link, _ := proxy.ToShareLink(nil)

			if !strings.HasPrefix(link, tt.protocol+"://") {
				t.Errorf("ToShareLink() should start with %s://, got %v", tt.protocol, link)
			}

			// Verify it can be parsed back
			_, err = ParseProxyLine(link)
			if err != nil {
				t.Errorf("Generated link cannot be parsed back: %v", err)
			}
		})
	}
}

// TestTUICProxyGenerateLink tests TUIC link generation
func TestTUICProxyGenerateLink(t *testing.T) {
	tests := []struct {
		name      string
		inputLink string
	}{
		{
			name:      "TUIC with password",
			inputLink: "tuic://12345678-1234-1234-1234-123456789012:password@example.com:443#Test",
		},
		{
			name:      "TUIC without password",
			inputLink: "tuic://12345678-1234-1234-1234-123456789012@example.com:443#Test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proxy, err := ParseProxyLine(tt.inputLink)
			if err != nil {
				t.Fatalf("ParseProxyLine error = %v", err)
			}

			link, _ := proxy.ToShareLink(nil)

			if !strings.HasPrefix(link, "tuic://") {
				t.Errorf("ToShareLink() should start with tuic://, got %v", link)
			}

			// Verify it can be parsed back
			_, err = ParseProxyLine(link)
			if err != nil {
				t.Errorf("Generated link cannot be parsed back: %v", err)
			}
		})
	}
}

func TestProcessRemark(t *testing.T) {
	remarks := map[string]int{}
	// First call returns original
	if ProcessRemark("Test", remarks) != "Test" {
		t.Error("ProcessRemark failed")
	}
	// Second call appends _2 (count+1)
	if ProcessRemark("Test", remarks) != "Test_2" {
		t.Error("ProcessRemark duplicate failed")
	}
}

func TestParseProxyLine(t *testing.T) {
	lines := []string{
		"ss://YWVzLTI1Ni1nY206cGFzc3dvcmQ=@example.com:443#Test",
		"trojan://password@example.com:443#Test",
	}
	for _, line := range lines {
		proxy, err := ParseProxyLine(line)
		if err != nil || proxy == nil {
			t.Errorf("ParseProxyLine failed for %s", line[:10])
		}
	}
}

func TestParseContentSkipsComments(t *testing.T) {
	content := `# comment line
ss://YWVzLTI1Ni1nY206cGFzc3dvcmQ=@example.com:443#Test`
	sp := &SubParser{
		Index: 0,
		URL:   "",
		Proxy: "",
	}
	result, err := sp.parseContent(content)
	if err != nil {
		t.Fatalf("parseContent returned error: %v", err)
	}
	if len(result.Proxies) != 1 {
		t.Fatalf("expected 1 proxy, got %d", len(result.Proxies))
	}
}

func TestSubParser_Parse_RawLine(t *testing.T) {
	sp := &SubParser{
		Index: 1,
		URL:   "ss://YWVzLTI1Ni1nY206cGFzc3dvcmQ=@example.com:443#Test",
	}
	sc, err := sp.Parse()
	assert.NoError(t, err)
	assert.Len(t, sc.Proxies, 1)
	assert.Equal(t, "Test", sc.Proxies[0].GetRemark())
}
