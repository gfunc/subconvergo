package parser

import (
	"testing"
)

func TestParseHysteria(t *testing.T) {
	tests := []struct {
		name           string
		link           string
		wantErr        bool
		expectedServer string
		expectedPort   int
		expectedType   string
	}{
		{
			name:           "Hysteria v1 basic",
			link:           "hysteria://example.com:443?auth=password&insecure=1&peer=example.com&upmbps=100&downmbps=100#Hysteria",
			wantErr:        false,
			expectedServer: "example.com",
			expectedPort:   443,
			expectedType:   "hysteria",
		},
		{
			name:           "Hysteria2 basic",
			link:           "hysteria2://password@example.com:443?insecure=1&sni=example.com#Hysteria2",
			wantErr:        false,
			expectedServer: "example.com",
			expectedPort:   443,
			expectedType:   "hysteria2",
		},
		{
			name:           "Hysteria2 short prefix (hy2)",
			link:           "hy2://password@example.com:8443?obfs=salamander&obfs-password=pass#HY2",
			wantErr:        false,
			expectedServer: "example.com",
			expectedPort:   8443,
			expectedType:   "hysteria2",
		},
		{
			name:    "Invalid hysteria link",
			link:    "hysteria://invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proxy, err := parseHysteria(tt.link)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseHysteria() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if proxy.Server != tt.expectedServer {
					t.Errorf("expected server %s, got %s", tt.expectedServer, proxy.Server)
				}
				if proxy.Port != tt.expectedPort {
					t.Errorf("expected port %d, got %d", tt.expectedPort, proxy.Port)
				}
				if proxy.Type != tt.expectedType {
					t.Errorf("expected type %s, got %s", tt.expectedType, proxy.Type)
				}
				if proxy.MihomoProxy == nil {
					t.Error("mihomo proxy should not be nil")
				}
			}
		})
	}
}

func TestParseTUIC(t *testing.T) {
	tests := []struct {
		name           string
		link           string
		wantErr        bool
		expectedServer string
		expectedPort   int
		expectedUUID   string
	}{
		{
			name:           "TUIC basic",
			link:           "tuic://12345678-1234-1234-1234-123456789012:password@example.com:443?sni=example.com&alpn=h3&congestion_control=bbr#TUIC",
			wantErr:        false,
			expectedServer: "example.com",
			expectedPort:   443,
			expectedUUID:   "12345678-1234-1234-1234-123456789012",
		},
		{
			name:           "TUIC without password",
			link:           "tuic://12345678-1234-1234-1234-123456789012@example.com:8443?allow_insecure=1#TUIC-NoPass",
			wantErr:        false,
			expectedServer: "example.com",
			expectedPort:   8443,
			expectedUUID:   "12345678-1234-1234-1234-123456789012",
		},
		{
			name:    "Invalid TUIC link",
			link:    "tuic://invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proxy, err := parseTUIC(tt.link)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseTUIC() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if proxy.Server != tt.expectedServer {
					t.Errorf("expected server %s, got %s", tt.expectedServer, proxy.Server)
				}
				if proxy.Port != tt.expectedPort {
					t.Errorf("expected port %d, got %d", tt.expectedPort, proxy.Port)
				}
				if proxy.UUID != tt.expectedUUID {
					t.Errorf("expected UUID %s, got %s", tt.expectedUUID, proxy.UUID)
				}
				if proxy.MihomoProxy == nil {
					t.Error("mihomo proxy should not be nil")
				}
			}
		})
	}
}

func TestParseFallbackProtocol(t *testing.T) {
	tests := []struct {
		name    string
		link    string
		wantErr bool
	}{
		{
			name:    "Hysteria protocol",
			link:    "hysteria://example.com:443?auth=pass#Test",
			wantErr: false,
		},
		{
			name:    "Hysteria2 protocol",
			link:    "hysteria2://pass@example.com:443#Test",
			wantErr: false,
		},
		{
			name:    "TUIC protocol",
			link:    "tuic://uuid@example.com:443#Test",
			wantErr: false,
		},
		{
			name:    "Truly unsupported protocol",
			link:    "unknown://example.com:443#Test",
			wantErr: true,
		},
		{
			name:    "Invalid link format",
			link:    "not-a-link",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseFallbackProtocol(tt.link)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseFallbackProtocol() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseProxyLineWithFallback(t *testing.T) {
	tests := []struct {
		name         string
		link         string
		wantErr      bool
		expectedType string
	}{
		{
			name:         "SS link",
			link:         "ss://YWVzLTI1Ni1nY206cGFzc3dvcmQ=@1.2.3.4:8388#SS",
			wantErr:      false,
			expectedType: "ss",
		},
		{
			name:         "Hysteria link",
			link:         "hysteria://example.com:443?auth=pass#Hysteria",
			wantErr:      false,
			expectedType: "hysteria",
		},
		{
			name:         "Hysteria2 link",
			link:         "hysteria2://pass@example.com:443#HY2",
			wantErr:      false,
			expectedType: "hysteria2",
		},
		{
			name:         "TUIC link",
			link:         "tuic://uuid@example.com:443#TUIC",
			wantErr:      false,
			expectedType: "tuic",
		},
		{
			name:    "Unsupported protocol",
			link:    "unknown://example.com:443#Unknown",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proxy, err := ParseProxyLine(tt.link)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseProxyLine() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && proxy.Type != tt.expectedType {
				t.Errorf("expected type %s, got %s", tt.expectedType, proxy.Type)
			}
		})
	}
}

func BenchmarkParseHysteria(b *testing.B) {
	link := "hysteria2://password@example.com:443?sni=example.com#Benchmark"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parseHysteria(link)
	}
}

func BenchmarkParseTUIC(b *testing.B) {
	link := "tuic://12345678-1234-1234-1234-123456789012:password@example.com:443?sni=example.com#Benchmark"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parseTUIC(link)
	}
}
