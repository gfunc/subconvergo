# Subconvergo Development Guide

> Complete guide for developing, building, testing, and deploying Subconvergo

## Table of Contents

- [Quick Start](#quick-start)
- [Architecture](#architecture)
- [Building](#building)
- [Testing](#testing)
- [Development Workflow](#development-workflow)
- [Configuration](#configuration)
- [Performance](#performance)

---

## Quick Start

### Prerequisites

- **Go 1.21+** - [Install Go](https://go.dev/doc/install)
- **Docker** - For containerized builds and testing (optional)
- **Python 3.8+** - For smoke tests

### Build and Run

```bash
# Clone and build
git clone https://github.com/gfunc/subconvergo.git
cd subconvergo
make build

# Run server
./subconvergo
# Server starts on http://localhost:25500
```

### Test Your Setup

```bash
# Check version
curl http://localhost:25500/version

# Convert a subscription
curl "http://localhost:25500/sub?target=clash&url=YOUR_SUB_URL"
```

---

## Architecture

### Project Structure

```
subconvergo/
├── main.go            # Entry point, HTTP server setup
├── config/            # Configuration management
│   └── config.go      # Load pref.yml/toml/ini
├── parser/            # Subscription & proxy parsing
│   └── parser.go      # Protocol parsers via mihomo
├── generator/         # Format conversion
│   └── generator.go   # Generate Clash/Surge/etc configs
├── handler/           # HTTP request handlers
│   └── handler.go     # API endpoints
├── base/              # Configuration & templates
│   ├── pref.toml      # Server configuration
│   ├── base/          # Client templates
│   ├── rules/         # Rule sets
│   └── config/        # Preset configs
└── tests/             # Testing infrastructure
    ├── smoke.py       # Integration/API tests
    ├── run-tests.sh   # Unit test runner
    └── mock-data/     # Test fixtures
```

### Key Components

#### Parser Layer (`parser/parser.go`)
- Parses subscription URLs and single proxy links
- Supports: SS, SSR, VMess, Trojan, VLESS, Hysteria, Hysteria2, TUIC, Clash YAML
- Uses [mihomo](https://github.com/metacubex/mihomo) for protocol validation
- Auto-upgrades protocols via mihomo updates

#### Generator Layer (`generator/generator.go`)
- Converts parsed proxies to target formats
- Supports: Clash, Surge, Quantumult X, Loon, sing-box, single links
- Template rendering with Go `text/template`
- Customizable via base templates in `base/base/`

#### Handler Layer (`handler/handler.go`)
- HTTP endpoints: `/sub`, `/version`, `/readconf`, `/render`, `/getprofile`, `/getruleset`
- Configuration loading and merging
- Proxy filtering (include/exclude, regex support)
- Remark processing (renames, emojis, sorting)

#### Config Layer (`config/config.go`)
- Loads `pref.yml` → `pref.toml` → `pref.ini` (priority order)
- Supports imports and environment variable overrides
- Global settings accessible via `config.Global`

### Request Flow

```
HTTP Request
    ↓
handler.HandleSub
    ↓
Fetch subscription(s) → parser.ParseSubscription
    ↓
Filter/Rename proxies → handler.filterProxies / processRemarks
    ↓
Load base template → handler.loadBaseFile
    ↓
Generate output → generator.Generate
    ↓
HTTP Response
```

---

## Building

### Local Build

```bash
# Standard build
make build

# Build with version info
make build VERSION=v1.0.0

# Multi-platform builds
make build-all
# Outputs: subconvergo-{linux,darwin,windows}-{amd64,arm64}[.exe]
```

### Development Mode

```bash
# Run without building
make dev

# Watch and rebuild on changes (requires entr)
make watch
```

### Docker Build

```bash
# Build image
make docker-build

# Run container
make docker-run
# Mounts base/ and pref.toml; exposes port 25500
```

---

## Testing

### Test Types

1. **Unit Tests** - Go package tests (`*_test.go`)
2. **Smoke Tests** - Integration/API tests via Python (`tests/smoke.py`)
3. **Benchmarks** - Performance tests

### Running Tests

```bash
# Unit tests (fast, no Docker)
make test-unit
# Or: ./tests/run-tests.sh unit

# Smoke tests (integration + API with Docker)
make test
# Or: python -m tests.smoke

# All tests
make test-all

# Coverage report
make coverage
make coverage-view  # Opens HTML report
```

### Smoke Test Details

`tests/smoke.py` orchestrates:
- Docker Compose stack (subconvergo + subconverter + mock-subscription)
- Per-scenario pref.yml generation
- Structural validation of YAML/JSON outputs
- Comparison with subconverter (non-fatal)

**Test scenarios:**
- `version` - Version endpoint check
- `sub` - Basic subscription conversion
- `render` - Template rendering
- `profile` - Profile handling
- `ruleset_remote` - Remote ruleset fetch
- `ruleset_compare` - Compare with subconverter
- `filters_regex` - Regex filtering
- `sub_with_external_config` - External config merging

**Run manually:**
```bash
cd /path/to/subconvergo
python -m tests.smoke
# Results: tests/results/smoke_summary.json
```

### Benchmarks

```bash
make bench
# Results: coverage/benchmark.txt
```

---

## Development Workflow

### Adding a New Feature

1. **Plan**: Update task tracking if complex
2. **Implement**: Follow existing patterns in corresponding layer
3. **Test**: Add unit tests in `*_test.go`
4. **Smoke**: Add scenario to `tests/smoke.py` if user-facing
5. **Document**: Update relevant doc sections

### Code Style

- Follow Go conventions: `gofmt`, `go vet`
- Run linter: `make lint` (requires `golangci-lint`)
- Format code: `make fmt`

### Adding Protocol Support

**Automatic via mihomo:**
```bash
# Update mihomo version in go.mod
go get github.com/metacubex/mihomo@latest
go mod tidy

# Rebuild and test
make build
make test-unit
```

**Manual parser (if needed):**
1. Add `parse<Protocol>()` function in `parser/parser.go`
2. Register in `parseSingleProxy()` switch
3. Validate with mihomo adapter
4. Add tests in `parser/parser_test.go`

### Adding Output Format

1. Add generator function in `generator/generator.go`
2. Register in `Generate()` switch
3. Create base template in `base/base/<format>_base.tpl`
4. Add smoke test scenario in `tests/smoke.py`

---

## Configuration

### Config File Priority

1. `pref.yml` (YAML)
2. `pref.toml` (TOML)
3. `pref.ini` (INI)

Auto-copies from `pref.example.*` on first run.

### Key Sections

```yaml
common:
  api_mode: true
  api_access_token: "password"
  default_url: ["http://example.com/sub"]
  include_remarks: []
  exclude_remarks: ["expired"]

node_pref:
  clash_use_new_field_name: true
  append_sub_userinfo: false

rulesets:
  enabled: true
  rulesets:
    - ruleset: "rules/custom.list"
      group: "Auto"

proxy_groups:
  custom_proxy_group:
    - name: "Auto"
      type: "select"
      rule: [".*"]

managed_config:
  write_managed_config: false
  managed_config_prefix: ""
  config_update_interval: 86400

server:
  listen: "0.0.0.0"
  port: 25500
```

### Environment Variables

Override config at runtime:

```bash
export API_MODE=true
export MANAGED_PREFIX=http://example.com:25500
export API_TOKEN=your_secret_token
export PORT=8080

./subconvergo
```

---

## Performance

### Benchmarks (Nov 2024)

| Operation | Time | Memory |
|-----------|------|--------|
| SS Parse | ~7.6µs | 3.2KB |
| VMess Parse | ~24.4µs | 8.1KB |
| Hysteria Parse | ~10.9µs | 4.5KB |
| TUIC Parse | ~16.1µs | 6.2KB |
| Clash Generate | ~2.1ms | 145KB |

**Test coverage: 81.8%**

### Optimization Tips

1. **Caching**: Enable `proxy_subscription: NONE` to skip unnecessary proxy fetches
2. **Filtering**: Use regex filters early to reduce processing
3. **Rulesets**: Set `update_ruleset_on_request: false` for static rulesets
4. **Logging**: Set `log_level: warn` in production

### Resource Limits

- Max subscriptions per request: configurable
- Max ruleset size: configurable
- Request timeout: 30s default

---

## API Reference

### Endpoints

#### `GET /sub`
Convert subscription(s) to target format.

**Parameters:**
- `target` - Output format: `clash`, `surge`, `quanx`, `loon`, `singbox`, `ss`, `ssr`, `v2ray`, `trojan`
- `url` - Subscription URL (pipe-separated for multiple)
- `config` - External config URL or file path
- `include` - Include filter (regex with `/...../`)
- `exclude` - Exclude filter (regex with `/...../`)
- `udp`, `tfo`, `scv` - Protocol flags (`true`/`false`)

**Example:**
```bash
curl "http://localhost:25500/sub?target=clash&url=https://example.com/sub&udp=true"
```

#### `GET /version`
Get version info.

#### `GET /readconf?token=<token>`
Reload configuration (requires `api_access_token`).

#### `GET /getruleset?url=<base64_url>&type=<clash|surge>`
Fetch and format ruleset.

#### `GET /render?path=<template_path>&token=<token>`
Render Go template with current config.

#### `GET /getprofile?name=<profile_name>&token=<token>`
Load profile from `base/profiles/<name>.ini`.

---

## Migration from C++ Subconverter

### Compatibility

- ✅ Same API endpoints and parameters
- ✅ Same config format (`pref.ini/yml/toml`)
- ✅ Same base templates and rules
- ✅ Equivalent proxy protocol support

### Differences

| Feature | C++ | Go |
|---------|-----|-----|
| Template Engine | inja | Go text/template |
| Script Engine | QuickJS | Not implemented |
| Cron Jobs | libcron | Not implemented |
| Regex Engine | PCRE2 | Go regexp (RE2) |

### Running Both Side-by-Side

```bash
# C++ version (port 25500)
cd subconverter
./subconverter

# Go version (port 25501)
cd subconvergo
PORT=25501 ./subconvergo
```

Test both with identical requests and compare outputs.

---

## Troubleshooting

### Build Issues

**Missing dependencies:**
```bash
go mod download
go mod tidy
```

**CGO errors:**
```bash
export CGO_ENABLED=0
go build
```

### Runtime Issues

**Config not found:**
- Ensure `base/pref.{yml,toml,ini}` exists
- Check working directory matches `base_path` setting

**Proxy parsing fails:**
- Verify proxy URL format
- Check mihomo version supports protocol: `go list -m github.com/metacubex/mihomo`

**Port already in use:**
```bash
# Change port
export PORT=8080
./subconvergo
```

### Test Failures

**Smoke tests timeout:**
- Check Docker is running: `docker ps`
- Pull images manually: `docker pull golang:alpine`

**Unit tests fail:**
- Check Go version: `go version` (requires 1.21+)
- Clean cache: `go clean -testcache`

---

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/my-feature`
3. Make changes following Go conventions
4. Add tests: unit + smoke scenario if applicable
5. Run full test suite: `make test-all`
6. Commit with clear messages
7. Push and open pull request

---

## License

GPL-3.0 - Same as original subconverter project
