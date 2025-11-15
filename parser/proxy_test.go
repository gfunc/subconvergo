package parser

import (
	"strings"
	"testing"
)

// TestProxyInterfaceGenerateLink tests that GenerateLink works correctly for all proxy types
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
				t.Fatalf("ParseProxyLineWithInterface() error = %v", err)
			}

			// Verify the type
			if proxyInterface.GetType() != tt.expectedType {
				t.Errorf("GetType() = %v, want %v", proxyInterface.GetType(), tt.expectedType)
			}

			// Generate a new link
			generatedLink, _ := proxyInterface.GenerateLink()
			if generatedLink == "" {
				t.Errorf("GenerateLink() returned empty string")
			}

			// Parse the generated link back and verify it matches
			proxyInterface2, err := ParseProxyLine(generatedLink)
			if err != nil {
				t.Fatalf("ParseProxyLineWithInterface() of generated link error = %v", err)
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
		t.Fatalf("ParseProxyLineWithInterface() error = %v", err)
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

// TestProxyInterfaceToLegacyProxy tests conversion to legacy Proxy struct
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
			proxy, err := parseShadowsocks(tt.inputLink)
			if err != nil {
				t.Fatalf("parseShadowsocks error = %v", err)
			}

			link, _ := proxy.GenerateLink()
			if !strings.HasPrefix(link, "ss://") {
				t.Errorf("GenerateLink() should start with ss://, got %v", link)
			}

			// Verify it can be parsed back
			_, err = parseShadowsocks(link)
			if err != nil {
				t.Errorf("Generated link cannot be parsed back: %v", err)
			}
		})
	}
}

// TestVMessProxyGenerateLink tests VMess link generation
func TestVMessProxyGenerateLink(t *testing.T) {
	inputLink := "vmess://eyJ2IjoiMiIsInBzIjoidGVzdCIsImFkZCI6ImV4YW1wbGUuY29tIiwicG9ydCI6IjQ0MyIsImlkIjoiMTIzNDU2NzgtMTIzNC0xMjM0LTEyMzQtMTIzNDU2Nzg5MDEyIiwiYWlkIjoiMCIsIm5ldCI6IndzIiwicGF0aCI6Ii9wYXRoIiwiaG9zdCI6ImV4YW1wbGUuY29tIiwidGxzIjoidGxzIn0="

	proxy, err := parseVMess(inputLink)
	if err != nil {
		t.Fatalf("parseVMess error = %v", err)
	}

	link, _ := proxy.GenerateLink()
	if !strings.HasPrefix(link, "vmess://") {
		t.Errorf("GenerateLink() should start with vmess://, got %v", link)
	}

	// Verify it can be parsed back
	proxy2, err := parseVMess(link)
	if err != nil {
		t.Errorf("Generated link cannot be parsed back: %v", err)
	}

	proxyVMess, err := proxy.(*MihomoProxy).GetVmessProxy()
	if err != nil {
		t.Fatalf("proxy is not of type *VMessProxy")
	}

	proxy2VMess, err := proxy2.(*MihomoProxy).GetVmessProxy()
	if err != nil {
		t.Fatalf("proxy2 is not of type *VMessProxy")
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
			proxy, err := parseTrojan(tt.inputLink)
			if err != nil {
				t.Fatalf("parseTrojan error = %v", err)
			}

			link, _ := proxy.GenerateLink()
			if !strings.HasPrefix(link, "trojan://") {
				t.Errorf("GenerateLink() should start with trojan://, got %v", link)
			}

			// Verify it can be parsed back
			_, err = parseTrojan(link)
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
			proxy, err := parseVLESS(tt.inputLink)
			if err != nil {
				t.Fatalf("parseVLESS error = %v", err)
			}

			link, _ := proxy.GenerateLink()
			if !strings.HasPrefix(link, "vless://") {
				t.Errorf("GenerateLink() should start with vless://, got %v", link)
			}

			// Verify it can be parsed back
			_, err = parseVLESS(link)
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
			proxy, err := parseHysteria(tt.inputLink)
			if err != nil {
				t.Fatalf("parseHysteria error = %v", err)
			}

			link, _ := proxy.GenerateLink()
			if !strings.HasPrefix(link, tt.protocol+"://") {
				t.Errorf("GenerateLink() should start with %s://, got %v", tt.protocol, link)
			}

			// Verify it can be parsed back
			_, err = parseHysteria(link)
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
			proxy, err := parseTUIC(tt.inputLink)
			if err != nil {
				t.Fatalf("parseTUIC error = %v", err)
			}

			link, _ := proxy.GenerateLink()
			if !strings.HasPrefix(link, "tuic://") {
				t.Errorf("GenerateLink() should start with tuic://, got %v", link)
			}

			// Verify it can be parsed back
			_, err = parseTUIC(link)
			if err != nil {
				t.Errorf("Generated link cannot be parsed back: %v", err)
			}
		})
	}
}
