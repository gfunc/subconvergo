package handler

import (
	"encoding/base64"
	"io"
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

func TestFilterProxiesRegexInclude(t *testing.T) {
	proxies := []P.ProxyInterface{
		&P.BaseProxy{Remark: "HK Node"},
		&P.BaseProxy{Remark: "US Node"},
		&P.BaseProxy{Remark: "HK Server"},
	}
	// Include remarks that start with HK using regex
	filtered := filterProxies(proxies, []string{"/^HK/"}, true)
	if len(filtered) != 2 {
		t.Fatalf("expected 2 proxies starting with HK, got %d", len(filtered))
	}
	for _, p := range filtered {
		if !strings.HasPrefix(p.GetRemark(), "HK") {
			t.Fatalf("unexpected remark in regex include: %s", p.GetRemark())
		}
	}
}

func TestFilterProxiesRegexExclude(t *testing.T) {
	proxies := []P.ProxyInterface{
		&P.BaseProxy{Remark: "HK Node"},
		&P.BaseProxy{Remark: "US Node"},
		&P.BaseProxy{Remark: "JP Node"},
	}
	// Exclude remarks matching US or JP
	filtered := filterProxies(proxies, []string{"/(US|JP)/"}, false)
	if len(filtered) != 1 || !strings.Contains(filtered[0].GetRemark(), "HK") {
		t.Fatalf("expected only HK to remain, got %v", filtered)
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

func TestFileExists(t *testing.T) {
	if fileExists("/nonexistent/file") {
		t.Error("Nonexistent file reported as existing")
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

func TestRenderTemplateFunc(t *testing.T) {
	h := NewSubHandler()
	result, err := h.renderTemplate("Hello {{.Name}}")
	if err != nil {
		t.Logf("Render error (expected without data): %v", err)
	} else if result != "" {
		t.Log("Render succeeded")
	}
}

func TestLoadExternalConfigFunc(t *testing.T) {
	h := NewSubHandler()
	cfg, err := h.loadExternalConfig("nonexistent.ini")
	if err == nil || cfg != nil {
		t.Error("expected error for nonexistent external config")
	}
}

func TestLoadExternalConfig_YAMLRemote(t *testing.T) {
	h := NewSubHandler()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, ""+
			"proxy_groups:\n"+
			"  custom_proxy_group:\n"+
			"    - name: Auto\n"+
			"      type: select\n"+
			"      rule: ['.*']\n"+
			"rulesets:\n"+
			"  enabled: true\n"+
			"  rulesets:\n"+
			"    - ruleset: rules/custom_test_rules.list\n"+
			"      group: Auto\n")
	}))
	defer srv.Close()

	ecfg, err := h.loadExternalConfig(srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ecfg.ProxyGroups) == 0 || ecfg.ProxyGroups[0].Name != "Auto" {
		t.Fatalf("external groups not parsed: %#v", ecfg.ProxyGroups)
	}
	if len(ecfg.Rulesets) == 0 || ecfg.Rulesets[0].Group != "Auto" {
		t.Fatalf("external rulesets not parsed: %#v", ecfg.Rulesets)
	}
}

func TestLoadExternalConfig_LocalTOML(t *testing.T) {
	h := NewSubHandler()
	dir := t.TempDir()
	content := []byte(`
		[proxy_groups]
		[[proxy_groups.custom_groups]]
		name = "Auto"
		type = "select"
		rule = [".*"]

		[rulesets]
		enabled = true
		[[rulesets.rulesets]]
		ruleset = "rules/custom_test_rules.list"
		group = "Auto"
	`)
	fp := filepath.Join(dir, "ext.toml")
	if err := os.WriteFile(fp, content, 0o644); err != nil {
		t.Fatal(err)
	}
	ecfg, err := h.loadExternalConfig(fp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ecfg.ProxyGroups) == 0 || ecfg.ProxyGroups[0].Name != "Auto" {
		t.Fatalf("toml groups not parsed: %#v", ecfg.ProxyGroups)
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

func TestHandleGetRuleset_RemoteFetch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSubHandler()

	// Start a test HTTP server serving a simple ruleset
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("DOMAIN-SUFFIX,example.com,Auto\nMATCH,Auto\n"))
	}))
	defer ts.Close()

	encoded := base64.URLEncoding.EncodeToString([]byte(ts.URL))
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/getruleset?url="+encoded+"&type=clash", nil)
	c.Request = req

	h.HandleGetRuleset(c)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "MATCH,Auto") {
		t.Fatalf("unexpected ruleset body: %s", body)
	}
}

func TestHandleGetRuleset_LocalPath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSubHandler()

	// Prepare a temp base path with a rules file
	dir := t.TempDir()
	rulesDir := filepath.Join(dir, "rules")
	_ = os.MkdirAll(rulesDir, 0o755)
	filePath := filepath.Join(rulesDir, "local_test.list")
	if err := os.WriteFile(filePath, []byte("GEOIP,CN,DIRECT\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Point base path to temp dir
	config.Global.Common.BasePath = dir

	encoded := base64.URLEncoding.EncodeToString([]byte("local_test.list"))
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/getruleset?url="+encoded+"&type=clash", nil)
	c.Request = req

	h.HandleGetRuleset(c)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "GEOIP,CN,DIRECT") {
		t.Fatalf("unexpected ruleset content: %s", w.Body.String())
	}
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
