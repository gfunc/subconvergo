// Package fetcher is the single hardened path for all outbound HTTP fetches
// made from user-supplied URLs (subscriptions, external configs, rulesets).
//
// Protections:
//   - http/https schemes only;
//   - loopback/private/link-local/reserved IPv4+IPv6 destinations (including
//     the cloud metadata address 169.254.169.254) refused at dial time: a
//     custom DialContext revalidates the actually-dialed IP, so there is no
//     TOCTOU gap between a preflight DNS lookup and the dial;
//   - every redirect hop revalidated and redirects capped;
//   - overall client timeout plus transport-level timeouts;
//   - response body capped via io.LimitReader(max+1) with an explicit
//     over-limit error.
//
// Tests can simulate public vs. internal destinations without real egress via
// WithAllowedHosts / WithDialContext and the process-wide SetGlobal override.
package fetcher

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	// DefaultMaxBodyBytes bounds one downloaded body when the config leaves
	// max_allowed_download_size unset (32 MiB).
	DefaultMaxBodyBytes int64 = 32 << 20
	// DefaultTimeout bounds the whole fetch, including reading the body.
	DefaultTimeout = 30 * time.Second
	// DefaultMaxRedirects bounds redirect hops per fetch.
	DefaultMaxRedirects = 5

	dialTimeout           = 10 * time.Second
	tlsHandshakeTimeout   = 10 * time.Second
	responseHeaderTimeout = 20 * time.Second
	idleConnTimeout       = 90 * time.Second
)

// ErrTooManyRedirects is returned when a fetch exceeds the redirect cap.
var ErrTooManyRedirects = errors.New("fetcher: too many redirects")

// SchemeError is returned for non-http(s) URLs.
type SchemeError struct{ Scheme string }

func (e *SchemeError) Error() string {
	return fmt.Sprintf("fetcher: scheme %q is not allowed (http/https only)", e.Scheme)
}

// BlockedAddressError is returned when a destination resolves to a
// loopback/private/link-local/reserved address.
type BlockedAddressError struct {
	Host string
	IP   netip.Addr
}

func (e *BlockedAddressError) Error() string {
	if e.IP.IsValid() {
		return fmt.Sprintf("fetcher: refusing to connect to %s (blocked address %s)", e.Host, e.IP)
	}
	return fmt.Sprintf("fetcher: refusing to connect to %s (blocked address)", e.Host)
}

// ResponseTooLargeError is returned when a body exceeds the configured cap.
type ResponseTooLargeError struct{ Limit int64 }

func (e *ResponseTooLargeError) Error() string {
	return fmt.Sprintf("fetcher: response body exceeds limit of %d bytes", e.Limit)
}

// HTTPStatusError is returned for non-2xx responses.
type HTTPStatusError struct {
	URL        string
	StatusCode int
	Status     string
}

func (e *HTTPStatusError) Error() string {
	return fmt.Sprintf("fetcher: GET %s: HTTP %d: %s", e.URL, e.StatusCode, e.Status)
}

// Limits holds the per-fetch resource bounds.
type Limits struct {
	MaxBodyBytes int64
	Timeout      time.Duration
	MaxRedirects int
}

func (l Limits) withDefaults() Limits {
	if l.MaxBodyBytes <= 0 {
		l.MaxBodyBytes = DefaultMaxBodyBytes
	}
	if l.Timeout <= 0 {
		l.Timeout = DefaultTimeout
	}
	if l.MaxRedirects <= 0 {
		l.MaxRedirects = DefaultMaxRedirects
	}
	return l
}

// DefaultLimits returns the finite default limits.
func DefaultLimits() Limits {
	return Limits{}.withDefaults()
}

// LimitsFromConfig derives fetcher limits from the advanced config value
// max_allowed_download_size (bytes); unset/non-positive values fall back to
// the finite defaults.
func LimitsFromConfig(maxAllowedDownloadSize int) Limits {
	l := DefaultLimits()
	if maxAllowedDownloadSize > 0 {
		l.MaxBodyBytes = int64(maxAllowedDownloadSize)
	}
	return l
}

// DialContextFunc matches the signature of net.Dialer.DialContext.
type DialContextFunc func(ctx context.Context, network, address string) (net.Conn, error)

// Fetcher performs hardened outbound GET fetches.
type Fetcher struct {
	limits      Limits
	allowHosts  map[string]struct{}
	dialContext DialContextFunc // injected dialer (test seam); nil uses the system resolver
	proxyFunc   func(*http.Request) (*url.URL, error)
	dialer      *net.Dialer
	client      *http.Client
}

// Option customizes a Fetcher.
type Option func(*Fetcher)

// WithAllowedHosts exempts the given hostnames/IPs (case-insensitive exact
// match) from the blocked-address checks. Intended for tests.
func WithAllowedHosts(hosts ...string) Option {
	return func(f *Fetcher) {
		if f.allowHosts == nil {
			f.allowHosts = make(map[string]struct{})
		}
		for _, h := range hosts {
			f.allowHosts[strings.ToLower(strings.TrimSpace(h))] = struct{}{}
		}
	}
}

// WithDialContext injects a custom dialer. IP-literal destinations are still
// validated before the dialer runs; hostname destinations are delegated to
// the dialer as-is (tests can map simulated public names to local servers).
func WithDialContext(d DialContextFunc) Option {
	return func(f *Fetcher) { f.dialContext = d }
}

// WithProxy routes fetches through an upstream proxy: "" or "NONE" means
// direct, "SYSTEM" uses the environment (http_proxy etc.), anything else is
// parsed as a proxy URL.
func WithProxy(proxy string) Option {
	return func(f *Fetcher) {
		switch {
		case proxy == "" || strings.EqualFold(proxy, "NONE"):
			// direct
		case strings.EqualFold(proxy, "SYSTEM"):
			f.proxyFunc = http.ProxyFromEnvironment
		default:
			if u, err := url.Parse(proxy); err == nil && u.Scheme != "" {
				f.proxyFunc = http.ProxyURL(u)
			}
		}
	}
}

// New builds a Fetcher with the given limits and options.
func New(l Limits, opts ...Option) *Fetcher {
	f := &Fetcher{
		limits: l.withDefaults(),
		dialer: &net.Dialer{Timeout: dialTimeout, KeepAlive: 30 * time.Second},
	}
	for _, opt := range opts {
		opt(f)
	}
	f.client = &http.Client{
		Transport: &http.Transport{
			Proxy:                 f.proxyFunc,
			DialContext:           f.safeDialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          32,
			TLSHandshakeTimeout:   tlsHandshakeTimeout,
			ResponseHeaderTimeout: responseHeaderTimeout,
			IdleConnTimeout:       idleConnTimeout,
			ExpectContinueTimeout: time.Second,
		},
		Timeout:       f.limits.Timeout,
		CheckRedirect: f.checkRedirect,
	}
	return f
}

// Get fetches rawURL and returns the response body, enforcing scheme,
// destination, redirect, timeout and size limits. Non-2xx responses are
// returned as *HTTPStatusError.
func (f *Fetcher) Get(rawURL string, headers map[string]string) ([]byte, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("fetcher: invalid URL %q: %w", rawURL, err)
	}
	if err := f.validateURL(u); err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("fetcher: invalid URL %q: %w", rawURL, err)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := f.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, &HTTPStatusError{URL: rawURL, StatusCode: resp.StatusCode, Status: resp.Status}
	}
	max := f.limits.MaxBodyBytes
	body, err := io.ReadAll(io.LimitReader(resp.Body, max+1))
	if err != nil {
		return nil, err
	}
	if int64(len(body)) > max {
		return nil, &ResponseTooLargeError{Limit: max}
	}
	return body, nil
}

// validateURL enforces the scheme allowlist and, for IP-literal hosts, the
// blocked-address rules. Hostname destinations are revalidated at dial time.
func (f *Fetcher) validateURL(u *url.URL) error {
	if u.Scheme != "http" && u.Scheme != "https" {
		return &SchemeError{Scheme: u.Scheme}
	}
	return f.checkHost(u.Hostname())
}

// checkRedirect revalidates every redirect hop and caps the hop count
// (MaxRedirects counts followed redirects, not the initial request).
func (f *Fetcher) checkRedirect(req *http.Request, via []*http.Request) error {
	if len(via) > f.limits.MaxRedirects {
		return ErrTooManyRedirects
	}
	return f.validateURL(req.URL)
}

func (f *Fetcher) checkHost(host string) error {
	if ip, err := netip.ParseAddr(host); err == nil {
		if !f.ipAllowed(host, ip) {
			return &BlockedAddressError{Host: host, IP: ip.Unmap()}
		}
	}
	return nil
}

// safeDialContext validates the dialed destination at dial time. Hostnames
// are resolved here and every resolved IP is revalidated; the connection is
// then made to a validated IP, closing the DNS-rebinding TOCTOU gap.
func (f *Fetcher) safeDialContext(ctx context.Context, network, address string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}

	if _, err := netip.ParseAddr(host); err == nil {
		// IP literal: validate, then dial it directly.
		if err := f.checkHost(host); err != nil {
			return nil, err
		}
		if f.dialContext != nil {
			return f.dialContext(ctx, network, address)
		}
		return f.dialer.DialContext(ctx, network, address)
	}

	if f.dialContext != nil {
		// Custom dialer: hostname destinations are delegated as-is.
		return f.dialContext(ctx, network, address)
	}

	addrs, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, err
	}
	if len(addrs) == 0 {
		return nil, fmt.Errorf("fetcher: no such host: %s", host)
	}
	for _, a := range addrs {
		ip, ok := netip.AddrFromSlice(a.IP)
		if !ok {
			continue
		}
		if !f.ipAllowed(host, ip) {
			return nil, &BlockedAddressError{Host: host, IP: ip.Unmap()}
		}
	}
	var lastErr error
	for _, a := range addrs {
		conn, err := f.dialer.DialContext(ctx, network, net.JoinHostPort(a.IP.String(), port))
		if err == nil {
			return conn, nil
		}
		lastErr = err
	}
	return nil, lastErr
}

func (f *Fetcher) ipAllowed(host string, ip netip.Addr) bool {
	if _, ok := f.allowHosts[strings.ToLower(host)]; ok {
		return true
	}
	return !isBlockedIP(ip)
}

var blockedPrefixes = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"),     // unspecified / "this host"
	netip.MustParsePrefix("100.64.0.0/10"), // RFC 6598 shared address space (CGNAT)
	netip.MustParsePrefix("192.0.0.0/24"),  // IETF protocol assignments
	netip.MustParsePrefix("198.18.0.0/15"), // benchmarking
	netip.MustParsePrefix("240.0.0.0/4"),   // reserved for future use
}

// isBlockedIP reports whether ip is loopback, private, link-local,
// multicast, unspecified, or otherwise reserved. 169.254.169.254 (cloud
// metadata) is covered by the link-local checks.
func isBlockedIP(ip netip.Addr) bool {
	ip = ip.Unmap()
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsUnspecified() ||
		ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsMulticast() {
		return true
	}
	for _, p := range blockedPrefixes {
		if p.Contains(ip) {
			return true
		}
	}
	return false
}

var (
	globalMu      sync.RWMutex
	globalFetcher *Fetcher
)

// SetGlobal installs a process-wide fetcher override (used by tests to
// simulate public vs. internal destinations without real egress). Passing
// nil clears the override.
func SetGlobal(f *Fetcher) {
	globalMu.Lock()
	defer globalMu.Unlock()
	globalFetcher = f
}

// Global returns the process-wide override, or nil when unset.
func Global() *Fetcher {
	globalMu.RLock()
	defer globalMu.RUnlock()
	return globalFetcher
}

// ForConfig returns the fetcher to use for one outbound fetch: the
// process-wide override when set, otherwise a fetcher whose limits derive
// from the advanced config values.
func ForConfig(maxAllowedDownloadSize int, proxy string) *Fetcher {
	if g := Global(); g != nil {
		return g
	}
	return New(LimitsFromConfig(maxAllowedDownloadSize), WithProxy(proxy))
}
