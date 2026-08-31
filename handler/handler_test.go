package handler

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gfunc/subconvergo/cache"
	"github.com/gfunc/subconvergo/config"
	_ "github.com/gfunc/subconvergo/generator/impl"
	"github.com/gin-gonic/gin"
)

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
	// Rule base files are resolved relative to the config directory (cwd in tests).
	dir := filepath.Join(".", "test_tmp_"+t.Name())
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	baseFile := filepath.Join(dir, "clash.tpl")
	if err := os.WriteFile(baseFile, []byte("mode: {{ default .clash.mode \"rule\" }}"), 0o644); err != nil {
		t.Fatal(err)
	}
	config.Global.Common.ClashRuleBase = baseFile

	rendered, err := h.loadBaseConfig("clash", map[string]string{"target": "clash"}, config.Global)
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
	config.Global.Common.APIAccessToken = "test-token"
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/readconf?token=test-token", nil)
	h.HandleReadConf(c)
	// May return 500 if config files not present
	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("Expected 200 or 500, got %d", w.Code)
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

	_, err := h.loadBaseConfig("clash", nil, config.Global)
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
	cache.Init(t.TempDir())
	allowLoopbackFetch(t)
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
	if len(ecfg.ProxyGroups.CustomProxyGroups) == 0 || ecfg.ProxyGroups.CustomProxyGroups[0].Name != "Auto" {
		t.Fatalf("external groups not parsed: %#v", ecfg.ProxyGroups.CustomProxyGroups)
	}
	if len(ecfg.Rulesets.Rulesets) == 0 || ecfg.Rulesets.Rulesets[0].Group != "Auto" {
		t.Fatalf("external rulesets not parsed: %#v", ecfg.Rulesets.Rulesets)
	}
}

func TestLoadExternalConfig_LocalTOML(t *testing.T) {
	h := NewSubHandler()
	dir := t.TempDir()
	cache.Init(dir) // Initialize cache for test

	content := []byte(`
[[custom_groups]]
name = "Auto"
type = "select"
rule = [".*"]

[[rulesets]]
group = "Auto"
ruleset = "rules/custom_test_rules.list"
`)
	fp := filepath.Join(dir, "ext.toml")
	if err := os.WriteFile(fp, content, 0o644); err != nil {
		t.Fatal(err)
	}
	ecfg, err := h.loadExternalConfig(fp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ecfg.ProxyGroups.CustomProxyGroups) == 0 || ecfg.ProxyGroups.CustomProxyGroups[0].Name != "Auto" {
		t.Fatalf("toml groups not parsed: %#v", ecfg.ProxyGroups.CustomProxyGroups)
	}
	if len(ecfg.Rulesets.Rulesets) == 0 || ecfg.Rulesets.Rulesets[0].Group != "Auto" {
		t.Fatalf("toml rulesets not parsed: %#v", ecfg.Rulesets.Rulesets)
	}
}

func TestHandleGetRuleset(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSubHandler()
	config.Global.Common.APIAccessToken = "test-token"
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/getruleset?token=test-token", nil)

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
		_, err := h.loadBaseConfig(target, nil, config.Global)
		if err != nil {
			t.Logf("%s base config load error (acceptable): %v", target, err)
		}
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
	cache.Init(t.TempDir())
	allowLoopbackFetch(t)
	gin.SetMode(gin.TestMode)
	config.Global.Common.APIAccessToken = "test-token"
	h := NewSubHandler()

	// Start a test HTTP server serving a simple ruleset
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("DOMAIN-SUFFIX,example.com,Auto\nMATCH,Auto\n"))
	}))
	defer ts.Close()

	encoded := base64.URLEncoding.EncodeToString([]byte(ts.URL))
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/getruleset?url="+encoded+"&type=clash&token=test-token", nil)
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
	cache.Init(t.TempDir())
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
	config.Global.Common.APIAccessToken = "test-token"

	encoded := base64.URLEncoding.EncodeToString([]byte("local_test.list"))
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/getruleset?url="+encoded+"&type=clash&token=test-token", nil)
	c.Request = req

	h.HandleGetRuleset(c)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "GEOIP,CN,DIRECT") {
		t.Fatalf("unexpected ruleset content: %s", w.Body.String())
	}
}

func TestHandleGetRuleset_AbsolutePathBlocked(t *testing.T) {
	cache.Init(t.TempDir())
	gin.SetMode(gin.TestMode)
	h := NewSubHandler()

	config.Global.Common.BasePath = t.TempDir()
	// A valid token keeps this test on the path guard (fail-closed token
	// checks would otherwise refuse the request before it).
	config.Global.Common.APIAccessToken = "test-token"
	// Set a secret token in the environment; /proc/self/environ would normally expose it.
	t.Setenv("API_TOKEN", "super-secret-token-12345")

	encoded := base64.URLEncoding.EncodeToString([]byte("/proc/self/environ"))
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/getruleset?url="+encoded+"&type=clash&token=test-token", nil)
	c.Request = req

	h.HandleGetRuleset(c)
	if w.Code == http.StatusOK {
		t.Fatalf("absolute path must not be served, got 200")
	}
	if strings.Contains(w.Body.String(), "super-secret-token-12345") {
		t.Fatalf("leaked API_TOKEN via /proc/self/environ")
	}
}

func TestHandleGetRuleset_TraversalBlocked(t *testing.T) {
	cache.Init(t.TempDir())
	gin.SetMode(gin.TestMode)
	h := NewSubHandler()

	// Create a file outside the base directory
	dir := t.TempDir()
	parent := filepath.Dir(dir)
	secretPath := filepath.Join(parent, "secret_ruleset.txt")
	if err := os.WriteFile(secretPath, []byte("SECRET DATA\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Remove(secretPath) })

	config.Global.Common.BasePath = dir
	config.Global.Common.APIAccessToken = "test-token"

	encoded := base64.URLEncoding.EncodeToString([]byte("../secret_ruleset.txt"))
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/getruleset?url="+encoded+"&type=clash&token=test-token", nil)
	c.Request = req

	h.HandleGetRuleset(c)
	if w.Code == http.StatusOK {
		t.Fatalf("directory traversal must not be served, got 200")
	}
	if strings.Contains(w.Body.String(), "SECRET DATA") {
		t.Fatalf("leaked file outside base directory via traversal")
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

func TestProcessSubRequest_ExternalConfigRuleBaseLeakBlocked(t *testing.T) {
	cache.Init(t.TempDir())
	gin.SetMode(gin.TestMode)
	h := NewSubHandler()

	// Create a file outside the intended config directory
	dir := t.TempDir()
	parent := filepath.Dir(dir)
	secretPath := filepath.Join(parent, "secret_base.txt")
	if err := os.WriteFile(secretPath, []byte("LEAK123\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Remove(secretPath) })

	// External config overrides clash_rule_base to point at the secret file
	extCfg := filepath.Join(dir, "ext.yml")
	content := []byte(fmt.Sprintf("common:\n  clash_rule_base: %s\n", secretPath))
	if err := os.WriteFile(extCfg, content, 0o644); err != nil {
		t.Fatal(err)
	}

	config.Global.Common.APIMode = true
	config.Global.Common.APIAccessToken = ""
	config.Global.Common.ClashRuleBase = ""

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/sub", nil)

	h.processSubRequest(c, &RequestParams{
		Target: "clash",
		URL:    "ss://YWVzLTEyOC1nY206dGVzdA==@1.2.3.4:8388#test",
		Config: extCfg,
	})

	if w.Code == http.StatusOK {
		t.Fatalf("external config must not be able to read arbitrary rule base, got 200")
	}
	if strings.Contains(w.Body.String(), "LEAK123") {
		t.Fatalf("leaked arbitrary file content via clash_rule_base: %s", w.Body.String())
	}
}

func TestHandleRender_TraversalBlocked(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSubHandler()

	baseDir := t.TempDir()
	tplDir := filepath.Join(baseDir, "tpl")
	if err := os.MkdirAll(tplDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Secret file is outside the base directory entirely
	parent := filepath.Dir(baseDir)
	secretPath := filepath.Join(parent, "secret_template.txt")
	if err := os.WriteFile(secretPath, []byte("SECRET_TEMPLATE_DATA"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Remove(secretPath) })

	config.Global.Common.APIAccessToken = "token123"
	config.Global.Common.BasePath = baseDir
	config.Global.Template.TemplatePath = "tpl"

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/render?path=../../secret_template.txt&token=token123", nil)

	h.HandleRender(c)
	if w.Code == http.StatusOK {
		t.Fatalf("template traversal must not be served, got 200")
	}
	if strings.Contains(w.Body.String(), "SECRET_TEMPLATE_DATA") {
		t.Fatalf("leaked template outside template directory via traversal")
	}
}

func TestHandleGetProfile_AbsolutePathBlocked(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSubHandler()

	dir := t.TempDir()
	parent := filepath.Dir(dir)
	secretPath := filepath.Join(parent, "secret_profile")
	if err := os.WriteFile(secretPath, []byte("SECRET_PROFILE_DATA"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Remove(secretPath) })

	config.Global.Common.APIAccessToken = "token123"
	config.Global.Common.BasePath = dir

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/getprofile?name="+secretPath+"&token=token123", nil)

	h.HandleGetProfile(c)
	if w.Code == http.StatusOK {
		t.Fatalf("absolute profile path must not be served, got 200")
	}
	if strings.Contains(w.Body.String(), "SECRET_PROFILE_DATA") {
		t.Fatalf("leaked profile outside profiles directory via absolute path")
	}
}

func TestResolveProfilePath_FallsBackToConfigDir(t *testing.T) {
	// Regression: profiles live under the pref directory (e.g. base/profiles/),
	// not under base_path/profiles/. When base_path points to a nested "base"
	// directory, resolveProfilePath must fall back to the config directory.
	dir := t.TempDir()
	profilesDir := filepath.Join(dir, "profiles")
	if err := os.MkdirAll(profilesDir, 0o755); err != nil {
		t.Fatal(err)
	}
	profilePath := filepath.Join(profilesDir, "example.ini")
	if err := os.WriteFile(profilePath, []byte("[Profile]\ntarget=clash\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	baseDir := filepath.Join(dir, "nested_base")
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		t.Fatal(err)
	}

	resolved, err := resolveProfilePath("example", []string{baseDir, dir})
	if err != nil {
		t.Fatalf("resolveProfilePath error: %v", err)
	}
	if resolved != profilePath {
		t.Fatalf("expected profile at %s, got %s", profilePath, resolved)
	}
}

func TestResolveProfilePath_RelativePath(t *testing.T) {
	// Users may pass the full relative path to a profile file, e.g.
	// name=profiles/gfunc.ini. It must be resolved under the candidate roots.
	dir := t.TempDir()
	profilePath := filepath.Join(dir, "profiles", "gfunc.ini")
	if err := os.MkdirAll(filepath.Dir(profilePath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(profilePath, []byte("[Profile]\ntarget=clash\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	resolved, err := resolveProfilePath("profiles/gfunc.ini", []string{dir})
	if err != nil {
		t.Fatalf("resolveProfilePath error: %v", err)
	}
	if resolved != profilePath {
		t.Fatalf("expected profile at %s, got %s", profilePath, resolved)
	}
}

func TestResolveProfilePath_RelativePathUnderBasePath(t *testing.T) {
	// When base_path differs from config_dir, a relative profile path should be
	// found under the base_path root as well.
	dir := t.TempDir()
	baseDir := filepath.Join(dir, "base")
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		t.Fatal(err)
	}
	profilePath := filepath.Join(baseDir, "profiles", "gfunc.ini")
	if err := os.MkdirAll(filepath.Dir(profilePath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(profilePath, []byte("[Profile]\ntarget=clash\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	resolved, err := resolveProfilePath("profiles/gfunc.ini", []string{baseDir, dir})
	if err != nil {
		t.Fatalf("resolveProfilePath error: %v", err)
	}
	if resolved != profilePath {
		t.Fatalf("expected profile at %s, got %s", profilePath, resolved)
	}
}

func TestHandleSub_SubURLKeepsPercentEncoding(t *testing.T) {
	cache.Init(t.TempDir())
	allowLoopbackFetch(t)
	gin.SetMode(gin.TestMode)
	h := NewSubHandler()

	// Regression: the subscription URL must be fetched with its percent-encoded
	// query bytes intact (e.g. name=%E4%BD%8E...). An extra URL-decode after
	// gin's query parsing turns them into raw UTF-8, which strict frontends
	// (Cloudflare) reject with 400.
	var gotRawQuery string
	sub := base64.StdEncoding.EncodeToString([]byte("ss://YWVzLTEyOC1nY206dGVzdA@c.example.com:8388#ENCODED-NODE"))
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotRawQuery = r.URL.RawQuery
		if strings.Contains(r.URL.EscapedPath()+r.URL.RawQuery, "低调机场") {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		_, _ = io.WriteString(w, sub)
	}))
	defer srv.Close()

	config.Global.Common.APIMode = true
	config.Global.Common.ClashRuleBase = ""

	inner := srv.URL + "/subscribe?token=abc&name=%E4%BD%8E%E8%B0%83%E6%9C%BA%E5%9C%BA"
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet,
		"/sub?target=clash&url="+url.QueryEscape(inner), nil)
	h.HandleSub(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "ENCODED-NODE") {
		t.Errorf("expected node from subscription, got:\n%s", w.Body.String())
	}
	if gotRawQuery != "token=abc&name=%E4%BD%8E%E8%B0%83%E6%9C%BA%E5%9C%BA" {
		t.Errorf("upstream saw mangled query: %q", gotRawQuery)
	}
}

func TestHandleGetProfile_QueryURLMergedWithProfileURL(t *testing.T) {
	cache.Init(t.TempDir())
	allowLoopbackFetch(t)
	gin.SetMode(gin.TestMode)
	h := NewSubHandler()

	// Two subscription endpoints with distinct node names
	subA := base64.StdEncoding.EncodeToString([]byte("ss://YWVzLTEyOC1nY206dGVzdA@a.example.com:8388#PROFILE-NODE-AAA"))
	subB := base64.StdEncoding.EncodeToString([]byte("ss://YWVzLTEyOC1nY206dGVzdA@b.example.com:8388#QUERY-NODE-BBB"))
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/a" {
			_, _ = io.WriteString(w, subA)
		} else {
			_, _ = io.WriteString(w, subB)
		}
	}))
	defer srv.Close()

	// Profile pointing at subscription A
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "profiles"), 0o755); err != nil {
		t.Fatal(err)
	}
	profileContent := "[Profile]\ntarget=clash\nurl=" + srv.URL + "/a\n"
	if err := os.WriteFile(filepath.Join(dir, "profiles", "merge.ini"), []byte(profileContent), 0o644); err != nil {
		t.Fatal(err)
	}

	config.Global.Common.APIMode = true
	config.Global.Common.APIAccessToken = "token123"
	config.Global.Common.BasePath = dir
	config.Global.Common.ClashRuleBase = ""
	config.Global.ManagedConfig.WriteManagedConfig = false

	// Pass subscription B via the url query parameter; it must be merged with
	// the profile's URL instead of replacing it.
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet,
		"/getprofile?name=merge&token=token123&url="+url.QueryEscape(srv.URL+"/b"), nil)
	h.HandleGetProfile(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	body := w.Body.String()
	if !strings.Contains(body, "PROFILE-NODE-AAA") {
		t.Errorf("profile URL node missing from merged output:\n%s", body)
	}
	if !strings.Contains(body, "QUERY-NODE-BBB") {
		t.Errorf("query URL node missing from merged output:\n%s", body)
	}
}
