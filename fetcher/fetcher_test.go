package fetcher

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

const fetchTestMarker = "FETCHER-TEST-MARKER-CONTENT"

func TestDefaultLimitsAreFinite(t *testing.T) {
	l := DefaultLimits()
	if l.MaxBodyBytes != 32<<20 {
		t.Errorf("expected default body cap %d, got %d", int64(32<<20), l.MaxBodyBytes)
	}
	if l.Timeout != 30*time.Second {
		t.Errorf("expected default timeout 30s, got %v", l.Timeout)
	}
	if l.MaxRedirects != 5 {
		t.Errorf("expected default redirect cap 5, got %d", l.MaxRedirects)
	}
}

func TestLimitsFromConfig(t *testing.T) {
	if got := LimitsFromConfig(0).MaxBodyBytes; got != DefaultMaxBodyBytes {
		t.Errorf("unset size must fall back to finite default %d, got %d", DefaultMaxBodyBytes, got)
	}
	if got := LimitsFromConfig(-5).MaxBodyBytes; got != DefaultMaxBodyBytes {
		t.Errorf("negative size must fall back to finite default %d, got %d", DefaultMaxBodyBytes, got)
	}
	if got := LimitsFromConfig(1024).MaxBodyBytes; got != 1024 {
		t.Errorf("configured size must be honored, got %d", got)
	}
}

func TestFetcher_BlocksNonHTTPSchemes(t *testing.T) {
	f := New(DefaultLimits())
	for _, rawURL := range []string{
		"file:///etc/passwd",
		"ftp://example.com/x",
		"gopher://example.com/",
		"file://C:/Windows/win.ini",
	} {
		_, err := f.Get(rawURL, nil)
		var se *SchemeError
		if !errors.As(err, &se) {
			t.Errorf("%s: expected *SchemeError, got %v", rawURL, err)
		}
	}
}

func TestFetcher_BlocksIPLiterals(t *testing.T) {
	f := New(DefaultLimits())
	cases := []struct {
		host string
		ip   string
	}{
		{"127.0.0.1", "127.0.0.1"},
		{"10.1.2.3", "10.1.2.3"},
		{"172.16.0.9", "172.16.0.9"},
		{"192.168.1.9", "192.168.1.9"},
		{"169.254.169.254", "169.254.169.254"}, // cloud metadata
		{"100.64.0.1", "100.64.0.1"},           // CGNAT
		{"198.18.0.1", "198.18.0.1"},           // benchmarking
		{"224.0.0.1", "224.0.0.1"},             // multicast
		{"240.0.0.1", "240.0.0.1"},             // reserved
		{"0.0.0.0", "0.0.0.0"},                 // unspecified
		{"[::1]", "::1"},                       // v6 loopback
		{"[fc00::1]", "fc00::1"},               // v6 ULA
		{"[fe80::1]", "fe80::1"},               // v6 link-local
		{"[ff02::1]", "ff02::1"},               // v6 multicast
		{"[::ffff:127.0.0.1]", "127.0.0.1"},    // v4-mapped loopback
	}
	for _, tc := range cases {
		_, err := f.Get("http://"+tc.host+":9/", nil)
		var be *BlockedAddressError
		if !errors.As(err, &be) {
			t.Errorf("%s: expected *BlockedAddressError, got %v", tc.host, err)
			continue
		}
		if want := netip.MustParseAddr(tc.ip); be.IP != want {
			t.Errorf("%s: expected blocked IP %s, got %s", tc.host, want, be.IP)
		}
	}
}

func TestFetcher_BlocksLocalhostByName(t *testing.T) {
	// Hostnames are resolved at dial time; localhost resolves to loopback and
	// must be refused even though it is not an IP literal.
	f := New(DefaultLimits())
	_, err := f.Get("http://localhost:9/", nil)
	var be *BlockedAddressError
	if !errors.As(err, &be) {
		t.Fatalf("expected *BlockedAddressError for localhost, got %v", err)
	}
}

func TestFetcher_RedirectHopToBlockedAddressRefused(t *testing.T) {
	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", "http://169.254.169.254/latest/meta-data")
		w.WriteHeader(http.StatusFound)
	}))
	defer origin.Close()

	// The origin itself is allowlisted (simulating a public destination); the
	// redirect target must still be refused at the hop, before any dial.
	f := New(DefaultLimits(), WithAllowedHosts("127.0.0.1", "::1"))
	_, err := f.Get(origin.URL, nil)
	var be *BlockedAddressError
	if !errors.As(err, &be) {
		t.Fatalf("expected *BlockedAddressError at redirect hop, got %v", err)
	}
	if want := netip.MustParseAddr("169.254.169.254"); be.IP != want {
		t.Fatalf("expected blocked redirect target %s, got %s", want, be.IP)
	}
}

func TestFetcher_RedirectsCapped(t *testing.T) {
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		w.Header().Set("Location", "/loop")
		w.WriteHeader(http.StatusFound)
	}))
	defer srv.Close()

	limits := DefaultLimits()
	limits.MaxRedirects = 2
	f := New(limits, WithAllowedHosts("127.0.0.1", "::1"))
	_, err := f.Get(srv.URL+"/loop", nil)
	if !errors.Is(err, ErrTooManyRedirects) {
		t.Fatalf("expected ErrTooManyRedirects, got %v", err)
	}
	if got := atomic.LoadInt32(&hits); got != 3 {
		t.Fatalf("expected 3 requests (initial + 2 follows), got %d", got)
	}
}

func TestFetcher_EnforcesBodyCap(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, fetchTestMarker+strings.Repeat("x", 4096))
	}))
	defer srv.Close()

	limits := DefaultLimits()
	limits.MaxBodyBytes = 1024
	f := New(limits, WithAllowedHosts("127.0.0.1", "::1"))
	body, err := f.Get(srv.URL, nil)
	var tl *ResponseTooLargeError
	if !errors.As(err, &tl) {
		t.Fatalf("expected *ResponseTooLargeError, got body=%d bytes err=%v", len(body), err)
	}
	if tl.Limit != 1024 {
		t.Fatalf("expected limit 1024 in error, got %d", tl.Limit)
	}
	if strings.Contains(string(body), fetchTestMarker) {
		t.Fatalf("over-limit body must not be returned")
	}
}

func TestFetcher_TimeoutCutsSlowOrigin(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(600 * time.Millisecond)
		_, _ = io.WriteString(w, fetchTestMarker)
	}))
	defer srv.Close()

	limits := DefaultLimits()
	limits.Timeout = 150 * time.Millisecond
	f := New(limits, WithAllowedHosts("127.0.0.1", "::1"))
	start := time.Now()
	_, err := f.Get(srv.URL, nil)
	elapsed := time.Since(start)
	if err == nil {
		t.Fatalf("expected timeout error, got success")
	}
	if elapsed > 500*time.Millisecond {
		t.Fatalf("fetch was not cut off in time: %v", elapsed)
	}
}

func TestFetcher_StatusError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	f := New(DefaultLimits(), WithAllowedHosts("127.0.0.1", "::1"))
	_, err := f.Get(srv.URL, nil)
	var he *HTTPStatusError
	if !errors.As(err, &he) {
		t.Fatalf("expected *HTTPStatusError, got %v", err)
	}
	if he.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", he.StatusCode)
	}
}

func TestFetcher_CustomDialContextSimulatesPublic(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, fetchTestMarker)
	}))
	defer srv.Close()
	u, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	port := u.Port()

	var dials int32
	f := New(DefaultLimits(), WithDialContext(func(ctx context.Context, network, address string) (net.Conn, error) {
		atomic.AddInt32(&dials, 1)
		host, p, err := net.SplitHostPort(address)
		if err != nil {
			return nil, err
		}
		if host == "public.example.com" {
			// Simulate a public destination without real egress.
			return (&net.Dialer{}).DialContext(ctx, network, net.JoinHostPort("127.0.0.1", p))
		}
		return nil, fmt.Errorf("unexpected dial host %s", host)
	}))

	body, err := f.Get("http://public.example.com:"+port+"/", nil)
	if err != nil {
		t.Fatalf("simulated public fetch failed: %v", err)
	}
	if !strings.Contains(string(body), fetchTestMarker) {
		t.Fatalf("unexpected body: %q", body)
	}
	if atomic.LoadInt32(&dials) == 0 {
		t.Fatalf("custom dialer was not used")
	}

	// An IP-literal loopback URL must be refused before the injected dialer runs.
	before := atomic.LoadInt32(&dials)
	_, err = f.Get("http://127.0.0.1:"+port+"/", nil)
	var be *BlockedAddressError
	if !errors.As(err, &be) {
		t.Fatalf("expected *BlockedAddressError, got %v", err)
	}
	if got := atomic.LoadInt32(&dials); got != before {
		t.Fatalf("injected dialer ran for a blocked literal (%d -> %d)", before, got)
	}
}

func TestFetcher_HeadersForwarded(t *testing.T) {
	var gotUA string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		_, _ = io.WriteString(w, "ok")
	}))
	defer srv.Close()

	f := New(DefaultLimits(), WithAllowedHosts("127.0.0.1", "::1"))
	if _, err := f.Get(srv.URL, map[string]string{"User-Agent": "subconvergo-test/1.0"}); err != nil {
		t.Fatalf("fetch failed: %v", err)
	}
	if gotUA != "subconvergo-test/1.0" {
		t.Fatalf("expected UA header forwarded, got %q", gotUA)
	}
}

func TestFetcher_AllowedHostsPermitsLoopback(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, fetchTestMarker)
	}))
	defer srv.Close()

	f := New(DefaultLimits(), WithAllowedHosts("127.0.0.1", "::1"))
	body, err := f.Get(srv.URL, nil)
	if err != nil {
		t.Fatalf("allowlisted destination failed: %v", err)
	}
	if !strings.Contains(string(body), fetchTestMarker) {
		t.Fatalf("unexpected body: %q", body)
	}
}

func TestSetGlobalOverride(t *testing.T) {
	if Global() != nil {
		t.Fatalf("expected nil global override initially")
	}
	f := New(DefaultLimits())
	SetGlobal(f)
	if Global() != f {
		t.Fatalf("expected global override to be installed")
	}
	SetGlobal(nil)
	if Global() != nil {
		t.Fatalf("expected global override cleared")
	}
}

func TestForConfigUsesOverride(t *testing.T) {
	f := New(DefaultLimits())
	SetGlobal(f)
	defer SetGlobal(nil)
	if got := ForConfig(1234, "NONE"); got != f {
		t.Fatalf("ForConfig must return the global override when set")
	}
	if got := ForConfig(1234, "NONE"); got == nil {
		t.Fatalf("ForConfig must never return nil")
	}
}
