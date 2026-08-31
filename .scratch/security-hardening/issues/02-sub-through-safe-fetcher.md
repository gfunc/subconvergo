# 02: Route /sub url= and config= through the safe fetcher

**What to build:** subscription fetching and external-config fetching (`config=` query param) go through the ticket-01 safe fetcher — same IP/redirect/timeout/size protections. The number of `|`-separated source URLs per `/sub` request is capped, and oversized/invalid source URLs are rejected before any fetch.

**Blocked by:** 01 (Safe outbound fetcher)

**Status:** done

## TDD design

Seam under test: the `/sub` HTTP endpoint via gin test context + `httptest` origin servers (allowed via injected dialer/allowlist), plus a simulated internal target.

Red-first cases (each must FAIL against pre-ticket code — record as negative control):

- [ ] `TestSub_RefusesInternalSubscriptionURL`: `url=http://127.0.0.1:<port>/sub` → refused, upstream not hit; today it is fetched (negative control: current handler returns the subscription content).
- [ ] `TestSub_RefusesInternalConfigURL`: `config=http://169.254.169.254/...` or loopback → refused; today `loadExternalConfig` fetches it.
- [ ] `TestSub_CapsSourceCount`: request with more `|`-separated URLs than the cap → 4xx before any fetch; today all are fetched.
- [ ] `TestSub_RefusesRedirectToInternal`: allowed origin redirects to internal → refused at the hop.
- [ ] `TestSub_EnforcesDownloadSizeCap`: origin streams beyond cap → rejected; today unbounded `io.ReadAll`.
- [ ] Green-path: `TestSub_PublicSubscriptionUnchanged` — an existing happy-path subscription conversion test still passes through the new fetcher (use existing fixtures).

## Acceptance criteria

- [ ] All red-first cases pass and were observed failing pre-fix.
- [ ] No remaining direct `http.Get`/`http.NewRequest`+`client.Do` outside the shared fetcher in the subscription/config paths.
- [ ] Existing `/sub` tests pass unchanged apart from the new enforcement behavior.

## Comments

All four fetch paths now go through the ticket-01 `fetcher` package:

- `parser/parser.go` `fetchSubscription` (subscription fetch; proxy + User-Agent preserved;
  `SYSTEM` proxy now maps to `http.ProxyFromEnvironment` instead of a broken scheme-less URL);
- `handler/handler.go` `loadExternalConfig` (`config=` fetch; cache + stale-cache preserved);
- `handler/handler.go` `/getruleset` (ticket 01);
- `generator/utils/utils.go` `FetchRuleset` remote branch only (local-file branch untouched,
  owned by ticket 03). Behavior change there: non-2xx now errors instead of returning the body.

`/sub` request bounds (handler/handler.go, before any fetch): at most 32 `|`-separated
source URLs (`maxSubSourceURLs`), each at most 2048 bytes (`maxSubSourceURLLength`);
violations are 400.

### Negative controls (red, run against pre-fix code — fetcher package present but consumers not yet rewired)

```
$ go test ./handler/ -run 'TestGetRuleset_|TestSub_|TestSafeFetch_' -count=1 -v
--- FAIL: TestSub_RefusesInternalSubscriptionURL (0.00s)
    fetch_security_test.go:243: internal subscription URL must be refused, got 200
--- FAIL: TestSub_RefusesInternalConfigURL (0.00s)
    fetch_security_test.go:270: internal config URL was fetched 1 time(s)
--- FAIL: TestSub_CapsSourceCount (0.00s)
    fetch_security_test.go:290: expected 400 for 33 source URLs, got 200
--- FAIL: TestSub_SourceURLLengthBounded (0.00s)
    fetch_security_test.go:306: expected 400 for oversized source URL, got 200
--- FAIL: TestSub_RefusesRedirectToInternal (0.00s)
    fetch_security_test.go:330: subscription redirect to internal target must be refused, got 200
--- FAIL: TestSub_EnforcesDownloadSizeCap (0.00s)
    fetch_security_test.go:355: oversized subscription body must be rejected, got 200
--- PASS: TestSub_PublicSubscriptionUnchanged (0.00s)   # green-path guard, passes pre- and post-fix
```

### Green (post-fix)

```
--- PASS: TestSub_RefusesInternalSubscriptionURL (0.00s)
--- PASS: TestSub_RefusesInternalConfigURL (0.00s)
--- PASS: TestSub_CapsSourceCount (0.00s)
--- PASS: TestSub_SourceURLLengthBounded (0.00s)
--- PASS: TestSub_RefusesRedirectToInternal (0.00s)
--- PASS: TestSub_EnforcesDownloadSizeCap (0.00s)
--- PASS: TestSub_PublicSubscriptionUnchanged (0.00s)
ok  	github.com/gfunc/subconvergo/handler	1.218s
```

Existing httptest-based tests (`TestLoadExternalConfig_YAMLRemote`,
`TestHandleGetRuleset_RemoteFetch`, `TestHandleSub_SubURLKeepsPercentEncoding`,
`TestHandleGetProfile_QueryURLMergedWithProfileURL`, `parser.TestSubParser_UserAgent`)
got a one-line seam install (`allowLoopbackFetch(t)` / inline `fetcher.SetGlobal` with
loopback allowlisted) since the fetcher now refuses loopback by default; their
assertions are unchanged. A refused `config=` fetch keeps the existing log-and-continue
semantics in `processSubRequest` (fetch refused => external config not applied; the
request proceeds with the global config) — test asserts upstream never hit and the
attacker group absent from output.

`go build ./...`, `go vet ./...`, `go test ./...` all green (2026-09-01).
