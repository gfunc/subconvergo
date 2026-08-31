# 04: Sanitize subscription-controlled fields in text output

**What to build:** subscription-controlled scalar fields (remarks, hosts, credentials, paths, group names/URLs) can no longer inject configuration directives into line-oriented outputs (Surge, Loon, Quantumult X, single-link formats). CR/LF/NUL are rejected or stripped at the parse/normalization boundary; each text target uses syntax-aware encoding instead of raw string interpolation. Verified end-to-end exploit today: a Shadowsocks fragment `#legit%0A[rewrite_local]%0A...` produces an injected INI section in Quantumult X output.

**Blocked by:** None (can start immediately)

**Status:** done

## TDD design

Seam under test: `/sub?target=quanx|surge|loon` via gin test context, subscription body served by an `httptest` origin. Assertions on the rendered output text (line structure), not on internal sanitizer functions.

Red-first cases (each must FAIL pre-fix — negative control already demonstrated once for quanx; reproduce per target):

- [ ] `TestSub_QuanX_RejectsNewlineInRemark`: node with `%0A`-bearing fragment → output line count unchanged, no injected `[rewrite_local]` section; today the section appears.
- [ ] `TestSub_Surge_RejectsNewlineInRemark` and `TestSub_Loon_RejectsNewlineInRemark`: same payload through `%0A`/`%0D` variants → no extra config lines.
- [ ] `TestSub_TextTargets_SanitizeCredentialAndGroupFields`: CR/LF payload placed in password/host/group-name fields → no injected lines in any text target.
- [ ] Green-path: legitimate remarks containing commas, `=`, quotes, and unicode still render (quoted/escaped per format) — pinned by tests so sanitization doesn't mangle normal names.
- [ ] Clash/sing-box regression: structured targets unchanged (already serialized safely).

## Acceptance criteria

- [ ] All red-first cases pass and were observed failing pre-fix (attach one captured exploit output to the ticket comments as evidence).
- [ ] Sanitization lives at the parse/normalization boundary or a shared serializer — not patched per-generator.
- [ ] Full test suite green.

## Comments

### Negative control (red-first evidence, pre-fix @ a45b859, 2026-09-01)

Tests live in `generator/impl/security_output_test.go` (parser+generator seam: crafted subscription content through `parser/sub` → public `Generator` interface) and `parser/utils/sanitize_test.go` (boundary-helper seam). Run pre-fix:

```
$ go test ./generator/impl/ -run 'TestSub_' -v
--- FAIL: TestSub_QuanX_RejectsNewlineInRemark (0.00s)
--- FAIL: TestSub_Surge_RejectsNewlineInRemark (0.00s)
--- FAIL: TestSub_Loon_RejectsNewlineInRemark (0.00s)
--- FAIL: TestSub_TextTargets_SanitizeCredentialAndGroupFields (0.00s)  [quanx, surge, loon subtests all FAIL]
--- PASS: TestSub_TextTargets_LegitNamesRenderCorrectly (0.00s)         [pinning test; must STAY green]
--- FAIL: TestSub_StructuredTargets_SanitizeWithoutCorruption (0.00s)
FAIL	github.com/gfunc/subconvergo/generator/impl	0.019s
```

Captured exploit output (QuanX, pre-fix) — the injected INI section appears:

```
[server_local]
shadowsocks=1.2.3.4:8388, method=aes-128-gcm, password=pass, tag=legit
[rewrite_local]
^https://attacker url reject
```

Surge pre-fix (remark splits into 3 output lines):

```
[Proxy]
DIRECT = direct
legit
[rewrite_local]
^https://attacker url reject = ss, 1.2.3.4, 8388, encrypt-method=aes-128-gcm, password=pass
```

Credential/group injection via a Clash-format source (password, sni, group name carrying `\n[rewrite_local]\n^https://evil url reject`) produced equivalent injected lines in all three text targets pre-fix (full log in the test run).

Helper seam pre-fix (expected compile failure):

```
$ go test ./parser/utils/ -run 'TestSanitize|TestToMihomo|TestIsSafeGroupFilter' -v
parser/utils/sanitize_test.go:34:28: undefined: SanitizeScalarField
parser/utils/sanitize_test.go:145:18: undefined: IsSafeGroupFilter
parser/utils/sanitize_test.go:166:17: undefined: MaxGroupFilterLength
FAIL	github.com/gfunc/subconvergo/parser/utils [build failed]
```

### Design decision (where the sanitization lives and why)

Central point: **`parser/utils` (parse/normalization boundary)**, not per-generator and not `proxy/core`:

- Every `parser/proxy/*.go` Parse* method (13 protocols × single-link/Clash/Surge/SSD/SSTap/Netch/SSAndroid/v2ray sources) funnels through `utils.ToMihomoProxy` / `ToMihomoProxyWithSetting` / `ToMihomoProxyFromClash`. One hook there covers all of them; no generator changes needed.
- `proxy/core` was the ticket's alternative but is outside this ticket's file scope; `parser/utils` is in scope and is the actual normalization funnel.
- Two construction sites bypass the funnel and call the same exported `SanitizeProxy` directly: `parser/sub/ssr.go` (`parseSSRNode`) and `parser/sub/clash.go` (`parseMihomoProxy`).

Policy: **strip, don't reject** (legitimate nodes keep working). `SanitizeScalarField` removes ASCII control chars (0x00–0x1F incl. CR/LF/NUL, plus DEL) from every subscription-controlled string field via a reflection walk (`SanitizeProxy`) covering remarks, servers, credentials, groups, underlying-proxy, plugin names/options, SNI/host/path, UUIDs, ALPN/DNS slices and `url.Values` params. With no CR/LF able to exist in any field, no field can ever add a new output LINE in any text target.

Syntax-aware note (commas/quotes/`=`/unicode): these chars are **not** stripped — they cannot create new lines. Surge/Loon place the remark before `" = "` and QuanX appends `tag=` last, so such names render intact (pinned by `TestSub_TextTargets_LegitNamesRenderCorrectly`). Pre-existing upstream quirk (unchanged, not a regression): a QuanX tag containing `,` is truncated at the comma by the *client* parser; Loon quotes passwords as `"..."` without escaping embedded quotes. No escaping is added because none of these can inject lines, and inventing per-format escaping would diverge from C++ subconverter output.

### Implementation (2026-09-01)

- `parser/utils/utils.go`: added `SanitizeScalarField` (strips 0x00–0x1F/0x7F), `SanitizeProxy` (reflection walk over all string fields incl. embedded BaseProxy, `map[string]interface{}` plugin opts, `map[string]string`, `url.Values`-style `map[string][]string`, string slices; handles `*impl.MihomoProxy` by sanitizing the wrapped proxy and its `Options` map). Hooked into all three funnel functions (`ToMihomoProxy`, `ToMihomoProxyWithSetting`, `ToMihomoProxyFromClash`), which every `parser/proxy/*.go` Parse* method returns through.
- `parser/sub/clash.go`: `parseMihomoProxy` (raw-construction bypass) now calls `utils.SanitizeProxy`; source proxy-group `Name`/`URL` and group member names are sanitized.
- `parser/sub/ssr.go`: `parseSSRNode` constructs proxies without the funnel; both construction sites now call `utils.SanitizeProxy`.
- No generator or `proxy/impl` sink changes: inputs are clean at the boundary, so raw interpolation can no longer produce extra lines. `proxy/impl` files were inspected (all 13 types × Surge/Loon/QuanX) but deliberately left unmodified.

### Green evidence (post-fix)

```
$ go test ./generator/impl/ -run 'TestSub_' -v
--- PASS: TestSub_QuanX_RejectsNewlineInRemark
--- PASS: TestSub_Surge_RejectsNewlineInRemark
--- PASS: TestSub_Loon_RejectsNewlineInRemark
--- PASS: TestSub_TextTargets_SanitizeCredentialAndGroupFields [quanx/surge/loon]
--- PASS: TestSub_TextTargets_LegitNamesRenderCorrectly [quanx/surge/loon]
--- PASS: TestSub_StructuredTargets_SanitizeWithoutCorruption
ok  	github.com/gfunc/subconvergo/generator/impl
```

Post-fix exploit output (single line, inert text, node preserved):

```
[server_local]
shadowsocks=1.2.3.4:8388, method=aes-128-gcm, password=pass, tag=legit[rewrite_local]^https://attacker url reject
```

`go build ./...`, `go vet ./...`, `go test ./...` all green (full suite, including pre-existing tests).

## Follow-up (main agent, post-smoke)

The smoke-test wave found a residual bypass: the top-level `url=` query param's
`#fragment` becomes `SubParser.Tag` and is applied via `SetRemark` AFTER the
sanitized parse funnel (`parser/parser.go` parseURL). Live-verified injection via
`/sub?target=quanx&url=ss://...%23%250A%5Brewrite_local%5D...`.

Fixed: `sp.Tag = parserutils.SanitizeScalarField(u.Fragment)` in `parseURL`.
Regression test `TestSubParser_URLFragmentTagIsSanitized`
(parser/parser_security_test.go) — red pre-fix (remark contained
`"\n[rewrite_local]\n^https://evil"`), green post-fix.
