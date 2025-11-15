package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/handler"
	"github.com/gin-gonic/gin"
)

var (
	version    = "0.1.0"
	configFile = flag.String("f", "", "Path to configuration file")
	genMode    = flag.Bool("g", false, "Generator mode")
	artifact   = flag.String("artifact", "", "Profile name for generator mode")
	logFile    = flag.String("l", "", "Log file path")
)

func main() {
	flag.Parse()

	// Set up logging
	if *logFile != "" {
		f, err := os.OpenFile(*logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalf("Failed to open log file: %v", err)
		}
		defer f.Close()
		log.SetOutput(f)
	}

	log.Printf("subconvergo %s starting up...", version)

	// Change to config directory
	if *configFile != "" {
		// TODO: Store pref path
		dir := filepath.Dir(*configFile)
		if dir != "." && dir != "" {
			if err := os.Chdir(dir); err != nil {
				log.Fatalf("Failed to change directory: %v", err)
			}
		}
	}

	// Load configuration
	if configFile, err := config.LoadConfig(); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	} else {
		log.Printf("Configuration loaded from: %s", configFile)
	}

	// Check for environment variable overrides
	checkEnvOverrides()

	// Generator mode
	if *genMode {
		log.Println("Generator mode not yet implemented")
		return
	}

	// Start HTTP server
	startServer()
}

func checkEnvOverrides() {
	if apiMode := os.Getenv("API_MODE"); apiMode != "" {
		config.Global.Common.APIMode = apiMode == "true"
	}

	if managedPrefix := os.Getenv("MANAGED_PREFIX"); managedPrefix != "" {
		config.Global.ManagedConfig.ManagedConfigPrefix = managedPrefix
	}

	if token := os.Getenv("API_TOKEN"); token != "" {
		config.Global.Common.APIAccessToken = token
	}

	if port := os.Getenv("PORT"); port != "" {
		fmt.Sscanf(port, "%d", &config.Global.Server.Port)
	}
}

func startServer() {
	// Set gin mode
	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()

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
	router.GET("/readconf", h.HandleReadConf)
	router.GET("/getruleset", h.HandleGetRuleset)
	router.GET("/getprofile", h.HandleGetProfile)
	router.GET("/render", h.HandleRender)

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

	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
