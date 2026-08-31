# Security Hardening Spec

Source: full security review of subconvergo (2026-08-31), six-area parallel audit
(handlers, parsers, generator/templates, outbound fetch/TLS, config/deployment,
dependencies). Findings verified against HEAD `a45b859`.

Threat model: unauthenticated remote users hitting the HTTP API; malicious
subscription-server content; multi-tenant public deployments.

Headline findings driving these tickets:

1. Unauthenticated SSRF with response exfiltration via `/getruleset` (Critical)
2. SSRF via `/sub` `url=` and `config=`; unlimited sources per request (High)
3. Arbitrary local file reads: `file://`/`os.Stat` path in the subscription parser,
   local ruleset paths from remote external configs, `!!import:` in remote INI
   configs; `api_mode` documented but never enforced (High)
4. Config directive injection: `%0A` in proxy remarks reaches line-oriented
   Surge/Loon/Quantumult X output unescaped (High, verified end-to-end)
5. No resource controls: unbounded `io.ReadAll`, `http.Get` without timeouts,
   unused `MaxAllowedDownloadSize`/concurrency config fields (High)
6. Reachable known-vulnerable deps (govulncheck): x/text, x/net, x/crypto,
   quic-go, plus stdlib vulns in the toolchain; Dockerfile pins old Go (High)
7. Shipped default token `api_access_token=password`, fails open on empty token;
   `make docker-run` mounts config where the image never reads it (High)
8. Credential-bearing URLs and full request headers logged (Medium)
9. Lexical-only path confinement bypassable via symlinks in render/base-config (Medium)
10. Source-provided proxy-group regex filters reach downstream mihomo `regexp2`
    with no timeout (Low/Medium)

Every ticket is TDD: failing test at a public seam first, negative control
required (prove the test fails against pre-fix code), then minimal implementation.
