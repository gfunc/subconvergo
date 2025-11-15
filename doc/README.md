# Documentation

Welcome to the subconvergo documentation!

## 📖 Available Documentation

### [Testing Guide](./TESTING.md)
Complete testing guide with all test modes, workflows, and troubleshooting.

**Contents:**
- Quick start - run all tests
- Test infrastructure overview
- Test modes (docker/local/api/bench/all)
- Test coverage and analysis (62% overall)
- Command requirements and detection
- Docker testing workflows
- API endpoint testing
- Troubleshooting common issues
- Testing workflows and best practices

---

### [Quick Start Guide](./QUICKSTART.md)
Get started quickly with subconvergo - installation, configuration, and basic usage.

**Contents:**
- Prerequisites
- Installation steps
- Basic configuration
- Running the server
- First subscription conversion
- Common use cases

---

### [Development Guide](./DEVELOPMENT.md)
Development setup, guidelines, and best practices.

**Contents:**
- Development environment setup
- Code structure and conventions
- Adding new features
- Testing guidelines
- Pull request process
- Code review guidelines

---

### [Protocol Support](./PROTOCOL_SUPPORT.md)
Complete guide to all supported proxy protocols with examples and specifications.

**Contents:**
- Shadowsocks (ss://)
- ShadowsocksR (ssr://)
- VMess (vmess://)
- Trojan (trojan://)
- VLESS (vless://)
- Hysteria (hysteria://)
- Hysteria2 (hysteria2://, hy2://)
- TUIC (tuic://)
- Clash YAML format
- Protocol examples and usage
- Performance benchmarks

---

## Quick Links

- **Main README**: [../README.md](../README.md)
- **Testing Guide**: [TESTING.md](./TESTING.md)
- **Source Code**: [../parser/](../parser/)
- **Test Infrastructure**: [../tests/](../tests/)
- **Test Scripts**: [../tests/run-tests.sh](../tests/run-tests.sh)

---

## Getting Started

1. **New to the project?** Start with the [Quick Start Guide](./QUICKSTART.md)
2. **Setting up development?** Check [Development Guide](./DEVELOPMENT.md)
3. **Want to add protocol support?** See [Protocol Support](./PROTOCOL_SUPPORT.md)
4. **Need to run tests?** Read [Testing Guide](./TESTING.md)

---

## Contributing

When updating documentation:

1. Keep examples up-to-date with code
2. Update version numbers when dependencies change
3. Add new protocols to PROTOCOL_SUPPORT.md
4. Update test coverage in TESTING.md
5. Add new test workflows to TESTING.md
6. Update quick start guide for major changes

---

**Project**: subconvergo  
**Repository**: https://github.com/gfunc/subconvergo  
**Version**: Go 1.25.3+, mihomo v1.19.16  
**License**: MIT
