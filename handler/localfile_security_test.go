package handler

import (
	"encoding/base64"
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
)

const (
	fileSubMarker       = "FILE-SUB-MARKER-7701"
	remoteRulesetMarker = "REMOTECFG-RULESET-MARKER-7702"
	remoteImportMarker  = "REMOTEINI-IMPORT-MARKER-7703"
	localImportGroup    = "LOCAL-IMPORT-GROUP-7704"
)

// writeMarkerSubscription writes a valid base64 subscription containing a
// unique marker node to a temp file and returns its path. The marker must
// never appear in any /sub response in api_mode.
func writeMarkerSubscription(t *testing.T, marker string) string {
	t.Helper()
	sub := base64.StdEncoding.EncodeToString([]byte("ss://YWVzLTEyOC1nY206dGVzdA@1.2.3.4:8388#" + marker))
	path := filepath.Join(t.TempDir(), "sub.txt")
	if err := os.WriteFile(path, []byte(sub), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestSub_APIModeRejectsFileURL(t *testing.T) {
	setupSubTestEnv(t) // sets APIMode=true
	h := NewSubHandler()
	path := writeMarkerSubscription(t, fileSubMarker)

	w := doSub(h, "/sub?target=clash&url="+url.QueryEscape("file://"+path))

	if w.Code == http.StatusOK {
		t.Fatalf("file:// subscription URL must be refused in api_mode, got 200")
	}
	if strings.Contains(w.Body.String(), fileSubMarker) {
		t.Fatalf("local file content leaked into response:\n%s", w.Body.String())
	}
}

func TestSub_APIModeRejectsPlainLocalPath(t *testing.T) {
	setupSubTestEnv(t) // sets APIMode=true
	h := NewSubHandler()
	path := writeMarkerSubscription(t, fileSubMarker)

	// Bare absolute path: os.Stat succeeds and routes to os.ReadFile.
	w := doSub(h, "/sub?target=clash&url="+url.QueryEscape(path))

	if w.Code == http.StatusOK {
		t.Fatalf("plain local path subscription must be refused in api_mode, got 200")
	}
	if strings.Contains(w.Body.String(), fileSubMarker) {
		t.Fatalf("local file content leaked into response:\n%s", w.Body.String())
	}
}

// TestSub_NonAPIModeLocalFileAllowed pins the trusted-deployment opt-in: with
// api_mode=false a local subscription file must still load.
func TestSub_NonAPIModeLocalFileAllowed(t *testing.T) {
	setupSubTestEnv(t)
	config.Global.Common.APIMode = false
	h := NewSubHandler()
	path := writeMarkerSubscription(t, fileSubMarker)

	w := doSub(h, "/sub?target=clash&url="+url.QueryEscape("file://"+path))

	if w.Code != http.StatusOK {
		t.Fatalf("non-API mode must allow local subscription files, got %d: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), fileSubMarker) {
		t.Fatalf("local subscription node missing from output:\n%s", w.Body.String())
	}
}

func TestSub_RemoteConfigCannotReadLocalRuleset(t *testing.T) {
	setupSubTestEnv(t)
	allowLoopbackFetch(t)
	config.Global.Rulesets.Enabled = true
	h := NewSubHandler()

	// Local marker ruleset file that a remote external config must never read.
	markerFile := filepath.Join(t.TempDir(), "marker_rules.list")
	if err := os.WriteFile(markerFile, []byte("DOMAIN-SUFFIX,"+remoteRulesetMarker+".invalid,Auto\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Attacker-hosted external config pointing a ruleset at the local file.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "rulesets:\n  enabled: true\n  rulesets:\n    - group: Auto\n      ruleset: "+markerFile+"\n")
	}))
	defer srv.Close()

	rawURL := "/sub?target=clash" +
		"&url=" + url.QueryEscape("ss://YWVzLTEyOC1nY206dGVzdA@1.2.3.4:8388#plainnode") +
		"&config=" + url.QueryEscape(srv.URL)
	w := doSub(h, rawURL)

	if w.Code != http.StatusOK {
		t.Fatalf("subscription itself must still succeed, got %d: %s", w.Code, w.Body.String())
	}
	if strings.Contains(w.Body.String(), remoteRulesetMarker) {
		t.Fatalf("remote external config read a local ruleset file:\n%s", w.Body.String())
	}
}

func TestSub_RemoteINIImportRejected(t *testing.T) {
	setupSubTestEnv(t)
	allowLoopbackFetch(t)
	config.Global.Rulesets.Enabled = true
	h := NewSubHandler()

	// Local marker file; a remote INI config's !!import: must never read it.
	importFile := filepath.Join(t.TempDir(), "imported.ini")
	if err := os.WriteFile(importFile, []byte("Auto,[]DOMAIN-SUFFIX,"+remoteImportMarker+".invalid\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Attacker-hosted INI external config importing the local file.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "[rulesets]\nenabled=true\nruleset=!!import:"+importFile+"\n")
	}))
	defer srv.Close()

	rawURL := "/sub?target=clash" +
		"&url=" + url.QueryEscape("ss://YWVzLTEyOC1nY206dGVzdA@1.2.3.4:8388#plainnode") +
		"&config=" + url.QueryEscape(srv.URL)
	w := doSub(h, rawURL)

	if w.Code != http.StatusOK {
		t.Fatalf("subscription itself must still succeed, got %d: %s", w.Code, w.Body.String())
	}
	if strings.Contains(w.Body.String(), remoteImportMarker) {
		t.Fatalf("remote INI external config imported a local file:\n%s", w.Body.String())
	}
}

// TestLoadExternalConfig_LocalINIImportAllowed pins the trusted path: a local
// (admin-controlled) INI config keeps !!import: for files under base_path.
func TestLoadExternalConfig_LocalINIImportAllowed(t *testing.T) {
	saveGlobal(t)
	cache.Init(t.TempDir())
	h := NewSubHandler()

	dir := t.TempDir()
	config.Global.Common.BasePath = dir

	if err := os.WriteFile(filepath.Join(dir, "imported.ini"),
		[]byte(localImportGroup+"`select`[]DIRECT\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfgPath := filepath.Join(dir, "ext.ini")
	if err := os.WriteFile(cfgPath,
		[]byte("[proxy_groups]\ncustom_proxy_group=!!import:imported.ini\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	ecfg, err := h.loadExternalConfig(cfgPath)
	if err != nil {
		t.Fatalf("local config load failed: %v", err)
	}
	// The INI branch reports parsed groups via the top-level CustomGroups
	// field; Merge() folds them into ProxyGroups.CustomProxyGroups.
	for _, g := range ecfg.CustomGroups {
		if g.Name == localImportGroup {
			return
		}
	}
	t.Fatalf("imported group %q missing from local external config: %#v", localImportGroup, ecfg.CustomGroups)
}
