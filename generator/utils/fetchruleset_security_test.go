package utils

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gfunc/subconvergo/config"
)

const rulesetEscapeMarker = "RULESET-ESCAPE-MARKER-8801"

func setRulesetBasePath(t *testing.T, dir string) {
	t.Helper()
	saved := config.Global.Common.BasePath
	config.Global.Common.BasePath = dir
	t.Cleanup(func() { config.Global.Common.BasePath = saved })
}

func TestFetchRuleset_RejectsAbsoluteLocalPath(t *testing.T) {
	dir := t.TempDir()
	marker := filepath.Join(dir, "abs_marker.list")
	if err := os.WriteFile(marker, []byte("DOMAIN-SUFFIX,"+rulesetEscapeMarker+".invalid\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	setRulesetBasePath(t, t.TempDir())

	content, err := FetchRuleset(marker)
	if err == nil {
		t.Fatalf("absolute local ruleset path must be refused, got content: %q", content)
	}
	if strings.Contains(content, rulesetEscapeMarker) {
		t.Fatalf("marker file content leaked via absolute path")
	}
}

func TestFetchRuleset_RejectsTraversal(t *testing.T) {
	dir := t.TempDir()
	base := filepath.Join(dir, "base")
	if err := os.MkdirAll(base, 0o755); err != nil {
		t.Fatal(err)
	}
	// Marker file lives outside the configured base directory.
	if err := os.WriteFile(filepath.Join(dir, "marker.list"), []byte("DOMAIN-SUFFIX,"+rulesetEscapeMarker+".invalid\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	setRulesetBasePath(t, base)

	content, err := FetchRuleset("../marker.list")
	if err == nil {
		t.Fatalf("traversal ruleset path must be refused, got content: %q", content)
	}
	if strings.Contains(content, rulesetEscapeMarker) {
		t.Fatalf("marker file content leaked via traversal")
	}
}

func TestFetchRuleset_SymlinkEscapeRefused(t *testing.T) {
	dir := t.TempDir()
	base := filepath.Join(dir, "base")
	rulesDir := filepath.Join(base, "rules")
	if err := os.MkdirAll(rulesDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Marker file lives outside the configured base directory.
	outside := filepath.Join(dir, "marker.list")
	if err := os.WriteFile(outside, []byte("DOMAIN-SUFFIX,"+rulesetEscapeMarker+".invalid\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// A symlink inside the rules root points outside it.
	if err := os.Symlink(outside, filepath.Join(rulesDir, "evil.list")); err != nil {
		t.Fatal(err)
	}
	setRulesetBasePath(t, base)

	content, err := FetchRuleset("evil.list")
	if err == nil {
		t.Fatalf("symlink escaping the rules root must be refused, got content: %q", content)
	}
	if strings.Contains(content, rulesetEscapeMarker) {
		t.Fatalf("marker file content leaked via symlink escape")
	}
}

// TestFetchRuleset_InRootSymlinkAllowed pins the green path: a symlink that
// resolves to a file still under the rules root keeps working.
func TestFetchRuleset_InRootSymlinkAllowed(t *testing.T) {
	base := t.TempDir()
	rulesDir := filepath.Join(base, "rules")
	if err := os.MkdirAll(rulesDir, 0o755); err != nil {
		t.Fatal(err)
	}
	want := "DOMAIN-SUFFIX,innocent.example\n"
	if err := os.WriteFile(filepath.Join(rulesDir, "real.list"), []byte(want), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("real.list", filepath.Join(rulesDir, "alias.list")); err != nil {
		t.Fatal(err)
	}
	setRulesetBasePath(t, base)

	content, err := FetchRuleset("alias.list")
	if err != nil {
		t.Fatalf("in-root symlink must keep working: %v", err)
	}
	if content != want {
		t.Fatalf("unexpected ruleset content: %q", content)
	}
}

// TestFetchRuleset_LocalFileUnderBaseStillWorks pins the plain green path for
// trusted local rulesets under base_path/rules.
func TestFetchRuleset_LocalFileUnderBaseStillWorks(t *testing.T) {
	base := t.TempDir()
	rulesDir := filepath.Join(base, "rules")
	if err := os.MkdirAll(rulesDir, 0o755); err != nil {
		t.Fatal(err)
	}
	want := "GEOIP,CN,DIRECT\n"
	if err := os.WriteFile(filepath.Join(rulesDir, "local.list"), []byte(want), 0o644); err != nil {
		t.Fatal(err)
	}
	setRulesetBasePath(t, base)

	content, err := FetchRuleset("local.list")
	if err != nil {
		t.Fatalf("local ruleset under base path must load: %v", err)
	}
	if content != want {
		t.Fatalf("unexpected ruleset content: %q", content)
	}
}
