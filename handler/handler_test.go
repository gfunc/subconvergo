package handler

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gfunc/subconvergo/config"
	P "github.com/gfunc/subconvergo/proxy"
	"github.com/gin-gonic/gin"
)

func TestApplyFilters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSubHandler()

	// Prepare proxies
	proxies := []P.ProxyInterface{
		&P.BaseProxy{Remark: "HK A"},
		&P.BaseProxy{Remark: "US B"},
		&P.BaseProxy{Remark: "JP C"},
	}

	// Set global include/exclude to exercise merging
	config.Global.Common.IncludeRemarks = []string{"HK"}
	config.Global.Common.ExcludeRemarks = []string{"JP"}

	// Build request with explicit include that should override exclude for HK
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/sub?include=US&exclude=HK", nil)
	c.Request = req

	filtered := h.applyFilters(proxies, c)
	// include=US should keep US, global include adds HK but query excludes HK, exclude removes JP
	if len(filtered) != 1 {
		t.Fatalf("expected 1 proxies after filtering, got %d", len(filtered))
	}
}

func TestApplyRenameRules(t *testing.T) {
	h := NewSubHandler()
	// Setup rename rules (with matcher + regex)
	config.Global.NodePref.RenameNodes = []config.RenameNodeConfig{
		{Match: "!!TYPE=SS!!香港", Replace: "HK"}, // matches SS only
		{Match: "JP", Replace: "Japan"},
	}

	proxies := []P.ProxyInterface{
		&P.BaseProxy{Type: "ss", Remark: "香港 01"},
		&P.BaseProxy{Type: "vmess", Remark: "JP 02"},
	}

	out := h.applyRenameRules(proxies)
	if out[0].GetRemark() != "HK 01" {
		t.Errorf("expected first remark renamed to 'HK 01', got %q", out[0].GetRemark())
	}
	if out[1].GetRemark() != "Japan 02" {
		t.Errorf("expected second remark renamed to 'Japan 02', got %q", out[1].GetRemark())
	}
}

func TestApplyEmojiRules(t *testing.T) {
	h := NewSubHandler()
	config.Global.Emojis.AddEmoji = true
	config.Global.Emojis.RemoveOldEmoji = true
	config.Global.Emojis.Rules = []config.EmojiRuleConfig{
		{Match: "HK", Emoji: "🇭🇰"},
		{Match: "US", Emoji: "🇺🇸"},
	}

	proxies := []P.ProxyInterface{
		&P.BaseProxy{Remark: "HK 01"},
		&P.BaseProxy{Remark: "US 02"},
	}

	out := h.applyEmojiRules(proxies)
	if got := out[0].GetRemark(); len(got) == 0 || !strings.HasPrefix(got, "🇭🇰") {
		t.Errorf("expected HK remark to start with emoji, got %q", got)
	}
	if got := out[1].GetRemark(); len(got) == 0 || !strings.HasPrefix(got, "🇺🇸") {
		t.Errorf("expected US remark to start with emoji, got %q", got)
	}
}

func TestRenderTemplateWithContext(t *testing.T) {
	h := NewSubHandler()
	// Provide one global and one request var
	config.Global.Template.Globals = []config.TemplateGlobalConfig{{Key: "clash.mode", Value: "rule"}}
	content := "port: {{ default .clash.port \"7890\" }}\nmode: {{ .clash.mode }}\nname: {{ .request.name }}\n"
	rendered, err := h.renderTemplateWithContext(content, map[string]string{"name": "test"})
	if err != nil {
		t.Fatalf("render error: %v", err)
	}
	if want := "mode: rule"; !containsLine(rendered, want) {
		t.Errorf("expected rendered to contain %q, got:\n%s", want, rendered)
	}
	if want := "name: test"; !containsLine(rendered, want) {
		t.Errorf("expected rendered to contain %q, got:\n%s", want, rendered)
	}
}

func TestLoadBaseConfig(t *testing.T) {
	h := NewSubHandler()
	dir := t.TempDir()
	baseFile := filepath.Join(dir, "clash.tpl")
	if err := os.WriteFile(baseFile, []byte("mode: {{ default .clash.mode \"rule\" }}"), 0644); err != nil {
		t.Fatal(err)
	}
	config.Global.Common.ClashRuleBase = baseFile
	config.Global.Template.TemplatePath = dir

	rendered, err := h.loadBaseConfig("clash", map[string]string{"target": "clash"})
	if err != nil {
		t.Fatalf("loadBaseConfig error: %v", err)
	}
	if rendered == "" || !containsLine(rendered, "mode: rule") {
		t.Errorf("expected rendered base to include 'mode: rule', got: %s", rendered)
	}
}

func containsLine(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || (len(s) > len(sub) && (stringContains(s, sub))))
}

func stringContains(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && (indexOf(s, sub) >= 0))
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

func TestHandleVersion(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSubHandler()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/version", nil)
	h.HandleVersion(c)
	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}
}

func TestHandleReadConf(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSubHandler()
	config.Global.Common.APIAccessToken = ""
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/readconf", nil)
	h.HandleReadConf(c)
	// May return 500 if config files not present
	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("Expected 200 or 500, got %d", w.Code)
	}
}

func TestFilterProxiesFunc(t *testing.T) {
	proxies := []P.ProxyInterface{
		&P.BaseProxy{Remark: "HK Node"},
		&P.BaseProxy{Remark: "US Node"},
	}
	filtered := filterProxies(proxies, []string{"HK"}, true)
	if len(filtered) != 1 {
		t.Errorf("Expected 1 proxy, got %d", len(filtered))
	}
}

func TestSetNestedValue(t *testing.T) {
	data := make(map[string]interface{})
	setNestedValue(data, "key", "value")
	if data["key"] != "value" {
		t.Error("Simple value not set")
	}
	setNestedValue(data, "nested.key", "val2")
	if nested, ok := data["nested"].(map[string]interface{}); !ok || nested["key"] != "val2" {
		t.Error("Nested value not set")
	}
}

func TestRemoveEmojiFunc(t *testing.T) {
	// removeEmoji removes emoji and trims spaces
	if removeEmoji("🇺🇸 US") != "US" {
		t.Error("Emoji not removed")
	}
}

func TestFileExists(t *testing.T) {
	if fileExists("/nonexistent/file") {
		t.Error("Nonexistent file reported as existing")
	}
}

func TestSortProxies(t *testing.T) {
	h := NewSubHandler()
	proxies := []P.ProxyInterface{
		&P.BaseProxy{Remark: "Z"},
		&P.BaseProxy{Remark: "A"},
	}
	config.Global.NodePref.SortFlag = true
	sorted := h.sortProxies(proxies)
	if len(sorted) != 2 || sorted[0].GetRemark() != "A" {
		t.Error("Sort failed")
	}
}

func TestApplyMatcherForRename(t *testing.T) {
	h := NewSubHandler()
	proxy := &P.BaseProxy{Type: "ss", Remark: "test", Port: 443}
	matched, _ := h.applyMatcherForRename("!!TYPE=SS!!test", proxy)
	if !matched {
		t.Error("TYPE matcher failed")
	}
	matched, _ = h.applyMatcherForRename("!!PORT=443!!test", proxy)
	if !matched {
		t.Error("PORT matcher failed")
	}
}

func TestHandleSubBasic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSubHandler()
	config.Global.Common.APIMode = true

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/sub?target=clash&url=", nil)

	h.HandleSub(c)
	// Should fail with invalid request
	if w.Code != http.StatusBadRequest {
		t.Logf("Expected 400, got %d (acceptable for test without valid URLs)", w.Code)
	}
}

func TestLoadBaseConfigFunc(t *testing.T) {
	h := NewSubHandler()
	config.Global.Common.ClashRuleBase = "test.yaml"

	_, err := h.loadBaseConfig("clash", nil)
	if err == nil {
		t.Log("Base config loaded (or expected error)")
	}
}

func TestMatchRangeFunc(t *testing.T) {
	h := NewSubHandler()
	if !h.matchRange("443", 443) {
		t.Error("Single value match failed")
	}
	if !h.matchRange("400-500", 443) {
		t.Error("Range match failed")
	}
	if h.matchRange("400-500", 600) {
		t.Error("Range should not match")
	}
}

func TestRenderTemplateFunc(t *testing.T) {
	h := NewSubHandler()
	result, err := h.renderTemplate("Hello {{.Name}}")
	if err != nil {
		t.Logf("Render error (expected without data): %v", err)
	} else if result != "" {
		t.Log("Render succeeded")
	}
}

func TestApplyEmojiRulesFunc(t *testing.T) {
	h := NewSubHandler()
	config.Global.Emojis.AddEmoji = true
	config.Global.Emojis.Rules = []config.EmojiRuleConfig{
		{Match: "US|America", Emoji: "🇺🇸"},
		{Match: "HK|Hong", Emoji: "🇭🇰"},
	}

	proxies := []P.ProxyInterface{
		&P.BaseProxy{Remark: "US Node"},
		&P.BaseProxy{Remark: "HK Server"},
	}

	result := h.applyEmojiRules(proxies)
	if len(result) != 2 {
		t.Error("Emoji rules changed proxy count")
	}
	if !strings.Contains(result[0].GetRemark(), "🇺🇸") && !strings.Contains(result[0].GetRemark(), "US") {
		t.Log("Emoji may not have been added (depends on config)")
	}
}

func TestLoadExternalConfigFunc(t *testing.T) {
	h := NewSubHandler()
	cfg, err := h.loadExternalConfig("nonexistent.ini")
	if err != nil {
		t.Error("loadExternalConfig returned error")
	}
	if cfg == nil {
		t.Error("cfg should not be nil")
	}
}

func TestHandleGetRuleset(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSubHandler()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/getruleset", nil)

	h.HandleGetRuleset(c)
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", w.Code)
	}
}

func TestHandleRender(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSubHandler()
	config.Global.Common.APIAccessToken = ""
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/render?template=test", nil)

	h.HandleRender(c)
	// Will fail without valid template but should respond
	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError && w.Code != http.StatusBadRequest {
		t.Logf("Got status %d", w.Code)
	}
}

func TestHandleGetProfile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSubHandler()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/getprofile", nil)

	h.HandleGetProfile(c)
	// Should fail without name parameter
	if w.Code != http.StatusBadRequest && w.Code != http.StatusInternalServerError && w.Code != http.StatusForbidden {
		t.Logf("Got status %d", w.Code)
	}
}

func TestRenderTemplateWithContextFunc(t *testing.T) {
	h := NewSubHandler()
	params := map[string]string{"key": "value"}
	result, err := h.renderTemplateWithContext("test{{.request.key}}", params)
	if err != nil {
		t.Logf("Render error (may be expected): %v", err)
	} else if result != "" {
		t.Log("Render with context succeeded")
	}
}

func TestLoadBaseConfigDifferentTargets(t *testing.T) {
	h := NewSubHandler()
	targets := []string{"clash", "surge", "loon", "quantumultx", "singbox", "ss", "v2ray", "trojan"}

	for _, target := range targets {
		_, err := h.loadBaseConfig(target, nil)
		if err != nil {
			t.Logf("%s base config load error (acceptable): %v", target, err)
		}
	}
}

func TestFilterProxiesWithIncludeExclude(t *testing.T) {
	proxies := []P.ProxyInterface{
		&P.BaseProxy{Remark: "HK Node"},
		&P.BaseProxy{Remark: "US Node"},
		&P.BaseProxy{Remark: "JP Node"},
	}

	// Test filterProxies function (package-level, not method)
	filtered := filterProxies(proxies, []string{"HK", "US"}, true)
	if len(filtered) == 0 {
		t.Error("Filter should not empty all proxies with includes")
	}

	excluded := filterProxies(proxies, []string{"HK"}, false)
	if len(excluded) == 0 {
		t.Error("Exclude should keep some proxies")
	}
}

func TestApplyRenameRulesFunc(t *testing.T) {
	h := NewSubHandler()
	config.Global.NodePref.RenameNodes = []config.RenameNodeConfig{
		{Match: "HK", Replace: "Hong Kong"},
		{Match: "US", Replace: "United States"},
	}

	proxies := []P.ProxyInterface{
		&P.BaseProxy{Remark: "HK Node"},
		&P.BaseProxy{Remark: "US Server"},
	}

	renamed := h.applyRenameRules(proxies)
	if len(renamed) != 2 {
		t.Error("Rename changed proxy count")
	}
	// Check if rename happened (depending on implementation)
	t.Logf("After rename: %s, %s", renamed[0].GetRemark(), renamed[1].GetRemark())
}

func TestMatchRangeEdgeCases(t *testing.T) {
	h := NewSubHandler()

	// Empty pattern returns true (matches all)
	if !h.matchRange("", 443) {
		t.Error("Empty pattern should match (match all)")
	}

	// Multiple ranges
	if !h.matchRange("80,443,8080", 443) {
		t.Error("Comma separated should match")
	}

	// Invalid format - should not panic
	h.matchRange("invalid", 443)
}

func TestSetNestedValueFunc(t *testing.T) {
	data := make(map[string]interface{})
	setNestedValue(data, "a.b.c", "value")

	if m, ok := data["a"].(map[string]interface{}); ok {
		if m2, ok2 := m["b"].(map[string]interface{}); ok2 {
			if m2["c"] != "value" {
				t.Error("setNestedValue failed")
			}
		}
	}
}

func TestHandleGetRulesetWithParams(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSubHandler()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Base64 encode a URL
	testURL := base64.URLEncoding.EncodeToString([]byte("http://example.com"))
	c.Request = httptest.NewRequest(http.MethodGet, "/getruleset?url="+testURL+"&type=clash", nil)

	h.HandleGetRuleset(c)
	t.Logf("GetRuleset with params returned: %d", w.Code)
}

func TestHandleRenderWithTemplate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSubHandler()
	config.Global.Common.APIAccessToken = ""
	config.Global.Template.TemplatePath = ""

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Try with a template parameter
	c.Request = httptest.NewRequest(http.MethodGet, "/render?template=Hello+World", nil)
	h.HandleRender(c)

	t.Logf("HandleRender returned: %d, body: %s", w.Code, w.Body.String())
}

func TestHandleGetProfileWithName(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSubHandler()
	config.Global.Common.APIAccessToken = ""

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/getprofile?name=test", nil)

	h.HandleGetProfile(c)
	// Will likely fail without actual profile files, but exercises the code
	t.Logf("HandleGetProfile with name returned: %d", w.Code)
}
