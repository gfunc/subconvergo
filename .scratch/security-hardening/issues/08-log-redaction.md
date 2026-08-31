# 08: Log redaction

**What to build:** logs no longer contain credential-bearing material: raw subscription/config URLs (which embed tokens), URL userinfo, sensitive query parameters (`token`, `key`, `auth`, `password`, `uuid`, …), or full request-header maps. What remains: scheme/host origin, sizes, counts, and a stable hash for correlation. Explicit log files are created with `0600`.

**Blocked by:** None (can start immediately)

**Status:** done

## TDD design

Seam under test: captured log output. Redirect the standard logger to a buffer, drive requests through gin test contexts, assert on what the buffer does and doesn't contain. Assertions target observable log content, not internal redaction helpers.

Red-first cases (each must FAIL pre-fix — secrets currently land in logs):

- [ ] `TestSub_DoesNotLogSubscriptionSecrets`: request with `url=https://provider.example/sub?token=SECRET-MARKER` → buffer contains no `SECRET-MARKER`; today it appears in the request-received log line.
- [ ] `TestSub_DoesNotLogAuthorizationHeader`: request carrying `Authorization: Bearer SECRET-MARKER` → marker absent from logs; today the full header map is printed.
- [ ] `TestParser_DoesNotLogURLUserinfo`: subscription URL with `user:SECRET-MARKER@host` → marker absent from parser fetch logs.
- [ ] `TestLogs_StillCorrelatable`: the redacted log line still carries origin host and a stable hash/count so debugging remains possible (guards against over-redaction).
- [ ] Log file mode: when `-l` is used, the created file is `0600`.

## Acceptance criteria

- [ ] All red-first cases pass and were observed failing pre-fix.
- [ ] Redaction is centralized (one URL/header redaction path used by handler and parser logging), not per-call-site string surgery.
- [ ] Full test suite green.

## Comments

### Negative control (red-first), 2026-09-01, pre-fix code

`go test ./handler/ -run 'TestSub_DoesNotLog|TestLogs_StillCorrelatable'`:

```
--- FAIL: TestSub_DoesNotLogSubscriptionSecrets (0.00s)
    log_redaction_test.go:56: subscription credential leaked into logs:
--- FAIL: TestSub_DoesNotLogAuthorizationHeader (0.00s)
    log_redaction_test.go:79: Authorization header leaked into logs:
--- FAIL: TestLogs_StillCorrelatable (0.00s)
    log_redaction_test.go:105: credential leaked into logs:
```

(pre-fix the /sub request-received line prints `url=<raw>&config=<raw> ...
headers=<full header map>` — the `token=SECRET-MARKER` query and the
`Authorization: Bearer SECRET-MARKER` header both land in the buffer.)

`go test ./parser/ -run TestParser_DoesNotLogURLUserinfo -v`:

```
--- FAIL: TestParser_DoesNotLogURLUserinfo (0.00s)
    log_redaction_test.go:48: URL userinfo leaked into logs:
        [parser.fetchSubscription] index=0 url=http://user:SECRET-MARKER@127.0.0.1:46745/sub size=68
        [parser.ParseSubscription] index=0 url=http://user:SECRET-MARKER@127.0.0.1:46745/sub parsed=1 proxies
        [parser.SubParser.Parse] index=0 url=http://user:SECRET-MARKER@127.0.0.1:46745/sub proxies=1
```

Log-file mode: `go test . -run TestLogFile_CreatedWith0600` — red via build
failure (the testable seam did not exist pre-fix; the `0644` open was inline
in `main()`):

```
./main_logfile_test.go:13:12: undefined: openLogFile
FAIL github.com/gfunc/subconvergo [build failed]
```

Pre-fix the open call in main.go used mode `0644`.

### What changed (2026-09-01)

- `utils/redact.go` (new): the single redaction path —
  `RedactURL` (scheme://host + `/<redacted>` + non-sensitive query keys kept,
  keys matching `token|key|auth|password|uuid|secret` case-insensitively
  dropped, userinfo never emitted, non-http(s) schemes reduced to
  `scheme:<redacted>`, stable 8-hex SHA-256 `sig=` correlation hash),
  `RedactURLOrPath` (URLs redacted, local paths pass through),
  `RedactHeaders` (allowlist: User-Agent only).
- `handler/handler.go`: every request-received/fetch/cache log line now goes
  through those helpers (HandleSub entry/summary/parse lines,
  loadExternalConfig, HandleVersion, HandleReadConf, HandleGetRuleset
  including fetch lines, HandleRender, HandleGetProfile, HandleSurge2Clash,
  HandleFlushCache). Full-header `%v` dumps are gone everywhere.
- `parser/parser.go`: Parse/ParseSubscription/fetchSubscription log
  `RedactURL(sp.URL)`; the ParseProxy fallback logs
  `protocol=<scheme> line=<scheme>:<redacted> sig=...` instead of the raw
  credential-bearing line. `ParseSubscriptionFile` still logs the local file
  path (filesystem paths carry no credentials).
- `main.go`: `openLogFile` creates the `-l` log file with mode 0600.
- New tests: `handler/log_redaction_test.go`
  (TestSub_DoesNotLogSubscriptionSecrets, TestSub_DoesNotLogAuthorizationHeader,
  TestLogs_StillCorrelatable), `parser/log_redaction_test.go`
  (TestParser_DoesNotLogURLUserinfo), `main_logfile_test.go`
  (TestLogFile_CreatedWith0600).

### Test results (post-fix)

```
--- PASS: TestSub_DoesNotLogSubscriptionSecrets (0.00s)
--- PASS: TestSub_DoesNotLogAuthorizationHeader (0.00s)
--- PASS: TestLogs_StillCorrelatable (0.00s)
--- PASS: TestParser_DoesNotLogURLUserinfo (0.00s)
--- PASS: TestLogFile_CreatedWith0600 (0.00s)
```

`go build ./...`, `go vet ./...`, and full `go test ./...` all green.

Known leftovers (outside ticket scope): `/sub` error responses echo the raw
subscription URL back to the requesting client (not a log line); generator
package logs config-provided local ruleset paths on fetch failure.
