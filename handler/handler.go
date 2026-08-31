package handler

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"strings"
	"text/template"

	pc "github.com/gfunc/subconvergo/proxy/core"

	"github.com/BurntSushi/toml"
	"github.com/gfunc/subconvergo/cache"
	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/fetcher"
	"github.com/gfunc/subconvergo/generator"
	"github.com/gfunc/subconvergo/generator/core"
	"github.com/gfunc/subconvergo/generator/transformers"
	"github.com/gfunc/subconvergo/parser"
	parserutils "github.com/gfunc/subconvergo/parser/utils"
	"github.com/gfunc/subconvergo/utils"
	"github.com/gfunc/subconvergo/version"
	"github.com/gin-gonic/gin"
	"gopkg.in/ini.v1"
	"gopkg.in/yaml.v3"
)

// SubHandler handles subscription conversion requests
type SubHandler struct{}

const (
	// maxSubSourceURLs caps the number of |-separated source URLs per /sub request.
	maxSubSourceURLs = 32
	// maxSubSourceURLLength bounds each source URL's length in bytes.
	maxSubSourceURLLength = 2048
)

// NewSubHandler creates a new subscription handler
func NewSubHandler() *SubHandler {
	return &SubHandler{}
}

// RequestParams holds the parameters for subscription conversion
type RequestParams struct {
	Target       string `form:"target"`
	URL          string `form:"url"`
	Config       string `form:"config"`
	UserAgent    string `form:"ua"`
	Group        string `form:"group"`
	Include      string `form:"include"`
	Exclude      string `form:"exclude"`
	UDP          *bool  `form:"udp"`
	TFO          *bool  `form:"tfo"`
	SCV          *bool  `form:"scv"`
	NewName      *bool  `form:"new_name"`
	SurgeVer     *int   `form:"ver"`
	IgnoreSource *bool  `form:"ignore_source"`
}

// HandleSub processes /sub endpoint
func (h *SubHandler) HandleSub(c *gin.Context) {
	var params RequestParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.String(http.StatusBadRequest, "Invalid parameters: "+err.Error())
		return
	}
	h.processSubRequest(c, &params)
}

// processSubRequest processes /sub with parsed parameters
func (h *SubHandler) processSubRequest(c *gin.Context, params *RequestParams) {
	target := params.Target
	urlParam := params.URL
	configParam := params.Config
	uaParam := params.UserAgent

	log.Printf("[handler.HandleSub] Request received: target=%s url=%s config=%s ua=%s client=%s %s", target, utils.RedactURL(urlParam), utils.RedactURLOrPath(configParam), uaParam, c.ClientIP(), utils.HeaderSummary(c.Request.Header))

	// Handle auto target detection
	var clashNewName *bool
	var surgeVer int = -1

	if target == "auto" {
		ua := c.Request.Header.Get("User-Agent")
		matchedTarget, cnn, sv := matchUserAgent(ua)
		if matchedTarget != "" {
			target = matchedTarget
			clashNewName = cnn
			surgeVer = sv
			log.Printf("[handler.HandleSub] Auto-detected target=%s from UA=%s", target, ua)
		}
	}

	// Validate required parameters
	if target == "" {
		c.String(http.StatusBadRequest, "Invalid target!")
		return
	}

	// Handle target aliases
	switch target {
	case "quan":
		target = "quanx"
	case "clashr":
		target = "clash"
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

	// Bound the number and size of |-separated source URLs before any fetch.
	sourceURLs := strings.Split(urlParam, "|")
	if len(sourceURLs) > maxSubSourceURLs {
		c.String(http.StatusBadRequest, fmt.Sprintf("Too many source URLs (max %d)", maxSubSourceURLs))
		return
	}
	for _, u := range sourceURLs {
		if len(u) > maxSubSourceURLLength {
			c.String(http.StatusBadRequest, fmt.Sprintf("Source URL exceeds max length of %d", maxSubSourceURLLength))
			return
		}
	}

	// Note: gin has already URL-decoded the query parameters once. Do NOT
	// unescape urlParam again here — a second decode turns percent-encoded
	// bytes inside the subscription URL (e.g. name=%E4%BD%8F...) into raw
	// UTF-8 on the wire, which strict frontends like Cloudflare reject.

	// Reload config on request if enabled
	if config.Global.Common.ReloadConfOnRequest {
		if _, err := config.ReloadConfig(); err == nil {
			// Config reloaded successfully
		}
	}

	// Handle insert URLs first if enabled
	var urlsToProcess []string
	if config.Global.Common.EnableInsert && len(config.Global.Common.InsertURL) > 0 {
		if config.Global.Common.PrependInsertURL {
			urlsToProcess = append(urlsToProcess, config.Global.Common.InsertURL...)
		}
	}

	// Add main URLs
	urlsToProcess = append(urlsToProcess, sourceURLs...)

	// Append insert URLs if needed
	if config.Global.Common.EnableInsert && len(config.Global.Common.InsertURL) > 0 {
		if !config.Global.Common.PrependInsertURL {
			urlsToProcess = append(urlsToProcess, config.Global.Common.InsertURL...)
		}
	}
	log.Printf("[handler.HandleSub] target=%s urls=%d urlLen=%d config=%s client=%s", target, len(urlsToProcess), len(urlParam), utils.RedactURLOrPath(configParam), c.ClientIP())

	// Create request-scoped config initialized with global settings
	reqConfig := *config.Global

	// Load external config if specified
	if configParam != "" {
		// Load external config (can be URL or file path)
		extConfig, err := h.loadExternalConfig(configParam)
		if err != nil {
			log.Printf("[handler.HandleSub] failed to load external config %s: %v", utils.RedactURLOrPath(configParam), err)
		} else if extConfig != nil {
			log.Printf("[handler.HandleSub] loaded external config %s", utils.RedactURLOrPath(configParam))
			// Merge external config into request config
			reqConfig.Merge(extConfig)
		}
	}

	ignoreSource := true
	if reqConfig.Common.IgnoreSource != nil {
		ignoreSource = *reqConfig.Common.IgnoreSource
	}
	if params.IgnoreSource != nil {
		ignoreSource = *params.IgnoreSource
	}

	// Parse subscription URLs (support multiple URLs separated by |)
	var allProxies []pc.ProxyInterface
	var otherProxyGroups []config.ProxyGroupConfig
	var rawRules []string
	for index, url := range urlsToProcess {
		url = strings.TrimSpace(url)
		if url == "" {
			continue
		}
		sp := &parser.SubParser{
			Index:     index,
			URL:       url,
			Proxy:     reqConfig.Common.ProxySubscription,
			UserAgent: uaParam,
		}
		custom, err := sp.Parse()
		if err == nil {
			log.Printf("[handler.HandleSub] Parsed URL %s: %d proxies, %d groups, %d rules", utils.RedactURL(url), len(custom.Proxies), len(custom.Groups), len(custom.RawRules))
			allProxies = append(allProxies, custom.Proxies...)
			if !ignoreSource {
				otherProxyGroups = append(otherProxyGroups, custom.Groups...)
				rawRules = append(rawRules, custom.RawRules...)
			}
			continue
		} else if !reqConfig.Advanced.SkipFailedLinks {
			c.String(http.StatusBadRequest, fmt.Sprintf("Failed to parse subscription (%s): %v", url, err))
			return
		} else {
			log.Printf("[handler.HandleSub] failed to parse subscription (index=%d url=%s): %v", index, utils.RedactURL(url), err)
		}
	}

	log.Printf("[handler.HandleSub] Parsed %d proxies, %d groups from %d URLs", len(allProxies), len(reqConfig.ProxyGroups.CustomProxyGroups), len(urlsToProcess))

	if len(allProxies) == 0 {
		log.Printf("[handler.HandleSub] no valid proxies parsed from %d url(s)", len(urlsToProcess))
		c.String(http.StatusBadRequest, "No valid proxies found")
		return
	}

	// Check custom group name override. The query param bypasses the parse
	// boundary, so it goes through the same scalar sanitizer here.
	if params.Group != "" {
		group := parserutils.SanitizeScalarField(params.Group)
		for _, p := range allProxies {
			p.SetGroup(group)
		}
	}

	// Prepare filter patterns
	include := params.Include
	exclude := params.Exclude

	var includePatterns []string
	if len(reqConfig.Common.IncludeRemarks) > 0 {
		includePatterns = append(includePatterns, reqConfig.Common.IncludeRemarks...)
	}
	if include != "" {
		includePatterns = append(includePatterns, include)
	}

	var excludePatterns []string
	if len(reqConfig.Common.ExcludeRemarks) > 0 {
		excludePatterns = append(excludePatterns, reqConfig.Common.ExcludeRemarks...)
	}
	if exclude != "" {
		excludePatterns = append(excludePatterns, exclude)
	}

	// Construct transformation pipeline
	pipeline := []transformers.Transformer{
		transformers.NewFilterTransformer(includePatterns, excludePatterns),
		transformers.NewRenameTransformer(reqConfig.NodePref.RenameNodes),
		transformers.NewEmojiTransformer(reqConfig.Emojis),
		transformers.NewSortTransformer(reqConfig.NodePref.SortFlag),
		transformers.NewDeduplicateTransformer(),
	}

	proxyGroups := reqConfig.ProxyGroups.CustomProxyGroups
	if len(otherProxyGroups) > 0 {
		proxyGroups = append(proxyGroups, otherProxyGroups...)
	}

	// Deduplicate proxy groups
	uniqueGroups := make([]config.ProxyGroupConfig, 0, len(proxyGroups))
	seenGroups := make(map[string]bool)
	for _, g := range proxyGroups {
		if !seenGroups[g.Name] {
			uniqueGroups = append(uniqueGroups, g)
			seenGroups[g.Name] = true
		}
	}
	proxyGroups = uniqueGroups

	// Prepare generator options
	rulesets, rawRules := enforceAdvancedCaps(reqConfig.Rulesets.Rulesets, rawRules)
	opts := core.GeneratorOptions{
		Target:                 target,
		ProxyGroups:            proxyGroups,
		Rulesets:               rulesets,
		RawRules:               rawRules,
		AppendProxyType:        reqConfig.Common.AppendProxyType,
		EnableRuleGen:          reqConfig.Rulesets.Enabled,
		OverwriteOriginalRules: false, // Default to false, will be updated below
		Pipelines:              pipeline,

		ProxySetting: config.ProxySetting{
			ClashProxiesStyle:   reqConfig.NodePref.ClashProxiesStyle,
			ClashGroupsStyle:    reqConfig.NodePref.ClashProxyGroupsStyle,
			SingBoxAddClashMode: reqConfig.NodePref.SingBoxAddClashModes,
			ClashUseNewField:    reqConfig.NodePref.ClashUseNewField,
		},
	}

	if reqConfig.Rulesets.OverwriteOriginalRules != nil {
		opts.OverwriteOriginalRules = *reqConfig.Rulesets.OverwriteOriginalRules
	}

	// Apply auto-detected settings
	if clashNewName != nil {
		opts.ProxySetting.ClashUseNewField = *clashNewName
	}
	if surgeVer != -1 {
		opts.ProxySetting.SurgeVer = surgeVer
	}

	// Apply node preferences to generator options
	opts.UDP = reqConfig.NodePref.UDPFlag
	opts.TFO = reqConfig.NodePref.TCPFastOpenFlag
	opts.SCV = reqConfig.NodePref.SkipCertVerifyFlag
	opts.TLS13 = reqConfig.NodePref.TLS13Flag

	// Parse boolean options
	if params.UDP != nil {
		opts.UDP = params.UDP
	}
	if params.TFO != nil {
		opts.TFO = params.TFO
	}
	if params.SCV != nil {
		opts.SCV = params.SCV
	}
	if params.NewName != nil {
		opts.ProxySetting.ClashUseNewField = *params.NewName
	}
	if params.SurgeVer != nil {
		opts.ProxySetting.SurgeVer = *params.SurgeVer
	}

	// Default Surge version to 3 if not set
	if opts.ProxySetting.SurgeVer == 0 {
		opts.ProxySetting.SurgeVer = 3
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
	baseConfig, err := h.loadBaseConfig(target, requestParams, &reqConfig)
	if err != nil {
		log.Printf("[handler.HandleSub] loadBaseConfig target=%s err=%v", target, err)
		c.String(http.StatusInternalServerError, fmt.Sprintf("Failed to load base config: %v", err))
		return
	}

	// Generate output
	output, err := generator.Generate(allProxies, opts, baseConfig)
	if err != nil {
		log.Printf("[handler.HandleSub] generator failed target=%s proxies=%d err=%v", target, len(allProxies), err)
		if err.Error() == "No valid proxies found" {
			c.String(http.StatusBadRequest, err.Error())
		} else {
			c.String(http.StatusInternalServerError, fmt.Sprintf("Failed to generate config: %v", err))
		}
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

func (h *SubHandler) loadBaseConfig(target string, requestParams map[string]string, reqConfig *config.Settings) (string, error) {
	var basePath string

	// Use request config which might have been overridden
	switch target {
	case "clash", "clashr":
		basePath = reqConfig.Common.ClashRuleBase
	case "surge":
		basePath = reqConfig.Common.SurgeRuleBase
	case "surfboard":
		basePath = reqConfig.Common.SurfboardRuleBase
	case "mellow":
		basePath = reqConfig.Common.MellowRuleBase
	case "quan":
		basePath = reqConfig.Common.QuanRuleBase
	case "quanx":
		basePath = reqConfig.Common.QuanXRuleBase
	case "loon":
		basePath = reqConfig.Common.LoonRuleBase
	case "sssub":
		basePath = reqConfig.Common.SSSubRuleBase
	case "singbox":
		basePath = reqConfig.Common.SingBoxRuleBase
	default:
		return "", nil
	}

	if basePath == "" {
		return "", nil
	}

	// Resolve path relative to the configuration directory and forbid escapes.
	root := config.GetConfigDir()
	resolved, err := utils.ResolveUnderRoot(basePath, root)
	if err != nil {
		return "", fmt.Errorf("invalid rule base path %q: %w", basePath, err)
	}

	data, err := os.ReadFile(resolved)
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
		// Also add to "global" namespace
		setNestedValue(data, "global."+g.Key, g.Value)
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

// loadExternalConfig loads external configuration from URL or file
func (h *SubHandler) loadExternalConfig(path string) (*config.Settings, error) {
	var data []byte

	// Remote configs come from request-controlled URLs and are untrusted:
	// they must not gain filesystem capabilities (local file imports).
	// Local configs are admin-controlled and trusted to use !!import:.
	isRemote := strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://")

	// Determine source: http(s) or local file
	if isRemote {
		// Check cache first
		cacheKey := ""
		if config.Global.Advanced.EnableCache {
			cacheKey = cache.GlobalManager.GetKey(path)
			if cachedData, ok := cache.GlobalManager.Get(cacheKey, config.Global.Advanced.CacheConfig); ok {
				log.Printf("[handler.loadExternalConfig] path=%s served from cache", utils.RedactURL(path))
				data = cachedData
			}
		}

		if data == nil {
			body, err := fetcher.ForConfig(config.Global.Advanced.MaxAllowedDownloadSize, "").Get(path, nil)
			if err != nil {
				log.Printf("[handler.loadExternalConfig] http fetch failed path=%s err=%v", utils.RedactURL(path), err)
				// Try stale cache
				if config.Global.Advanced.EnableCache && config.Global.Advanced.ServeCacheOnFetchFail {
					if cachedData, ok := cache.GlobalManager.GetStale(cacheKey); ok {
						log.Printf("[handler.loadExternalConfig] path=%s served from stale cache", utils.RedactURL(path))
						data = cachedData
					} else {
						return nil, err
					}
				} else {
					return nil, err
				}
			} else {
				data = body
				// Save to cache
				if config.Global.Advanced.EnableCache {
					if err := cache.GlobalManager.Set(cacheKey, data); err != nil {
						log.Printf("[handler.loadExternalConfig] failed to save cache: %v", err)
					}
				}
			}
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
		if isRemote {
			mergeExternalConfigCollections(&extSettings)
		} else if err := extSettings.ProcessImports(); err != nil {
			log.Printf("[handler.loadExternalConfig] failed to process imports in YAML: %v", err)
		}
		return &extSettings, nil
	}

	if _, err := toml.Decode(string(data), &extSettings); err == nil {
		if isRemote {
			mergeExternalConfigCollections(&extSettings)
		} else if err := extSettings.ProcessImports(); err != nil {
			log.Printf("[handler.loadExternalConfig] failed to process imports in TOML: %v", err)
		}
		return &extSettings, nil
	}

	if cfg, err := ini.LoadSources(ini.LoadOptions{AllowShadows: true}, data); err == nil {
		if err := cfg.MapTo(&extSettings); err == nil {
			// Manually parse custom_proxy_group and ruleset if MapTo didn't pick them up
			// (e.g. due to struct tags or section name mismatches)

			if len(extSettings.CustomGroups) == 0 {
				extSettings.CustomGroups = parseINICustomGroups(cfg, !isRemote)
			}

			if len(extSettings.CustomRulesets) == 0 {
				extSettings.CustomRulesets = parseINIRulesets(cfg, !isRemote)
			}

			// Check for overwrite_original_rules in rulesets section manually
			if section, err := cfg.GetSection("rulesets"); err == nil {
				if section.HasKey("overwrite_original_rules") {
					val := section.Key("overwrite_original_rules").MustBool(false)
					extSettings.Rulesets.OverwriteOriginalRules = &val
				}
			}

			return &extSettings, nil
		}
	}

	// If all failed, return empty (non-nil) config to avoid breaking caller
	return &config.Settings{}, nil
}

// mergeExternalConfigCollections applies only the in-memory collection merges
// of ProcessImports for remote (untrusted) external configs, skipping every
// filesystem import directive. Import fields are left in place but never
// processed, so a remote config cannot read local files.
func mergeExternalConfigCollections(s *config.Settings) {
	if len(s.CustomGroups) > 0 {
		s.ProxyGroups.CustomProxyGroups = append(s.ProxyGroups.CustomProxyGroups, s.CustomGroups...)
		s.CustomGroups = nil
	}
	if len(s.CustomRulesets) > 0 {
		s.Rulesets.Rulesets = append(s.Rulesets.Rulesets, s.CustomRulesets...)
		s.CustomRulesets = nil
	}
}

// enforceAdvancedCaps applies the advanced max_allowed_rulesets /
// max_allowed_rules limits (0 or negative = unlimited) by truncation: excess
// entries are dropped and logged rather than failing the request. This is the
// single funnel where request- and config-supplied rulesets/rules enter
// generation; note max_allowed_rules caps the inline rule list (ruleset file
// contents are fetched later inside the generators and are not counted here).
func enforceAdvancedCaps(rulesets []config.RulesetConfig, rawRules []string) ([]config.RulesetConfig, []string) {
	if max := config.Global.Advanced.MaxAllowedRulesets; max > 0 && len(rulesets) > max {
		log.Printf("[handler.enforceAdvancedCaps] truncating %d rulesets to max_allowed_rulesets=%d", len(rulesets), max)
		rulesets = rulesets[:max]
	}
	if max := config.Global.Advanced.MaxAllowedRules; max > 0 && len(rawRules) > max {
		log.Printf("[handler.enforceAdvancedCaps] truncating %d raw rules to max_allowed_rules=%d", len(rawRules), max)
		rawRules = rawRules[:max]
	}
	return rulesets, rawRules
}

// HandleVersion processes /version endpoint
func (h *SubHandler) HandleVersion(c *gin.Context) {
	log.Printf("[handler.HandleVersion] Request received client=%s %s", c.ClientIP(), utils.HeaderSummary(c.Request.Header))
	c.String(http.StatusOK, fmt.Sprintf("subconvergo v%s backend\n", version.Version))
}

// HandleReadConf processes /readconf endpoint
func (h *SubHandler) HandleReadConf(c *gin.Context) {
	log.Printf("[handler.HandleReadConf] Request received client=%s %s", c.ClientIP(), utils.HeaderSummary(c.Request.Header))

	// Token gate: fails closed when no token is configured.
	if !requireAPIToken(c) {
		return
	}

	// Reload configuration
	if config, err := config.ReloadConfig(); err != nil {
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

	log.Printf("[handler.HandleGetRuleset] Request received: url=%s type=%s client=%s %s", utils.RedactURL(urlParam), rulesetType, c.ClientIP(), utils.HeaderSummary(c.Request.Header))

	// Token gate (consistent with other protected endpoints): fails closed
	// when no token is configured.
	if !requireAPIToken(c) {
		return
	}

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
		// Check cache first
		cacheKey := ""
		if config.Global.Advanced.EnableCache {
			cacheKey = cache.GlobalManager.GetKey(target)
			if cachedData, ok := cache.GlobalManager.Get(cacheKey, config.Global.Advanced.CacheRuleset); ok {
				log.Printf("[handler.HandleGetRuleset] url=%s served from cache", utils.RedactURL(target))
				content = cachedData
			}
		}

		if content == nil {
			body, err := fetcher.ForConfig(config.Global.Advanced.MaxAllowedDownloadSize, "").Get(target, nil)
			if err != nil {
				log.Printf("[handler.HandleGetRuleset] fetch failed url=%s err=%v", utils.RedactURL(target), err)
				// Try stale cache
				if config.Global.Advanced.EnableCache && config.Global.Advanced.ServeCacheOnFetchFail {
					if cachedData, ok := cache.GlobalManager.GetStale(cacheKey); ok {
						log.Printf("[handler.HandleGetRuleset] url=%s served from stale cache", utils.RedactURL(target))
						content = cachedData
					} else {
						c.String(http.StatusBadRequest, fmt.Sprintf("Failed to fetch ruleset: %v", err))
						return
					}
				} else {
					c.String(http.StatusBadRequest, fmt.Sprintf("Failed to fetch ruleset: %v", err))
					return
				}
			} else {
				content = body
				// Save to cache
				if config.Global.Advanced.EnableCache {
					if err := cache.GlobalManager.Set(cacheKey, content); err != nil {
						log.Printf("[handler.HandleGetRuleset] failed to save cache: %v", err)
					}
				}
			}
		}
	} else {
		// Local path resolution attempts (restricted to base path)
		p, err := utils.ResolveRulesetPath(target)
		if err != nil {
			c.String(http.StatusBadRequest, fmt.Sprintf("Ruleset not found: %v", err))
			return
		}
		data, err := os.ReadFile(p)
		if err != nil {
			c.String(http.StatusNotFound, fmt.Sprintf("Ruleset not found: %v", err))
			return
		}
		content = data
	}

	// For now, return content as-is for supported types
	_ = rulesetType // placeholder for future conversions
	c.Data(http.StatusOK, "text/plain; charset=utf-8", content)
}

// HandleRender processes /render endpoint for template rendering
func (h *SubHandler) HandleRender(c *gin.Context) {
	// Token gate: fails closed when no token is configured.
	if !requireAPIToken(c) {
		return
	}

	// Get template path from query
	templatePath := c.Query("path")
	log.Printf("[handler.HandleRender] Request received: path=%s client=%s %s", templatePath, c.ClientIP(), utils.HeaderSummary(c.Request.Header))

	if templatePath == "" {
		c.String(http.StatusBadRequest, "Missing template path\n")
		return
	}

	// Resolve template root under base_path.
	basePath := config.GetBasePath()
	templateRoot := basePath
	if config.Global.Template.TemplatePath != "" {
		root, err := utils.ResolveUnderRoot(config.Global.Template.TemplatePath, basePath)
		if err != nil {
			c.String(http.StatusInternalServerError, "Invalid template configuration\n")
			return
		}
		templateRoot = root
	}

	// Resolve and validate user-supplied template path.
	resolved, err := utils.ResolveUnderRoot(templatePath, templateRoot)
	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("Invalid template path: %v\n", err))
		return
	}

	// Read template file
	data, err := os.ReadFile(resolved)
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

	log.Printf("[handler.HandleGetProfile] Request received: name=%s client=%s %s", name, c.ClientIP(), utils.HeaderSummary(c.Request.Header))

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

	// Resolve the first profile under the configured profile roots.
	// A bare profile name is resolved to <root>/profiles/<name>.ini. A relative
	// path (e.g. profiles/gfunc.ini) is resolved directly under the root.
	// base_path is tried first for backward compatibility, then the pref directory.
	profileRoots := []string{
		config.GetBasePath(),
		config.GetConfigDir(),
	}

	firstProfile, err := resolveProfilePath(profiles[0], profileRoots)
	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("Invalid profile name: %v", err))
		return
	}
	if firstProfile == "" {
		c.String(http.StatusNotFound, "Profile not found")
		return
	}
	profilePath := firstProfile

	// Parse first profile
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

	// Validate token (constant-time; empty configured tokens never match)
	profileToken, hasProfileToken := contents["profile_token"]
	if len(profiles) == 1 && hasProfileToken {
		// Single profile with its own token
		if !tokenEqual(token, profileToken) {
			c.String(http.StatusForbidden, "Forbidden")
			return
		}
		token = config.Global.Common.APIAccessToken
	} else {
		// Multiple profiles or no profile token - use global token
		if !tokenEqual(token, config.Global.Common.APIAccessToken) {
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
			profileName := profiles[i]

			additionalPath, err := resolveProfilePath(profileName, profileRoots)
			if err != nil || additionalPath == "" {
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

	// Merge URLs passed via query parameter with profile URLs
	if queryURL := c.Query("url"); queryURL != "" {
		allURLs = append(allURLs, strings.Split(queryURL, "|")...)
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
			profileName := profiles[i]

			additionalPath, err := resolveProfilePath(profileName, profileRoots)
			if err != nil || additionalPath == "" {
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

	// Copy all original query parameters (query params override profile params,
	// except "url" which was merged into the profile URLs above)
	for key, values := range c.Request.URL.Query() {
		if len(values) > 0 && key != "name" && key != "url" {
			contents[key] = values[0]
		}
	}

	// Add token
	contents["token"] = token

	// Build merged params
	paramsMap := make(map[string]string)
	for key, value := range contents {
		paramsMap[key] = value
	}

	// Convert map to RequestParams
	params := &RequestParams{
		Target:    paramsMap["target"],
		URL:       paramsMap["url"],
		Config:    paramsMap["config"],
		UserAgent: paramsMap["ua"],
		Group:     paramsMap["group"],
		Include:   paramsMap["include"],
		Exclude:   paramsMap["exclude"],
	}

	if v, ok := paramsMap["udp"]; ok {
		b := v == "true"
		params.UDP = &b
	}
	if v, ok := paramsMap["tfo"]; ok {
		b := v == "true"
		params.TFO = &b
	}
	if v, ok := paramsMap["scv"]; ok {
		b := v == "true"
		params.SCV = &b
	}
	if v, ok := paramsMap["new_name"]; ok {
		b := v == "true"
		params.NewName = &b
	}
	if v, ok := paramsMap["ver"]; ok {
		if i, err := strconv.Atoi(v); err == nil {
			params.SurgeVer = &i
		}
	}
	if v, ok := paramsMap["ignore_source"]; ok {
		b := v == "true"
		params.IgnoreSource = &b
	}

	// Forward to /sub handler
	h.processSubRequest(c, params)
}

// HandleSurge2Clash processes /surge2clash endpoint
func (h *SubHandler) HandleSurge2Clash(c *gin.Context) {
	log.Printf("[handler.HandleSurge2Clash] Request received client=%s %s", c.ClientIP(), utils.HeaderSummary(c.Request.Header))
	var params RequestParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.String(http.StatusBadRequest, "Invalid parameters: "+err.Error())
		return
	}
	// Force target to clash
	params.Target = "clash"
	h.processSubRequest(c, &params)
}

// HandleFlushCache processes /flushcache endpoint
func (h *SubHandler) HandleFlushCache(c *gin.Context) {
	log.Printf("[handler.HandleFlushCache] Request received client=%s %s", c.ClientIP(), utils.HeaderSummary(c.Request.Header))
	// Token gate: fails closed when no token is configured.
	if !requireAPIToken(c) {
		return
	}

	if err := cache.GlobalManager.Flush(); err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Failed to flush cache: %v\n", err))
		return
	}

	c.String(http.StatusOK, "Cache flushed\n")
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// resolveProfilePath validates a profile name and returns the path to the
// profile file under one of the candidate roots.
//
// If name contains a path separator it is treated as a relative path and
// resolved directly under each root (with an optional .ini suffix). Otherwise
// it is treated as a bare profile name and resolved to <root>/profiles/<name>.ini.
func resolveProfilePath(name string, candidateRoots []string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("empty profile name")
	}

	// Relative profile path (e.g. "profiles/gfunc.ini").
	if strings.ContainsRune(name, '/') || strings.ContainsRune(name, '\\') {
		for _, root := range candidateRoots {
			for _, tryName := range []string{name, name + ".ini"} {
				resolved, err := utils.ResolveUnderRoot(tryName, root)
				if err != nil {
					continue
				}
				if _, err := os.Stat(resolved); err == nil {
					return resolved, nil
				}
			}
		}
		return "", nil // not found, but not an error
	}

	// Bare profile name: look in <root>/profiles/<name>.ini.
	name = strings.TrimSuffix(name, ".ini")
	for _, root := range candidateRoots {
		profilesDir := filepath.Join(root, "profiles")
		resolved, err := utils.ResolveUnderRoot(name+".ini", profilesDir)
		if err != nil {
			continue
		}
		if _, err := os.Stat(resolved); err == nil {
			return resolved, nil
		}
	}
	return "", nil // not found, but not an error
}

// applyFilters applies include/exclude filters to proxies

// importINILines loads a !!import: target from a local (trusted) INI config:
// the path is confined to the configured base directory via the shared
// canonical resolver (no CWD fallback) and the file's lines are returned.
// Remote-fetched configs are untrusted — with allowImport=false the import is
// refused and nil is returned. logTag identifies the calling parser in logs.
func importINILines(path string, allowImport bool, logTag string) []string {
	if !allowImport {
		log.Printf("[%s] refused local import from remote config: %s", logTag, path)
		return nil
	}

	fullPath, err := utils.ResolveUnderRoot(path, config.GetBasePath())
	if err != nil {
		log.Printf("[%s] refused import path %s: %v", logTag, path, err)
		return nil
	}

	content, err := os.ReadFile(fullPath)
	if err != nil {
		log.Printf("[%s] failed to import %s: %v", logTag, fullPath, err)
		return nil
	}

	return strings.Split(string(content), "\n")
}

func parseINICustomGroups(cfg *ini.File, allowImport bool) []config.ProxyGroupConfig {
	var groups []config.ProxyGroupConfig

	// Helper to parse a single line
	var parseLine func(value string)
	parseLine = func(value string) {
		value = strings.TrimSpace(value)
		if value == "" || strings.HasPrefix(value, ";") || strings.HasPrefix(value, "#") {
			return
		}

		if strings.HasPrefix(value, "!!import:") {
			path := strings.TrimPrefix(value, "!!import:")
			for _, line := range importINILines(path, allowImport, "parseINICustomGroups") {
				parseLine(line)
			}
			return
		}

		// Format: Name`Type`Rule...
		if strings.Contains(value, "`") {
			parts := strings.Split(value, "`")
			if len(parts) >= 2 {
				name := parts[0]
				groupType := parts[1]
				proxies := parts[2:]

				groups = append(groups, config.ProxyGroupConfig{
					Name:    name,
					Type:    groupType,
					Proxies: proxies,
				})
			}
		}
	}

	// 1. Check [custom_proxy_group] section (Key is Name, Value is Type,Proxies)
	if section, err := cfg.GetSection("custom_proxy_group"); err == nil {
		for _, key := range section.Keys() {
			name := key.Name()
			value := key.String()

			// Format: type,content
			parts := strings.SplitN(value, ",", 2)
			if len(parts) >= 1 {
				groupType := parts[0]
				content := ""
				if len(parts) > 1 {
					content = parts[1]
				}

				proxies := []string{}
				if content != "" {
					proxies = strings.Split(content, ",")
				}

				groups = append(groups, config.ProxyGroupConfig{
					Name:    name,
					Type:    groupType,
					Proxies: proxies,
				})
			}
		}
	}

	// 2. Check custom_proxy_group keys in [proxy_groups] and [common]
	for _, secName := range []string{"proxy_groups", "common"} {
		if section, err := cfg.GetSection(secName); err == nil {
			if key, err := section.GetKey("custom_proxy_group"); err == nil {
				for _, value := range key.ValueWithShadows() {
					parseLine(value)
				}
			}
		}
	}

	return groups
}

func parseINIRulesets(cfg *ini.File, allowImport bool) []config.RulesetConfig {
	var rulesets []config.RulesetConfig

	var parseLine func(value string)
	parseLine = func(value string) {
		value = strings.TrimSpace(value)
		if value == "" || strings.HasPrefix(value, ";") || strings.HasPrefix(value, "#") {
			return
		}

		if strings.HasPrefix(value, "!!import:") {
			path := strings.TrimPrefix(value, "!!import:")
			for _, line := range importINILines(path, allowImport, "parseINIRulesets") {
				parseLine(line)
			}
			return
		}

		parts := strings.Split(value, ",")
		if len(parts) < 2 {
			// Check if it's a single rule line (e.g. MATCH,DIRECT)
			// If it is, treat it as inline rule
			if isRuleType(strings.ToUpper(parts[0])) {
				rulesets = append(rulesets, config.RulesetConfig{
					Rule: "[]" + value,
				})
			}
			return
		}

		// Check for implicit inline rule (starts with rule type)
		firstPart := strings.ToUpper(strings.TrimSpace(parts[0]))
		if isRuleType(firstPart) {
			rulesets = append(rulesets, config.RulesetConfig{
				Rule: "[]" + value,
			})
			return
		}

		rc := config.RulesetConfig{
			Group: strings.TrimSpace(parts[0]),
		}

		// Check if last part is an integer (interval)
		var ruleParts []string
		if len(parts) > 2 {
			lastPart := strings.TrimSpace(parts[len(parts)-1])
			if interval, err := strconv.Atoi(lastPart); err == nil {
				rc.Interval = interval
				ruleParts = parts[1 : len(parts)-1]
			} else {
				ruleParts = parts[1:]
			}
		} else {
			ruleParts = parts[1:]
		}

		ruleContent := strings.Join(ruleParts, ",")
		ruleContent = strings.TrimSpace(ruleContent)

		if strings.HasPrefix(ruleContent, "[]") {
			rc.Rule = ruleContent
		} else {
			rc.Ruleset = ruleContent
		}

		rulesets = append(rulesets, rc)
	}

	if section, err := cfg.GetSection("rulesets"); err == nil {
		if key, err := section.GetKey("ruleset"); err == nil {
			for _, value := range key.ValueWithShadows() {
				parseLine(value)
			}
		}
	}
	return rulesets
}

func isRuleType(t string) bool {
	switch t {
	case "DOMAIN", "DOMAIN-SUFFIX", "DOMAIN-KEYWORD", "DOMAIN-SET",
		"IP-CIDR", "IP-CIDR6", "GEOIP", "MATCH", "PROCESS-NAME",
		"DST-PORT", "SRC-PORT", "IN-PORT", "SRC-IP-CIDR", "SCRIPT":
		return true
	}
	return false
}
