package parser

import (
	"encoding/base64"
	"testing"
)

func TestParseShadowsocks(t *testing.T) {
	tests := []struct {
		name    string
		link    string
		wantErr bool
	}{
		{
			name:    "valid ss link (new format)",
			link:    "ss://YWVzLTI1Ni1nY206cGFzc3dvcmQ=@example.com:8388#Test%20Server",
			wantErr: false,
		},
		{
			name:    "valid ss link (old format)",
			link:    "ss://YWVzLTI1Ni1nY206cGFzc3dvcmRAZXhhbXBsZS5jb206ODM4OA==#Test%20Server",
			wantErr: false,
		},
		{
			name:    "invalid ss link",
			link:    "ss://invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proxy, err := parseShadowsocks(tt.link)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseShadowsocks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if proxy.Type != "ss" {
					t.Errorf("expected type ss, got %s", proxy.Type)
				}
				if proxy.Server == "" {
					t.Errorf("server should not be empty")
				}
			}
		})
	}
}

func TestParseShadowsocksR(t *testing.T) {
	tests := []struct {
		name    string
		link    string
		wantErr bool
	}{
		{
			name:    "valid ssr link",
			link:    "ssr://ZXhhbXBsZS5jb206ODM4ODpvcmlnaW46YWVzLTI1Ni1jZmI6cGxhaW46Y0dGemMzZHZjbVE",
			wantErr: false,
		},
		{
			name:    "invalid ssr link",
			link:    "ssr://invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proxy, err := parseShadowsocksR(tt.link)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseShadowsocksR() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if proxy.Server == "" {
					t.Errorf("server should not be empty")
				}
			}
		})
	}
}

func TestParseVMess(t *testing.T) {
	// Valid VMess link (JSON format)
	vmessJSON := `{"v":"2","ps":"test","add":"example.com","port":"443","id":"12345678-1234-1234-1234-123456789012","aid":"0","net":"tcp","type":"none","host":"","path":"","tls":"tls"}`
	vmessLink := "vmess://" + urlSafeBase64Encode(vmessJSON)

	tests := []struct {
		name    string
		link    string
		wantErr bool
	}{
		{
			name:    "valid vmess link",
			link:    vmessLink,
			wantErr: false,
		},
		{
			name:    "invalid vmess link",
			link:    "vmess://invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proxy, err := parseVMess(tt.link)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseVMess() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if proxy.Type != "vmess" {
					t.Errorf("expected type vmess, got %s", proxy.Type)
				}
				if proxy.Server == "" {
					t.Errorf("server should not be empty")
				}
			}
		})
	}
}

func TestParseTrojan(t *testing.T) {
	tests := []struct {
		name    string
		link    string
		wantErr bool
	}{
		{
			name:    "valid trojan link",
			link:    "trojan://password@example.com:443#Test%20Server",
			wantErr: false,
		},
		{
			name:    "valid trojan link with ws",
			link:    "trojan://password@example.com:443?type=ws&path=/path&sni=example.com#Test",
			wantErr: false,
		},
		{
			name:    "invalid trojan link",
			link:    "trojan://invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proxy, err := parseTrojan(tt.link)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseTrojan() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if proxy.Type != "trojan" {
					t.Errorf("expected type trojan, got %s", proxy.Type)
				}
				if proxy.Server == "" {
					t.Errorf("server should not be empty")
				}
			}
		})
	}
}

func TestParseVLESS(t *testing.T) {
	tests := []struct {
		name    string
		link    string
		wantErr bool
	}{
		{
			name:    "valid vless link",
			link:    "vless://12345678-1234-1234-1234-123456789012@example.com:443?type=tcp&security=tls&sni=example.com#Test%20Server",
			wantErr: false,
		},
		{
			name:    "invalid vless link",
			link:    "vless://invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proxy, err := parseVLESS(tt.link)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseVLESS() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if proxy.Type != "vless" {
					t.Errorf("expected type vless, got %s", proxy.Type)
				}
				if proxy.Server == "" {
					t.Errorf("server should not be empty")
				}
			}
		})
	}
}

// Helper function for tests
func urlSafeBase64Encode(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}
