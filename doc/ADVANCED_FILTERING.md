# Advanced Proxy Filtering and Ruleset Features

This document describes the advanced filtering and ruleset features implemented in subconvergo, matching the functionality of the C++ subconverter.

## Proxy Filtering

### Basic Regex Filtering

Filter proxies by their remark (name) using regular expressions:

```ini
[proxy_groups]
custom_proxy_group=US Nodes`select`.*US.*
custom_proxy_group=HK Nodes`select`.*HK.*|.*Hong Kong.*
```

### Special Matchers

Subconvergo supports advanced matchers that allow filtering by proxy attributes beyond just the name:

#### 1. Type Matcher: `!!TYPE=`

Filter proxies by protocol type:

```ini
# Select only Shadowsocks proxies
custom_proxy_group=SS Only`select`!!TYPE=SS

# Select VMess or Trojan proxies
custom_proxy_group=V2Ray`select`!!TYPE=VMESS|TROJAN

# Combine type filter with regex on name
custom_proxy_group=US SS`select`!!TYPE=SS!!.*US.*
```

Supported types: `SS`, `SSR`, `VMESS`, `TROJAN`, `VLESS`, `HYSTERIA`, `HYSTERIA2`, `TUIC`, `SNELL`, `HTTP`, `HTTPS`, `SOCKS5`, `WIREGUARD`

#### 2. Group Matcher: `!!GROUP=`

Filter proxies by their subscription group:

```ini
# Select proxies from groups matching "Premium"
custom_proxy_group=Premium Nodes`select`!!GROUP=Premium

# Combine with regex on name
custom_proxy_group=Premium US`select`!!GROUP=Premium!!.*US.*
```

#### 3. Port Matcher: `!!PORT=`

Filter proxies by port number or range:

```ini
# Only port 443 proxies
custom_proxy_group=HTTPS Proxies`select`!!PORT=443

# Port range 8000-9000
custom_proxy_group=High Ports`select`!!PORT=8000-9000

# Multiple ports
custom_proxy_group=Common Ports`select`!!PORT=443,8080,8443

# Complex patterns
custom_proxy_group=Mixed Ports`select`!!PORT=1-100,443,8000-9000
```

#### 4. Server Matcher: `!!SERVER=`

Filter proxies by server address using regex:

```ini
# Servers in example.com domain
custom_proxy_group=Example Servers`select`!!SERVER=.*\\.example\\.com

# Specific server patterns
custom_proxy_group=US Servers`select`!!SERVER=us[0-9]+\\..*
```

#### 5. Group ID Matcher: `!!GROUPID=`

Filter by subscription group ID (for advanced use):

```ini
# Group ID 0
custom_proxy_group=Group Zero`select`!!GROUPID=0

# Group ID range
custom_proxy_group=Groups 1-5`select`!!GROUPID=1-5
```

### Direct Node Inclusion

Use `[]` prefix to include special nodes like DIRECT or REJECT:

```ini
custom_proxy_group=Final`select`[]DIRECT`[]REJECT`.*
```

### Combining Matchers

Matchers can be combined with `!!` separator. The format is:

```
!!MATCHER=pattern!!regex_for_name
```

Examples:

```ini
# US Shadowsocks proxies on port 443
custom_proxy_group=US SS 443`select`!!TYPE=SS!!PORT=443!!.*US.*

# Premium group VMess proxies
custom_proxy_group=Premium VMess`select`!!GROUP=Premium!!TYPE=VMESS
```

## Ruleset Features

### Ruleset Formats

Subconvergo supports multiple ruleset formats:

#### 1. Clash Payload Format

```yaml
payload:
  - 'example.com'
  - '.google.com'
  - '+.facebook.com'
  - '1.1.1.1/32'
  - '2001:db8::/32'
```

Automatic conversion:
- `example.com` → `DOMAIN,example.com`
- `.google.com` → `DOMAIN-SUFFIX,google.com`
- `+.facebook.com` → `DOMAIN-SUFFIX,facebook.com`
- `.example.com.*` → `DOMAIN-KEYWORD,example.com`
- `1.1.1.1/32` → `IP-CIDR,1.1.1.1/32`
- `2001:db8::/32` → `IP-CIDR6,2001:db8::/32`

#### 2. Surge Format

```
DOMAIN-SUFFIX,google.com
DOMAIN,example.com
IP-CIDR,1.1.1.1/32
DOMAIN-KEYWORD,test
```

Used as-is, already in correct format.

#### 3. QuantumultX Format

```
HOST,example.com,PROXY
HOST-SUFFIX,google.com,PROXY
HOST-KEYWORD,test,PROXY
IP6-CIDR,2001:db8::/32,PROXY
```

Automatic conversion:
- `HOST` → `DOMAIN`
- `HOST-SUFFIX` → `DOMAIN-SUFFIX`
- `HOST-KEYWORD` → `DOMAIN-KEYWORD`
- `IP6-CIDR` → `IP-CIDR6`

### Ruleset Sources

Rulesets can be loaded from:

1. **Remote URLs:**
   ```ini
   [rulesets]
   ruleset=PROXY,https://example.com/rules/proxy.list
   ```

2. **Local files (relative to base path):**
   ```ini
   ruleset=PROXY,rules/proxy.list
   ```

3. **Direct rules:**
   ```ini
   rule=DOMAIN-SUFFIX,google.com,PROXY
   ```

### Ruleset Configuration

```ini
[rulesets]
enabled=true

# Remote ruleset
ruleset=PROXY,https://raw.githubusercontent.com/ACL4SSR/ACL4SSR/master/Clash/ProxyLite.list

# Local ruleset
ruleset=DIRECT,rules/direct.list

# Direct rule
rule=DOMAIN-SUFFIX,cn,DIRECT

# With type specification
ruleset=REJECT,type:http://example.com/ads.list,classical
```

### Rule Processing

1. Rules are fetched from URLs or loaded from local files
2. Format is auto-detected and converted to Clash format
3. Rules are merged with the target group
4. Comments and empty lines are filtered out
5. Final MATCH/FINAL rule is automatically added

## Configuration Examples

### Example 1: Advanced Filtering

```ini
[proxy_groups]
# All proxies
custom_proxy_group=🔰 Select`select`.*

# Only Shadowsocks
custom_proxy_group=⚡ SS`select`!!TYPE=SS

# US nodes on port 443
custom_proxy_group=🇺🇸 US 443`select`!!PORT=443!!.*US.*

# Premium VMess nodes
custom_proxy_group=💎 Premium V2Ray`url-test`!!GROUP=Premium!!TYPE=VMESS`http://www.gstatic.com/generate_204`300

# Final with fallback
custom_proxy_group=🌐 Final`select`[]DIRECT`[]REJECT`🔰 Select
```

### Example 2: Complete Configuration with Rulesets

```ini
[common]
enable_filter=true

[proxy_groups]
custom_proxy_group=🔰 Proxy`select`.*
custom_proxy_group=⚡ Auto`url-test`.*`http://www.gstatic.com/generate_204`300
custom_proxy_group=🎯 Direct`select`[]DIRECT
custom_proxy_group=🛑 Reject`select`[]REJECT

[rulesets]
enabled=true

# Ad blocking
ruleset=🛑 Reject,https://raw.githubusercontent.com/ACL4SSR/ACL4SSR/master/Clash/BanAD.list

# Proxy services
ruleset=🔰 Proxy,https://raw.githubusercontent.com/ACL4SSR/ACL4SSR/master/Clash/ProxyLite.list

# China direct
ruleset=🎯 Direct,https://raw.githubusercontent.com/ACL4SSR/ACL4SSR/master/Clash/ChinaDomain.list
ruleset=🎯 Direct,https://raw.githubusercontent.com/ACL4SSR/ACL4SSR/master/Clash/ChinaCompanyIp.list

# Local rules
ruleset=🎯 Direct,rules/local-direct.list

# Direct rule
rule=GEOIP,CN,🎯 Direct
```

### Example 3: Multi-Region Setup

```ini
[proxy_groups]
# Main selection
custom_proxy_group=🌍 Global`select`🇭🇰 HK`🇺🇸 US`🇯🇵 JP`🇸🇬 SG

# Regional groups with type filtering
custom_proxy_group=🇭🇰 HK`url-test`!!TYPE=SS|VMESS|TROJAN!!.*(HK|Hong Kong|香港).*`http://www.gstatic.com/generate_204`300
custom_proxy_group=🇺🇸 US`url-test`!!TYPE=SS|VMESS|TROJAN!!.*(US|United States|美国).*`http://www.gstatic.com/generate_204`300
custom_proxy_group=🇯🇵 JP`url-test`!!TYPE=SS|VMESS|TROJAN!!.*(JP|Japan|日本).*`http://www.gstatic.com/generate_204`300
custom_proxy_group=🇸🇬 SG`url-test`!!TYPE=SS|VMESS|TROJAN!!.*(SG|Singapore|新加坡).*`http://www.gstatic.com/generate_204`300

# Streaming optimized (port 443 only)
custom_proxy_group=🎬 Streaming`select`!!PORT=443!!.*(HK|US|JP).*

# Gaming optimized (low port numbers, trojan preferred)
custom_proxy_group=🎮 Gaming`url-test`!!TYPE=TROJAN!!PORT=1-1000!!.*`http://www.gstatic.com/generate_204`300,,50
```

## Performance Notes

- Regex matching is cached for performance
- Ruleset fetching supports caching (configure via `cache_ruleset` in advanced settings)
- Large rulesets are automatically optimized
- Proxy filtering is done in a single pass for efficiency

## Compatibility

This implementation matches the C++ subconverter behavior for:
- ✅ All special matchers (TYPE, GROUP, PORT, SERVER, GROUPID)
- ✅ Regex filtering on proxy names
- ✅ Direct node inclusion with `[]` prefix
- ✅ Multiple matcher combinations
- ✅ Clash payload, Surge, and QuantumultX ruleset formats
- ✅ Remote and local ruleset loading
- ✅ Auto-detection and conversion of ruleset formats
