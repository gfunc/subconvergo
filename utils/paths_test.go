package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gfunc/subconvergo/config"
)

// setRulesetBasePath points base_path at dir for the duration of the test.
func setRulesetBasePath(t *testing.T, dir string) {
	t.Helper()
	saved := config.Global.Common.BasePath
	config.Global.Common.BasePath = dir
	t.Cleanup(func() { config.Global.Common.BasePath = saved })
}

// writeFile creates path (and its parents) with the given content.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestResolveRulesetPath_ContainerLayout pins the layout the Docker image
// ships: the image copies the repository's base directory to /base, so the
// rules live next to the config file at <config-dir>/rules while base_path
// ("base") points one level deeper, at <config-dir>/base.
//
// A ruleset target is written relative to the directory holding the rules
// directory ("rules/custom.list"), so it must resolve to
// <config-dir>/rules/custom.list.
func TestResolveRulesetPath_ContainerLayout(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)

	// base_path is "base", so GetBasePath() is <root>/base — the templates
	// directory, which holds no rules.
	setRulesetBasePath(t, "base")
	writeFile(t, filepath.Join(root, "base", "all_base.tpl"), "proxies: ~\n")
	writeFile(t, filepath.Join(root, "rules", "custom.list"), "DOMAIN-SUFFIX,example.invalid\n")

	got, err := ResolveRulesetPath("rules/custom.list")
	if err != nil {
		t.Fatalf("container-layout ruleset must resolve: %v", err)
	}
	want := filepath.Join(root, "rules", "custom.list")
	if got != want {
		t.Fatalf("resolved to %q, want %q", got, want)
	}
}

// TestResolveRulesetPath_CheckoutLayout pins the layout of a checkout running
// from the repository root, where the rules live under base_path/rules.
func TestResolveRulesetPath_CheckoutLayout(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)

	setRulesetBasePath(t, "base")
	writeFile(t, filepath.Join(root, "base", "rules", "custom.list"), "DOMAIN-SUFFIX,example.invalid\n")

	got, err := ResolveRulesetPath("rules/custom.list")
	if err != nil {
		t.Fatalf("checkout-layout ruleset must resolve: %v", err)
	}
	want := filepath.Join(root, "base", "rules", "custom.list")
	if got != want {
		t.Fatalf("resolved to %q, want %q", got, want)
	}
}

// TestResolveRulesetPath_ConfinedToRulesRoot pins that resolution stays
// inside the rules directory: a target naming a file beside it (the config
// file itself) must not resolve.
func TestResolveRulesetPath_ConfinedToRulesRoot(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)

	setRulesetBasePath(t, "base")
	writeFile(t, filepath.Join(root, "pref.toml"), "api_access_token = \"secret\"\n")
	writeFile(t, filepath.Join(root, "rules", "custom.list"), "DOMAIN-SUFFIX,example.invalid\n")

	if got, err := ResolveRulesetPath("pref.toml"); err == nil {
		t.Fatalf("target outside the rules root must not resolve, got %q", got)
	}
	if got, err := ResolveRulesetPath("rules/../pref.toml"); err == nil {
		t.Fatalf("traversal out of the rules root must not resolve, got %q", got)
	}
	if _, err := ResolveRulesetPath("/etc/passwd"); err == nil {
		t.Fatal("absolute target must not resolve")
	}
}
