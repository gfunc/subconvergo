package handler

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gfunc/subconvergo/cache"
	"github.com/gfunc/subconvergo/config"
	"github.com/gin-gonic/gin"
)

const (
	renderSymlinkMarker  = "RENDER-SYMLINK-MARKER-9901"
	baseCfgSymlinkMarker = "BASECFG-SYMLINK-MARKER-9902"
	profileSymlinkMarker = "PROFILE-SYMLINK-MARKER-9903"
	inRootSymlinkMarker  = "IN-ROOT-SYMLINK-OK-9904"
)

func TestRender_SymlinkEscapeRefused(t *testing.T) {
	saveGlobal(t)
	gin.SetMode(gin.TestMode)
	h := NewSubHandler()

	baseDir := t.TempDir()
	tplDir := filepath.Join(baseDir, "templates")
	if err := os.MkdirAll(tplDir, 0o755); err != nil {
		t.Fatal(err)
	}
	outside := t.TempDir()
	if err := os.WriteFile(filepath.Join(outside, "marker.txt"), []byte(renderSymlinkMarker), 0o644); err != nil {
		t.Fatal(err)
	}
	// A symlink inside the template root points outside it.
	if err := os.Symlink(outside, filepath.Join(tplDir, "logs")); err != nil {
		t.Fatal(err)
	}

	config.Global.Common.APIAccessToken = "token123"
	config.Global.Common.BasePath = baseDir
	config.Global.Template.TemplatePath = "templates"

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/render?path=logs/marker.txt&token=token123", nil)
	h.HandleRender(c)

	if w.Code == http.StatusOK {
		t.Fatalf("template symlink escape must be refused, got 200")
	}
	if strings.Contains(w.Body.String(), renderSymlinkMarker) {
		t.Fatalf("file outside template root leaked via symlink")
	}
}

// TestRender_InRootSymlinkAllowed pins the green path: a symlink resolving to
// a file still under the template root keeps working.
func TestRender_InRootSymlinkAllowed(t *testing.T) {
	saveGlobal(t)
	gin.SetMode(gin.TestMode)
	h := NewSubHandler()

	baseDir := t.TempDir()
	tplDir := filepath.Join(baseDir, "templates")
	if err := os.MkdirAll(tplDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tplDir, "real.txt"), []byte(inRootSymlinkMarker), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("real.txt", filepath.Join(tplDir, "alias.txt")); err != nil {
		t.Fatal(err)
	}

	config.Global.Common.APIAccessToken = "token123"
	config.Global.Common.BasePath = baseDir
	config.Global.Template.TemplatePath = "templates"

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/render?path=alias.txt&token=token123", nil)
	h.HandleRender(c)

	if w.Code != http.StatusOK {
		t.Fatalf("in-root symlink must keep working, got %d: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), inRootSymlinkMarker) {
		t.Fatalf("in-root symlink target content missing: %q", w.Body.String())
	}
}

func TestBaseConfig_SymlinkEscapeRefused(t *testing.T) {
	saveGlobal(t)
	h := NewSubHandler()

	// Rule base files resolve under the config directory (CWD in tests);
	// build the symlink tree inside it.
	dir := filepath.Join(".", "test_tmp_"+t.Name())
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })

	outside := t.TempDir()
	if err := os.WriteFile(filepath.Join(outside, "marker.tpl"), []byte(baseCfgSymlinkMarker), 0o644); err != nil {
		t.Fatal(err)
	}
	// A symlink inside the config directory points outside it.
	if err := os.Symlink(outside, filepath.Join(dir, "link")); err != nil {
		t.Fatal(err)
	}

	config.Global.Template.TemplatePath = ""
	reqCfg := *config.Global
	reqCfg.Common.ClashRuleBase = filepath.Join(dir, "link", "marker.tpl")

	content, err := h.loadBaseConfig("clash", map[string]string{"target": "clash"}, &reqCfg)
	if err == nil {
		t.Fatalf("symlink escape via rule base must be refused, got content: %q", content)
	}
	if strings.Contains(content, baseCfgSymlinkMarker) {
		t.Fatalf("file outside config directory leaked via symlinked rule base")
	}
}

func TestProfile_SymlinkEscapeRefused(t *testing.T) {
	setupSubTestEnv(t)
	config.Global.Common.APIAccessToken = "token123"
	h := NewSubHandler()

	baseDir := t.TempDir()
	profilesDir := filepath.Join(baseDir, "profiles")
	if err := os.MkdirAll(profilesDir, 0o755); err != nil {
		t.Fatal(err)
	}
	outside := t.TempDir()
	profileContent := "[Profile]\ntarget=clash\nurl=ss://YWVzLTEyOC1nY206dGVzdA@1.2.3.4:8388#" + profileSymlinkMarker + "\n"
	if err := os.WriteFile(filepath.Join(outside, "marker.ini"), []byte(profileContent), 0o644); err != nil {
		t.Fatal(err)
	}
	// A symlink inside the profiles directory points outside it.
	if err := os.Symlink(outside, filepath.Join(profilesDir, "link")); err != nil {
		t.Fatal(err)
	}
	config.Global.Common.BasePath = baseDir

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/getprofile?name=profiles/link/marker.ini&token=token123", nil)
	h.HandleGetProfile(c)

	if w.Code == http.StatusOK {
		t.Fatalf("profile symlink escape must be refused, got 200")
	}
	if strings.Contains(w.Body.String(), profileSymlinkMarker) {
		t.Fatalf("profile outside profiles root leaked via symlink")
	}
}

// TestGetRuleset_InRootSymlinkAllowed pins the green path for the local
// ruleset endpoint once it converges on the shared resolver: a symlink
// resolving to a file still under the rules root keeps working.
func TestGetRuleset_InRootSymlinkAllowed(t *testing.T) {
	saveGlobal(t)
	cache.Init(t.TempDir())
	gin.SetMode(gin.TestMode)
	h := NewSubHandler()

	dir := t.TempDir()
	rulesDir := filepath.Join(dir, "rules")
	if err := os.MkdirAll(rulesDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(rulesDir, "real.list"), []byte("GEOIP,CN,DIRECT\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("real.list", filepath.Join(rulesDir, "alias.list")); err != nil {
		t.Fatal(err)
	}

	config.Global.Common.APIAccessToken = "token123"
	config.Global.Common.BasePath = dir
	config.Global.Advanced.EnableCache = false

	encoded := base64.URLEncoding.EncodeToString([]byte("alias.list"))
	w := doGetRuleset(h, "/getruleset?url="+encoded+"&type=clash&token=token123")

	if w.Code != http.StatusOK {
		t.Fatalf("in-root symlink must keep working, got %d: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "GEOIP,CN,DIRECT") {
		t.Fatalf("unexpected ruleset content: %q", w.Body.String())
	}
}

// TestGetRuleset_SymlinkEscapeRefused pins the existing confinement behavior
// of the local ruleset endpoint against regressions from the shared-resolver
// convergence.
func TestGetRuleset_SymlinkEscapeRefused(t *testing.T) {
	saveGlobal(t)
	cache.Init(t.TempDir())
	gin.SetMode(gin.TestMode)
	h := NewSubHandler()

	dir := t.TempDir()
	rulesDir := filepath.Join(dir, "rules")
	if err := os.MkdirAll(rulesDir, 0o755); err != nil {
		t.Fatal(err)
	}
	outside := t.TempDir()
	if err := os.WriteFile(filepath.Join(outside, "marker.list"), []byte(rulesetEscapeMarker02), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(rulesDir, "evil")); err != nil {
		t.Fatal(err)
	}

	config.Global.Common.APIAccessToken = "token123"
	config.Global.Common.BasePath = dir
	config.Global.Advanced.EnableCache = false

	encoded := base64.URLEncoding.EncodeToString([]byte("evil/marker.list"))
	w := doGetRuleset(h, "/getruleset?url="+encoded+"&type=clash&token=token123")

	if w.Code == http.StatusOK {
		t.Fatalf("ruleset symlink escape must be refused, got 200")
	}
	if strings.Contains(w.Body.String(), rulesetEscapeMarker02) {
		t.Fatalf("file outside rules root leaked via symlink")
	}
}

const rulesetEscapeMarker02 = "GETRULESET-SYMLINK-MARKER-9905"
