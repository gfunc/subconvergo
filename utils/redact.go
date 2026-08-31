package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Log redaction.
//
// Logs must never carry credential-bearing material: subscription/config URLs
// embed access tokens (in the query, the path, or the userinfo), and request
// headers carry Authorization values. Every log line that mentions a URL or
// request headers goes through the helpers below so redaction lives in one
// place instead of per-call-site string surgery.

// sensitiveQueryKey matches query keys that commonly carry credentials.
var sensitiveQueryKey = regexp.MustCompile(`(?i)token|key|auth|password|uuid|secret`)

// URLHash returns a stable short hash of raw so repeated log lines about the
// same URL can be correlated without revealing the URL itself.
func URLHash(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])[:8]
}

// RedactURL renders raw for logging: scheme://host with userinfo stripped,
// the path replaced by /<redacted> (subscription paths can embed tokens),
// query keys matching token|key|auth|password|uuid|secret (case-insensitive)
// dropped, and a stable correlation hash appended as sig=<hash>.
//
// Non-http(s) URLs (ss://, vmess://, trojan://, ...) embed credentials in
// what url.Parse reports as the host or opaque part, so only their scheme is
// shown. Unparseable input degrades to a bare redacted marker plus the hash.
func RedactURL(raw string) string {
	sig := URLHash(raw)
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" {
		return "<redacted-url sig=" + sig + ">"
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return u.Scheme + ":<redacted> sig=" + sig
	}

	var b strings.Builder
	b.WriteString(u.Scheme)
	b.WriteString("://")
	// u.Host excludes the userinfo by construction; u.User is never read.
	b.WriteString(u.Host)
	b.WriteString("/<redacted>")

	q := u.Query()
	kept := make([]string, 0, len(q))
	for k := range q {
		if sensitiveQueryKey.MatchString(k) {
			continue
		}
		kept = append(kept, k+"="+q.Get(k))
	}
	sort.Strings(kept)
	if len(kept) > 0 {
		b.WriteString("?")
		b.WriteString(strings.Join(kept, "&"))
	}
	b.WriteString(" sig=")
	b.WriteString(sig)
	return b.String()
}

// RedactURLOrPath redacts p when it looks like a URL and otherwise returns it
// unchanged: local file paths carry no credentials and stay useful in logs.
func RedactURLOrPath(p string) string {
	if strings.Contains(p, "://") {
		return RedactURL(p)
	}
	return p
}

// HeaderSummary renders the allowlisted request headers for logging. Only
// User-Agent is on the allowlist; everything else (Authorization, Cookie,
// tokens echoed by clients) stays out of the logs.
func HeaderSummary(h http.Header) string {
	return "ua=" + strconv.Quote(h.Get("User-Agent"))
}
