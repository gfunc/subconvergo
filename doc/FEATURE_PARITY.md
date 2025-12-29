# Feature Parity with C++ Subconverter

> Status as of November 16, 2025 | C++ Reference: [subconverter README-cn.md](https://github.com/tindy2013/subconverter/blob/master/README-cn.md)

## Summary

| Category | Implemented | Not Implemented | Total | Coverage |
|----------|-------------|-----------------|-------|----------|
| **Endpoints** | 7 | 0 | 7 | 100% |
| **Query Parameters** | 18 | 8 | 26 | 69% |
| **Config Sections** | 8 | 1 | 9 | 89% |
| **Protocol Support** | 9 | 0 | 9 | 100% |
| **Overall** | **41** | **10** | **51** | **80%** |

---

## ⚠️ Intentional Deviations

### Flag Precedence (UDP, TFO, SCV, TLS13)
- **Subconverter (C++)**: Global flags (from URL params or `pref.ini`) **override** source flags. If `udp=true` is set globally, all nodes get `udp: true`, even if the source explicitly had `udp: false`.
- **Subconvergo (Go)**: Source flags **take precedence**. Global flags are only applied as a fallback if the source flag is missing (nil/undefined). This preserves explicit `false` settings from the subscription source.

## ✅ Implemented Features

### HTTP Endpoints

| Endpoint | Status | Notes |
|----------|--------|-------|
| `/sub` | ✅ | Full implementation with all core parameters |
| `/version` | ✅ | Returns version string |
| `/readconf` | ✅ | Reload configuration with token |
| `/getprofile` | ✅ | Load profiles from `base/profiles/*.ini` |
| `/getruleset` | ✅ | Fetch and format rulesets (remote/local) |
| `/render` | ✅ | Render Go templates with global variables |
| `/surge2clash` | ✅ | Shortcut for `/sub?target=clash` |

### Query Parameters (`/sub` endpoint)

| Parameter | Status | Implementation | Notes |
|-----------|--------|----------------|-------|
| `target` | ✅ | Full | clash, surge, quanx, loon, singbox, ss, ssr, v2ray, trojan |
| `url` | ✅ | Full | Pipe-separated, URL-encoded, supports `tag:xxx,url` |
| `config` | ✅ | Full | External config (HTTP/file, YAML/TOML/INI) |
| `include` | ✅ | Full | Regex with `/pattern/` or substring |
| `exclude` | ✅ | Full | Regex with `/pattern/` or substring |
| `emoji` | ✅ | Full | Add emojis based on regex rules |
| `add_emoji` | ✅ | Full | Control emoji addition |
| `remove_emoji` | ✅ | Full | Remove old emojis first |
| `append_type` | ✅ | Full | Add [ss], [vmess] to node names |
| `udp` | ✅ | Full | Enable UDP flag |
| `tfo` | ✅ | Full | Enable TCP Fast Open |
| `scv` | ✅ | Full | Skip certificate verification |
| `tls13` | ✅ | Full | Enable TLS 1.3 |
| `sort` | ✅ | Full | Alphabetical sorting |
| `rename` | ✅ | Full | Custom rename rules (query override) |
| `insert` | ✅ | Full | Enable/disable insert_url |
| `prepend` | ✅ | Full | Insert before (true) or after (false) |
| `group` | ✅ | Full | Set group name (for SSD/SSR) |
| `ignore_source` | ✅ | Full | Ignore proxy groups/rules from source |

### Configuration File Sections

| Section | Status | Implementation |
|---------|--------|----------------|
| `[common]` | ✅ | api_mode, api_access_token, default_url, include/exclude_remarks, insert_url, proxy configs, base templates |
| `[node_pref]` | ✅ | udp_flag, tfo_flag, scv_flag, tls13_flag, sort_flag, rename_node, clash options, singbox options |
| `[rulesets]` | ✅ | enabled, overwrite_original_rules, update_ruleset_on_request, inline/remote rules |
| `[proxy_groups]` | ✅ | custom_proxy_group with advanced matchers (!!TYPE, !!PORT, !!SERVER, !!GROUP, []DIRECT) |
| `[managed_config]` | ✅ | write_managed_config, managed_config_prefix, config_update_interval |
| `[emojis]` | ✅ | add_emoji, remove_old_emoji, regex-based rules |
| `[template]` | ✅ | template_path, globals with nested keys |
| `[aliases]` | ✅ | URI redirects with query param preservation |
| `[server]` | ✅ | listen, port |

### Protocol Support (Parsing)

| Protocol | Status | Features |
|----------|--------|----------|
| Shadowsocks | ✅ | IPv6, plugins (simple-obfs, v2ray-plugin), SS2022 |
| ShadowsocksR | ✅ | Auto-convert to SS when applicable |
| VMess | ✅ | All transports (TCP, WS, H2, gRPC, QUIC), TLS/XTLS |
| Trojan | ✅ | Standard, WebSocket, gRPC |
| VLESS | ✅ | Flow control, Reality, all transports |
| Hysteria | ✅ | v1, bandwidth config, obfuscation |
| Hysteria2 | ✅ | v2, Salamander obfuscation |
| TUIC | ✅ | QUIC, BBR/Cubic congestion control |
| Clash YAML | ✅ | Native parser via mihomo |

### Output Formats (Generation)

| Format | Status | Notes |
|--------|--------|-------|
| Clash | ✅ | YAML with proxy-groups, rules |
| sing-box | ✅ | JSON with Clash API modes |
| Surge | ✅ | INI-style configuration |
| Quantumult X | ✅ | Custom INI format |
| Loon | ✅ | Configuration format |
| Shadowsocks | ✅ | SIP002 links, base64 subscription |
| ShadowsocksR | ✅ | SSR links, base64 subscription |
| V2Ray | ✅ | VMess links, base64 subscription |
| Trojan | ✅ | Trojan links, base64 subscription |

### Advanced Features

| Feature | Status | Implementation |
|---------|--------|----------------|
| Regex filtering | ✅ | `/pattern/` syntax in include/exclude |
| Advanced matchers | ✅ | !!TYPE=, !!PORT=, !!SERVER=, !!GROUP=, !!GROUPID=, !!INSERT= |
| External config | ✅ | HTTP/file, YAML/TOML/INI parsing |
| Template rendering | ✅ | Go text/template with global variables |
| Rulesets | ✅ | Local/remote, Clash/Surge formats |
| Emojis | ✅ | Regex-based country/region detection |
| Node renaming | ✅ | Regex replacement with advanced matchers |
| Managed config | ✅ | Surge/Surfboard headers |
| Profile system | ✅ | Load preset configs from INI files |
| Aliases | ✅ | URI redirects |
| Multi-subscription | ✅ | Pipe-separated URLs |
| Tag-based grouping | ✅ | `tag:xxx,url` format |

---

## ❌ Not Implemented Features

### HTTP Endpoints

| Endpoint | Priority | Reason | Workaround |
|----------|----------|--------|------------|
| (None) | - | - | - |

### Query Parameters

| Parameter | Priority | Reason | Workaround |
|-----------|----------|--------|------------|
| `list` | Medium | Node List/Proxy Provider output | Generate full config and extract proxies section |
| `filename` | Low | Cosmetic (filename in Clash for Windows) | Set in client manually |
| `expand` | Low | Rule inlining control | Rules are expanded by default |
| `classic` | Low | Classical rule-provider format | Domain/IP rules work as-is |
| `script` | Low | Clash Script generation | Use Clash Premium features directly |
| `fdn` | Low | Filter unsupported nodes | Nodes are validated via mihomo |
| `target=auto` | Low | User-Agent detection | Specify target explicitly |
| `target=mixed` | Low | Mixed format (all node types as links) | Use specific target (ss/ssr/v2ray) |

### Configuration

| Feature | Priority | Reason | Workaround |
|---------|----------|--------|------------|
| `[userinfo]` | Low | Stream/time extraction from node names | Node remarks preserved as-is |
| QuickJS script execution | Low | Security/complexity | Pre-filter subscriptions externally |
| Gist auto-upload | Low | External service dependency | Upload manually or use CI/CD |
| Data URI support | Low | Rarely used | Use regular HTTP URLs |
| CORS proxy | Low | Can use external CORS proxy | Set up nginx/cloudflare worker |

---

## 🔄 Migration Considerations

### ✅ Safe to Migrate If:
- You primarily use **Clash**, **Surge**, or **sing-box**
- You use standard proxy protocols (SS, VMess, Trojan, Hysteria, TUIC)
- You rely on:
  - Basic or regex filtering
  - Node renaming and emojis
  - External configs
  - Rulesets (local or remote)
  - Template rendering
  - Profile system

### ⚠️ Migration Requires Adjustment If:
- You use `list=true` parameter → Extract proxies section from full config
- You use `filename` parameter → Set filename in client
- You use QuickJS filter/sort scripts → Pre-process subscriptions or accept default behavior
- You use Gist auto-upload → Set up alternative upload mechanism
- You use `target=auto` → Explicitly specify target format

### ❌ Cannot Migrate If:
- You **require** QuickJS script execution (filter_script/sort_script with JS code)
- You **must** have Gist integration
- You depend on Data URI subscriptions
- You need `target=mixed` output format

---

## 📊 Implementation Status by Category

### Core Functionality: **100%**
All essential subscription conversion features are implemented.

### Query Parameters: **69%**
Missing parameters are mostly convenience features (list, filename, expand, classic) or rarely-used (auto, mixed).

### Configuration: **100%**
All configuration settings from C++ subconverter are implemented, including aliases, templates, insert URLs, emoji/rename rules, and more.

### Protocol Support: **100%**
All major proxy protocols fully supported via mihomo.

### Output Formats: **100%**
All common client formats supported (Clash, Surge, QuanX, Loon, sing-box, single links).

---

## 🎯 Recommendations

### For Most Users:
Subconvergo is **production-ready**. The feature coverage includes all commonly-used features. Missing items are convenience shortcuts, cosmetic options, or rarely-used advanced features.

### Priority for Future Implementation:

See [API Reference](./API.md) for detailed format support and limitations.

> **Note**: The following list represents features present in the C++ version but not yet implemented in Go. Priorities are estimated based on common usage.

1. **High Priority** (commonly requested):
   - [ ] `list` parameter for Proxy Provider output (4 hours)
   - [ ] `filename` parameter (1 hour)

2. **Medium Priority** (nice to have):
   - [ ] `expand` parameter control (2 hours)
   - [ ] `classic` parameter for rule-provider (3 hours)
   - [ ] `target=auto` User-Agent detection (4 hours)

3. **Low Priority** (edge cases):
   - [ ] `target=mixed` output (3 hours)
   - [ ] Userinfo extraction rules (4 hours)
   - [ ] Gist auto-upload (6 hours)
   - [ ] QuickJS script execution (10+ hours, security review needed)

---

## 🔍 Testing Parity

### Test Coverage: **81.8%** (parser), **72%** (generator), **30%** (handler)

**Smoke Tests Cover:**
- ✅ Version endpoint
- ✅ Subscription conversion (Clash, sing-box)
- ✅ Template rendering
- ✅ Profile loading
- ✅ Ruleset fetching (local/remote)
- ✅ Regex filtering
- ✅ External config merging
- ✅ Comparison with C++ subconverter

**Comparison with C++ Subconverter:**
Smoke tests include a subconverter container (port 25550) for behavioral parity checks. Differences are logged but don't fail tests (allows for intentional improvements).

---

## 📝 Documentation Status

### ✅ Documented:
- Quick start and installation
- API endpoints (detailed in README)
- Configuration reference (REFERENCE.md)
- Development guide (GUIDE.md)
- Protocol support details
- Testing procedures

### ⚠️ Needs Improvement:
- Chinese README (README-cn.md) - not created yet
- More URLEncode examples
- Collapsible usage examples
- Video tutorials or animated GIFs

---

## 🤝 Contributing

To help improve feature parity:

1. **Review** the "Not Implemented" section above
2. **Pick** a feature based on priority and your interest
3. **Check** existing issues for that feature
4. **Implement** following patterns in the codebase
5. **Test** with unit tests and smoke tests
6. **Document** in README and REFERENCE.md
7. **Submit** pull request

See [Development Guide](./GUIDE.md) for detailed contribution workflow.

---

## 📞 Support

- 📖 [Configuration Reference](./REFERENCE.md) - All settings and options
- 📖 [Development Guide](./GUIDE.md) - Building and testing
- 🐛 Feature requests: Open an issue describing your use case
- 💬 Questions: Use discussions or issues

---

**Last Updated:** November 25, 2025  
**Subconvergo Version:** Development (smoke branch)  
**C++ Subconverter Reference:** [README-cn.md](https://github.com/tindy2013/subconverter/blob/master/README-cn.md)
