package handler

import (
	"encoding/base64"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gfunc/subconvergo/cache"
	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/fetcher"
	_ "github.com/gfunc/subconvergo/generator/impl"
	"github.com/gin-gonic/gin"
)

const (
	rulesetMarker      = "SECRET-RULESET-MARKER"
	internalNodeMarker = "INTERNAL-SSRF-NODE"
	configMarkerGroup  = "PWNED-INTERNAL-GROUP"
	publicNodeMarker   = "PUBLIC-NODE-MARKER"
)

// saveGlobal snapshots config.Global and restores it after the test.
func saveGlobal(t *testing.T) {
	t.Helper()
	saved := *config.Global
	t.Cleanup(func() { *config.Global = saved })
}

// allowLoopbackFetch installs a process-wide fetcher override that treats
// loopback as allowed, simulating a public destination without real egress.
func allowLoopbackFetch(t *testing.T) {
	t.Helper()
	fetcher.SetGlobal(fetcher.New(fetcher.DefaultLimits(), fetcher.WithAllowedHosts("127.0.0.1", "::1", "localhost")))
	t.Cleanup(func() { fetcher.SetGlobal(nil) })
}

// setupSubTestEnv pins the global config fields the /sub tests depend on.
func setupSubTestEnv(t *testing.T) {
	t.Helper()
	saveGlobal(t)
	cache.Init(t.TempDir())
	gin.SetMode(gin.TestMode)
	config.Global.Common.APIMode = true
	config.Global.Common.APIAccessToken = ""
	config.Global.Common.ClashRuleBase = ""
	config.Global.Common.EnableInsert = false
	config.Global.Common.InsertURL = nil
	config.Global.ProxyGroups.CustomProxyGroups = []config.ProxyGroupConfig{
		{Name: "Auto", Type: "select", Rule: []string{".*"}},
	}
	config.Global.Advanced.EnableCache = false
	config.Global.Advanced.MaxAllowedDownloadSize = 0
	config.Global.ManagedConfig.WriteManagedConfig = false
}

func doGetRuleset(h *SubHandler, rawURL string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, rawURL, nil)
	h.HandleGetRuleset(c)
	return w
}

func TestGetRuleset_RefusesLoopbackURL(t *testing.T) {
	saveGlobal(t)
	cache.Init(t.TempDir())
	gin.SetMode(gin.TestMode)
	// A valid token keeps this test on the fetch guard (fail-closed token
	// checks would otherwise refuse the request before it).
	config.Global.Common.APIAccessToken = "test-token"
	config.Global.Advanced.EnableCache = false
	h := NewSubHandler()

	var hits int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		_, _ = io.WriteString(w, "DOMAIN-SUFFIX,internal.example,"+rulesetMarker+"\n")
	}))
	defer ts.Close()

	encoded := base64.URLEncoding.EncodeToString([]byte(ts.URL))
	w := doGetRuleset(h, "/getruleset?url="+encoded+"&type=clash&token=test-token")

	if w.Code == http.StatusOK {
		t.Fatalf("loopback URL must be refused, got 200 with body %q", w.Body.String())
	}
	if strings.Contains(w.Body.String(), rulesetMarker) {
		t.Fatalf("internal ruleset content leaked in response: %q", w.Body.String())
	}
	if n := atomic.LoadInt32(&hits); n != 0 {
		t.Fatalf("internal upstream was contacted %d time(s)", n)
	}
}

func TestGetRuleset_RefusesRedirectToInternal(t *testing.T) {
	saveGlobal(t)
	cache.Init(t.TempDir())
	gin.SetMode(gin.TestMode)
	config.Global.Common.APIAccessToken = "test-token"
	config.Global.Advanced.EnableCache = false
	h := NewSubHandler()

	var secretHits int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/secret" {
			atomic.AddInt32(&secretHits, 1)
			_, _ = io.WriteString(w, "DOMAIN-SUFFIX,internal.example,"+rulesetMarker+"\n")
			return
		}
		w.Header().Set("Location", "/secret")
		w.WriteHeader(http.StatusFound)
	}))
	defer ts.Close()

	encoded := base64.URLEncoding.EncodeToString([]byte(ts.URL))
	w := doGetRuleset(h, "/getruleset?url="+encoded+"&type=clash&token=test-token")

	if w.Code == http.StatusOK {
		t.Fatalf("redirect to internal target must be refused, got 200 with body %q", w.Body.String())
	}
	if strings.Contains(w.Body.String(), rulesetMarker) {
		t.Fatalf("internal ruleset content leaked via redirect: %q", w.Body.String())
	}
	if n := atomic.LoadInt32(&secretHits); n != 0 {
		t.Fatalf("redirect target was fetched %d time(s)", n)
	}
}

func TestGetRuleset_RejectsOversizedBody(t *testing.T) {
	saveGlobal(t)
	cache.Init(t.TempDir())
	gin.SetMode(gin.TestMode)
	config.Global.Common.APIAccessToken = "test-token"
	config.Global.Advanced.EnableCache = false
	// 64-byte cap; the origin serves far more.
	config.Global.Advanced.MaxAllowedDownloadSize = 64
	h := NewSubHandler()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, rulesetMarker+strings.Repeat("x", 256))
	}))
	defer ts.Close()

	encoded := base64.URLEncoding.EncodeToString([]byte(ts.URL))
	w := doGetRuleset(h, "/getruleset?url="+encoded+"&type=clash&token=test-token")

	if w.Code == http.StatusOK {
		t.Fatalf("oversized ruleset body must be rejected, got 200")
	}
	if strings.Contains(w.Body.String(), rulesetMarker) {
		t.Fatalf("over-limit body leaked into response")
	}
}

func TestGetRuleset_RequiresToken(t *testing.T) {
	saveGlobal(t)
	cache.Init(t.TempDir())
	gin.SetMode(gin.TestMode)
	config.Global.Common.APIAccessToken = "secret-token"
	config.Global.Advanced.EnableCache = false
	h := NewSubHandler()

	var hits int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		_, _ = io.WriteString(w, rulesetMarker)
	}))
	defer ts.Close()

	encoded := base64.URLEncoding.EncodeToString([]byte(ts.URL))
	for _, rawURL := range []string{
		"/getruleset?url=" + encoded + "&type=clash",             // no token
		"/getruleset?url=" + encoded + "&type=clash&token=wrong", // wrong token
	} {
		w := doGetRuleset(h, rawURL)
		if w.Code != http.StatusForbidden && w.Code != http.StatusUnauthorized {
			t.Fatalf("%s: expected 401/403, got %d", rawURL, w.Code)
		}
		if strings.Contains(w.Body.String(), rulesetMarker) {
			t.Fatalf("%s: upstream content leaked without valid token", rawURL)
		}
	}
	if n := atomic.LoadInt32(&hits); n != 0 {
		t.Fatalf("upstream was contacted %d time(s) without a valid token", n)
	}
}

// TestSafeFetch_AllowedPublicURLStillWorks is the green-path guard: an
// allowlisted (simulated public) destination must still be served.
func TestSafeFetch_AllowedPublicURLStillWorks(t *testing.T) {
	saveGlobal(t)
	cache.Init(t.TempDir())
	gin.SetMode(gin.TestMode)
	config.Global.Common.APIAccessToken = "test-token"
	config.Global.Advanced.EnableCache = false
	allowLoopbackFetch(t)
	h := NewSubHandler()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "DOMAIN-SUFFIX,example.com,Auto\nMATCH,Auto\n")
	}))
	defer ts.Close()

	encoded := base64.URLEncoding.EncodeToString([]byte(ts.URL))
	w := doGetRuleset(h, "/getruleset?url="+encoded+"&type=clash&token=test-token")

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for allowed destination, got %d: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "MATCH,Auto") {
		t.Fatalf("unexpected ruleset body: %q", w.Body.String())
	}
}

func doSub(h *SubHandler, rawURL string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, rawURL, nil)
	h.HandleSub(c)
	return w
}

func TestSub_RefusesInternalSubscriptionURL(t *testing.T) {
	setupSubTestEnv(t)
	h := NewSubHandler()

	sub := base64.StdEncoding.EncodeToString([]byte("ss://YWVzLTEyOC1nY206dGVzdA@10.0.0.1:8388#" + internalNodeMarker))
	var hits int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		_, _ = io.WriteString(w, sub)
	}))
	defer ts.Close()

	w := doSub(h, "/sub?target=clash&url="+url.QueryEscape(ts.URL))

	if w.Code == http.StatusOK {
		t.Fatalf("internal subscription URL must be refused, got 200")
	}
	if strings.Contains(w.Body.String(), internalNodeMarker) {
		t.Fatalf("internal subscription content leaked in response")
	}
	if n := atomic.LoadInt32(&hits); n != 0 {
		t.Fatalf("internal subscription upstream was contacted %d time(s)", n)
	}
}

func TestSub_RefusesInternalConfigURL(t *testing.T) {
	setupSubTestEnv(t)
	h := NewSubHandler()

	var hits int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		_, _ = io.WriteString(w, "proxy_groups:\n  custom_proxy_group:\n    - name: "+configMarkerGroup+"\n      type: select\n      rule: ['.*']\n")
	}))
	defer ts.Close()

	rawURL := "/sub?target=clash" +
		"&url=" + url.QueryEscape("ss://YWVzLTEyOC1nY206dGVzdA@1.2.3.4:8388#plainnode") +
		"&config=" + url.QueryEscape(ts.URL)
	w := doSub(h, rawURL)

	if n := atomic.LoadInt32(&hits); n != 0 {
		t.Fatalf("internal config URL was fetched %d time(s)", n)
	}
	if strings.Contains(w.Body.String(), configMarkerGroup) {
		t.Fatalf("attacker-controlled external config was applied:\n%s", w.Body.String())
	}
}

func TestSub_CapsSourceCount(t *testing.T) {
	setupSubTestEnv(t)
	h := NewSubHandler()

	// 33 sources, one over the 32-source cap.
	links := make([]string, 33)
	for i := range links {
		links[i] = "ss://YWVzLTEyOC1nY206dGVzdA@1.2.3.4:8388#node" + strconv.Itoa(i)
	}
	rawURL := "/sub?target=clash&url=" + url.QueryEscape(strings.Join(links, "|"))
	w := doSub(h, rawURL)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for %d source URLs, got %d", len(links), w.Code)
	}
	if !strings.Contains(w.Body.String(), "Too many") {
		t.Fatalf("expected a source-count error, got: %q", w.Body.String())
	}
}

func TestSub_SourceURLLengthBounded(t *testing.T) {
	setupSubTestEnv(t)
	h := NewSubHandler()

	// A syntactically valid source URL far beyond the 2048-byte bound.
	long := "ss://YWVzLTEyOC1nY206dGVzdA@1.2.3.4:8388#" + strings.Repeat("a", 4096)
	w := doSub(h, "/sub?target=clash&url="+url.QueryEscape(long))

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for oversized source URL, got %d", w.Code)
	}
}

func TestSub_RefusesRedirectToInternal(t *testing.T) {
	setupSubTestEnv(t)
	h := NewSubHandler()

	sub := base64.StdEncoding.EncodeToString([]byte("ss://YWVzLTEyOC1nY206dGVzdA@10.0.0.1:8388#" + internalNodeMarker))
	var secretHits int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/secret" {
			atomic.AddInt32(&secretHits, 1)
			_, _ = io.WriteString(w, sub)
			return
		}
		w.Header().Set("Location", "/secret")
		w.WriteHeader(http.StatusFound)
	}))
	defer ts.Close()

	w := doSub(h, "/sub?target=clash&url="+url.QueryEscape(ts.URL))

	if w.Code == http.StatusOK {
		t.Fatalf("subscription redirect to internal target must be refused, got 200")
	}
	if strings.Contains(w.Body.String(), internalNodeMarker) {
		t.Fatalf("internal subscription content leaked via redirect")
	}
	if n := atomic.LoadInt32(&secretHits); n != 0 {
		t.Fatalf("redirect target was fetched %d time(s)", n)
	}
}

func TestSub_EnforcesDownloadSizeCap(t *testing.T) {
	setupSubTestEnv(t)
	// 64-byte cap; the subscription body is far larger.
	config.Global.Advanced.MaxAllowedDownloadSize = 64
	h := NewSubHandler()

	sub := base64.StdEncoding.EncodeToString([]byte("ss://YWVzLTEyOC1nY206dGVzdA@c.example.com:8388#" + internalNodeMarker + strings.Repeat("x", 256)))
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, sub)
	}))
	defer ts.Close()

	w := doSub(h, "/sub?target=clash&url="+url.QueryEscape(ts.URL))

	if w.Code == http.StatusOK {
		t.Fatalf("oversized subscription body must be rejected, got 200")
	}
	if strings.Contains(w.Body.String(), internalNodeMarker) {
		t.Fatalf("over-limit subscription content leaked into response")
	}
}

// TestSub_PublicSubscriptionUnchanged is the green-path guard: a normal
// subscription conversion must still work through the new fetcher.
func TestSub_PublicSubscriptionUnchanged(t *testing.T) {
	setupSubTestEnv(t)
	allowLoopbackFetch(t)
	h := NewSubHandler()

	sub := base64.StdEncoding.EncodeToString([]byte("ss://YWVzLTEyOC1nY206dGVzdA@c.example.com:8388#" + publicNodeMarker))
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, sub)
	}))
	defer ts.Close()

	w := doSub(h, "/sub?target=clash&url="+url.QueryEscape(ts.URL))

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for allowed subscription, got %d: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), publicNodeMarker) {
		t.Fatalf("subscription node missing from output:\n%s", w.Body.String())
	}
}

// TestSub_SlowOriginIsCutOff drives the external-config path (historically a
// bare http.Get with no timeout) against an origin that stalls; the fetch
// must fail within a bounded time.
func TestSub_SlowOriginIsCutOff(t *testing.T) {
	setupSubTestEnv(t)
	fetcher.SetGlobal(fetcher.New(
		fetcher.Limits{MaxBodyBytes: 1 << 20, Timeout: 200 * time.Millisecond, MaxRedirects: 5},
		fetcher.WithAllowedHosts("127.0.0.1", "::1"),
	))
	t.Cleanup(func() { fetcher.SetGlobal(nil) })
	h := NewSubHandler()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1200 * time.Millisecond)
		_, _ = io.WriteString(w, "proxy_groups:\n  custom_proxy_group: []\n")
	}))
	defer ts.Close()

	rawURL := "/sub?target=clash" +
		"&url=" + url.QueryEscape("ss://YWVzLTEyOC1nY206dGVzdA@1.2.3.4:8388#plainnode") +
		"&config=" + url.QueryEscape(ts.URL)

	start := time.Now()
	doSub(h, rawURL)
	elapsed := time.Since(start)

	if elapsed > 800*time.Millisecond {
		t.Fatalf("slow origin was not cut off: request took %v", elapsed)
	}
}
