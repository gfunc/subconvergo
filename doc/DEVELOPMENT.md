# Subconverter Go Implementation - Development Guide

## Overview

This Go implementation replaces the C++ codebase with a more maintainable solution using the mihomo package for automatic proxy protocol support.

Module path: `github.com/gfunc/subconvergo`

## Project Structure

```
subconvergo/
├── config/           # Configuration loading and management
│   └── config.go     # Loads pref.ini/yml/toml, manages global settings
├── parser/           # Subscription parsing
│   └── parser.go     # Uses mihomo to parse proxy subscriptions
├── generator/        # Format conversion
│   └── generator.go  # Converts to Clash, Surge, Quantumult X, etc.
├── handler/          # HTTP request handlers
│   └── handler.go    # Implements /sub, /version, /readconf endpoints
├── main.go           # Application entry point
├── go.mod            # Go module dependencies
├── build.sh          # Build script
├── Dockerfile        # Docker build configuration
└── README.md         # This file
```

## Key Advantages

### 1. Automatic Protocol Support
By using `github.com/metacubex/mihomo`, new proxy protocols are automatically supported:
- Just upgrade mihomo version in go.mod
- No need to manually implement parsers for new protocols
- Built-in configuration validation

### 2. Easier Maintenance
- Go's strong type system catches errors at compile time
- Standard library provides robust HTTP, JSON, YAML support
- Simpler codebase (~1000 lines vs ~10000 lines C++)

### 3. Full Compatibility
- Same HTTP API endpoints (/sub, /version, /readconf, etc.)
- Same configuration format (pref.ini/yml/toml)
- Uses same base/ directory for templates and rules
- Can run alongside C++ version for testing

## Implementation Details

### Configuration Loading (config/config.go)

Implements the same 3-format priority as C++ version:
1. Try pref.toml (TOML format)
2. Try pref.yml (YAML format)
3. Try pref.ini (INI format)

Key settings structure mirrors C++ `Settings` struct:
```go
type Settings struct {
    APIMode          bool
    ManagedConfigPrefix string
    ClashUseNewField bool
    CustomProxyGroups []ProxyGroupConfig
    // ... matches src/handler/settings.h
}
```

### Parser (parser/parser.go)

**COMPLETED**: Full implementation of all major proxy protocols with mihomo validation.

Implements share link parsing for:
- **Shadowsocks (ss://)**: Both old and new format, with plugin support
- **ShadowsocksR (ssr://)**: Full protocol/obfs support, auto-converts to SS when applicable
- **VMess (vmess://)**: JSON format with all transport protocols (ws, h2, grpc, quic)
- **Trojan (trojan://)**: Standard and WebSocket/gRPC variants
- **VLESS (vless://)**: Full protocol support with various transports
- **Clash YAML**: Native support via mihomo

Each parser:
```go
// Parse SS link and validate with mihomo
proxy, err := parseShadowsocks("ss://...")

// All parsers return validated Proxy struct
type Proxy struct {
    Type          string
    Remark        string
    Server        string
    Port          int
    // ... protocol-specific fields
    MihomoProxy   constant.Proxy  // Validated mihomo proxy
}
```

Parsing behavior matches subconverter C++ implementation:
- URL decoding and base64 decoding
- Query parameter extraction
- Regex matching for link formats
- Automatic protocol detection
- SSR to SS conversion when applicable

All parsers include comprehensive test coverage in `parser_test.go`.

### Generator (generator/generator.go)

Converts proxies to target formats:
- **Clash/ClashR**: YAML output with proxy-groups
- **Surge**: INI-style configuration
- **Quantumult X**: Custom INI format
- **sing-box**: JSON configuration
- **Single formats**: Base64-encoded proxy links

### HTTP Handler (handler/handler.go)

Implements all main endpoints:
- `GET /sub` - Main conversion endpoint
- `GET /version` - Version information
- `GET /readconf` - Reload configuration
- `GET /getruleset` - Serve rulesets
- `GET /render` - Template rendering

Query parameters match C++ version:
- `target` - Output format (clash, surge, quanx, etc.)
- `url` - Subscription URL (supports multiple with `|`)
- `config` - External configuration URL
- `include`/`exclude` - Regex filtering
- `udp`/`tfo`/`scv` - Protocol options

## Building

### Local Development
```bash
cd subconvergo
./build.sh
```

### Cross-compilation
```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o subconvergo-linux

# Windows
GOOS=windows GOARCH=amd64 go build -o subconvergo.exe

# macOS
GOOS=darwin GOARCH=amd64 go build -o subconvergo-mac
```

### Docker
```bash
cd subconvergo
docker build -t subconvergo:latest .
docker run -d -p 25500:25500 subconvergo:latest
```

## Usage

### Basic Usage
```bash
# Run from subconvergo/ directory
./subconvergo

# Specify custom config
./subconvergo -f /path/to/pref.ini

# Enable logging to file
./subconvergo -l /var/log/subconverter.log
```

### Example API Calls
```bash
# Convert subscription to Clash
curl "http://localhost:25500/sub?target=clash&url=https%3A%2F%2Fexample.com%2Fsub"

# With filtering
curl "http://localhost:25500/sub?target=surge&url=...&exclude=%E8%BF%87%E6%9C%9F"

# Check version
curl "http://localhost:25500/version"
```

### Environment Variables
```bash
export API_MODE=true
export MANAGED_PREFIX=http://example.com:25500
export API_TOKEN=your_secret_token
export PORT=8080

./subconvergo
```

## Migration Path

### Phase 1: Parallel Testing
1. Run Go version on different port (e.g., 25501)
2. Test with same requests as C++ version
3. Compare outputs for consistency

### Phase 2: Feature Parity
- [ ] Complete all format generators (Surge, QuanX, Loon)
- [ ] Implement template rendering support
- [ ] Add ruleset fetching and caching
- [ ] Implement filter/sort scripts (QuickJS alternative)

### Phase 3: Production Deployment
1. Deploy Go version alongside C++
2. Gradually migrate traffic
3. Monitor performance and compatibility
4. Full switchover once stable

## TODO: Remaining Work

### High Priority
- [ ] Complete Surge format generator
- [ ] Complete Quantumult X format generator
- [ ] Complete Loon format generator
- [ ] Implement proper regex filtering (PCRE2 equivalent)
- [ ] Implement template rendering (Go templates vs inja)

### Medium Priority
- [ ] Ruleset fetching and caching
- [ ] External config loading and merging
- [ ] User-Agent based auto-detection
- [ ] Upload to Gist functionality

### Low Priority
- [ ] JavaScript filter/sort scripts (consider goja instead of QuickJS)
- [ ] Cron job support for scheduled updates
- [ ] Generator mode for offline conversion

## Completed Work

### ✅ Parser Implementation (2024-11-13)
- Implemented all major proxy protocol parsers (SS, SSR, VMess, Trojan, VLESS)
- All parsers validate using mihomo adapter for correctness
- Parsing behavior matches subconverter C++ implementation
- Added comprehensive test coverage with passing tests
- Supports both old and new share link formats
- Handles query parameters, base64 encoding, URL encoding
- SSR automatically converts to SS when appropriate (matching subconverter behavior)

**Files modified:**
- `parser/parser.go`: Added ~700 lines implementing all protocol parsers
- `parser/parser_test.go`: Added comprehensive test suite

**Key functions implemented:**
- `parseShadowsocks()`: ss:// link parser with plugin support
- `parseShadowsocksR()`: ssr:// link parser with protocol/obfs support
- `parseVMess()`: vmess:// JSON format parser
- `parseTrojan()`: trojan:// link parser with WS/gRPC support
- `parseVLESS()`: vless:// link parser with full transport options
- Helper functions: `urlDecode()`, `urlSafeBase64Decode()`, `parsePluginOpts()`

All tests pass successfully.

## Performance Considerations

### Memory Usage
- Go's garbage collector handles memory automatically
- Expected lower memory footprint than C++
- No need for `malloc_trim()` equivalent

### Concurrency
- Gin framework handles concurrent requests efficiently
- Can easily add middleware for rate limiting
- Connection pooling for subscription fetching

### Caching
- Implement in-memory cache for:
  - Subscription content (TTL: 60s)
  - External configs (TTL: 300s)
  - Rulesets (TTL: 21600s)

## Testing

```bash
# Run tests
cd subconvergo
go test ./...

# Run with coverage
go test -cover ./...

# Benchmark
go test -bench=. ./...
```

## Contributing

When adding new features:
1. Follow Go standard project layout
2. Maintain API compatibility with C++ version
3. Add tests for new functionality
4. Update this documentation

## License

Same as parent project (GPL-3.0)
