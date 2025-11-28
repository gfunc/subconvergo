# API Reference

## Endpoints

### `/sub`
Convert subscription(s) to target format.

**Method:** `GET`

**Parameters:**

| Parameter | Type | Description | Default |
|-----------|------|-------------|---------|
| `target` | string | Output format (see [Target Formats](#target-formats)) | `clash` |
| `url` | string | Subscription URL(s), pipe-separated (`\|`), URL-encoded | - |
| `config` | string | External config URL or file path (URL-encoded) | - |
| `include` | string | Include filter (regex with `/pattern/` or substring) | - |
| `exclude` | string | Exclude filter (regex with `/pattern/` or substring) | - |
| `emoji` | bool | Add emojis to node names | `false` |
| `list` | bool | Return base64 encoded list of nodes (not implemented) | `false` |
| `udp` | bool | Enable UDP for nodes | `false` |
| `tfo` | bool | Enable TCP Fast Open | `false` |
| `scv` | bool | Skip Certificate Verification | `false` |
| `tls13` | bool | Enable TLS 1.3 | `false` |
| `append_type` | bool | Append protocol type to node names | `false` |
| `sort` | bool | Sort nodes alphabetically | `false` |
| `rename` | string | Custom rename rules (URL-encoded) | - |
| `insert` | bool | Insert additional URLs from config | `true` |
| `prepend` | bool | Insert additional URLs before subscription URLs | `false` |
| `group` | string | Set group name for SSD/SSR | - |

**Example:**
```bash
curl "http://localhost:25500/sub?target=clash&url=https%3A%2F%2Fexample.com%2Fsub&udp=true&emoji=true"
```

### `/surge2clash`
Shortcut for converting Surge subscription to Clash format. Equivalent to `/sub?target=clash&url=...`.

**Method:** `GET`

**Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| `url` | string | Surge subscription URL (URL-encoded) |
| ... | ... | Supports all other `/sub` parameters |

### `/version`
Get version info.

**Method:** `GET`

### `/readconf`
Reload configuration.

**Method:** `GET`

**Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| `token` | string | API Access Token (if configured) |

### `/getprofile`
Load preset configuration from `base/profiles/<name>.ini`.

**Method:** `GET`

**Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| `name` | string | Profile name (without .ini extension) |
| `token` | string | API Access Token (if configured) |

### `/getruleset`
Fetch and format ruleset.

**Method:** `GET`

**Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| `url` | string | Base64 encoded ruleset URL |
| `type` | string | Ruleset type (`clash` or `surge`) |

### `/render`
Render Go template with current config.

**Method:** `GET`

**Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| `path` | string | Path to template file (relative to base_path) |
| `token` | string | API Access Token (if configured) |

---

## Source Formats

Subconvergo supports parsing subscriptions in the following formats:

| Format | Description |
|--------|-------------|
| **Base64** | Base64 encoded list of proxy links (standard) |
| **Plain Text** | Line-separated list of proxy links |
| **Clash** | Clash YAML configuration file. **Note:** Subconvergo preserves `proxy-groups` and `rules` from the source. |
| **SIP002** | Shadowsocks URI scheme (`ss://`) |
| **SSR** | ShadowsocksR URI scheme (`ssr://`) |
| **VMess** | V2Ray URI scheme (`vmess://`) |
| **Trojan** | Trojan URI scheme (`trojan://`) |
| **VLESS** | VLESS URI scheme (`vless://`) |
| **Hysteria** | Hysteria URI scheme (`hysteria://`) |
| **Hysteria2** | Hysteria2 URI scheme (`hy2://`, `hysteria2://`) |
| **TUIC** | TUIC URI scheme (`tuic://`) |
| **SSD** | SSD URI scheme (`ssd://`) |
| **Netch** | Netch URI scheme (`Netch://`) |

**Note:** Telegram links (`tg://`, `https://t.me/`) are parsed as single proxies.

---

## Target Formats

Subconvergo supports generating configurations in the following formats:

| Target | Description | Limitations |
|--------|-------------|-------------|
| `clash` | Clash YAML configuration | Full support for all protocols via mihomo |
| `clashr` | Alias for `clash` | - |
| `surge` | Surge INI configuration | VLESS, Hysteria, TUIC may not be supported by standard Surge |
| `quanx` | Quantumult X configuration | - |
| `loon` | Loon configuration | - |
| `singbox` | sing-box JSON configuration | - |
| `ss` | Shadowsocks SIP002 links | Only SS proxies |
| `ssr` | ShadowsocksR links | Only SSR proxies |
| `v2ray` | VMess links | Only VMess/VLESS proxies |
| `trojan` | Trojan links | Only Trojan proxies |
| `mixed` | Mixed list of links | Returns original link format if available |

**Note:** When converting to single link formats (`ss`, `ssr`, `v2ray`, `trojan`), only proxies of that specific type are included in the output. `mixed` target includes all proxies that can be converted to a link format.
