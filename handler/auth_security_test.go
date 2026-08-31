package handler

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gfunc/subconvergo/cache"
	"github.com/gfunc/subconvergo/config"
	"github.com/gin-gonic/gin"
)

// protectedEndpointCase is one token-guarded endpoint with a request that,
// once past the token gate, fails cheaply for another reason (missing params,
// missing files, unwritable reload) with a non-403 status.
type protectedEndpointCase struct {
	name   string
	rawURL string
	handle func(h *SubHandler, c *gin.Context)
}

func exercise(h func(h *SubHandler, c *gin.Context), rawURL string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, rawURL, nil)
	h(NewSubHandler(), c)
	return w
}

func withToken(rawURL, token string) string {
	sep := "?"
	if strings.Contains(rawURL, "?") {
		sep = "&"
	}
	return rawURL + sep + "token=" + token
}

// setupPlainProfile creates a base dir holding one profile without a
// profile_token, so /getprofile reaches the global-token comparison. It
// returns the base dir; callers must re-apply it to config.Global because
// /readconf's ReloadConfig replaces the Global pointer.
func setupPlainProfile(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "profiles"), 0o755); err != nil {
		t.Fatal(err)
	}
	content := "[Profile]\ntarget=clash\nurl=ss://YWVzLTEyOC1nY206dGVzdA@1.2.3.4:8388#plain\n"
	if err := os.WriteFile(filepath.Join(dir, "profiles", "plain.ini"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

var protectedEndpoints = []protectedEndpointCase{
	{"readconf", "/readconf", func(h *SubHandler, c *gin.Context) { h.HandleReadConf(c) }},
	{"getruleset", "/getruleset?url=&type=", func(h *SubHandler, c *gin.Context) { h.HandleGetRuleset(c) }},
	{"render", "/render?path=x.tpl", func(h *SubHandler, c *gin.Context) { h.HandleRender(c) }},
	{"getprofile", "/getprofile?name=plain", func(h *SubHandler, c *gin.Context) { h.HandleGetProfile(c) }},
	{"flushcache", "/flushcache", func(h *SubHandler, c *gin.Context) { h.HandleFlushCache(c) }},
}

// TestProtectedEndpoint_EmptyTokenFailsClosed: with no api_access_token
// configured, every protected endpoint must refuse ALL requests — an empty
// configured token must not mean "allow everyone".
func TestProtectedEndpoint_EmptyTokenFailsClosed(t *testing.T) {
	saveGlobal(t)
	cache.Init(t.TempDir())
	gin.SetMode(gin.TestMode)
	baseDir := setupPlainProfile(t)

	for _, tc := range protectedEndpoints {
		for _, token := range []string{"anything", ""} {
			// Re-apply per case: HandleReadConf reloads config, which
			// replaces the Global pointer with fresh defaults.
			config.Global.Common.APIAccessToken = ""
			config.Global.Common.BasePath = baseDir

			rawURL := tc.rawURL
			if token != "" {
				rawURL = withToken(rawURL, token)
			}
			w := exercise(tc.handle, rawURL)
			if w.Code != http.StatusForbidden {
				t.Errorf("%s (token param %q): empty configured token must fail closed with 403, got %d body=%q",
					tc.name, token, w.Code, w.Body.String())
			}
		}
	}
}

// TestProtectedEndpoint_WrongTokenRefused pins both directions of the gate:
// a wrong token is refused, the correct token passes the gate (the request
// then fails for an unrelated reason or succeeds, but never with 403).
func TestProtectedEndpoint_WrongTokenRefused(t *testing.T) {
	saveGlobal(t)
	cache.Init(t.TempDir())
	gin.SetMode(gin.TestMode)
	baseDir := setupPlainProfile(t)

	for _, tc := range protectedEndpoints {
		config.Global.Common.APIAccessToken = "s3cret-token"
		config.Global.Common.BasePath = baseDir

		w := exercise(tc.handle, withToken(tc.rawURL, "wrong-token"))
		if w.Code != http.StatusForbidden {
			t.Errorf("%s: wrong token must be refused with 403, got %d body=%q", tc.name, w.Code, w.Body.String())
		}

		config.Global.Common.APIAccessToken = "s3cret-token"
		config.Global.Common.BasePath = baseDir
		w = exercise(tc.handle, tc.rawURL)
		if w.Code != http.StatusForbidden {
			t.Errorf("%s: missing token must be refused with 403, got %d body=%q", tc.name, w.Code, w.Body.String())
		}

		config.Global.Common.APIAccessToken = "s3cret-token"
		config.Global.Common.BasePath = baseDir
		w = exercise(tc.handle, withToken(tc.rawURL, "s3cret-token"))
		if w.Code == http.StatusForbidden {
			t.Errorf("%s: correct token must pass the gate, got 403", tc.name)
		}
	}
}
