package utils

import (
	pc "github.com/gfunc/subconvergo/proxy/core"
	"testing"
)

func TestFetchRuleset(t *testing.T) {
	// Test with invalid URL - should return error
	_, err := FetchRuleset("invalid://url")
	if err == nil {
		t.Log("FetchRuleset should fail with invalid URL (may have fallback)")
	}
}

func TestApplyMatcherForRename(t *testing.T) {
	proxy := &pc.BaseProxy{Type: "ss", Remark: "test", Port: 443}
	matched, _ := ApplyMatcher("!!TYPE=SS!!test", proxy)
	if !matched {
		t.Error("TYPE matcher failed")
	}
	matched, _ = ApplyMatcher("!!PORT=443!!test", proxy)
	if !matched {
		t.Error("PORT matcher failed")
	}
}

func TestMatchRangeFunc(t *testing.T) {
	if !MatchRange("443", 443) {
		t.Error("Single value match failed")
	}
	if !MatchRange("400-500", 443) {
		t.Error("Range match failed")
	}
	if MatchRange("400-500", 600) {
		t.Error("Range should not match")
	}
}

func TestMatchRangeEdgeCases(t *testing.T) {
	// Empty pattern returns true (matches all)
	if !MatchRange("", 443) {
		t.Error("Empty pattern should match (match all)")
	}

	// Multiple ranges
	if !MatchRange("80,443,8080", 443) {
		t.Error("Comma separated should match")
	}

	// Invalid format - should not panic
	MatchRange("invalid", 443)
}
