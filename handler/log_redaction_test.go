package handler

import (
	"bytes"
	"encoding/base64"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

const logSecretMarker = "SECRET-MARKER"

// captureLogs redirects the standard logger into a buffer for the test.
func captureLogs(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })
	return &buf
}

// newSubServer serves one fixed base64 subscription body.
func newSubServer(t *testing.T, body string) *httptest.Server {
	t.Helper()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, body)
	}))
	t.Cleanup(ts.Close)
	return ts
}

var testSubBody = base64.StdEncoding.EncodeToString([]byte("ss://YWVzLTEyOC1nY206dGVzdA@c.example.com:8388#node"))

// TestSub_DoesNotLogSubscriptionSecrets: a subscription URL carrying a
// credential query parameter must not land in the logs.
func TestSub_DoesNotLogSubscriptionSecrets(t *testing.T) {
	setupSubTestEnv(t)
	allowLoopbackFetch(t)
	buf := captureLogs(t)
	h := NewSubHandler()

	ts := newSubServer(t, testSubBody)
	w := doSub(h, "/sub?target=clash&url="+url.QueryEscape(ts.URL+"/sub?token="+logSecretMarker))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if strings.Contains(buf.String(), logSecretMarker) {
		t.Fatalf("subscription credential leaked into logs:\n%s", buf.String())
	}
}

// TestSub_DoesNotLogAuthorizationHeader: request headers must not be dumped
// wholesale; only an allowlist (User-Agent) may be logged.
func TestSub_DoesNotLogAuthorizationHeader(t *testing.T) {
	setupSubTestEnv(t)
	allowLoopbackFetch(t)
	buf := captureLogs(t)
	h := NewSubHandler()

	ts := newSubServer(t, testSubBody)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/sub?target=clash&url="+url.QueryEscape(ts.URL), nil)
	c.Request.Header.Set("Authorization", "Bearer "+logSecretMarker)
	h.HandleSub(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if strings.Contains(buf.String(), logSecretMarker) {
		t.Fatalf("Authorization header leaked into logs:\n%s", buf.String())
	}
}

var sigRe = regexp.MustCompile(`sig=([0-9a-f]+)`)

// TestLogs_StillCorrelatable: redaction must keep logs debuggable — the
// origin host, non-sensitive query parameters, and a stable per-URL hash
// survive; only credentials are stripped.
func TestLogs_StillCorrelatable(t *testing.T) {
	setupSubTestEnv(t)
	allowLoopbackFetch(t)
	buf := captureLogs(t)
	h := NewSubHandler()

	ts := newSubServer(t, testSubBody)
	rawURL := "/sub?target=clash&url=" + url.QueryEscape(ts.URL+"/sub?foo=bar&token="+logSecretMarker)
	for i := 0; i < 2; i++ {
		w := doSub(h, rawURL)
		if w.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d: %s", i, w.Code, w.Body.String())
		}
	}

	out := buf.String()
	if strings.Contains(out, logSecretMarker) {
		t.Fatalf("credential leaked into logs:\n%s", out)
	}
	// The redacted URL keeps scheme://host but not the credential path.
	redactedOrigin := "url=http://" + strings.TrimPrefix(ts.URL, "http://") + "/<redacted>"
	if !strings.Contains(out, redactedOrigin) {
		t.Fatalf("redacted log line must keep the origin host %q:\n%s", redactedOrigin, out)
	}
	// Non-sensitive query parameters survive for debugging.
	if !strings.Contains(out, "foo=bar") {
		t.Fatalf("non-sensitive query parameter must survive redaction:\n%s", out)
	}
	// The raw URL with its real path must be gone.
	if strings.Contains(out, ts.URL+"/sub?") {
		t.Fatalf("raw subscription URL leaked into logs:\n%s", out)
	}
	// The correlation hash is present and stable across requests of the
	// same URL.
	sigs := sigRe.FindAllStringSubmatch(out, -1)
	if len(sigs) < 2 {
		t.Fatalf("expected correlation hashes in logs, got %d:\n%s", len(sigs), out)
	}
	for _, s := range sigs[1:] {
		if s[1] != sigs[0][1] {
			t.Fatalf("correlation hash must be stable for the same URL: %q vs %q", sigs[0][1], s[1])
		}
	}
}
