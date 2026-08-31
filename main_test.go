package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gfunc/subconvergo/config"
	"github.com/gin-gonic/gin"
)

func TestServer_HasFiniteTimeouts(t *testing.T) {
	srv := buildHTTPServer("127.0.0.1:0", http.NotFoundHandler())
	if srv.ReadHeaderTimeout <= 0 {
		t.Errorf("ReadHeaderTimeout must be finite, got %v", srv.ReadHeaderTimeout)
	}
	if srv.ReadTimeout <= 0 {
		t.Errorf("ReadTimeout must be finite, got %v", srv.ReadTimeout)
	}
	if srv.WriteTimeout <= 0 {
		t.Errorf("WriteTimeout must be finite, got %v", srv.WriteTimeout)
	}
	if srv.IdleTimeout <= 0 {
		t.Errorf("IdleTimeout must be finite, got %v", srv.IdleTimeout)
	}
}

func TestAdvancedLimits_AreEnforced(t *testing.T) {
	gin.SetMode(gin.TestMode)
	saved := *config.Global
	t.Cleanup(func() { *config.Global = saved })

	// One in-flight request, no queue: a second concurrent request must be
	// rejected with 503 instead of being served.
	config.Global.Advanced.MaxConcurrentThreads = 1
	config.Global.Advanced.MaxPendingConnections = 0

	router := gin.New()
	router.Use(concurrencyLimitMiddleware(
		config.Global.Advanced.MaxConcurrentThreads,
		config.Global.Advanced.MaxPendingConnections,
	))
	started := make(chan struct{})
	release := make(chan struct{})
	router.GET("/block", func(c *gin.Context) {
		close(started)
		<-release
		c.String(http.StatusOK, "ok")
	})
	router.GET("/fast", func(c *gin.Context) {
		c.String(http.StatusOK, "fast")
	})

	srv := httptest.NewServer(router)
	defer srv.Close()

	first := make(chan int, 1)
	go func() {
		resp, err := http.Get(srv.URL + "/block")
		if err != nil {
			first <- -1
			return
		}
		defer resp.Body.Close()
		first <- resp.StatusCode
	}()
	<-started

	resp, err := http.Get(srv.URL + "/fast")
	if err != nil {
		t.Fatalf("second request failed: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 while the single in-flight slot is occupied, got %d", resp.StatusCode)
	}

	close(release)
	if code := <-first; code != http.StatusOK {
		t.Fatalf("blocked request must complete with 200, got %d", code)
	}

	resp, err = http.Get(srv.URL + "/fast")
	if err != nil {
		t.Fatalf("third request failed: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 after the slot was released, got %d", resp.StatusCode)
	}
}
