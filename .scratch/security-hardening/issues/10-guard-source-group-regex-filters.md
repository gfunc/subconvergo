# 10: Guard source-provided group regex filters

**What to build:** when `ignore_source=false`, Clash `proxy-groups` `filter` values arriving from subscription content are validated before reaching generated output: only a safe regex subset (RE2-compatible, no nested/repeated quantifier constructs that cause catastrophic backtracking in downstream engines) survives; anything else is rejected or stripped with the group preserved in a safe form. A malicious subscription can no longer ship a filter that makes the *client's* mihomo hang on config load (`regexp2` has no default timeout downstream).

**Blocked by:** 04 (Sanitize subscription-controlled fields — the field-sanitization boundary this plugs into is built there)

**Status:** done

## TDD design

Seam under test: `/sub?target=clash&ignore_source=false` with an attacker-controlled subscription containing pathological `proxy-groups` filters; assertions on the generated YAML. A compile-level property check (candidate filters must compile under Go's RE2 `regexp`) is acceptable as a second seam if the validation helper is exported.

Red-first cases (each must FAIL pre-fix):

- [ ] `TestSub_Clash_StripsPathologicalGroupFilter`: source config with `filter: "^(a+)+$"` (and variants like `(a|a)*`, `(a*)+`) → generated output contains no such filter; today it is passed through verbatim.
- [ ] `TestSub_Clash_FilterLengthAndCountCapped`: over-long filter / excessive group count → rejected or truncated per policy.
- [ ] Green-path: ordinary RE2-safe filters (`^(HK|SG)`, keyword matches) survive into output unchanged — pinned by test so the guard doesn't break legitimate source groups.

## Acceptance criteria

- [ ] All red-first cases pass and were observed failing pre-fix.
- [ ] Policy decision (reject vs. strip vs. rewrite) is documented in the ticket comments and applied consistently to all source-derived group fields.
- [ ] Full test suite green.

## Comments

### Empirical correction to the ticket premise (verified @ a45b859, 2026-09-01)

The ticket assumes source `filter` values are "passed through verbatim" today. Verified empirically (scratch test parsing a Clash source with `filter: "^(a+)+$"` through `sub.ParseMihomoConfig`): **the Go port silently DROPS `filter` at parse time** — `parser/sub/clash.go` JSON-roundtrips each group map into `config.ProxyGroupConfig`, which has no `Filter` field, so no filter text ever reaches generated output:

```
GROUP: name="evil-group" type="select" rule=[]string{"[]node-a"} url=""   # filter gone
```

Consequence: `TestSub_Clash_StripsPathologicalGroupFilter`'s literal assertion ("output contains no such filter") cannot fail pre-fix — it passes vacuously. The red-first controls for this ticket are therefore:

1. The exported validation helper seam (explicitly allowed by the ticket): `IsSafeGroupFilter` / `MaxGroupFilterLength` / `MaxSourceProxyGroups` in `parser/utils` — pre-fix run fails to compile:
   ```
   parser/utils/sanitize_test.go:145:18: undefined: IsSafeGroupFilter
   parser/utils/sanitize_test.go:166:17: undefined: MaxGroupFilterLength
   FAIL	github.com/gfunc/subconvergo/parser/utils [build failed]
   ```
2. Behavioral red on the green-path case (`TestSub_Clash_SafeGroupFilterSurvives`): a valid filter `^(HK|SG)` has NO filtering effect today (group falls back to all nodes):
   ```
   --- FAIL: TestSub_Clash_SafeGroupFilterSurvives (0.00s)
       []interface {}{"HK-01", "SG-02", "US-03"} should not contain "US-03"
       safe RE2 filter must still filter group members
   ```
3. Behavioral red on caps (`TestSub_Clash_FilterLengthAndCountCapped`): 151 source groups rendered pre-fix:
   ```
   "151" is not less than or equal to "100"
   source-provided group count must be capped, got 151
   ```

### Policy (reject vs. strip vs. rewrite)

**Strip-as-text, preserve-as-behavior.** Source-provided filters are *never emitted as text* into generated output (subconvergo expands group membership server-side, so the client's mihomo never receives any source regex — the `regexp2`-no-timeout threat is eliminated structurally). A **validated** filter is preserved as behavior by appending it to the group's `Rule` (subconvergo's native remark-regex group filtering, `generator/utils.FilterProxiesByRules`, Go RE2 = linear time). An invalid/pathological/over-long filter is **stripped** (logged); the group is preserved and falls back to its explicit proxies or, if none, the unfiltered node list — exactly today's de-facto behavior for all filters, so pathological inputs cause no behavioral regression.

Validation (`parser/utils.IsSafeGroupFilter`): non-empty, length ≤ `MaxGroupFilterLength` (256), compiles under Go's RE2 `regexp`, and the `regexp/syntax` AST contains (a) no nested quantifiers — a `* + ? {m,n}` node with another quantifier in its subtree (`^(a+)+$`, `(a*)+`, `(a?){2}`) and (b) no quantifier directly over an alternation with duplicate or empty-matchable branches (`(a|a)*`, `(a|)*`). Disjoint-branch quantified alternations like `^HK(a|b){2}$` are allowed.

Caps: at most `MaxSourceProxyGroups` (100) proxy-groups per subscription; excess are dropped at parse (logged). Over-long filters stripped (not truncated — a truncated regex is a different regex).

Applied to all source-derived group fields in `parser/sub/clash.go`: `filter` (validated→Rule or stripped), group `name`/`url` sanitized via ticket 04's `SanitizeScalarField`, and non-`[]` entries in source-provided `rule` lists are dropped unless they pass `IsSafeGroupFilter` (a source group must not inject subconvergo-internal rule syntax such as `!!GROUP=` matchers either).

### Implementation (2026-09-01)

- `parser/utils/utils.go`: exported guard `IsSafeGroupFilter` (empty/length cap `MaxGroupFilterLength=256` → RE2 `regexp.Compile` → two-layer backtracking-shape check) and `MaxSourceProxyGroups=100`.
  - Layer 1 (AST): reject a quantifier whose subtree contains another quantifier (`^(a+)+$`, `(a*)+`, `(a?){2}`) or an alternation with an empty-matchable branch (catches `(a|ab)*` post-factoring).
  - Layer 2 (raw text): reject a quantified group whose top-level alternation has duplicate or empty branches (`(a|a)*`, `(a|)*`). Needed because `regexp/syntax`'s parser factors `a|a` → `a`, erasing the duplicate signal from the AST.
  - Green-path shapes verified safe: `^(HK|SG)`, `HK`, `HK|香港`, `(?i)hk|sg`, `^.* Premium .*$`, `(ab)+`, `x{0,1}y{2}`, `^HK(a|b){2}$`.
- `parser/sub/clash.go` (group loop): per-subscription cap of 100 source groups; `filter` validated via `IsSafeGroupFilter` → appended to `group.Rule` (safe) or stripped with a log (unsafe/over-long); source-provided `rule` entries (never legitimate in clash sources — that is pref syntax) are kept only if `[]name` or filter-safe, dropping `!!GROUP=`-style internal matchers and pathological regexes; names/URLs/member names sanitized per ticket 04.
- No generator changes: validated filters execute server-side under Go RE2 (`generator/utils.FilterProxiesByRules`), so no source regex is ever emitted into client-facing output.

### Green evidence (post-fix)

```
$ go test ./generator/impl/ -run 'TestSub_Clash' -v
--- PASS: TestSub_Clash_StripsPathologicalGroupFilter   (group preserved, filter effect removed, no filter text in output)
--- PASS: TestSub_Clash_SafeGroupFilterSurvives         (asia=[HK-01 SG-02], keyword=[HK-01])
--- PASS: TestSub_Clash_FilterLengthAndCountCapped      (151 source groups -> 100 rendered; 704-char filter stripped)
$ go test ./parser/utils/
ok  (IsSafeGroupFilter table + length boundary)
```

`go build ./...`, `go vet ./...`, `go test ./...` all green.
