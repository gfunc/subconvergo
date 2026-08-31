# 09: Symlink-safe path resolution everywhere

**What to build:** template/base-config/profile path resolution gains real filesystem confinement: the configured root and the requested file are both resolved with `EvalSymlinks`, and the resolved target must remain beneath the resolved root — matching what `/getruleset` local resolution already does. A symlink inside the templates/base directory can no longer escape the root.

**Blocked by:** None (can start immediately)

**Status:** done

## TDD design

Seam under test: `/render` (and the base-config/profile load paths) via gin test context with a temp-dir root containing a symlink pointing outside it.

Red-first cases (each must FAIL pre-fix):

- [ ] `TestRender_SymlinkEscapeRefused`: temp root `templates/` containing `logs -> /tmp/outside-marker-dir`; request `path=logs/marker.txt` → today the lexical check passes and the outside file's marker content is returned; after fix: 4xx, marker absent.
- [ ] `TestBaseConfig_SymlinkEscapeRefused`: same shape through the request-scoped `*_rule_base` path → outside file never read.
- [ ] `TestProfile_SymlinkEscapeRefused`: same shape through profile resolution.
- [ ] Green-path: legitimate in-root symlinks (resolving to a file still under the root) continue to work, matching `/getruleset`'s current behavior.

## Acceptance criteria

- [ ] All red-first cases pass and were observed failing pre-fix.
- [ ] One shared resolver used by render, base-config, profile, and ruleset paths (converge with ticket 03's resolver if both land — whoever lands second refactors to the shared one).
- [ ] Existing traversal/absolute-path tests still pass.

## Comments

### Negative control (red-first evidence, pre-fix code)

`go test ./handler/ -run 'TestRender_SymlinkEscapeRefused|TestRender_InRootSymlinkAllowed|TestBaseConfig_SymlinkEscapeRefused|TestProfile_SymlinkEscapeRefused|TestGetRuleset_InRootSymlinkAllowed|TestGetRuleset_SymlinkEscapeRefused' -v` against the lexical-only `resolvePathUnderRoot`:

```
--- FAIL: TestRender_SymlinkEscapeRefused (0.00s)
    symlink_security_test.go:53: template symlink escape must be refused, got 200
        (200 body is the outside marker file's content: "RENDER-SYMLINK-MARKER-9901")
--- FAIL: TestBaseConfig_SymlinkEscapeRefused (0.00s)
    symlink_security_test.go:123: symlink escape via rule base must be refused, got content: "BASECFG-SYMLINK-MARKER-9902"
--- FAIL: TestProfile_SymlinkEscapeRefused (0.00s)
    symlink_security_test.go:157: profile symlink escape must be refused, got 200
--- PASS: TestRender_InRootSymlinkAllowed (0.00s)        # green-path pins, pass pre-fix by design
--- PASS: TestGetRuleset_InRootSymlinkAllowed (0.00s)
--- PASS: TestGetRuleset_SymlinkEscapeRefused (0.00s)    # /getruleset already confined; pin against regression
```

Implementation note: ticket 03 landed first with the shared resolver at `utils.ResolveUnderRoot` (utils/paths.go); this ticket converges `resolvePathUnderRoot` (render / base-config / profile) and `resolveRulesetLocalPath` (/getruleset) onto it.

### What changed (post-fix)

- `handler/handler.go`:
  - `resolvePathUnderRoot` is now a thin delegate to `utils.ResolveUnderRoot` — used by `/render` (template root + user path), `loadBaseConfig` (`*_rule_base` under config dir), and `resolveProfilePath` (bare names and relative profile paths). Root and target are both EvalSymlinks-resolved; the resolved target must stay under the resolved root.
  - `resolveRulesetLocalPath` converged onto the same shared resolver over candidate roots `<base>` and `<base>/rules`, returning the canonical resolved path (replacing its bespoke stat+EvalSymlinks check, which compared against the unresolved root).
  - `generator/utils.FetchRuleset` (ticket 03) and trusted INI `!!import:` (ticket 03) use the same resolver, so all five path domains share one implementation.
- Resolver semantics note: the target must exist (`EvalSymlinks` fails otherwise). All call sites only resolve paths they immediately read or probe, so this is behavior-preserving; `/render` now answers 400 (instead of 404) for non-existent template paths — still 4xx, no content disclosure.

### Test results (post-fix)

```
go build ./... && go vet ./...                      # OK
go test ./handler/ -run 'Symlink' -v
--- PASS: TestRender_SymlinkEscapeRefused (0.00s)
--- PASS: TestRender_InRootSymlinkAllowed (0.00s)
--- PASS: TestBaseConfig_SymlinkEscapeRefused (0.00s)
--- PASS: TestProfile_SymlinkEscapeRefused (0.00s)
--- PASS: TestGetRuleset_InRootSymlinkAllowed (0.00s)
--- PASS: TestGetRuleset_SymlinkEscapeRefused (0.00s)
go test ./...                                       # all packages ok, incl. pre-existing
                                                    # TestHandleGetRuleset_{LocalPath,AbsolutePathBlocked,TraversalBlocked},
                                                    # TestHandleRender_TraversalBlocked, TestHandleGetProfile_AbsolutePathBlocked,
                                                    # TestResolveProfilePath_* (3), TestProcessSubRequest_ExternalConfigRuleBaseLeakBlocked
```
