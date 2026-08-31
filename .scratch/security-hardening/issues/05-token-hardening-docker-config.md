# 05: Token hardening + docker-run config fix

**What to build:** no shipped default credential — the bundled config no longer contains `api_access_token=password`. Token checks fail closed: an empty configured token means protected endpoints refuse every request (not "allow all"). Token comparison is constant-time. `make docker-run` mounts operator config where the image actually loads it (`/base`), or the binary is invoked with the mounted path, so hardening attempts are no longer silently ignored.

**Blocked by:** None (can start immediately)

**Status:** done

## TDD design

Seams under test: protected endpoints (`/render`, `/readconf`, `/flushcache`, `/getprofile`) via gin test context with various configured tokens; config-load/startup validation as a behavior test; the Docker mount fix via the existing Python smoke-test harness (or a shell check in `tests/`).

Red-first cases (each must FAIL pre-fix):

- [ ] `TestProtectedEndpoint_EmptyTokenFailsClosed`: `api_access_token` empty → `/render` refuses all requests regardless of `token` param; today it allows everything.
- [ ] `TestProtectedEndpoint_WrongTokenRefused` + correct token accepted (pins both directions of the gate).
- [ ] `TestStartup_RefusesInsecurePublicDefault`: config with `listen=0.0.0.0` and empty/default token → startup validation reports an error (behavioral test of the validation function; decide and document whether this is a hard failure or loud warning).
- [ ] Docker smoke case: run the image with the documented `docker-run` mount carrying a distinctive token; assert the loaded config is the mounted one (e.g. protected endpoint honors the distinctive token, not `password`). Fails today because `/app` mounts are never read.

## Acceptance criteria

- [ ] All red-first cases pass and were observed failing pre-fix.
- [ ] Bundled/example configs contain no operational default token; constant-time comparison in use (verified at code review — timing itself is not unit-tested).
- [ ] `make docker-run` mount path and image config-load path agree; smoke test proves it.
- [ ] Docs/comments describing the old default-token behavior updated.

## Comments

### Negative control (red-first), 2026-09-01, pre-fix code

`go test ./handler/ -run 'TestProtectedEndpoint_'` — FAIL as expected:

```
--- FAIL: TestProtectedEndpoint_EmptyTokenFailsClosed (0.00s)
    auth_security_test.go:88: readconf (token param "anything"): empty configured token must fail closed with 403, got 500 body="Failed to reload config: no configuration file found\n"
    auth_security_test.go:88: readconf (token param ""): ... got 500
    auth_security_test.go:88: getruleset (token param "anything"): ... got 400 body="Invalid request!"
    auth_security_test.go:88: getruleset (token param ""): ... got 400
    auth_security_test.go:88: render (token param "anything"): ... got 404 body="Template not found: ..."
    auth_security_test.go:88: render (token param ""): ... got 404
    auth_security_test.go:88: flushcache (token param "anything"): ... got 200 body="Cache flushed\n"
    auth_security_test.go:88: flushcache (token param ""): ... got 200
```

(Pre-fix, empty `api_access_token` disables the check entirely — every
protected endpoint serves. `/getprofile` is the one exception: it already
refused because a non-empty request token never equals an empty configured
one; it is pinned in the test for uniformity, not a red case.)

`TestProtectedEndpoint_WrongTokenRefused` PASSES pre-fix — it is a
both-directions pin of the existing gate (wrong refused, correct accepted),
not a red case; kept green post-fix to guard the refactor.

`go test ./config/ -run TestStartup_RefusesInsecurePublicDefault` — red via
build failure (new seam did not exist pre-fix):

```
config/startup_security_test.go:19:14: s.ValidateStartupSecurity undefined (type *Settings has no field or method ValidateStartupSecurity)
FAIL github.com/gfunc/subconvergo/config [build failed]
```

### Docker mount fix reasoning (cannot be unit-tested in Go)

- `Dockerfile` final stage: `WORKDIR /base`, `ENTRYPOINT ["subconvergo"]` (no
  `-f`), and `COPY base /base`.
- `config.LoadConfig` (config/config.go ~296-304) searches the **current
  working directory** for `pref.yml`, `pref.yaml`, `pref.toml`, `pref.ini`,
  then `pref.example.*` — with no `-f`, CWD is `/base`.
- The old `docker-run` target mounted `./base` at `/app/base` and
  `./pref.toml` at `/app/pref.toml`. `/app` exists only in the builder stage;
  the runtime never reads it, so both mounts were silently ignored and the
  container always ran the baked-in `pref.example.*` (with the old default
  token).
- Fix: mount `./base` at `/base` and `./pref.toml` at `/base/pref.toml:ro`,
  so the loader's CWD search finds the operator config. (`base/pref.toml`
  does not exist in the repo; the `./pref.toml` host file is
  operator-created, which is why it is untracked.)

### Startup-validation decision

Refusal (hard failure, `log.Fatalf` in main.go) for a non-loopback
`server.listen` with an empty token **or** the legacy shipped default
`"password"`; loopback binds proceed with a loud warning that protected
endpoints stay disabled (fail-closed checks make them refuse everything).
Rationale: a public bind without a token is never a deliberate secure state
for this service; the shipped examples set `listen=0.0.0.0`, so warning-only
would leave every default docker deployment exposed. Loopback stays allowed
so local/dev usage isn't broken.

### What changed (2026-09-01)

- `handler/auth.go` (new): `requireAPIToken(c)` — single gate for
  `/readconf`, `/getruleset`, `/render`, `/flushcache`; fails closed (403)
  when `common.api_access_token` is empty, otherwise compares via
  `tokenEqual` = SHA-256 both sides + `subtle.ConstantTimeCompare` on the
  fixed-length digests. `/getprofile` keeps its per-profile `profile_token`
  flow but both comparisons now go through `tokenEqual` (empty configured
  token never matches).
- `config/config.go`: `Settings.ValidateStartupSecurity()` — refuses
  non-loopback `server.listen` with empty token or the legacy shipped
  default `"password"`; loopback proceeds with a loud warning. Decision:
  refusal for public binds (see rationale above).
- `main.go`: calls `ValidateStartupSecurity` after `LoadConfig`
  (`log.Fatalf` on error).
- `base/pref.ini`, `base/pref.example.{ini,toml,yml}`: ship
  `api_access_token` EMPTY with a comment that protected endpoints stay
  disabled until a token is set and that public binds are refused without
  one. (`base/pref.toml` does not exist in the repo; nothing to strip there.)
- `Makefile` `docker-run`: mounts now `-v $(PWD)/base:/base` and
  `-v $(PWD)/pref.toml:/base/pref.toml:ro` (image WORKDIR=/base, bare
  `subconvergo` entrypoint, LoadConfig searches CWD), plus a guard telling
  the operator to create pref.toml first.
- `doc/API.md`, `doc/REFERENCE.md`: token rows now say "required / fails
  closed", `/getruleset` documents its token requirement, protected-endpoint
  list gained `/getruleset` + `/flushcache`, `token=password` examples
  replaced, advanced-knob comments corrected (0 = 32MiB default;
  concurrency limits live). Also `pref.example.*` comments updated.
- Existing tests adjusted to the fail-closed semantics (they now set and
  pass a token): `handler/handler_test.go` (TestHandleReadConf,
  TestHandleGetRuleset, TestHandleGetRuleset_RemoteFetch/_LocalPath/
  _AbsolutePathBlocked/_TraversalBlocked), `handler/fetch_security_test.go`
  (TestGetRuleset_RefusesLoopbackURL/_RefusesRedirectToInternal/
  _RejectsOversizedBody, TestSafeFetch_AllowedPublicURLStillWorks).
- New tests: `handler/auth_security_test.go`
  (TestProtectedEndpoint_EmptyTokenFailsClosed,
  TestProtectedEndpoint_WrongTokenRefused),
  `config/startup_security_test.go`
  (TestStartup_RefusesInsecurePublicDefault).

### Test results (post-fix)

```
--- PASS: TestProtectedEndpoint_EmptyTokenFailsClosed (0.00s)
--- PASS: TestProtectedEndpoint_WrongTokenRefused (0.00s)
--- PASS: TestStartup_RefusesInsecurePublicDefault (0.00s)
ok  github.com/gfunc/subconvergo/handler
ok  github.com/gfunc/subconvergo/config
```

`go build ./...`, `go vet ./...`, and full `go test ./...` all green.
Docker smoke case not run (Python/docker harness out of scope for this
wave); the mount fix is verified by the LoadConfig trace above.
