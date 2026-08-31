# 07: Dependency upgrades + vulnerability gate

**What to build:** the dependency graph moves past the govulncheck-confirmed reachable advisories: `golang.org/x/text` ≥ v0.39.0, `golang.org/x/net` ≥ v0.55.0 (or current compatible), `golang.org/x/crypto` ≥ v0.52.0, `quic-go` ≥ v0.59.1, and a mihomo version that permits them. The Dockerfile builder pins a patched Go toolchain instead of `golang:1.25-alpine` floating. A `govulncheck ./...` gate exists as a Makefile target so regressions are caught.

**Blocked by:** None (can start immediately)

**Status:** done

## TDD design

This ticket's "tests" are gates rather than red-green unit tests; the negative control is the gate failing on the current graph.

Red-first checks (must FAIL pre-fix — govulncheck currently reports 15 reachable advisories):

- [ ] `govulncheck ./...` reports zero reachable vulnerabilities after the upgrade (capture the current failing report as the negative control in the ticket comments).
- [ ] `go mod verify` passes; no `replace`/`exclude` hacks introduced to silence the scanner.
- [ ] Dockerfile builder image pinned to a patched Go version by digest or explicit patch version.
- [ ] `make vulncheck` (or equivalent) target exists and fails the build on reachable findings.

## Acceptance criteria

- [ ] All gate checks pass; the pre-fix govulncheck output is recorded in the ticket comments as evidence.
- [ ] Full `go test ./...` green after the upgrade (mihomo major bumps may require adapter adjustments — keep them minimal).
- [ ] Python smoke tests in `tests/` still pass if they are runnable in the dev environment.

## Comments

### Negative control (pre-fix govulncheck, 2026-08-31 15:50 UTC, local toolchain go1.26.4)

Command: `go run golang.org/x/vuln/cmd/govulncheck@latest ./...` → exit status 3, 15 reachable vulnerabilities.

```
=== Symbol Results ===

Vulnerability #1: GO-2026-6218
    Avoid quadratic complexity in resolvePath in net/url
  More info: https://pkg.go.dev/vuln/GO-2026-6218
  Standard library
    Found in: net/url@go1.26.4
    Fixed in: net/url@go1.26.6
    Example traces found:
      #1: parser/parser.go:247:24: parser.SubParser.fetchSubscription calls http.Client.Do, which eventually calls url.URL.Parse

Vulnerability #2: GO-2026-6090
    Limit handshake messages we are willing to accept post-handshake in
    crypto/tls
  More info: https://pkg.go.dev/vuln/GO-2026-6090
  Standard library
    Found in: crypto/tls@go1.26.4
    Fixed in: crypto/tls@go1.26.6
    Example traces found:
      #1: handler/handler.go:769:5: handler.SubHandler.HandleGetRuleset calls v2ray.Mux.Close, which eventually calls tls.Conn.Handshake
      #2: handler/handler.go:769:5: handler.SubHandler.HandleGetRuleset calls v2ray.Mux.Close, which eventually calls tls.Conn.HandshakeContext
      #3: handler/handler.go:785:29: handler.SubHandler.HandleGetRuleset calls io.ReadAll, which calls tls.Conn.Read
      #4: handler/handler.go:769:5: handler.SubHandler.HandleGetRuleset calls v2ray.Mux.Close, which calls tls.Conn.Write
      #5: parser/parser.go:247:24: parser.SubParser.fetchSubscription calls http.Client.Do, which eventually calls tls.Dialer.DialContext

Vulnerability #3: GO-2026-6089
    Apply ReadHeaderTimeout when doing unencrypted HTTP/2 check in net/http
  More info: https://pkg.go.dev/vuln/GO-2026-6089
  Standard library
    Found in: net/http@go1.26.4
    Fixed in: net/http@go1.26.6
    Example traces found:
      #1: main.go:125:22: subconvergo.startServer calls gin.Engine.Run, which calls http.ListenAndServe

Vulnerability #4: GO-2026-5972
    Enforce maximum recursion depth in encoding/asn1
  More info: https://pkg.go.dev/vuln/GO-2026-5972
  Standard library
    Found in: encoding/asn1@go1.26.4
    Fixed in: encoding/asn1@go1.26.6
    Example traces found:
      #1: parser/utils/utils.go:55:41: utils.ToMihomoProxyWithSetting calls adapter.ParseProxy, which eventually calls asn1.Unmarshal

Vulnerability #5: GO-2026-5970
    Infinite loop on invalid input in golang.org/x/text
  More info: https://pkg.go.dev/vuln/GO-2026-5970
  Module: golang.org/x/text
    Found in: golang.org/x/text@v0.30.0
    Fixed in: golang.org/x/text@v0.39.0
    Example traces found:
      #1: generator/impl/clash.go:425:24: impl.validateClashRule calls rules.ParseRule, which eventually calls norm.Form.Bytes
      #2: generator/impl/clash.go:425:24: impl.validateClashRule calls rules.ParseRule, which eventually calls norm.Form.IsNormalString
      #3: generator/impl/clash.go:425:24: impl.validateClashRule calls rules.ParseRule, which eventually calls norm.Form.QuickSpan
      #4: generator/impl/clash.go:425:24: impl.validateClashRule calls rules.ParseRule, which eventually calls norm.Form.String

Vulnerability #6: GO-2026-5856
    Invoking Encrypted Client Hello privacy leak in crypto/tls
  More info: https://pkg.go.dev/vuln/GO-2026-5856
  Standard library
    Found in: crypto/tls@go1.26.4
    Fixed in: crypto/tls@go1.26.5
    Example traces found:
      #1: handler/handler.go:769:5: handler.SubHandler.HandleGetRuleset calls v2ray.Mux.Close, which eventually calls tls.Conn.Handshake
      #2: handler/handler.go:769:5: handler.SubHandler.HandleGetRuleset calls v2ray.Mux.Close, which eventually calls tls.Conn.HandshakeContext
      #3: handler/handler.go:785:29: handler.SubHandler.HandleGetRuleset calls io.ReadAll, which calls tls.Conn.Read
      #4: handler/handler.go:769:5: handler.SubHandler.HandleGetRuleset calls v2ray.Mux.Close, which calls tls.Conn.Write
      #5: parser/parser.go:247:24: parser.SubParser.fetchSubscription calls http.Client.Do, which eventually calls tls.Dialer.DialContext

Vulnerability #7: GO-2026-5676
    HTTP/3 QPACK Trailer Expansion Memory Exhaustion in
    github.com/quic-go/quic-go
  More info: https://pkg.go.dev/vuln/GO-2026-5676
  Module: github.com/quic-go/quic-go
    Found in: github.com/quic-go/quic-go@v0.56.0
    Fixed in: github.com/quic-go/quic-go@v0.59.1
    Example traces found:
      #1: handler/handler.go:785:29: handler.SubHandler.HandleGetRuleset calls io.ReadAll, which eventually calls http3.ConfigureTLSConfig
      #2: handler/handler.go:1137:67: handler.SubHandler.HandleSurge2Clash calls http3.Error.Error
      #3: handler/handler.go:785:29: handler.SubHandler.HandleGetRuleset calls io.ReadAll, which eventually calls http3.countingByteReader.Read

Vulnerability #8: GO-2026-5026
    Invoking failure to reject ASCII-only Punycode-encoded labels in
    golang.org/x/net/idna
  More info: https://pkg.go.dev/vuln/GO-2026-5026
  Module: golang.org/x/net
    Found in: golang.org/x/net@v0.46.0
    Fixed in: golang.org/x/net@v0.55.0
    Example traces found:
      #1: generator/impl/clash.go:425:24: impl.validateClashRule calls rules.ParseRule, which eventually calls idna.ToASCII

  Standard library
    Found in: net/http@go1.26.4
    Fixed in: net/http@go1.26.6
    Example traces found:
      #1: handler/handler.go:769:5: handler.SubHandler.HandleGetRuleset calls smux.stream.Close, which eventually calls http.Client.CloseIdleConnections
      #2: parser/parser.go:247:24: parser.SubParser.fetchSubscription calls http.Client.Do
      #3: handler/handler.go:753:25: handler.SubHandler.HandleGetRuleset calls http.Get

Vulnerability #9: GO-2026-5020
    Invoking infinite loop on large channel writes in golang.org/x/crypto/ssh
  More info: https://pkg.go.dev/vuln/GO-2026-5020
  Module: golang.org/x/crypto
    Found in: golang.org/x/crypto@v0.43.0
    Fixed in: golang.org/x/crypto@v0.52.0
    Example traces found:
      #1: parser/utils/utils.go:55:41: utils.ToMihomoProxyWithSetting calls adapter.ParseProxy, which eventually calls ssh.NewClientConn
      #2: handler/handler.go:769:5: handler.SubHandler.HandleGetRuleset calls v2ray.Mux.Close, which calls ssh.channel.Write

Vulnerability #10: GO-2026-5019
    Invoking bypass of FIDO/U2F security keys physical interaction in
    golang.org/x/crypto/ssh
  More info: https://pkg.go.dev/vuln/GO-2026-5019
  Module: golang.org/x/crypto
    Found in: golang.org/x/crypto@v0.43.0
    Fixed in: golang.org/x/crypto@v0.52.0
    Example traces found:
      #1: parser/utils/utils.go:55:41: utils.ToMihomoProxyWithSetting calls adapter.ParseProxy, which eventually calls ssh.NewClientConn

Vulnerability #11: GO-2026-5018
    Invoking pathological RSA/DSA parameters may cause DoS in
    golang.org/x/crypto/ssh
  More info: https://pkg.go.dev/vuln/GO-2026-5018
  Module: golang.org/x/crypto
    Found in: golang.org/x/crypto@v0.43.0
    Fixed in: golang.org/x/crypto@v0.52.0
    Example traces found:
      #1: parser/utils/utils.go:55:41: utils.ToMihomoProxyWithSetting calls adapter.ParseProxy, which eventually calls ssh.NewClientConn
      #2: parser/utils/utils.go:55:41: utils.ToMihomoProxyWithSetting calls adapter.ParseProxy, which eventually calls ssh.ParseAuthorizedKey
      #3: parser/utils/utils.go:55:41: utils.ToMihomoProxyWithSetting calls adapter.ParseProxy, which eventually calls ssh.ParsePrivateKey
      #4: parser/utils/utils.go:55:41: utils.ToMihomoProxyWithSetting calls adapter.ParseProxy, which eventually calls ssh.ParsePrivateKeyWithPassphrase

Vulnerability #12: GO-2026-5017
    Invoking client can cause server deadlock on unexpected responses in
    golang.org/x/crypto/ssh
  More info: https://pkg.go.dev/vuln/GO-2026-5017
  Module: golang.org/x/crypto
    Found in: golang.org/x/crypto@v0.43.0
    Fixed in: golang.org/x/crypto@v0.52.0
    Example traces found:
      #1: parser/utils/utils.go:55:41: utils.ToMihomoProxyWithSetting calls adapter.ParseProxy, which eventually calls ssh.NewClientConn

Vulnerability #13: GO-2026-5013
    Invoking byte arithmetic causes underflow and panic in
    golang.org/x/crypto/ssh
  More info: https://pkg.go.dev/vuln/GO-2026-5013
  Module: golang.org/x/crypto
    Found in: golang.org/x/crypto@v0.43.0
    Fixed in: golang.org/x/crypto@v0.52.0
    Example traces found:
      #1: parser/utils/utils.go:55:41: utils.ToMihomoProxyWithSetting calls adapter.ParseProxy, which eventually calls ssh.NewClientConn

Vulnerability #14: GO-2026-4918
    Infinite loop in HTTP/2 transport when given bad SETTINGS_MAX_FRAME_SIZE in
    net/http/internal/http2 in golang.org/x/net
  More info: https://pkg.go.dev/vuln/GO-2026-4918
  Module: golang.org/x/net
    Found in: golang.org/x/net@v0.46.0
    Fixed in: golang.org/x/net@v0.53.0
    Example traces found:
      #1: parser/utils/utils.go:55:41: utils.ToMihomoProxyWithSetting calls adapter.ParseProxy, which eventually calls http2.Transport.NewClientConn
      #2: parser/parser.go:247:24: parser.SubParser.fetchSubscription calls http.Client.Do, which eventually calls http2.Transport.RoundTrip
      #3: parser/parser.go:247:24: parser.SubParser.fetchSubscription calls http.Client.Do, which eventually calls http2.noDialH2RoundTripper.RoundTrip
      #4: parser/parser.go:247:24: parser.SubParser.fetchSubscription calls http.Client.Do, which eventually calls http2.unencryptedTransport.RoundTrip

Vulnerability #15: GO-2025-4233
    HTTP/3 QPACK Header Expansion DoS in github.com/quic-go/quic-go
  More info: https://pkg.go.dev/vuln/GO-2025-4233
  Module: github.com/quic-go/quic-go
    Found in: github.com/quic-go/quic-go@v0.56.0
    Fixed in: github.com/quic-go/quic-go@v0.57.0
    Example traces found:
      #1: handler/handler.go:785:29: handler.SubHandler.HandleGetRuleset calls io.ReadAll, which eventually calls http3.ConfigureTLSConfig
      #2: handler/handler.go:1137:67: handler.SubHandler.HandleSurge2Clash calls http3.Error.Error
      #3: handler/handler.go:785:29: handler.SubHandler.HandleGetRuleset calls io.ReadAll, which eventually calls http3.countingByteReader.Read

Your code is affected by 15 vulnerabilities from 4 modules and the Go standard library.
This scan also found 11 vulnerabilities in packages you import and 14
vulnerabilities in modules you require, but your code doesn't appear to call
these vulnerabilities.
Use '-show verbose' for more details.
exit status 3
```

### Resolution (2026-08-31 15:59 UTC)

**Dependency upgrades** (via `go get` + `go mod tidy`, no replace/exclude directives):

- `github.com/metacubex/mihomo` v1.19.16 → **v1.19.30** (latest v1.19.x; no major bump needed, no adapter code changes — all four use sites compile unchanged)
- `github.com/quic-go/quic-go` v0.56.0 → **v0.62.0** (pulled in by gin's http3, NOT by mihomo — mihomo v1.19.30 moved to its own fork `metacubex/quic-go` v0.61.1-pre, itself past v0.59.1)
- `golang.org/x/text` v0.30.0 → **v0.41.0**, `golang.org/x/net` v0.46.0 → **v0.58.0**, `golang.org/x/crypto` v0.43.0 → **v0.55.0**
- go directive: 1.25.3 → **1.26.7** (quic-go v0.62.0 needs ≥1.26.0; pinned to the patched patch-release so any GOTOOLCHAIN=auto build/scan uses a fixed stdlib)
- transitive bumps of note: gin stays v1.11.0 (compiles fine against quic-go v0.62.0), testify v1.11.1 → v1.12.1

**Dockerfile**: builder pinned `golang:1.25-alpine` → `golang:1.26.7-alpine` (explicit patch version; tag verified to exist on Docker Hub) with a keep-in-sync comment. `GOPROXY=https://goproxy.cn` kept, with a comment documenting why (proxy.golang.org unreachable from CN build networks; goproxy.cn still proxies the sum.golang.org checksum DB so go.sum verification applies). Final stage untouched.

**Makefile**: added `vulncheck` target (`go run golang.org/x/vuln/cmd/govulncheck@latest ./...`) + .PHONY entry. govulncheck exits 3 on reachable findings, so the target fails the build. `docker-run` untouched.

**Green evidence:**

Post-fix `go run golang.org/x/vuln/cmd/govulncheck@latest ./...` → exit 0:

```
=== Symbol Results ===

No vulnerabilities found.

Your code is affected by 0 vulnerabilities.
This scan also found 0 vulnerabilities in packages you import and 3
vulnerabilities in modules you require, but your code doesn't appear to call
these vulnerabilities.
Use '-show verbose' for more details.
```

`make vulncheck` → MAKE_EXIT=0, same "No vulnerabilities found" report.

```
$ go mod verify
all modules verified

$ go build ./... && go vet ./...
(both exit 0, no output)

$ go test ./...
ok  	github.com/gfunc/subconvergo/config
ok  	github.com/gfunc/subconvergo/generator
ok  	github.com/gfunc/subconvergo/generator/core
ok  	github.com/gfunc/subconvergo/generator/impl
ok  	github.com/gfunc/subconvergo/generator/transformers
ok  	github.com/gfunc/subconvergo/generator/utils
ok  	github.com/gfunc/subconvergo/handler
ok  	github.com/gfunc/subconvergo/parser
ok  	github.com/gfunc/subconvergo/parser/proxy
ok  	github.com/gfunc/subconvergo/parser/sub
ok  	github.com/gfunc/subconvergo/parser/utils
ok  	github.com/gfunc/subconvergo/proxy/core
ok  	github.com/gfunc/subconvergo/proxy/impl
ok  	github.com/gfunc/subconvergo/proxy/utils
(remaining packages: no test files)
```

Notes for later waves:

- mihomo v1.19.30's own go.mod carries `replace google.golang.org/protobuf => github.com/metacubex/protobuf-go`; dependency replaces do not propagate, so this module still builds against stock protobuf.
- 3 vulns remain in *required but not reachable* modules (informational; govulncheck exits 0). The gate only fails on reachable findings.
- No Go source files were modified by this ticket; handler/generator/parser logic untouched.
