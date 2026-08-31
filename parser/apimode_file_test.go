package parser

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gfunc/subconvergo/config"
)

const apiModeFileMarker = "APIMODE-FILE-MARKER-6601"

// writeLocalSubFile writes a valid base64 subscription with a unique marker
// node to a temp file and returns its path.
func writeLocalSubFile(t *testing.T) string {
	t.Helper()
	sub := base64.StdEncoding.EncodeToString([]byte("ss://YWVzLTEyOC1nY206dGVzdA@1.2.3.4:8388#" + apiModeFileMarker))
	path := filepath.Join(t.TempDir(), "sub.txt")
	if err := os.WriteFile(path, []byte(sub), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func setAPIMode(t *testing.T, on bool) {
	t.Helper()
	saved := config.Global.Common.APIMode
	config.Global.Common.APIMode = on
	t.Cleanup(func() { config.Global.Common.APIMode = saved })
}

func TestParse_APIModeRejectsFileURL(t *testing.T) {
	setAPIMode(t, true)
	path := writeLocalSubFile(t)

	sp := &SubParser{URL: "file://" + path}
	sc, err := sp.Parse()
	if err == nil {
		t.Fatalf("file:// URL must be rejected in api_mode, parsed %d proxies", len(sc.Proxies))
	}
}

func TestParse_APIModeRejectsPlainLocalPath(t *testing.T) {
	setAPIMode(t, true)
	path := writeLocalSubFile(t)

	// Bare absolute path: os.Stat succeeds and routes to a local read.
	sp := &SubParser{URL: path}
	sc, err := sp.Parse()
	if err == nil {
		t.Fatalf("plain local path must be rejected in api_mode, parsed %d proxies", len(sc.Proxies))
	}
}

// TestParse_NonAPIModeLocalFileAllowed pins the trusted-deployment opt-in:
// with api_mode=false a local subscription file must still be parsed.
func TestParse_NonAPIModeLocalFileAllowed(t *testing.T) {
	setAPIMode(t, false)
	path := writeLocalSubFile(t)

	sp := &SubParser{URL: "file://" + path}
	sc, err := sp.Parse()
	if err != nil {
		t.Fatalf("non-API mode must allow local subscription files: %v", err)
	}
	if len(sc.Proxies) != 1 || !strings.Contains(sc.Proxies[0].GetRemark(), apiModeFileMarker) {
		t.Fatalf("expected marker node from local file, got %d proxies", len(sc.Proxies))
	}
}
