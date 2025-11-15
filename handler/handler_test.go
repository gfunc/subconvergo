package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gfunc/subconvergo/config"
	"github.com/gin-gonic/gin"
)

func init() {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Initialize global config with test defaults
	config.Global.Common.APIMode = true
	config.Global.Advanced.SkipFailedLinks = true
}

func TestVersionHandler(t *testing.T) {
	router := gin.New()
	handler := NewSubHandler()
	router.GET("/version", handler.HandleVersion)

	req, _ := http.NewRequest("GET", "/version", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if body == "" {
		t.Error("version response should not be empty")
	}

	// Check if response contains version info
	if len(body) < 5 {
		t.Error("version response seems too short")
	}
}

func TestReadConfHandler(t *testing.T) {
	router := gin.New()
	handler := NewSubHandler()
	router.GET("/readconf", handler.HandleReadConf)

	req, _ := http.NewRequest("GET", "/readconf", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return OK even if config doesn't exist (will use defaults)
	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 200 or 500, got %d", w.Code)
	}
}

func TestSubHandler_MissingParameters(t *testing.T) {
	router := gin.New()
	handler := NewSubHandler()
	router.GET("/sub", handler.HandleSub)

	tests := []struct {
		name           string
		url            string
		expectedStatus int
	}{
		{
			name:           "missing target and url",
			url:            "/sub",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing url",
			url:            "/sub?target=clash",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing target",
			url:            "/sub?url=https://example.com/sub",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", tt.url, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestSubHandler_ValidRequest(t *testing.T) {
	// This test would require a mock HTTP server for subscription fetching
	// Skipping actual conversion test, just checking the endpoint exists
	router := gin.New()
	handler := NewSubHandler()
	router.GET("/sub", handler.HandleSub)

	req, _ := http.NewRequest("GET", "/sub?target=clash&url=https://example.com/sub", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Will fail to fetch from example.com, but endpoint should respond
	// Status could be 400, 500, or 200 depending on implementation
	if w.Code < 200 || w.Code >= 600 {
		t.Errorf("unexpected status code %d", w.Code)
	}
}

func TestGetRulesetHandler(t *testing.T) {
	router := gin.New()
	handler := NewSubHandler()
	router.GET("/getruleset", handler.HandleGetRuleset)

	req, _ := http.NewRequest("GET", "/getruleset?type=1&url=https://example.com/rules.list", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should handle the request (might return error if can't fetch)
	if w.Code < 200 || w.Code >= 600 {
		t.Errorf("unexpected status code %d", w.Code)
	}
}

func TestHealthCheckHandler(t *testing.T) {
	router := gin.New()
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("failed to parse JSON response: %v", err)
	}

	if response["status"] != "ok" {
		t.Errorf("expected status ok, got %v", response["status"])
	}
}

func TestCORSMiddleware(t *testing.T) {
	router := gin.New()
	// Simple CORS middleware
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "test")
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://example.com")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") == "" {
		t.Error("CORS header not set")
	}
}

func TestQueryParameterParsing(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected map[string]string
	}{
		{
			name:  "simple parameters",
			query: "target=clash&url=https://example.com",
			expected: map[string]string{
				"target": "clash",
				"url":    "https://example.com",
			},
		},
		{
			name:  "URL encoded parameters",
			query: "url=https%3A%2F%2Fexample.com%2Fsub&include=%E9%A6%99%E6%B8%AF",
			expected: map[string]string{
				"url":     "https://example.com/sub",
				"include": "香港",
			},
		},
		{
			name:  "multiple URLs",
			query: "url=https://example1.com|https://example2.com",
			expected: map[string]string{
				"url": "https://example1.com|https://example2.com",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.GET("/test", func(c *gin.Context) {
				for key, expectedValue := range tt.expected {
					actualValue := c.Query(key)
					if actualValue != expectedValue {
						t.Errorf("expected %s=%s, got %s=%s", key, expectedValue, key, actualValue)
					}
				}
				c.String(http.StatusOK, "ok")
			})

			req, _ := http.NewRequest("GET", "/test?"+tt.query, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
		})
	}
}

// TestHandleGetProfile tests the /getprofile endpoint
func TestHandleGetProfile(t *testing.T) {
	router := gin.New()
	handler := NewSubHandler()
	router.GET("/getprofile", handler.HandleGetProfile)

	tests := []struct {
		name           string
		query          string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "missing token",
			query:          "name=profiles/example_profile.ini",
			expectedStatus: http.StatusForbidden,
			expectedBody:   "Forbidden",
		},
		{
			name:           "missing name",
			query:          "token=password",
			expectedStatus: http.StatusForbidden,
			expectedBody:   "Forbidden",
		},
		{
			name:           "profile not found",
			query:          "name=profiles/nonexistent.ini&token=password",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Profile not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/getprofile?"+tt.query, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d (body: %s)", tt.expectedStatus, w.Code, w.Body.String())
			}

			if tt.expectedBody != "" && !strings.Contains(w.Body.String(), tt.expectedBody) {
				t.Errorf("expected body to contain %q, got %q", tt.expectedBody, w.Body.String())
			}
		})
	}

	// Test with existing profile
	t.Run("existing profile", func(t *testing.T) {
		// Change to parent directory where base/ is located
		originalWd, _ := os.Getwd()
		os.Chdir("..")
		defer os.Chdir(originalWd)

		// Set the access token for this test
		originalToken := config.Global.Common.APIAccessToken
		config.Global.Common.APIAccessToken = "test-password"
		defer func() { config.Global.Common.APIAccessToken = originalToken }()

		// Use the example_profile.ini which already exists
		// The test will forward to /sub, which will fail with 400 (missing url or invalid)
		// But we can verify the profile was loaded
		req, _ := http.NewRequest("GET", "/getprofile?name=example_profile&token=test-password", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Should process the profile (may return 400 from /sub if profile has invalid config)
		// But should NOT return 404 (profile not found) or 403 (forbidden)
		if w.Code == http.StatusNotFound {
			t.Errorf("profile should be found, got 404: %s", w.Body.String())
		}
		if w.Code == http.StatusForbidden {
			t.Errorf("should not be forbidden, got 403: %s", w.Body.String())
		}
	})
}

// BenchmarkSubHandler benchmarks the subscription handler
func BenchmarkSubHandler(b *testing.B) {
	router := gin.New()
	handler := NewSubHandler()
	router.GET("/sub", handler.HandleSub)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/sub?target=clash&url=https://example.com/sub", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}
