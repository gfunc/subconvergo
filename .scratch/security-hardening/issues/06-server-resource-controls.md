# 06: Server-side resource controls

**What to build:** the HTTP server runs with explicit read-header/read/write/idle timeouts instead of bare `router.Run`; the existing advanced config fields (`MaxAllowedDownloadSize`, `MaxConcurrentThreads`, `MaxPendingConnections`) are wired into real enforcement (the size cap lands in the ticket-01 fetcher; concurrency/pending limits gate request handling); per-request source URL length is bounded.

**Blocked by:** 01 (Safe outbound fetcher — the size-cap plumbing and fetcher config land there)

**Status:** done

## TDD design

Seam under test: the constructed `http.Server` configuration (assert effective timeouts) plus endpoint behavior under slow/oversized conditions via `httptest`.

Red-first cases (each must FAIL pre-fix):

- [ ] `TestServer_HasFiniteTimeouts`: the server built by startup has nonzero `ReadHeaderTimeout`/read/write/idle timeouts; today none are set.
- [ ] `TestSub_SlowOriginIsCutOff`: an origin that drips bytes past the configured fetch timeout → request fails within a bounded time; today the `config=`/`/getruleset` paths can hang indefinitely (test via the external-config path which uses bare `http.Get`).
- [ ] `TestAdvancedLimits_AreEnforced`: set tiny `MaxAllowedDownloadSize`/concurrency in config → oversized or excess-concurrent requests are rejected/queued per the configured values (today these fields are dead — no consumer exists, which the negative control demonstrates).
- [ ] Green-path: normal `/sub` conversion completes under default limits (existing fixtures).

## Acceptance criteria

- [ ] All red-first cases pass and were observed failing pre-fix.
- [ ] Every advanced limit field either enforces something or is removed from config; no dead knobs remain.
- [ ] Defaults are finite and documented in the example configs.

## Comments

- `main.go`: bare `router.Run` replaced by `buildHTTPServer` — explicit `http.Server` with
  ReadHeaderTimeout=10s, ReadTimeout=30s, WriteTimeout=60s (exceeds the fetcher's 30s
  overall timeout), IdleTimeout=120s.
- `main.go`: `concurrencyLimitMiddleware(maxInFlight, maxPending)` wired via `router.Use`
  from `config.Global.Advanced.MaxConcurrentThreads` (in-flight semaphore; <=0 disables)
  and `MaxPendingConnections` (bounded wait queue; excess -> 503, and clients that
  disconnect while queued stop waiting). Both knobs now enforce something.
- `config/config.go`: `init()` defaults `MaxAllowedDownloadSize` to
  `fetcher.DefaultMaxBodyBytes` (32 MiB); `fetcher.LimitsFromConfig` also maps
  non-positive values to the finite default. Note: config files that explicitly set
  `max_allowed_download_size=0` (all current example configs do) now get 32 MiB, not
  "unlimited" — intentional per the ticket ("sane finite defaults when unset").
- Per-request source URL length bound landed with ticket 02 (`maxSubSourceURLLength=2048`
  in handler/handler.go).

### Negative controls (red, run against pre-fix code)

`TestSub_SlowOriginIsCutOff` fails behaviorally pre-fix (the `config=` path was a bare
`http.Get` with no timeout — the request hangs for the origin's full stall):

```
--- FAIL: TestSub_SlowOriginIsCutOff (1.20s)
    fetch_security_test.go:412: slow origin was not cut off: request took 1.201689577s
```

`TestServer_HasFiniteTimeouts` and `TestAdvancedLimits_AreEnforced` cannot even compile
pre-fix because the server had no explicit `http.Server` and the advanced knobs had no
consumer (dead knobs — exactly what the ticket describes):

```
$ go test . -run 'TestServer_HasFiniteTimeouts|TestAdvancedLimits_AreEnforced' -count=1 -v
# github.com/gfunc/subconvergo [github.com/gfunc/subconvergo.test]
./main_test.go:13:9: undefined: buildHTTPServer
./main_test.go:39:13: undefined: concurrencyLimitMiddleware
FAIL	github.com/gfunc/subconvergo [build failed]
```

(Pre-fix dead-knob demonstration for `MaxAllowedDownloadSize` is in ticket 01/02's red
runs: setting it to 64 changed nothing and oversized bodies were served with 200.)

### Green (post-fix)

```
$ go test . -count=1
ok  	github.com/gfunc/subconvergo	0.010s          # TestServer_HasFiniteTimeouts + TestAdvancedLimits_AreEnforced
$ go test ./handler/ -run 'TestSub_SlowOriginIsCutOff' -count=1 -v
--- PASS: TestSub_SlowOriginIsCutOff (1.20s)        # request cut at the 200ms fetch timeout; 1.2s is httptest Close waiting out the sleeping origin
```

`go build ./...`, `go vet ./...`, `go test ./...` all green (2026-09-01).

Follow-up for the next wave (out of my file scope): document the now-live semantics of
`max_concurrent_threads` (in-flight request cap, default 2), `max_pending_connections`
(queue cap, default 10240, excess -> 503) and `max_allowed_download_size` (0 -> 32 MiB
default, not unlimited) in `base/pref.example.{ini,toml,yml}` — I did not touch those
files to avoid colliding with ticket 05's config edits.
