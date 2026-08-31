package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLogFile_CreatedWith0600: explicit log files (-l flag) carry request
// metadata and must be created owner-only.
func TestLogFile_CreatedWith0600(t *testing.T) {
	path := filepath.Join(t.TempDir(), "subconvergo.log")
	f, err := openLogFile(path)
	if err != nil {
		t.Fatalf("openLogFile: %v", err)
	}
	defer f.Close()

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("log file mode = %o, want 600", got)
	}
}
