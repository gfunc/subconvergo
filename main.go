package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gfunc/subconvergo/cache"
	"github.com/gfunc/subconvergo/config"
	_ "github.com/gfunc/subconvergo/generator/impl"
	"github.com/gfunc/subconvergo/handler"
	"github.com/gfunc/subconvergo/version"
	"github.com/gin-gonic/gin"
)

var (
	configFile = flag.String("f", "", "Path to configuration file")
	genMode    = flag.Bool("g", false, "Generator mode")
	artifact   = flag.String("artifact", "", "Profile name for generator mode")
	logFile    = flag.String("l", "", "Log file path")
)

func main() {
	flag.Parse()

	// Set up logging
	if *logFile != "" {
		f, err := openLogFile(*logFile)
		if err != nil {
			log.Fatalf("Failed to open log file: %v", err)
		}
		defer f.Close()
		log.SetOutput(f)
	}

	log.Printf("subconvergo %s starting up...", version.Version)
	log.Printf("Command line flags: configFile=%s", *configFile)

	// Change to config directory
	if *configFile != "" {
		// TODO: Store pref path
		dir := filepath.Dir(*configFile)
		log.Printf("Changing directory to: %s", dir)
		if dir != "." && dir != "" {
			if err := os.Chdir(dir); err != nil {
				log.Fatalf("Failed to change directory: %v", err)
			}
		}
	}

	// Load configuration
	if configFile, err := config.LoadConfig(*configFile); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	} else {
		log.Printf("Configuration loaded from: %s", configFile)
	}

	// Refuse an insecure public bind (tokenless or legacy-default token on a
	// non-loopback listen address).
	if err := config.Global.ValidateStartupSecurity(); err != nil {
		log.Fatalf("Insecure configuration: %v", err)
	}

	// Initialize cache
	cache.Init(config.GetBasePath())

	// Generator mode
	if *genMode {
		log.Println("Generator mode not yet implemented")
		return
	}

	// Start HTTP server
	startServer()
}

// openLogFile opens the -l log file owner-only (0600): log lines carry
// request metadata (client IPs, redacted URLs, sizes) that must not be
// world-readable.
func openLogFile(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
}

func startServer() {
	// Set gin mode
	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()

	// Enforce the advanced concurrency/pending limits on every route.
	router.Use(concurrencyLimitMiddleware(
		config.Global.Advanced.MaxConcurrentThreads,
		config.Global.Advanced.MaxPendingConnections,
	))

	// Create handler
	h := handler.NewSubHandler()

	// Register aliases (redirects)
	for _, alias := range config.Global.Aliases {
		uri := alias.URI
		target := alias.Target
		router.GET(uri, func(c *gin.Context) {
			// Build redirect URL with query parameters
			redirectURL := target
			if len(c.Request.URL.Query()) > 0 {
				redirectURL += "?" + c.Request.URL.RawQuery
			}
			c.Redirect(http.StatusMovedPermanently, redirectURL)
		})
	}

	// Register routes
	router.GET("/version", h.HandleVersion)
	router.GET("/sub", h.HandleSub)
	router.HEAD("/sub", h.HandleSub)
	router.GET("/surge2clash", h.HandleSurge2Clash)
	router.GET("/readconf", h.HandleReadConf)
	router.GET("/getruleset", h.HandleGetRuleset)
	router.GET("/getprofile", h.HandleGetProfile)
	router.GET("/render", h.HandleRender)
	router.GET("/flushcache", h.HandleFlushCache)

	// Additional routes when not in API mode
	if !config.Global.Common.APIMode {
		router.GET("/get", func(c *gin.Context) {
			// TODO: Implement /get endpoint
			c.String(200, "Not implemented")
		})
		router.GET("/getlocal", func(c *gin.Context) {
			// TODO: Implement /getlocal endpoint
			c.String(200, "Not implemented")
		})
	}

	// Start server
	addr := fmt.Sprintf("%s:%d", config.Global.Server.Listen, config.Global.Server.Port)
	log.Printf("Startup completed. Serving HTTP @ http://%s", addr)
	log.Printf("Loaded %d alias(es)", len(config.Global.Aliases))

	srv := buildHTTPServer(addr, router)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// buildHTTPServer builds the HTTP server with finite timeouts; bare
// router.Run leaves all of these unlimited (slowloris / stuck-connection DoS).
func buildHTTPServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:    addr,
		Handler: handler,
		// WriteTimeout must exceed the fetcher's overall timeout (30s) so
		// responses that wait on upstream fetches are not cut mid-flight.
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
}

// concurrencyLimitMiddleware enforces the advanced limits:
// max_concurrent_threads bounds in-flight requests and
// max_pending_connections bounds requests queued waiting for a slot; anything
// beyond that gets 503 instead of queueing unboundedly. A non-positive
// maxInFlight disables the limit entirely.
func concurrencyLimitMiddleware(maxInFlight, maxPending int) gin.HandlerFunc {
	if maxInFlight <= 0 {
		return func(c *gin.Context) { c.Next() }
	}
	if maxPending < 0 {
		maxPending = 0
	}
	slots := make(chan struct{}, maxInFlight)
	queue := make(chan struct{}, maxPending)
	return func(c *gin.Context) {
		select {
		case slots <- struct{}{}:
			defer func() { <-slots }()
			c.Next()
			return
		default:
		}

		select {
		case queue <- struct{}{}:
			defer func() { <-queue }()
		default:
			c.AbortWithStatus(http.StatusServiceUnavailable)
			return
		}

		select {
		case slots <- struct{}{}:
			defer func() { <-slots }()
			c.Next()
		case <-c.Request.Context().Done():
			c.AbortWithStatus(http.StatusServiceUnavailable)
		}
	}
}
