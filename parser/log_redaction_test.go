package parser

import (
	"bytes"
	"encoding/base64"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/fetcher"
)

// TestParser_DoesNotLogURLUserinfo: a subscription URL carrying credentials
// in the userinfo part must not land in the parser's fetch/parse logs.
func TestParser_DoesNotLogURLUserinfo(t *testing.T) {
	saved := *config.Global
	t.Cleanup(func() { *config.Global = saved })
	config.Global.Advanced.EnableCache = false

	// The safe fetcher refuses loopback by default; allow it so the local
	// test server stands in for a public origin without real egress.
	fetcher.SetGlobal(fetcher.New(fetcher.DefaultLimits(), fetcher.WithAllowedHosts("127.0.0.1", "::1")))
	t.Cleanup(func() { fetcher.SetGlobal(nil) })

	var buf bytes.Buffer
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })

	sub := base64.StdEncoding.EncodeToString([]byte("ss://YWVzLTEyOC1nY206dGVzdA@c.example.com:8388#node"))
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, sub)
	}))
	defer ts.Close()

	sp := &SubParser{
		Index: 0,
		URL:   "http://user:SECRET-MARKER@" + strings.TrimPrefix(ts.URL, "http://") + "/sub",
	}
	if _, err := sp.Parse(); err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if strings.Contains(buf.String(), "SECRET-MARKER") {
		t.Fatalf("URL userinfo leaked into logs:\n%s", buf.String())
	}
}
