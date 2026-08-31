# 01: Safe outbound fetcher, wired into /getruleset

**What to build:** one shared outbound-fetch helper used as the single way the service fetches remote content: http/https only; loopback/private/link-local/reserved IPs (v4+v6, incl. cloud metadata 169.254.169.254) refused at dial time (custom `DialContext`, not a preflight DNS lookup); every redirect hop revalidated and redirects capped; overall timeout plus transport-level timeouts; response body capped via a limited reader with over-limit rejection. `/getruleset` fetches through it and requires the access token. This is the tracer bullet: the fetcher is the prefactor all later tickets reuse.

**Blocked by:** None (can start immediately)

**Status:** done

## TDD design

Seam under test: the `/getruleset` HTTP endpoint via gin test context + `httptest`, plus the new fetcher behind an injectable dialer/transport so tests can simulate public vs. internal destinations without real egress.

Red-first cases (each must FAIL against current code — that failure is the negative control):

- [ ] `TestGetRuleset_RefusesLoopbackURL`: target base64 of `http://127.0.0.1:<httptest port>/secret`; current code fetches and returns the body (negative control: marker string appears in response today); after fix: 4xx, marker absent, upstream never hit (hit counter on the test server).
- [ ] `TestGetRuleset_RefusesRedirectToInternal`: an "allowed" origin (mocked transport) 302s to a loopback/metadata address; fetch must be refused at the redirect hop.
- [ ] `TestGetRuleset_RejectsOversizedBody`: upstream streams more than the configured cap; response rejected with an over-limit error, body not returned, connection closed early.
- [ ] `TestGetRuleset_RequiresToken`: with `api_access_token` configured, no/wrong token → 401/403 and no outbound fetch (hit counter stays zero).
- [ ] `TestSafeFetch_AllowedPublicURLStillWorks`: green-path guard so the fix doesn't kill legitimate use — an allowlisted/injected-transport destination returns content as before.

## Acceptance criteria

- [ ] All red-first cases pass; each was observed failing on pre-fix code (record the failing output in the ticket comments).
- [ ] `/getruleset` no longer calls `http.Get`/`http.DefaultClient` directly.
- [ ] Existing `/getruleset` tests (local-path traversal suite) still pass.
- [ ] Fetcher config (timeouts, size cap, redirect limit) derives from config fields, wiring `MaxAllowedDownloadSize` rather than adding a parallel knob.

## Comments

Implemented together with tickets 02 and 06 (shared fetcher). New package `fetcher/`
(repo root): http/https only; loopback/private/link-local/reserved IPv4+IPv6 (incl.
169.254.169.254, CGNAT 100.64/10, 240/4, v4-mapped v6) refused at dial time via a
custom `DialContext` that resolves and revalidates every resolved IP and then dials a
validated IP (no preflight-DNS TOCTOU); `CheckRedirect` revalidates every hop and caps
redirects (default 5); overall client timeout (default 30s) + transport timeouts
(Dial 10s, TLSHandshake 10s, ResponseHeader 20s, IdleConn 90s); body capped via
`io.LimitReader(max+1)` with `*ResponseTooLargeError`. Limits derive from
`config.Global.Advanced.MaxAllowedDownloadSize` via `fetcher.LimitsFromConfig`
(<=0 -> finite default 32 MiB; `config.init()` also defaults the field to 32 MiB).
Test seams: `WithAllowedHosts`, `WithDialContext`, `SetGlobal`/`ForConfig`.

`/getruleset` now requires `token` when `api_access_token` is set (same inline pattern
as /readconf, /flushcache, /render) and fetches remote rulesets through
`fetcher.ForConfig(...)`; cache + stale-cache semantics preserved.

### Negative controls (red, run against pre-fix code — fetcher package present but consumers not yet rewired)

```
$ go test ./handler/ -run 'TestGetRuleset_|TestSub_|TestSafeFetch_' -count=1 -v
--- FAIL: TestGetRuleset_RefusesLoopbackURL (0.00s)
    fetch_security_test.go:90: loopback URL must be refused, got 200 with body "DOMAIN-SUFFIX,internal.example,SECRET-RULESET-MARKER\n"
--- FAIL: TestGetRuleset_RefusesRedirectToInternal (0.00s)
    fetch_security_test.go:124: redirect to internal target must be refused, got 200 with body "DOMAIN-SUFFIX,internal.example,SECRET-RULESET-MARKER\n"
--- FAIL: TestGetRuleset_RejectsOversizedBody (0.00s)
    fetch_security_test.go:153: oversized ruleset body must be rejected, got 200
--- FAIL: TestGetRuleset_RequiresToken (0.00s)
    fetch_security_test.go:182: /getruleset?url=aHR0cDovLzEyNy4wLjAuMTozOTIwNQ==&type=clash: expected 401/403, got 200
--- PASS: TestSafeFetch_AllowedPublicURLStillWorks (0.00s)   # green-path guard, passes pre- and post-fix
```

### Green (post-fix)

```
$ go test ./handler/ -run 'TestGetRuleset_|TestSub_|TestSafeFetch_' -count=1 -v
--- PASS: TestGetRuleset_RefusesLoopbackURL (0.00s)
--- PASS: TestGetRuleset_RefusesRedirectToInternal (0.00s)
--- PASS: TestGetRuleset_RejectsOversizedBody (0.00s)
--- PASS: TestGetRuleset_RequiresToken (0.00s)
--- PASS: TestSafeFetch_AllowedPublicURLStillWorks (0.00s)
ok  	github.com/gfunc/subconvergo/handler	1.218s
```

Note: the redirect-to-internal handler test's first hop is itself loopback, so post-fix
the refusal fires at the initial request; hop-level revalidation is pinned by
`fetcher.TestFetcher_RedirectHopToBlockedAddressRefused` (allowlisted origin 302s to
169.254.169.254 literal -> `*BlockedAddressError` before any dial).

`go build ./...`, `go vet ./...`, `go test ./...` all green (2026-09-01).
