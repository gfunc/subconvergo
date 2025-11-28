# Subconvergo

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Test Coverage](https://img.shields.io/badge/coverage-81.8%25-brightgreen)](#-testing)
[![mihomo](https://img.shields.io/badge/mihomo-v1.19.16+-blue)](https://github.com/metacubex/mihomo)

A high-performance Go reimplementation of [subconverter](https://github.com/tindy2013/subconverter) using [mihomo](https://github.com/metacubex/mihomo) for automatic proxy protocol support.

**Module**: `github.com/gfunc/subconvergo`

> **Status**: Production-ready for Clash format. Other formats are functional but less thoroughly tested.

---

## 🚀 Quick Start

```bash
# Clone and build
git clone https://github.com/gfunc/subconvergo.git
cd subconvergo
make build

# Run server
./subconvergo
# Server starts on http://localhost:25500

# Test it
curl http://localhost:25500/version
```

---

## ✨ Features

- **8+ Proxy Protocols**: SS, SSR, VMess, Trojan, VLESS, Hysteria, Hysteria2, TUIC
- **High Performance**: Sub-25µs parsing, 81.8% test coverage
- **Auto Protocol Updates**: New protocols supported via mihomo upgrades
- **Full Compatibility**: Same API and config format as C++ version
- **Production Ready**: 50+ tests, comprehensive benchmarks, Docker support
- **Extensible**: Fallback architecture for future protocols
- **Improved Observability**: Parser, handler, and generator log contextual info/errors to simplify troubleshooting

---

## 📚 Documentation

- **[Configuration Reference](./doc/REFERENCE.md)** - Complete config options, filtering, protocols
- **[API Reference](./doc/API.md)** - Endpoints, parameters, supported formats
- **[Development Guide](./doc/GUIDE.md)** - Building, testing, architecture, workflow
- **[Smoke Tests](./doc/SMOKE_TESTS.md)** - Detailed guide to the integration test suite
- **[Feature Parity](./doc/FEATURE_PARITY.md)** - Comparison with C++ subconverter

---

## ⚠️ Behavioral Differences

While Subconvergo aims for full compatibility with the C++ version, there are some intentional differences:

1.  **Clash Source Format**: When parsing a Clash configuration as a subscription source, Subconvergo **preserves** the `proxy-groups` and `rules` defined in the source file and adds them to the target output. This allows for easier migration of complex Clash configs.
2.  **Protocol Parsing**: Subconvergo uses `mihomo` adapters for parsing, which may have stricter or slightly different validation logic compared to the custom parsers in the C++ version.
3.  **Local File URI**: Subconvergo supports `file://` URIs for subscription sources

---

## 🧪 Testing

**Coverage**: 81.8% (parser) | 72% (generator) | 30% (handler) | **Status**: ✅ All Passing

### Quick Test

```bash
# Unit tests (fast, no Docker)
make test-unit

# Smoke tests (integration/API with Docker)
make test
```

See **[Smoke Tests](./doc/SMOKE_TESTS.md)** for details on the integration test suite and parity verification.

---

## 🛠️ Development

### Prerequisites

- **Go 1.25+**
- **Docker** (for smoke tests)
- **Python 3.8+** (for smoke tests)

### Build & Run

```bash
# Install dependencies
go mod download

# Build
make build

# Development mode (auto-reload)
make dev
```

---

## 🔄 Migration from C++ Version

This Go implementation is a **drop-in replacement** for most use cases.

### ✅ **Compatible**
- Same `base/` directory structure
- Same configuration format (pref.ini/yml/toml)
- Core API endpoints (`/sub`, `/version`, `/readconf`, `/getprofile`, `/getruleset`, `/render`)
- All proxy protocols (SS, SSR, VMess, Trojan, VLESS, Hysteria, etc.)

### ⚠️ **Not Implemented**
- `list` parameter (Node List/Proxy Provider output)
- `filename` parameter
- QuickJS filter/sort script execution

📖 **[Complete Feature Parity Status](./doc/FEATURE_PARITY.md)**

---

## 🤝 Contributing

Contributions welcome! Please ensure:

1. ✅ Unit tests pass: `make test-unit`
2. ✅ Smoke tests pass: `make test`
3. ✅ Code formatted: `make fmt`

📖 **[Development Guide](./doc/GUIDE.md)**



---

##  License

MIT License - See [LICENSE](./LICENSE) file

---

## 🙏 Acknowledgments

- [mihomo](https://github.com/metacubex/mihomo) - Proxy protocol support
- [subconverter](https://github.com/tindy2013/subconverter) - Original implementation


