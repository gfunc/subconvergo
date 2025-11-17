package handler

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"

	"strings"
	"text/template"

	"github.com/gfunc/subconvergo/proxy"

	"github.com/BurntSushi/toml"
	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/generator"
	"github.com/gfunc/subconvergo/parser"
	"github.com/gin-gonic/gin"
	"gopkg.in/ini.v1"
	"gopkg.in/yaml.v3"
)

// SubHandler handles subscription conversion requests
type SubHandler struct{}

// NewSubHandler creates a new subscription handler
func NewSubHandler() *SubHandler {
	return &SubHandler{}
}

// HandleSub processes /sub endpoint
func (h *SubHandler) HandleSub(c *gin.Context) {
	h.handleSubWithParams(c, nil)
}

// handleSubWithParams processes /sub with optional parameter overrides
func (h *SubHandler) handleSubWithParams(c *gin.Context, params map[string]string) {
	// Extract parameters from params map or query
	getParam := func(key string) string {
		if params != nil {
			if val, ok := params[key]; ok {
				return val
			}
		}
		return c.Query(key)
	}

	target := getParam("target")
	urlParam := getParam("url")
	configParam := getParam("config")

	// Validate required parameters
	if target == "" {
		c.String(http.StatusBadRequest, "Invalid target!")
		return
	}

	// Use default URL if empty and not in API mode
	if urlParam == "" {
		if !config.Global.Common.APIMode {
			urlParam = strings.Join(config.Global.Common.DefaultURL, "|")
		}
	}

	if urlParam == "" {
		c.String(http.StatusBadRequest, "Invalid request!")
		return
	}

	// URL decode
	urlParam, err := url.QueryUnescape(urlParam)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid URL encoding")
		return
	}

	// Handle insert URLs first if enabled
	var urlsToProcess []string
	if config.Global.Common.EnableInsert && len(config.Global.Common.InsertURL) > 0 {
		if config.Global.Common.PrependInsertURL {
			urlsToProcess = append(urlsToProcess, config.Global.Common.InsertURL...)
		}
	}

	// Add main URLs
	urlsToProcess = append(urlsToProcess, strings.Split(urlParam, "|")...)

	// Append insert URLs if needed
	if config.Global.Common.EnableInsert && len(config.Global.Common.InsertURL) > 0 {
		if !config.Global.Common.PrependInsertURL {
			urlsToProcess = append(urlsToProcess, config.Global.Common.InsertURL...)
		}
	}
	log.Printf("[handler.HandleSub] target=%s urls=%d urlLen=%d config=%s client=%s", target, len(urlsToProcess), len(urlParam), configParam, c.ClientIP())

	// Load external config if specified
	proxyGroups := config.Global.ProxyGroups.CustomProxyGroups
	rulesets := config.Global.Rulesets.Rulesets

	if configParam != "" {
		// Load external config (can be URL or file path)
		extConfig, err := h.loadExternalConfig(configParam)
		if err != nil {
			log.Printf("[handler.HandleSub] failed to load external config %s: %v", configParam, err)
		} else if extConfig != nil {
			log.Printf("[handler.HandleSub] loaded external config %s proxyGroups=%d rulesets=%d", configParam, len(extConfig.ProxyGroups), len(extConfig.Rulesets))
			// Merge external config
			if len(extConfig.ProxyGroups) > 0 {
				proxyGroups = extConfig.ProxyGroups
			}
			if len(extConfig.Rulesets) > 0 {
				rulesets = extConfig.Rulesets
			}
		}
	}

	// Parse subscription URLs (support multiple URLs separated by |)
	var allProxies []proxy.ProxyInterface
	var otherProxyGroups []config.ProxyGroupConfig
	var rawRules []string
	for index, url := range urlsToProcess {
		url = strings.TrimSpace(url)
		if url == "" {
			continue
		}
		sp := &parser.SubParser{
			Index: index,
			URL:   url,
			Proxy: config.Global.Common.ProxySubscription,
		}
		custom, err := sp.Parse()
		if err == nil {
			allProxies = append(allProxies, custom.Proxies...)
			otherProxyGroups = append(otherProxyGroups, custom.Groups...)
			rawRules = append(rawRules, custom.RawRules...)
			continue
		} else if !config.Global.Advanced.SkipFailedLinks {
			c.String(http.StatusBadRequest, fmt.Sprintf("Failed to parse subscription (%s): %v", url, err))
			return
		} else {
			log.Printf("[handler.HandleSub] failed to parse subscription (index=%d url=%s): %v", index, url, err)
		}
	}

	if len(allProxies) == 0 {
		log.Printf("[handler.HandleSub] no valid proxies parsed from %d url(s)", len(urlsToProcess))
		c.String(http.StatusBadRequest, "No valid proxies found")
		return
	}

	// Apply filters
	allProxies = h.applyFilters(allProxies, c)
	log.Printf("[handler.HandleSub] proxies after filters=%d", len(allProxies))

	// Reload config on request if enabled
	if config.Global.Common.ReloadConfOnRequest {
		if _, err := config.LoadConfig(); err == nil {
			// Config reloaded successfully
		}
	}

	if len(otherProxyGroups) > 0 {
		proxyGroups = append(proxyGroups, otherProxyGroups...)
	}

	// Prepare generator options
	opts := generator.GeneratorOptions{
		Target:          target,
		ProxyGroups:     proxyGroups,
		Rulesets:        rulesets,
		RawRules:        rawRules,
		AppendProxyType: config.Global.Common.AppendProxyType,
		EnableRuleGen:   config.Global.Rulesets.Enabled,
		RenameNodes:     config.Global.NodePref.RenameNodes,
		SortProxies:     config.Global.NodePref.SortFlag,
		Emoji:           config.Global.Emojis,

		ExtraSetting: config.ExtraSetting{
			ClashProxiesStyle:   config.Global.NodePref.ClashProxiesStyle,
			ClashGroupsStyle:    config.Global.NodePref.ClashProxyGroupsStyle,
			SingBoxAddClashMode: config.Global.NodePref.SingBoxAddClashModes,
		},
	}

	// Apply node preferences to generator options
	if config.Global.NodePref.UDPFlag != nil {
		opts.UDP = config.Global.NodePref.UDPFlag
	}
	if config.Global.NodePref.TCPFastOpenFlag != nil {
		opts.TFO = config.Global.NodePref.TCPFastOpenFlag
	}
	if config.Global.NodePref.SkipCertVerifyFlag != nil {
		opts.SkipCertVerify = config.Global.NodePref.SkipCertVerifyFlag
	}
	if config.Global.NodePref.TLS13Flag != nil {
		opts.TLS13 = config.Global.NodePref.TLS13Flag
	}

	// Parse boolean options
	if udp := getParam("udp"); udp != "" {
		val := udp == "true"
		opts.UDP = &val
	}
	if tfo := getParam("tfo"); tfo != "" {
		val := tfo == "true"
		opts.TFO = &val
	}
	if scv := getParam("scv"); scv != "" {
		val := scv == "true"
		opts.SkipCertVerify = &val
	}

	// Prepare request parameters for template rendering
	requestParams := map[string]string{
		"target": target,
	}
	// Add all query parameters to request context
	for key, values := range c.Request.URL.Query() {
		if len(values) > 0 {
			requestParams[key] = values[0]
		}
	}

	// Load base configuration
	baseConfig, err := h.loadBaseConfig(target, requestParams)
	if err != nil {
		log.Printf("[handler.HandleSub] loadBaseConfig target=%s err=%v", target, err)
		c.String(http.StatusInternalServerError, fmt.Sprintf("Failed to load base config: %v", err))
		return
	}

	// Generate output
	output, err := generator.Generate(allProxies, opts, baseConfig)
	if err != nil {
		log.Printf("[handler.HandleSub] generator failed target=%s proxies=%d err=%v", target, len(allProxies), err)
		c.String(http.StatusInternalServerError, fmt.Sprintf("Failed to generate config: %v", err))
		return
	}

	// Set appropriate content type
	contentType := "text/plain;charset=utf-8"
	switch target {
	case "clash", "clashr":
		contentType = "text/yaml;charset=utf-8"
	case "singbox":
		contentType = "application/json;charset=utf-8"
	}

	// Add managed config header for Surge/Surfboard
	if target == "surge" || target == "surfboard" {
		if config.Global.ManagedConfig.WriteManagedConfig && config.Global.ManagedConfig.ManagedConfigPrefix != "" {
			managedURL := config.Global.ManagedConfig.ManagedConfigPrefix + "/sub?" + c.Request.URL.RawQuery
			output = fmt.Sprintf("#!MANAGED-CONFIG %s interval=%d strict=%t\n%s",
				managedURL,
				config.Global.ManagedConfig.ConfigUpdateInterval,
				config.Global.ManagedConfig.ConfigUpdateStrict,
				output)
		}
	}

	// Add QuanX device ID header if configured
	if target == "quanx" && config.Global.ManagedConfig.QuanXDeviceID != "" {
		c.Header("profile-update-interval", fmt.Sprintf("%d", config.Global.ManagedConfig.ConfigUpdateInterval))
		c.Header("subscription-userinfo", "upload=0; download=0; total=10737418240; expire=4102329600")
	}

	// Append subscription userinfo if enabled
	if config.Global.NodePref.AppendSubUserinfo {
		// Check if we have userinfo from subscription headers
		if userinfo := c.GetHeader("subscription-userinfo"); userinfo != "" {
			c.Header("subscription-userinfo", userinfo)
		}
	}

	c.Data(http.StatusOK, contentType, []byte(output))
}

func (h *SubHandler) applyFilters(proxies []proxy.ProxyInterface, c *gin.Context) []proxy.ProxyInterface {
	// Get filter params
	include := c.Query("include")
	exclude := c.Query("exclude")

	// Apply exclude filter
	if exclude != "" || len(config.Global.Common.ExcludeRemarks) > 0 {
		patterns := append(config.Global.Common.ExcludeRemarks, exclude)
		proxies = filterProxies(proxies, patterns, false)
	}

	// Apply include filter
	if include != "" || len(config.Global.Common.IncludeRemarks) > 0 {
		patterns := append(config.Global.Common.IncludeRemarks, include)
		proxies = filterProxies(proxies, patterns, true)
	}

	return proxies
}

func filterProxies(proxies []proxy.ProxyInterface, patterns []string, include bool) []proxy.ProxyInterface {
	if len(patterns) == 0 {
		return proxies
	}

	// Pre-compile regexes for all patterns
	type compiledPat struct {
		raw string
		re  *regexp.Regexp
	}
	compiled := make([]compiledPat, len(patterns))
	for i, p := range patterns {
		compiled[i].raw = p
		if p == "" {
			continue
		}
		var expr string
		if strings.HasPrefix(p, "/") && strings.HasSuffix(p, "/") && len(p) > 2 {
			expr = p[1 : len(p)-1]
		} else {
			expr = regexp.QuoteMeta(p)
		}
		if re, err := regexp.Compile(expr); err == nil {
			compiled[i].re = re
		}
	}

	var result []proxy.ProxyInterface
	for _, pr := range proxies {
		matched := false
		for i, p := range patterns {
			if p == "" {
				continue
			}
			if compiled[i].re != nil {
				if compiled[i].re.MatchString(pr.GetRemark()) {
					matched = true
					break
				}
			} else {
				// regex failed to compile; fallback to substring contains
				if strings.Contains(pr.GetRemark(), p) {
					matched = true
					break
				}
			}
		}
		if include == matched {
			result = append(result, pr)
		}
	}

	return result
}

func (h *SubHandler) loadBaseConfig(target string, requestParams map[string]string) (string, error) {
	var basePath string

	switch target {
	case "clash", "clashr":
		basePath = config.Global.Common.ClashRuleBase
	case "surge":
		basePath = config.Global.Common.SurgeRuleBase
	case "surfboard":
		basePath = config.Global.Common.SurfboardRuleBase
	case "mellow":
		basePath = config.Global.Common.MellowRuleBase
	case "quan":
		basePath = config.Global.Common.QuanRuleBase
	case "quanx":
		basePath = config.Global.Common.QuanXRuleBase
	case "loon":
		basePath = config.Global.Common.LoonRuleBase
	case "sssub":
		basePath = config.Global.Common.SSSubRuleBase
	case "singbox":
		basePath = config.Global.Common.SingBoxRuleBase
	default:
		return "", nil
	}

	if basePath == "" {
		return "", nil
	}

	// Resolve path relative to base directory
	// if !filepath.IsAbs(basePath) {
	// 	basePath = filepath.Join(config.GetBasePath(), basePath)
	// }

	data, err := os.ReadFile(basePath)
	if err != nil {
		return "", err
	}

	baseContent := string(data)

	// Apply template rendering with request context
	if config.Global.Template.TemplatePath != "" || strings.Contains(baseContent, "{{") {
		rendered, err := h.renderTemplateWithContext(baseContent, requestParams)
		if err == nil {
			baseContent = rendered
		}
	}

	return baseContent, nil
}

// renderTemplate renders template with global variables and request context
func (h *SubHandler) renderTemplate(content string) (string, error) {
	return h.renderTemplateWithContext(content, nil)
}

// renderTemplateWithContext renders template with request context
func (h *SubHandler) renderTemplateWithContext(content string, requestParams map[string]string) (string, error) {
	// Create template data map with support for nested keys
	data := make(map[string]interface{})

	// Add global template settings directly to root (for compatibility)
	for _, g := range config.Global.Template.Globals {
		setNestedValue(data, g.Key, g.Value)
	}

	// Add request parameters under "request" namespace (nil-safe range)
	for key, value := range requestParams {
		setNestedValue(data, "request."+key, value)
	}

	// Define template functions
	funcMap := template.FuncMap{
		"default": func(value interface{}, defaultValue string) string {
			if value == nil {
				return defaultValue
			}
			if str, ok := value.(string); ok {
				if str == "" {
					return defaultValue
				}
				return str
			}
			return defaultValue
		},
		"toBool": func(value interface{}) bool {
			if value == nil {
				return false
			}
			if str, ok := value.(string); ok {
				return str == "true" || str == "1" || str == "yes"
			}
			if b, ok := value.(bool); ok {
				return b
			}
			return false
		},
		"eq": func(a, b interface{}) bool {
			return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
		},
		"ne": func(a, b interface{}) bool {
			return fmt.Sprintf("%v", a) != fmt.Sprintf("%v", b)
		},
		"or": func(args ...interface{}) bool {
			for _, arg := range args {
				if b, ok := arg.(bool); ok && b {
					return true
				}
			}
			return false
		},
		"and": func(args ...interface{}) bool {
			for _, arg := range args {
				if b, ok := arg.(bool); !ok || !b {
					return false
				}
			}
			return true
		},
	}

	// Parse and execute template
	tmpl, err := template.New("base").Funcs(funcMap).Parse(content)
	if err != nil {
		return content, fmt.Errorf("template parse error: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return content, fmt.Errorf("template execute error: %w", err)
	}

	return buf.String(), nil
}

// setNestedValue sets a value in a nested map using dotted key notation
func setNestedValue(data map[string]interface{}, key string, value string) {
	keys := strings.Split(key, ".")
	if len(keys) == 1 {
		data[key] = value
		return
	}

	// Create nested structure
	current := data
	for i := 0; i < len(keys)-1; i++ {
		if _, ok := current[keys[i]]; !ok {
			current[keys[i]] = make(map[string]interface{})
		}
		if nested, ok := current[keys[i]].(map[string]interface{}); ok {
			current = nested
		} else {
			// Handle case where key exists but isn't a map
			return
		}
	}
	current[keys[len(keys)-1]] = value
}

// ExternalConfig represents external configuration
type ExternalConfig struct {
	ProxyGroups []config.ProxyGroupConfig
	Rulesets    []config.RulesetConfig
	BasePath    string
}

// loadExternalConfig loads external configuration from URL or file
func (h *SubHandler) loadExternalConfig(path string) (*ExternalConfig, error) {
	var data []byte

	// Determine source: http(s) or local file
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		resp, err := http.Get(path)
		if err != nil {
			log.Printf("[handler.loadExternalConfig] http fetch failed path=%s err=%v", path, err)
			return nil, err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			log.Printf("[handler.loadExternalConfig] http fetch path=%s status=%d", path, resp.StatusCode)
			return nil, fmt.Errorf("fetch external config status %d", resp.StatusCode)
		}
		data, err = io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("[handler.loadExternalConfig] http read failed path=%s err=%v", path, err)
			return nil, err
		}
	} else {
		// resolve candidate paths
		candidates := []string{path}
		if !filepath.IsAbs(path) {
			candidates = append(candidates, filepath.Join(config.GetBasePath(), path))
			candidates = append(candidates, filepath.Join(config.GetBasePath(), "config", path))
		}
		var readErr error
		for _, p := range candidates {
			if b, err := os.ReadFile(p); err == nil {
				data = b
				readErr = nil
				break
			} else {
				log.Printf("[handler.loadExternalConfig] file read failed candidate=%s err=%v", p, err)
				readErr = err
			}
		}
		if data == nil {
			log.Printf("[handler.loadExternalConfig] file candidates exhausted for path=%s lastErr=%v", path, readErr)
			return nil, fmt.Errorf("external config not found: %v", readErr)
		}
	}

	// Try YAML -> TOML -> INI using the Settings struct to leverage existing tags
	var extSettings config.Settings
	if err := yaml.Unmarshal(data, &extSettings); err == nil {
		return &ExternalConfig{
			ProxyGroups: extSettings.ProxyGroups.CustomProxyGroups,
			Rulesets:    extSettings.Rulesets.Rulesets,
			BasePath:    extSettings.Common.BasePath,
		}, nil
	}

	if _, err := toml.Decode(string(data), &extSettings); err == nil {
		return &ExternalConfig{
			ProxyGroups: extSettings.ProxyGroups.CustomProxyGroups,
			Rulesets:    extSettings.Rulesets.Rulesets,
			BasePath:    extSettings.Common.BasePath,
		}, nil
	}

	if cfg, err := ini.Load(data); err == nil {
		if err := cfg.MapTo(&extSettings); err == nil {
			return &ExternalConfig{
				ProxyGroups: extSettings.ProxyGroups.CustomProxyGroups,
				Rulesets:    extSettings.Rulesets.Rulesets,
				BasePath:    extSettings.Common.BasePath,
			}, nil
		}
	}

	// If all failed, return empty (non-nil) config to avoid breaking caller
	return &ExternalConfig{}, nil
}

// HandleVersion processes /version endpoint
func (h *SubHandler) HandleVersion(c *gin.Context) {
	c.String(http.StatusOK, "subconvergo v0.1.0 backend\n")
}

// HandleReadConf processes /readconf endpoint
func (h *SubHandler) HandleReadConf(c *gin.Context) {
	// Check token
	if config.Global.Common.APIAccessToken != "" {
		token := c.Query("token")
		if token != config.Global.Common.APIAccessToken {
			c.String(http.StatusForbidden, "Forbidden\n")
			return
		}
	}

	// Reload configuration
	if config, err := config.LoadConfig(); err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Failed to reload config: %v\n", err))
		return
	} else {

		c.String(http.StatusOK, "done, loaded "+config+"\n")
	}

}

// HandleGetRuleset processes /getruleset endpoint
func (h *SubHandler) HandleGetRuleset(c *gin.Context) {
	urlParam := c.Query("url")
	rulesetType := c.Query("type")

	if urlParam == "" || rulesetType == "" {
		c.String(http.StatusBadRequest, "Invalid request!")
		return
	}

	// URL decode
	decoded, err := base64.URLEncoding.DecodeString(urlParam)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid URL encoding")
		return
	}

	target := string(decoded)
	var content []byte

	// Remote fetch
	if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") {
		resp, err := http.Get(target)
		if err != nil {
			c.String(http.StatusBadRequest, fmt.Sprintf("Failed to fetch ruleset: %v", err))
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			c.String(http.StatusBadRequest, fmt.Sprintf("Failed to fetch ruleset: status %d", resp.StatusCode))
			return
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			c.String(http.StatusInternalServerError, fmt.Sprintf("Failed to read ruleset: %v", err))
			return
		}
		content = body
	} else {
		// Local path resolution attempts
		candidates := []string{
			target,
			filepath.Join(config.GetBasePath(), target),
			filepath.Join(config.GetBasePath(), "rules", target),
		}
		var readErr error
		for _, p := range candidates {
			if data, err := os.ReadFile(p); err == nil {
				content = data
				readErr = nil
				break
			} else {
				readErr = err
			}
		}
		if content == nil {
			c.String(http.StatusNotFound, fmt.Sprintf("Ruleset not found: %v", readErr))
			return
		}
	}

	// For now, return content as-is for supported types
	_ = rulesetType // placeholder for future conversions
	c.Data(http.StatusOK, "text/plain; charset=utf-8", content)
}

// HandleRender processes /render endpoint for template rendering
func (h *SubHandler) HandleRender(c *gin.Context) {
	// Check token if required
	if config.Global.Common.APIAccessToken != "" {
		token := c.Query("token")
		if token != config.Global.Common.APIAccessToken {
			c.String(http.StatusForbidden, "Forbidden\n")
			return
		}
	}

	// Get template path from query
	templatePath := c.Query("path")
	if templatePath == "" {
		c.String(http.StatusBadRequest, "Missing template path\n")
		return
	}

	// Resolve template path
	if !filepath.IsAbs(templatePath) {
		templatePath = filepath.Join(config.GetBasePath(), config.Global.Template.TemplatePath, templatePath)
	}

	// Read template file
	data, err := os.ReadFile(templatePath)
	if err != nil {
		c.String(http.StatusNotFound, fmt.Sprintf("Template not found: %v\n", err))
		return
	}

	// Render template
	rendered, err := h.renderTemplate(string(data))
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Failed to render template: %v\n", err))
		return
	}

	c.String(http.StatusOK, rendered)
}

// HandleGetProfile processes /getprofile endpoint
// Loads profile configuration files and merges parameters before calling /sub
func (h *SubHandler) HandleGetProfile(c *gin.Context) {
	name := c.Query("name")
	token := c.Query("token")

	// Validate required parameters
	if token == "" || name == "" {
		c.String(http.StatusForbidden, "Forbidden")
		return
	}

	// Support multiple profiles separated by |
	profiles := strings.Split(name, "|")
	if len(profiles) == 0 {
		c.String(http.StatusForbidden, "Forbidden")
		return
	}

	// Load first profile
	firstProfile := profiles[0]

	// Try multiple path resolutions
	var profilePath string
	basePath := config.GetBasePath()

	// Try different path combinations
	pathsToTry := []string{
		firstProfile,
		filepath.Join(basePath, firstProfile),
		filepath.Join("base", firstProfile),
		filepath.Join(basePath, "profiles", firstProfile+".ini"),
		filepath.Join("base", "profiles", firstProfile+".ini"),
		filepath.Join("profiles", firstProfile+".ini"),
	}

	for _, path := range pathsToTry {
		if fileExists(path) {
			profilePath = path
			break
		}
	}

	if profilePath == "" {
		c.String(http.StatusNotFound, "Profile not found")
		return
	} // Parse first profile
	// Load INI with custom options to preserve # in URLs
	cfg, err := ini.LoadSources(ini.LoadOptions{
		IgnoreInlineComment: true, // Don't treat # as inline comment
	}, profilePath)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Load profile failed! Reason: %v", err))
		return
	}

	// Check if Profile section exists
	if !cfg.HasSection("Profile") {
		c.String(http.StatusInternalServerError, "Broken profile!")
		return
	}

	profileSection := cfg.Section("Profile")
	if len(profileSection.Keys()) == 0 {
		c.String(http.StatusInternalServerError, "Broken profile!")
		return
	}

	// Build contents map from profile
	contents := make(map[string]string)
	for _, key := range profileSection.Keys() {
		contents[key.Name()] = key.String()
	}

	// Validate token
	profileToken, hasProfileToken := contents["profile_token"]
	if len(profiles) == 1 && hasProfileToken {
		// Single profile with its own token
		if token != profileToken {
			c.String(http.StatusForbidden, "Forbidden")
			return
		}
		token = config.Global.Common.APIAccessToken
	} else {
		// Multiple profiles or no profile token - use global token
		if token != config.Global.Common.APIAccessToken {
			c.String(http.StatusForbidden, "Forbidden")
			return
		}
	}

	// Merge URLs from all profiles
	allURLs := []string{}
	if urlVal, ok := contents["url"]; ok {
		allURLs = append(allURLs, strings.Split(urlVal, "|")...)
	}

	// If multiple profiles, merge them
	if len(profiles) > 1 {
		for i := 1; i < len(profiles); i++ {
			var additionalPath string
			profileName := profiles[i]

			// Try multiple path resolutions
			if fileExists(profileName) {
				additionalPath = profileName
			} else if fileExists(filepath.Join("base", profileName)) {
				additionalPath = filepath.Join("base", profileName)
			} else if fileExists(filepath.Join(config.GetBasePath(), profileName)) {
				additionalPath = filepath.Join(config.GetBasePath(), profileName)
			} else {
				continue
			}

			additionalCfg, err := ini.LoadSources(ini.LoadOptions{
				IgnoreInlineComment: true,
			}, additionalPath)
			if err != nil || !additionalCfg.HasSection("Profile") {
				continue
			}

			additionalSection := additionalCfg.Section("Profile")
			if urlKey := additionalSection.Key("url"); urlKey != nil {
				urlVal := urlKey.String()
				if urlVal != "" {
					allURLs = append(allURLs, strings.Split(urlVal, "|")...)
				}
			}
		}
	}

	// Update URL in contents
	if len(allURLs) > 0 {
		contents["url"] = strings.Join(allURLs, "|")
	}

	// Merge rename, exclude, include from all profiles
	allRenames := []string{}
	allExcludes := []string{}
	allIncludes := []string{}

	if renameVal, ok := contents["rename"]; ok {
		allRenames = append(allRenames, strings.Split(renameVal, "`")...)
	}
	if excludeVal, ok := contents["exclude"]; ok {
		allExcludes = append(allExcludes, strings.Split(excludeVal, "`")...)
	}
	if includeVal, ok := contents["include"]; ok {
		allIncludes = append(allIncludes, strings.Split(includeVal, "`")...)
	}

	// Merge from additional profiles
	if len(profiles) > 1 {
		for i := 1; i < len(profiles); i++ {
			var additionalPath string
			profileName := profiles[i]

			// Try multiple path resolutions
			if fileExists(profileName) {
				additionalPath = profileName
			} else if fileExists(filepath.Join("base", profileName)) {
				additionalPath = filepath.Join("base", profileName)
			} else if fileExists(filepath.Join(config.GetBasePath(), profileName)) {
				additionalPath = filepath.Join(config.GetBasePath(), profileName)
			} else {
				continue
			}

			additionalCfg, err := ini.LoadSources(ini.LoadOptions{
				IgnoreInlineComment: true,
			}, additionalPath)
			if err != nil || !additionalCfg.HasSection("Profile") {
				continue
			}

			additionalSection := additionalCfg.Section("Profile")
			if renameKey := additionalSection.Key("rename"); renameKey != nil {
				if val := renameKey.String(); val != "" {
					allRenames = append(allRenames, strings.Split(val, "`")...)
				}
			}
			if excludeKey := additionalSection.Key("exclude"); excludeKey != nil {
				if val := excludeKey.String(); val != "" {
					allExcludes = append(allExcludes, strings.Split(val, "`")...)
				}
			}
			if includeKey := additionalSection.Key("include"); includeKey != nil {
				if val := includeKey.String(); val != "" {
					allIncludes = append(allIncludes, strings.Split(val, "`")...)
				}
			}
		}
	}

	// Update merged values
	if len(allRenames) > 0 {
		contents["rename"] = strings.Join(allRenames, "`")
	}
	if len(allExcludes) > 0 {
		contents["exclude"] = strings.Join(allExcludes, "`")
	}
	if len(allIncludes) > 0 {
		contents["include"] = strings.Join(allIncludes, "`")
	}

	// Add token and profile_data
	contents["token"] = token

	// Build profile_data URL
	profileDataURL := config.Global.ManagedConfig.ManagedConfigPrefix + "/getprofile?" + c.Request.URL.RawQuery
	contents["profile_data"] = base64.StdEncoding.EncodeToString([]byte(profileDataURL))

	// Copy all original query parameters (query params override profile params)
	for key, values := range c.Request.URL.Query() {
		if len(values) > 0 && key != "name" {
			contents[key] = values[0]
		}
	}

	// Add token
	contents["token"] = token

	// Build merged params
	params := make(map[string]string)
	for key, value := range contents {
		params[key] = value
	}

	// Store params in context for handleSubWithParams to use
	c.Set("_merged_params", params)

	// Forward to /sub handler
	h.handleSubWithParams(c, params)
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// applyFilters applies include/exclude filters to proxies
