# Subconvergo Configuration Reference

> Complete reference for all configuration options, filtering, and protocol support

## Table of Contents

- [Protocol Support](#protocol-support)
- [Configuration File](#configuration-file)
- [Proxy Filtering](#proxy-filtering)
- [Rulesets](#rulesets)
- [Templates](#templates)
- [Advanced Features](#advanced-features)

---

## Protocol Support

### Supported Protocols

| Protocol | Prefix | Status | Performance | Notes |
|----------|--------|--------|-------------|-------|
| **Shadowsocks** | `ss://` | ✅ Full | ~7.6µs | IPv6, plugins, SS2022 |
| **ShadowsocksR** | `ssr://` | ✅ Full | - | Auto-convert to SS |
| **VMess** | `vmess://` | ✅ Full | ~24.4µs | All transports, TLS |
| **Trojan** | `trojan://` | ✅ Full | - | WS/gRPC, SNI |
| **VLESS** | `vless://` | ✅ Full | - | Reality, flow control |
| **Hysteria** | `hysteria://` | ✅ Full | ~10.9µs | v1, bandwidth config |
| **Hysteria2** | `hy2://`, `hysteria2://` | ✅ Full | ~10.9µs | v2, obfuscation |
| **TUIC** | `tuic://` | ✅ Full | ~16.1µs | QUIC, BBR/Cubic |
| **AnyTLS** | `anytls://` | ✅ Full | - | TLS-based proxy |
| **Snell** | `snell://` | ✅ Full | - | Link, Surge, Clash format |
| **WireGuard** | `wireguard://`, `wg://` | ✅ Full | - | Surge/Clash format (no standalone link) |
| **SOCKS5** | `socks5://`, `socks://` | ✅ Full | - | Link, Surge, Telegram format |
| **HTTP/HTTPS** | `http://`, `https://` | ✅ Full | - | Link, Surge, Telegram format |
| **Clash YAML** | - | ✅ Full | - | Native parser |

✅ = Explicitly implemented and tested

### Protocol Details

All parsers validate using [mihomo](https://github.com/metacubex/mihomo) adapters for correctness. New protocols are automatically supported when mihomo is upgraded.

**Flag Precedence (UDP, TFO, SCV, TLS13):**
Unlike the C++ version, **source flags take precedence over global flags**. Global flags (e.g., `udp=true` in URL or `udp_flag` in config) are only applied if the source flag is missing (nil/undefined). This ensures that explicit `false` settings from the subscription source are preserved.

**Upgrade mihomo:**
```bash
go get github.com/metacubex/mihomo@latest
go mod tidy
make build
```

---

## Configuration File

### File Priority

Checked in order (first found wins):

1. `pref.yml` (YAML)
2. `pref.toml` (TOML)
3. `pref.ini` (INI)

Auto-copies from `pref.example.*` on first run.

### Complete Configuration Example

```yaml
# Common Settings
common:
  api_mode: true                          # Enable API mode
  api_access_token: ""                      # Token for protected endpoints; empty = those endpoints disabled (fail closed)
  default_url: []                         # Default subscription URLs
  base_path: "base"                       # Base directory for templates/rules
  
  # Subscription proxies
  proxy_subscription: "NONE"              # Proxy for fetching subscriptions
  proxy_config: "NONE"                    # Proxy for fetching configs
  proxy_ruleset: "NONE"                   # Proxy for fetching rulesets
  
  # Filtering
  include_remarks: []                     # Include filter (regex: /.../)
  exclude_remarks: []                     # Exclude filter (regex: /.../)
  
  # Base templates
  clash_rule_base: "base/all_base.tpl"
  surge_rule_base: "base/surge_base.conf"
  quanx_rule_base: "base/quanx_base.conf"
  loon_rule_base: "base/loon_base.conf"
  singbox_rule_base: "base/singbox_base.json"
  
  # Insert URLs
  enable_insert: false                    # Enable auto-insert URLs
  insert_url: []                          # URLs to insert
  prepend_insert_url: false               # Insert before (true) or after (false)
  
  # Proxy options
  append_proxy_type: false                # Add [ss], [vmess] to names

# Advanced Settings
advanced:
  log_level: "info"                       # Log level (debug, info, warn, error)
  print_debug_info: false                 # Print debug info to stdout
  max_pending_connections: 10240          # Max requests queued for a slot (excess gets HTTP 503)
  max_concurrent_threads: 64              # Max in-flight requests (0 = no limit)
  max_allowed_rulesets: 64                # Max rulesets per request, excess dropped (0 = unlimited)
  max_allowed_rules: 0                    # Max inline rules per request, excess dropped (0 = unlimited)
  max_allowed_download_size: 0            # Max bytes per remote download (0 = built-in default 32MiB)
  
  # Caching
  enable_cache: false                     # Enable file-based caching
  cache_subscription: 60                  # Subscription cache TTL (seconds)
  cache_config: 300                       # Config cache TTL (seconds)
  cache_ruleset: 21600                    # Ruleset cache TTL (seconds)
  serve_cache_on_fetch_fail: true         # Serve stale cache on fetch failure

# Node Preferences
node_pref:
  # Clash options
  clash_use_new_field_name: true          # Use 'proxies' vs 'Proxy'
  clash_proxies_style: "flow"             # 'flow' or 'block'
  clash_proxy_groups_style: "block"       # 'flow' or 'block'
  
  # UDP/TFO/SCV
  udp_flag: false                         # Enable UDP by default
  tfo_flag: false                         # Enable TCP Fast Open
  skip_cert_verify_flag: false            # Skip certificate verification
  tls13_flag: false                       # Enable TLS 1.3
  
  # sing-box
  singbox_add_clash_modes: true           # Add Clash API modes
  
  # Userinfo
  append_sub_userinfo: true               # Forward subscription-userinfo header
  
  # Sorting
  sort_flag: false                        # Enable alphabetical sorting
  
  # Rename rules (format: match@replace)
  rename_node: []

# Rulesets
rulesets:
  enabled: true
  overwrite_original_rules: false
  update_ruleset_on_request: false
  rulesets:
    - ruleset: "rules/custom.list"        # Local file or URL
      group: "Auto"
    - rule: "MATCH,Auto"                  # Inline rule
      group: "Auto"

# Proxy Groups
proxy_groups:
  custom_proxy_group:
    - name: "Auto"
      type: "select"                      # select, url-test, fallback, load-balance
      rule: [".*"]                        # Proxy filter rules
    - name: "HK"
      type: "url-test"
      rule: ["/^HK/"]                     # Regex filter
      url: "http://www.gstatic.com/generate_204"
      interval: 300

# Managed Config (Surge/Surfboard)
managed_config:
  write_managed_config: false
  managed_config_prefix: ""               # External URL prefix
  config_update_interval: 86400           # Update interval (seconds)
  config_update_strict: false

# Emojis
emojis:
  add_emoji: true
  remove_old_emoji: true
  rule:
    - match: "Hong Kong"
      emoji: "🇭🇰"
    - match: "United States"
      emoji: "🇺🇸"
    - match: "Japan"
      emoji: "🇯🇵"

# Templates
template:
  template_path: "base/base"              # Template directory
  globals:
    - key: "clash.http_port"
      value: 7890
    - key: "clash.socks_port"
      value: 7891
    - key: "clash.allow_lan"
      value: true

# Aliases (URI Redirects)
aliases:
  - uri: "/clash"
    target: "/sub?target=clash"
  - uri: "/surge"
    target: "/sub?target=surge"

# Server
server:
  listen: "0.0.0.0"
  port: 25500

# Advanced
advanced:
  log_level: "info"                       # debug, info, warn, error
```

### Environment Variable Overrides

Override at runtime:

| Variable | Config Key | Example |
|----------|------------|---------|
| `API_MODE` | `common.api_mode` | `API_MODE=true` |
| `API_TOKEN` | `common.api_access_token` | `API_TOKEN=secret` |
| `MANAGED_PREFIX` | `managed_config.managed_config_prefix` | `MANAGED_PREFIX=http://example.com` |
| `PORT` | `server.port` | `PORT=8080` |

```bash
API_MODE=true PORT=8080 ./subconvergo
```

---

## Proxy Filtering

### Basic Filtering

**Include/Exclude by remark:**

```yaml
common:
  include_remarks: ["HK", "US"]           # Substring match
  exclude_remarks: ["expired", "test"]    # Substring match
```

**Query parameters:**
```bash
# Include only HK proxies
curl "http://localhost:25500/sub?target=clash&url=...&include=HK"

# Exclude expired proxies
curl "http://localhost:25500/sub?target=clash&url=...&exclude=expired"
```

### Regex Filtering

Use `/pattern/` syntax for regex:

```yaml
common:
  include_remarks: ["/^HK-/"]            # Starts with HK-
  exclude_remarks: ["/x\\d+/"]           # Contains x followed by digits
```

```bash
# Regex in query
curl "http://localhost:25500/sub?target=clash&url=...&include=/^HK/"
```

### Advanced Matchers

Use special matchers in proxy group rules:

#### Type Matcher: `!!TYPE=`

Filter by protocol type:

```yaml
proxy_groups:
  custom_proxy_group:
    - name: "SS Only"
      type: "select"
      rule: ["!!TYPE=SS"]
    - name: "V2Ray"
      type: "select"
      rule: ["!!TYPE=VMESS|TROJAN"]      # Multiple types
    - name: "US SS"
      type: "select"
      rule: ["!!TYPE=SS!!.*US.*"]         # Type + name regex
```

**Supported types:** `SS`, `SSR`, `VMESS`, `TROJAN`, `VLESS`, `HYSTERIA`, `HYSTERIA2`, `TUIC`, `SNELL`, `HTTP`, `HTTPS`, `SOCKS5`, `WIREGUARD`

#### Server Matcher: `!!SERVER=`

Filter by server address (regex):

```yaml
proxy_groups:
  custom_proxy_group:
    - name: "Example Servers"
      type: "select"
      rule: ["!!SERVER=.*\\.example\\.com"]
    - name: "US Servers"
      type: "select"
      rule: ["!!SERVER=us[0-9]+\\..*"]
```

#### Port Matcher: `!!PORT=`

Filter by port number or range:

```yaml
proxy_groups:
  custom_proxy_group:
    - name: "HTTPS Proxies"
      type: "select"
      rule: ["!!PORT=443"]
    - name: "High Ports"
      type: "select"
      rule: ["!!PORT=8000-9000"]
    - name: "Multiple Ports"
      type: "select"
      rule: ["!!PORT=443,8080,8443"]
```

#### Group Matcher: `!!GROUP=`

Filter by subscription group:

```yaml
proxy_groups:
  custom_proxy_group:
    - name: "Premium Nodes"
      type: "select"
      rule: ["!!GROUP=Premium"]
    - name: "Premium US"
      type: "select"
      rule: ["!!GROUP=Premium!!.*US.*"]
```

#### Combining Matchers

Stack multiple matchers in one rule:

```yaml
proxy_groups:
  custom_proxy_group:
    - name: "US HTTPS VMess"
      type: "select"
      rule: ["!!TYPE=VMESS!!PORT=443!!SERVER=.*\\.us\\..*"]
```

#### Direct Node Inclusion

Use `[]` prefix for special nodes:

```yaml
proxy_groups:
  custom_proxy_group:
    - name: "Final"
      type: "select"
      rule: ["[]DIRECT", "[]REJECT", ".*"]
```

---

## Rulesets

### Configuration

```yaml
rulesets:
  enabled: true
  overwrite_original_rules: false         # Replace base rules
  update_ruleset_on_request: false        # Fetch on each request (vs cache)
  rulesets:
    # Remote ruleset
    - ruleset: "https://example.com/rules.list"
      group: "Auto"
    
    # Local file (relative to base/)
    - ruleset: "rules/custom.list"
      group: "Proxy"
    
    # Inline rule
    - rule: "DOMAIN-SUFFIX,google.com,Proxy"
      group: "Proxy"
    
    # Final rule
    - rule: "MATCH,DIRECT"
      group: "DIRECT"
```

### Ruleset Formats

**Clash format:**
```
DOMAIN-SUFFIX,google.com
DOMAIN-KEYWORD,youtube
IP-CIDR,192.168.0.0/16
GEOIP,CN
```

**Surge format:**
```
DOMAIN-SUFFIX,google.com
DOMAIN-KEYWORD,youtube
IP-CIDR,192.168.0.0/16,no-resolve
GEOIP,CN
```

### Fetching Rulesets

Endpoint: `GET /getruleset?url=<base64_url>&type=<clash|surge>&token=<token>`

Requires the API access token; like the other protected endpoints it fails closed when no token is configured.

**Example:**
```bash
# Encode URL
URL_BASE64=$(echo -n "https://example.com/rules.list" | base64)

# Fetch ruleset
curl "http://localhost:25500/getruleset?url=$URL_BASE64&type=clash&token=<your-token>"
```

---

## Templates

### Go Template Rendering

Base templates support Go `text/template` syntax when `template_path` is set.

**Template file (`base/base/my_template.tpl`):**
```yaml
port: {{ .clash.http_port }}
socks-port: {{ .clash.socks_port }}
allow-lan: {{ .clash.allow_lan }}
mode: {{ .clash.mode }}

proxies:
{{ range .proxies }}
  - name: {{ .name }}
    type: {{ .type }}
    server: {{ .server }}
    port: {{ .port }}
{{ end }}
```

**Configuration:**
```yaml
template:
  template_path: "base/base"
  globals:
    - key: "clash.http_port"
      value: 7890
    - key: "clash.mode"
      value: "rule"
```

**Render endpoint:**
```bash
curl "http://localhost:25500/render?path=/base/base/my_template.tpl&token=<your-token>"
```

### Available Template Variables

- `.clash.*` - Globals from `template.globals` with `clash.` prefix
- `.proxies` - Array of proxy objects
- `.groups` - Array of proxy group objects
- `.rules` - Array of rule strings
- `.request.*` - Query parameters from request

---

## Advanced Features

### Remark Processing

**Rename rules (match@replace):**

```yaml
node_pref:
  rename_node:
    - "Hong Kong@HK"                      # Simple replacement
    - "!!TYPE=SS!!(.*)@SS-$1"             # Type filter + regex
```

**Emoji rules (match,emoji):**

```yaml
emojis:
  add_emoji: true
  remove_old_emoji: true
  rule:
    - match: "/HK|Hong Kong|🇭🇰/"
      emoji: "🇭🇰"
    - match: "/US|United States|🇺🇸/"
      emoji: "🇺🇸"
```

**Append proxy type:**

```yaml
common:
  append_proxy_type: true                 # "HK Node" → "HK Node [ss]"
```

**Sorting:**

```yaml
node_pref:
  sort_flag: true                         # Alphabetical sort
```

### External Config Loading

Merge external configuration via `config` parameter:

```bash
curl "http://localhost:25500/sub?target=clash&url=...&config=https://example.com/config.yml"
```

**External config format:**
```yaml
proxy_groups:
  custom_proxy_group:
    - name: "Custom Group"
      type: "select"
      rule: [".*"]

rulesets:
  rulesets:
    - ruleset: "https://example.com/custom.list"
      group: "Proxy"
```

Supports: YAML, TOML, INI from HTTP URLs or local files.

### Insert URLs

Automatically combine multiple subscriptions:

```yaml
common:
  enable_insert: true
  insert_url:
    - "https://example.com/extra1.txt"
    - "https://example.com/extra2.txt"
  prepend_insert_url: false               # false = append, true = prepend
```

Result: Proxies from `url` param + insert URLs combined.

### Profiles

Store preset configurations in `base/profiles/<name>.ini`:

**Profile file (`base/profiles/my_profile.ini`):**
```ini
[Profile]
target=clash
url=https://example.com/sub
include=HK|US
exclude=expired
```

**Usage:**
```bash
curl "http://localhost:25500/getprofile?name=my_profile&token=<your-token>"
```

Query parameters override profile settings, except `url`: subscription URLs passed via `&url=` are **merged** with (appended to) the profile's URLs instead of replacing them. Multiple profiles can be combined with `name=a|b`, which merges their `url`, `rename`, `include`, and `exclude` values.

### API Security

Protected endpoints require a token:

```yaml
common:
  api_access_token: "my_secret_token"
```

**Protected endpoints:**
- `/readconf?token=<token>` - Reload config
- `/render?path=...&token=<token>` - Render template
- `/getprofile?name=...&token=<token>` - Load profile (a single profile's own `profile_token` is also accepted)
- `/getruleset?url=...&type=...&token=<token>` - Fetch ruleset
- `/flushcache?token=<token>` - Flush cache

Token behavior is fail closed: when `api_access_token` is empty, every protected endpoint refuses all requests (403). Token comparison is constant-time. Additionally, startup is refused when `server.listen` is a non-loopback address and the token is empty or still the legacy shipped default (`password`) — bind to `127.0.0.1` or set a strong token (env override: `API_TOKEN`).

---

## Query Parameter Reference

### `/sub` Endpoint

| Parameter | Type | Description | Example |
|-----------|------|-------------|---------|
| `target` | string | Output format | `clash`, `clashr`, `surge`, `quanx`, `loon`, `singbox`, `ss`, `ssr`, `v2ray`, `trojan`, `mixed`, `auto` |
| `url` | string | Subscription URL(s) | `https://example.com/sub` (pipe-separated for multiple) |
| `config` | string | External config URL/path | `https://example.com/config.yml` |
| `include` | string | Include filter | `HK` or `/^HK-/` |
| `exclude` | string | Exclude filter | `expired` or `/x\\d+/` |
| `udp` | bool | Enable UDP | `true`, `false` |
| `tfo` | bool | Enable TCP Fast Open | `true`, `false` |
| `scv` | bool | Skip certificate verification | `true`, `false` |
| `ignore_source` | bool | Ignore source proxy groups/rules | `true`, `false` |
| `group` | string | Group name for SSD/SSR | - |
| `ua` | string | Override User-Agent | - |
| `new_name` | bool | Use new Clash field names | `true`, `false` |
| `ver` | int | Surge version number | `3`, `4` |

**Example:**
```bash
curl "http://localhost:25500/sub?target=clash&url=https://example.com/sub&include=/^HK/&udp=true&emoji=true"
```

---

## Tips and Best Practices

### Performance Optimization

1. **Disable on-demand fetching:**
   ```yaml
   rulesets:
     update_ruleset_on_request: false
   ```

2. **Use local rulesets:**
   ```yaml
   rulesets:
     rulesets:
       - ruleset: "rules/local.list"     # Faster than remote
         group: "Proxy"
   ```

3. **Reduce logging:**
   ```yaml
   advanced:
     log_level: "warn"                   # info/warn/error
   ```

### Security

1. **Enable API token:**
   ```yaml
   common:
     api_access_token: "strong_random_token"
   ```

2. **Restrict proxies:**
   ```yaml
   common:
     proxy_subscription: "NONE"
     proxy_config: "NONE"
     proxy_ruleset: "NONE"
   ```

3. **Validate external configs:** Ensure `config` parameter only accepts trusted sources.

### Debugging

1. **Enable debug logging:**
   ```yaml
   advanced:
     log_level: "debug"
   ```

2. **Test filtering:**
   ```bash
   # Check which proxies match filter
   curl "http://localhost:25500/sub?target=clash&url=...&include=/^HK/" | grep "name:"
   ```

3. **Validate config:**
   ```bash
   # Reload and check logs
   curl "http://localhost:25500/readconf?token=<your-token>"
   ```

---

## See Also

- [Development Guide](GUIDE.md) - Building, testing, and development workflow
- [Feature Parity](FEATURE_PARITY.md) - Feature parity with C++ version
- [README](../README.md) - Quick start and overview
