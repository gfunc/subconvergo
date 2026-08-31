package parser

import (
	"strings"
	"testing"
)

// Regression: the top-level url= fragment (#tag) is applied via SetRemark AFTER
// the sanitized ToMihomoProxy funnel, so it must be sanitized at the parseURL
// boundary. Pre-fix, a fragment carrying %0A injected raw newlines into
// line-oriented outputs (verified live via /sub?target=quanx&url=ss://...%23%250A...).
func TestSubParser_URLFragmentTagIsSanitized(t *testing.T) {
	sp := &SubParser{
		URL: "ss://YWVzLTEyOC1nY206dGVzdA@1.2.3.4:8388#%0A[rewrite_local]%0A^https://evil",
	}
	sc, err := sp.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if len(sc.Proxies) == 0 {
		t.Fatal("expected at least one proxy")
	}
	for _, p := range sc.Proxies {
		remark := p.GetRemark()
		if strings.ContainsAny(remark, "\r\n\x00") {
			t.Fatalf("remark contains control characters after sanitization: %q", remark)
		}
	}
}
