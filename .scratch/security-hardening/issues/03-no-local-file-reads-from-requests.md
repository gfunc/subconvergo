# 03: Request-originated input can no longer read local files

**What to build:** with `api_mode=true` (the public-deployment mode), `file://` URLs and plain filesystem paths in `/sub` `url=` are rejected before reaching the parser. Remote external configs lose filesystem capabilities: a ruleset entry pointing at a local path is refused (local ruleset reads go through a canonical symlink-resolved resolver confined to the configured rules root), and `!!import:` of a local path from a remote config is refused. Trusted local sources remain possible for non-API deployments via explicit opt-in.

**Blocked by:** None (can start immediately)

**Status:** done

## TDD design

Seam under test: `/sub` HTTP endpoint with `api_mode=true`, using temp-dir marker files whose contents must never appear in responses.

Red-first cases (each must FAIL pre-fix — that is the negative control):

- [ ] `TestSub_APIModeRejectsFileURL`: temp file containing a valid subscription + unique marker; `url=file://<path>` → today the marker is parsed into output; after fix: 4xx, marker absent.
- [ ] `TestSub_APIModeRejectsPlainLocalPath`: same file passed as a bare absolute path → refused (today `os.Stat` success routes it to `os.ReadFile`).
- [ ] `TestSub_RemoteConfigCannotReadLocalRuleset`: attacker-hosted external config referencing `ruleset: <temp marker file>`; today the file's lines land in generated Clash `rules`; after fix: refused, marker absent.
- [ ] `TestSub_RemoteINIImportRejected`: remote INI config with `!!import:<temp marker file>` → no local read occurs (assert file-read via marker absence and an instrumented resolver in tests).
- [ ] Green-path: non-API mode with explicit opt-in can still load a local subscription (new test pins the trusted path so the feature isn't silently deleted).

## Acceptance criteria

- [ ] All red-first cases pass and were observed failing pre-fix.
- [ ] `api_mode` is actually consulted on the request path (closing the documented-vs-actual gap), or the docs are corrected if semantics intentionally change.
- [ ] Local ruleset resolution shares one canonical resolver (absolute/`..` rejection + `EvalSymlinks` confinement), not per-call-site checks.
- [ ] Existing traversal tests for `/getruleset` local paths still pass.

## Comments

### Negative control (red-first evidence, pre-fix code)

`go test ./handler/ -run 'TestSub_APIModeRejectsFileURL|TestSub_APIModeRejectsPlainLocalPath|TestSub_NonAPIModeLocalFileAllowed|TestSub_RemoteConfigCannotReadLocalRuleset|TestSub_RemoteINIImportRejected' -v`:

```
--- FAIL: TestSub_APIModeRejectsFileURL (0.00s)
    localfile_security_test.go:46: file:// subscription URL must be refused in api_mode, got 200
--- FAIL: TestSub_APIModeRejectsPlainLocalPath (0.00s)
    localfile_security_test.go:62: plain local path subscription must be refused in api_mode, got 200
--- FAIL: TestSub_RemoteConfigCannotReadLocalRuleset (0.00s)
    localfile_security_test.go:114: remote external config read a local ruleset file:
        rules:
            - DOMAIN-SUFFIX,REMOTECFG-RULESET-MARKER-7702.invalid,Auto
            - MATCH,DIRECT
--- FAIL: TestSub_RemoteINIImportRejected (0.00s)
    localfile_security_test.go:145: remote INI external config imported a local file:
        rules:
            - DOMAIN-SUFFIX,REMOTEINI-IMPORT-MARKER-7703.invalid,Auto
            - MATCH,DIRECT
--- PASS: TestSub_NonAPIModeLocalFileAllowed (0.00s)   # green-path pin, passes pre-fix by design
```

`go test ./parser/ -run 'TestParse_APIMode' -v` (parser-level pins):

```
--- FAIL: TestParse_APIModeRejectsFileURL (0.00s)
    apimode_file_test.go:41: file:// URL must be rejected in api_mode, parsed 1 proxies
--- FAIL: TestParse_APIModeRejectsPlainLocalPath (0.00s)
    apimode_file_test.go:53: plain local path must be rejected in api_mode, parsed 1 proxies
```

`go test ./generator/utils/ -run 'TestFetchRuleset_' -v` (ruleset local-branch confinement):

```
--- FAIL: TestFetchRuleset_RejectsAbsoluteLocalPath (0.00s)
    fetchruleset_security_test.go:31: absolute local ruleset path must be refused, got content: "DOMAIN-SUFFIX,RULESET-ESCAPE-MARKER-8801.invalid\n"
--- FAIL: TestFetchRuleset_RejectsTraversal (0.00s)
    fetchruleset_security_test.go:52: traversal ruleset path must be refused, got content: "DOMAIN-SUFFIX,RULESET-ESCAPE-MARKER-8801.invalid\n"
--- FAIL: TestFetchRuleset_SymlinkEscapeRefused (0.00s)
    fetchruleset_security_test.go:79: symlink escaping the rules root must be refused, got content: "DOMAIN-SUFFIX,RULESET-ESCAPE-MARKER-8801.invalid\n"
--- PASS: TestFetchRuleset_InRootSymlinkAllowed (0.00s)      # green pins, pass pre-fix by design
--- PASS: TestFetchRuleset_LocalFileUnderBaseStillWorks (0.00s)
```

Note: `TestSub_RemoteINIImportRejected` asserts marker absence in the `/sub` response; the "instrumented resolver" half of that assertion is covered indirectly — post-fix the only code path that could read the file goes through the canonical resolver, which refuses absolute/escaping import targets (unit-pinned via `TestFetchRuleset_*` and the ticket-09 resolver tests).

### What changed (post-fix)

- `parser/parser.go` (`SubParser.Parse`): after computing `isFile` (`file://` prefix or `os.Stat` hit), the parse is refused with an error when `config.Global.Common.APIMode` is true — before any read. Non-API mode keeps local files (explicit trusted-deployment opt-in, pinned by `TestSub_NonAPIModeLocalFileAllowed` / `TestParse_NonAPIModeLocalFileAllowed`). `api_mode` is now actually consulted on the request path.
- `generator/utils/utils.go` (`FetchRuleset` LOCAL branch): absolute paths rejected outright; relative paths resolve only under `config.GetBasePath()` and `<base>/rules` via the new shared canonical resolver `utils.ResolveUnderRoot` (EvalSymlinks on root and target, target must stay under root). No CWD fallback. Remote (http/https) branch untouched.
- `handler/handler.go` (`loadExternalConfig` + INI parse): explicit `isRemote` trust decision at the top of `loadExternalConfig`. For remote configs: YAML/TOML skip `ProcessImports` (new `mergeExternalConfigCollections` does only the in-memory collection merges — remote configs get zero filesystem capabilities, including `import:` on groups/rulesets/renames/emojis) and `parseINICustomGroups`/`parseINIRulesets` refuse every `!!import:` (new `allowImport=false` parameter). For local configs: `ProcessImports` unchanged; INI `!!import:` resolves through `utils.ResolveUnderRoot` confined to `config.GetBasePath()` — no CWD fallback, absolute imports refused.
- New shared resolver: `utils/paths.go` → `utils.ResolveUnderRoot(target, root)`.
- `generator/impl/clash_test.go` (`TestClashGenerator_Generate_WithRulesets`): updated to the new contract — the old version relied on the removed CWD-relative fallback (`os.Stat("test_rules.list")` in the package dir); it now writes the ruleset under a temp `base/rules/` and points `BasePath` at it. This is the only pre-existing test that needed changes.

### Test results (post-fix)

```
go build ./... && go vet ./...                      # OK
go test ./handler/ -run 'TestSub_APIMode|TestSub_NonAPI|TestSub_Remote|TestLoadExternalConfig_LocalINIImportAllowed' -v
--- PASS: TestSub_APIModeRejectsFileURL (0.00s)
--- PASS: TestSub_APIModeRejectsPlainLocalPath (0.00s)
--- PASS: TestSub_NonAPIModeLocalFileAllowed (0.00s)
--- PASS: TestSub_RemoteConfigCannotReadLocalRuleset (0.00s)
--- PASS: TestSub_RemoteINIImportRejected (0.00s)
--- PASS: TestLoadExternalConfig_LocalINIImportAllowed (0.00s)
go test ./parser/ -run 'TestParse_'                  # 3/3 PASS (incl. green pin)
go test ./generator/utils/ -run 'TestFetchRuleset'   # 6/6 PASS (incl. pre-existing TestFetchRuleset)
go test ./...                                        # all packages ok
```

Known remaining gap (out of scope, `config` package): `Settings.ProcessImports`/`resolveImportPath` in `config/config.go` still has a CWD fallback and no confinement. Remote YAML/TOML configs no longer reach it (gated in `loadExternalConfig`), but the main pref loading path still uses it for trusted configs — fine today, worth hardening separately if pref loading ever takes request-influenced paths.
