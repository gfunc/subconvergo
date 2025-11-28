# Documentation

Welcome to the subconvergo documentation!

## 📖 Available Documentation

### [Configuration Reference](./REFERENCE.md)
Complete reference for all configuration options, proxy filtering, rulesets, protocols, and templates.

**Contents:**
- Protocol support (SS, SSR, VMess, Trojan, VLESS, Hysteria, Hysteria2, TUIC, Clash)
- Configuration file format (YAML/TOML/INI)
- Proxy filtering (basic, regex, advanced matchers)
- Rulesets configuration and formats
- Template system with Go templates
- Query parameter reference
- Tips and best practices

---

### [API Reference](./API.md)
Detailed documentation of API endpoints, parameters, source formats, and target formats.

### [Development Guide](./GUIDE.md)
Complete guide for building, testing, and developing subconvergo.

**Contents:**
- Quick start and prerequisites
- Architecture overview (parser, generator, handler, config layers)
- Building (local, cross-compile, Docker)
- Testing (unit, smoke, benchmarks)
- Development workflow and code style
- Adding protocols and output formats
- Configuration system deep-dive
- Performance benchmarks and optimization
- API reference
- Migration from C++ subconverter
- Troubleshooting

---

### [Feature Parity Status](./FEATURE_PARITY.md)
Comprehensive comparison with C++ subconverter - what's implemented, what's not, and migration guidance.

**Contents:**
- Feature coverage by category (80% overall)
- Implemented vs. not implemented features
- Migration considerations and compatibility
- Priority recommendations for future work
- Detailed feature-by-feature comparison

---

## Quick Links

- **Main README**: [../README.md](../README.md)
- **Configuration Reference**: [REFERENCE.md](./REFERENCE.md)
- **API Reference**: [API.md](./API.md)
- **Development Guide**: [GUIDE.md](./GUIDE.md)
- **Source Code**: [../parser/](../parser/), [../generator/](../generator/), [../handler/](../handler/)
- **Tests**: [../tests/smoke.py](../tests/smoke.py), [../tests/run-tests.sh](../tests/run-tests.sh)

---

## Getting Started

1. **New to the project?** Start with the main [README](../README.md) Quick Start section
2. **Need configuration help?** Check [Configuration Reference](./REFERENCE.md)
3. **Setting up development?** Read [Development Guide](./GUIDE.md)

---

## Testing

**Unit tests:**
```bash
make test-unit
# Or: ./tests/run-tests.sh unit
```

**Smoke tests (integration/API with Docker):**
```bash
make test
# Or: python -m tests.smoke
```

See [Development Guide - Testing](./GUIDE.md#testing) for details.

---

## Contributing

When updating documentation:

1. Update [REFERENCE.md](./REFERENCE.md) for config/protocol changes
2. Update [GUIDE.md](./GUIDE.md) for architecture/workflow changes
3. Keep examples up-to-date with code
5. Add smoke test scenarios for user-facing features

---

**Project**: subconvergo  
**Repository**: https://github.com/gfunc/subconvergo  
**Version**: Go 1.25.3+, mihomo v1.19.16  
**License**: MIT
