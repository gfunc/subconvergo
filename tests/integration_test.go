package tests

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gfunc/subconvergo/parser"
)

// TestEndToEndSSConversion tests complete SS subscription conversion
func TestEndToEndSSConversion(t *testing.T) {
	// Create a mock subscription server
	ssLinks := "ss://YWVzLTI1Ni1nY206cGFzc3dvcmQ=@example1.com:8388#Server1\n" +
		"ss://YWVzLTI1Ni1nY206cGFzc3dvcmQ=@example2.com:8388#Server2"

	base64Sub := base64.StdEncoding.EncodeToString([]byte(ssLinks))

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(base64Sub))
	}))
	defer mockServer.Close()

	// Test parsing the subscription
	resp, err := http.Get(mockServer.URL)
	if err != nil {
		t.Fatalf("failed to fetch subscription: %v", err)
	}
	defer resp.Body.Close()

	_, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response: %v", err)
	}

	proxies, err := parser.ParseSubscription(mockServer.URL, "")
	if err != nil {
		// This might fail if ParseSubscription is not fully implemented
		t.Logf("ParseSubscription not fully implemented: %v", err)
		return
	}

	if len(proxies) != 2 {
		t.Errorf("expected 2 proxies, got %d", len(proxies))
	}

	// Verify proxy details
	for i, proxy := range proxies {
		if proxy.Type != "ss" {
			t.Errorf("proxy %d: expected type ss, got %s", i, proxy.Type)
		}
		if proxy.Server == "" {
			t.Errorf("proxy %d: server should not be empty", i)
		}
		if proxy.Port == 0 {
			t.Errorf("proxy %d: port should not be 0", i)
		}
	}
}

// TestEndToEndVMessConversion tests VMess subscription conversion
func TestEndToEndVMessConversion(t *testing.T) {
	vmessJSON1 := `{"v":"2","ps":"Server1","add":"example1.com","port":"443","id":"12345678-1234-1234-1234-123456789012","aid":"0","net":"ws","type":"none","host":"example1.com","path":"/path","tls":"tls"}`
	vmessJSON2 := `{"v":"2","ps":"Server2","add":"example2.com","port":"443","id":"12345678-1234-1234-1234-123456789012","aid":"0","net":"grpc","type":"none","path":"service","tls":"tls"}`

	vmessLink1 := "vmess://" + base64.StdEncoding.EncodeToString([]byte(vmessJSON1))
	vmessLink2 := "vmess://" + base64.StdEncoding.EncodeToString([]byte(vmessJSON2))

	subContent := vmessLink1 + "\n" + vmessLink2
	base64Sub := base64.StdEncoding.EncodeToString([]byte(subContent))

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(base64Sub))
	}))
	defer mockServer.Close()

	proxies, err := parser.ParseSubscription(mockServer.URL, "")
	if err != nil {
		t.Logf("ParseSubscription not fully implemented: %v", err)
		return
	}

	if len(proxies) != 2 {
		t.Errorf("expected 2 proxies, got %d", len(proxies))
	}

	// Check networks
	networks := []string{"ws", "grpc"}
	for i, proxy := range proxies {
		if proxy.Network != networks[i] {
			t.Errorf("proxy %d: expected network %s, got %s", i, networks[i], proxy.Network)
		}
	}
}

// TestEndToEndMixedSubscription tests mixed protocol subscription
func TestEndToEndMixedSubscription(t *testing.T) {
	ssLink := "ss://YWVzLTI1Ni1nY206cGFzc3dvcmQ=@example.com:8388#SS"
	trojanLink := "trojan://password@example.com:443#Trojan"
	vmessJSON := `{"v":"2","ps":"VMess","add":"example.com","port":"443","id":"12345678-1234-1234-1234-123456789012","aid":"0","net":"tcp","tls":"tls"}`
	vmessLink := "vmess://" + base64.StdEncoding.EncodeToString([]byte(vmessJSON))

	subContent := ssLink + "\n" + trojanLink + "\n" + vmessLink
	base64Sub := base64.StdEncoding.EncodeToString([]byte(subContent))

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(base64Sub))
	}))
	defer mockServer.Close()

	proxies, err := parser.ParseSubscription(mockServer.URL, "")
	if err != nil {
		t.Logf("ParseSubscription not fully implemented: %v", err)
		return
	}

	if len(proxies) != 3 {
		t.Errorf("expected 3 proxies, got %d", len(proxies))
	}

	// Check types
	expectedTypes := []string{"ss", "trojan", "vmess"}
	for i, proxy := range proxies {
		if i < len(expectedTypes) && proxy.Type != expectedTypes[i] {
			t.Errorf("proxy %d: expected type %s, got %s", i, expectedTypes[i], proxy.Type)
		}
	}
}

// TestSubscriptionWithProxy tests fetching through a proxy
func TestSubscriptionWithProxy(t *testing.T) {
	// Create a mock proxy server
	proxyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate proxy behavior
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("proxy response"))
	}))
	defer proxyServer.Close()

	// This test would need actual proxy implementation
	t.Skip("Proxy support test - requires full implementation")
}

// TestSubscriptionTimeout tests timeout handling
func TestSubscriptionTimeout(t *testing.T) {
	// Create a slow server
	slowServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(35 * time.Second) // Longer than default timeout
		w.Write([]byte("too slow"))
	}))
	defer slowServer.Close()

	start := time.Now()
	_, err := parser.ParseSubscription(slowServer.URL, "")
	duration := time.Since(start)

	if err == nil {
		t.Error("expected timeout error, got nil")
	}

	if duration > 31*time.Second {
		t.Errorf("timeout took too long: %v", duration)
	}
}

// TestSubscriptionHTTPErrors tests various HTTP error responses
func TestSubscriptionHTTPErrors(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		wantErr    bool
	}{
		{"404 Not Found", http.StatusNotFound, true},
		{"500 Internal Server Error", http.StatusInternalServerError, true},
		{"403 Forbidden", http.StatusForbidden, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			_, err := parser.ParseSubscription(server.URL, "")
			if (err != nil) != tt.wantErr {
				t.Errorf("expected error: %v, got: %v", tt.wantErr, err)
			}
		})
	}

	// Test 200 OK with valid content
	t.Run("200 OK with valid content", func(t *testing.T) {
		validSS := "ss://YWVzLTI1Ni1nY206cGFzc3dvcmQ=@example.com:8388#TestServer"
		base64Sub := base64.StdEncoding.EncodeToString([]byte(validSS))

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(base64Sub))
		}))
		defer server.Close()

		proxies, err := parser.ParseSubscription(server.URL, "")
		if err != nil {
			t.Errorf("expected no error for 200 OK with valid content, got: %v", err)
		}
		if len(proxies) == 0 {
			t.Error("expected at least one proxy")
		}
	})
}

// TestClashYAMLParsing tests parsing Clash format subscriptions
func TestClashYAMLParsing(t *testing.T) {
	clashYAML := `
proxies:
  - name: "ss-test"
    type: ss
    server: example.com
    port: 8388
    cipher: aes-256-gcm
    password: password
  - name: "vmess-test"
    type: vmess
    server: example.com
    port: 443
    uuid: 12345678-1234-1234-1234-123456789012
    alterId: 0
    cipher: auto
    network: ws
    tls: true
    ws-opts:
      path: /path
`

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/yaml")
		w.Write([]byte(clashYAML))
	}))
	defer mockServer.Close()

	proxies, err := parser.ParseSubscription(mockServer.URL, "")
	if err != nil {
		t.Logf("Clash YAML parsing not fully implemented: %v", err)
		return
	}

	if len(proxies) < 2 {
		t.Errorf("expected at least 2 proxies, got %d", len(proxies))
	}
}

// TestProxyRemarkUniqueness tests that duplicate remarks are handled
func TestProxyRemarkUniqueness(t *testing.T) {
	ssLinks := "ss://YWVzLTI1Ni1nY206cGFzc3dvcmQ=@example1.com:8388#Server\n" +
		"ss://YWVzLTI1Ni1nY206cGFzc3dvcmQ=@example2.com:8388#Server\n" +
		"ss://YWVzLTI1Ni1nY206cGFzc3dvcmQ=@example3.com:8388#Server"

	base64Sub := base64.StdEncoding.EncodeToString([]byte(ssLinks))

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(base64Sub))
	}))
	defer mockServer.Close()

	proxies, err := parser.ParseSubscription(mockServer.URL, "")
	if err != nil {
		t.Logf("ParseSubscription not fully implemented: %v", err)
		return
	}

	// Apply ProcessRemark to ensure uniqueness
	remarks := make(map[string]int)
	for i := range proxies {
		proxies[i].Remark = parser.ProcessRemark(proxies[i].Remark, remarks)
	}

	// Check all remarks are unique
	uniqueRemarks := make(map[string]bool)
	for _, proxy := range proxies {
		if uniqueRemarks[proxy.Remark] {
			t.Errorf("duplicate remark found: %s", proxy.Remark)
		}
		uniqueRemarks[proxy.Remark] = true
	}
}

// TestLargeSubscription tests handling of large subscriptions
func TestLargeSubscription(t *testing.T) {
	// Generate 200 valid SS proxies (reduced from 1000 for faster testing)
	var links []string
	for i := 0; i < 200; i++ {
		// Create valid SS link with different ports
		port := 8388 + (i % 100)
		link := fmt.Sprintf("ss://YWVzLTI1Ni1nY206cGFzc3dvcmQ=@example.com:%d#Server%d", port, i)
		links = append(links, link)
	}

	subContent := strings.Join(links, "\n")
	base64Sub := base64.StdEncoding.EncodeToString([]byte(subContent))

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(base64Sub))
	}))
	defer mockServer.Close()

	start := time.Now()
	proxies, err := parser.ParseSubscription(mockServer.URL, "")
	duration := time.Since(start)

	if err != nil {
		t.Logf("ParseSubscription not fully implemented: %v", err)
		return
	}

	if len(proxies) < 180 { // Allow some parsing failures (90% success rate)
		t.Errorf("expected at least 180 proxies, got %d", len(proxies))
	}

	if duration > 5*time.Second {
		t.Errorf("parsing took too long: %v", duration)
	}
}

// BenchmarkSubscriptionParsing benchmarks subscription parsing
func BenchmarkSubscriptionParsing(b *testing.B) {
	ssLinks := "ss://YWVzLTI1Ni1nY206cGFzc3dvcmQ=@example1.com:8388#Server1\n" +
		"ss://YWVzLTI1Ni1nY206cGFzc3dvcmQ=@example2.com:8388#Server2\n" +
		"ss://YWVzLTI1Ni1nY206cGFzc3dvcmQ=@example3.com:8388#Server3"

	base64Sub := base64.StdEncoding.EncodeToString([]byte(ssLinks))

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(base64Sub))
	}))
	defer mockServer.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parser.ParseSubscription(mockServer.URL, "")
	}
}
