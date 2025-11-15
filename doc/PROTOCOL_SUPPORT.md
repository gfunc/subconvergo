# Protocol Support - Subconvergo

## Supported Proxy Protocols

### Core Protocols (Explicitly Implemented)

#### 1. Shadowsocks (ss://)
- **Status**: ✅ Fully Supported
- **Features**:
  - New format: `ss://base64(method:password)@server:port`
  - Old format: `ss://base64(method:password@server:port)`
  - IPv6 support with bracket notation
  - Plugin support (simple-obfs, v2ray-plugin)
  - SS 2022 ciphers
- **Validation**: mihomo adapter
- **Performance**: ~7.6µs per proxy

#### 2. ShadowsocksR (ssr://)
- **Status**: ✅ Fully Supported
- **Features**:
  - Full SSR protocol parsing
  - Auto-conversion to SS when applicable
  - Protocol/obfs parameters
  - Custom protocol/obfs params
- **Validation**: mihomo adapter
- **Note**: Automatically converts to SS for plain obfs + origin protocol

#### 3. VMess (vmess://)
- **Status**: ✅ Fully Supported
- **Features**:
  - JSON format parsing
  - All transport types: TCP, WebSocket, HTTP/2, gRPC, QUIC
  - TLS/XTLS support
  - AlterID handling
  - Version 1 compatibility (host;path format)
- **Validation**: mihomo adapter
- **Performance**: ~24.4µs per proxy

#### 4. Trojan (trojan://)
- **Status**: ✅ Fully Supported
- **Features**:
  - Standard trojan protocol
  - WebSocket transport
  - gRPC transport
  - TLS configuration
  - SNI/ALPN support
  - allowInsecure option
- **Validation**: mihomo adapter

#### 5. VLESS (vless://)
- **Status**: ✅ Fully Supported
- **Features**:
  - VLESS protocol parsing
  - Flow control (xtls-rprx-direct, xtls-rprx-vision)
  - All transport types: TCP, WebSocket, gRPC, HTTP/2
  - TLS/Reality support
  - Custom SNI
- **Validation**: mihomo adapter

### Modern Protocols (Mihomo-based)

#### 6. Hysteria (hysteria://)
- **Status**: ✅ Fully Supported (v1.19.16+)
- **Prefixes**: `hysteria://`
- **Features**:
  - Bandwidth configuration (up/down)
  - Obfuscation support
  - ALPN/SNI configuration
  - Auth string support
  - Default bandwidth: 10 Mbps up, 50 Mbps down
- **Validation**: mihomo adapter
- **Performance**: ~10.9µs per proxy

#### 7. Hysteria2 (hysteria2://, hy2://)
- **Status**: ✅ Fully Supported (v1.19.16+)
- **Prefixes**: `hysteria2://`, `hy2://`
- **Features**:
  - Password authentication
  - Salamander obfuscation
  - ALPN/SNI configuration
  - Obfs password support
- **Validation**: mihomo adapter
- **Performance**: ~10.9µs per proxy

#### 8. TUIC (tuic://)
- **Status**: ✅ Fully Supported (v1.19.16+)
- **Features**:
  - UUID + password authentication
  - Congestion control (BBR, Cubic)
  - UDP relay modes
  - ALPN configuration
  - SNI support
- **Validation**: mihomo adapter
- **Performance**: ~16.1µs per proxy

### Additional Formats

#### 9. Clash YAML
- **Status**: ✅ Fully Supported
- **Features**:
  - Full Clash configuration parsing
  - Supports all proxy types in Clash format
  - mihomo's native parser
- **Validation**: mihomo config parser

## Fallback Mechanism

For any protocol not explicitly implemented, subconvergo attempts to parse it using mihomo's built-in parsers. This provides forward compatibility with new protocols that mihomo adds support for.

### Supported via Fallback:
- Any protocol that mihomo v1.19.16+ supports
- WireGuard (if mihomo adds support)
- Future protocols added to mihomo

### How It Works:
1. Check if protocol has explicit parser (ss, ssr, vmess, trojan, vless)
2. If not, extract protocol prefix (e.g., `hysteria://`)
3. Route to appropriate mihomo-based parser
4. If unknown, return unsupported error

## Protocol Detection

The parser automatically detects protocols based on URL scheme:

```
ss://...       → parseShadowsocks()
ssr://...      → parseShadowsocksR()
vmess://...    → parseVMess()
trojan://...   → parseTrojan()
vless://...    → parseVLESS()
hysteria://... → parseHysteria() [fallback]
hysteria2://.. → parseHysteria() [fallback]
hy2://...      → parseHysteria() [fallback]
tuic://...     → parseTUIC() [fallback]
other://...    → parseFallbackProtocol() → error if unsupported
```

## Usage Examples

### Shadowsocks
```
ss://YWVzLTI1Ni1nY206cGFzc3dvcmQ=@1.2.3.4:8388#MyProxy
ss://YWVzLTI1Ni1nY206cGFzc3dvcmRAZXhhbXBsZS5jb206ODM4OA==#OldFormat
ss://[2001:db8::1]:8388#IPv6
```

### VMess
```
vmess://base64({"v":"2","ps":"Name","add":"example.com","port":"443",...})
```

### Trojan
```
trojan://password@example.com:443?sni=example.com#MyTrojan
trojan://password@example.com:443?type=ws&path=/path#TrojanWS
```

### VLESS
```
vless://uuid@example.com:443?type=ws&security=tls&sni=example.com#MyVLESS
```

### Hysteria
```
hysteria://example.com:443?auth=password&peer=example.com&upmbps=100&downmbps=100#Hysteria
hysteria2://password@example.com:443?sni=example.com#Hysteria2
hy2://password@example.com:8443?obfs=salamander&obfs-password=pass#HY2
```

### TUIC
```
tuic://uuid:password@example.com:443?sni=example.com&alpn=h3&congestion_control=bbr#TUIC
```

## Testing Coverage

All protocols have comprehensive test coverage:

| Protocol | Test Cases | Coverage | Status |
|----------|-----------|----------|--------|
| Shadowsocks | 10 | ✅ | Complete |
| ShadowsocksR | 5 | ✅ | Complete |
| VMess | 9 | ✅ | Complete |
| Trojan | 8 | ✅ | Complete |
| VLESS | 6 | ✅ | Complete |
| Hysteria | 4 | ✅ | Complete |
| TUIC | 3 | ✅ | Complete |
| Fallback | 5 | ✅ | Complete |

**Overall**: 50+ test cases, 81.8% code coverage

## Performance Comparison

Based on benchmarks (AMD Ryzen 7 5800H):

| Protocol | Time/op | Memory/op | Allocs/op |
|----------|---------|-----------|-----------|
| Shadowsocks | 7.6µs | 4.7 KB | 57 |
| VMess | 24.4µs | 13.2 KB | 166 |
| Hysteria | 10.9µs | 6.7 KB | 63 |
| TUIC | 16.1µs | 6.7 KB | 69 |

All parsers are highly optimized for production use.

## Adding New Protocols

To add support for a new protocol:

### Option 1: Explicit Implementation
1. Create `parseNewProtocol(line string) (Proxy, error)` function
2. Add protocol detection in `parseProxyLine()`
3. Build mihomo config map
4. Validate with `adapter.ParseProxy()`
5. Add tests in `parser_test.go`

### Option 2: Mihomo Fallback
1. Add protocol case in `parseFallbackProtocol()`
2. Create parser function following existing patterns
3. Use mihomo's validation
4. Add tests

Example:
```go
func parseNewProtocol(line string) (Proxy, error) {
    // Extract fields from URL
    // Build mihomo config
    mihomoConfig := map[string]interface{}{
        "type": "newprotocol",
        "name": remark,
        // ... other fields
    }
    
    // Validate
    mihomoProxy, err := adapter.ParseProxy(mihomoConfig)
    if err != nil {
        return Proxy{}, fmt.Errorf("validation failed: %w", err)
    }
    
    return Proxy{
        Type: "newprotocol",
        // ... populate fields
        MihomoProxy: mihomoProxy,
    }, nil
}
```

## Dependencies

- **mihomo v1.19.16+**: Core validation and parsing
- **Go 1.25.3+**: Language runtime

## Validation

All parsed proxies are validated using mihomo's `adapter.ParseProxy()` to ensure:
- ✅ Correct configuration format
- ✅ Valid cipher/encryption methods
- ✅ Proper transport options
- ✅ Compatible with mihomo core
- ✅ Ready for conversion to any output format

## Compatibility

The parser maintains compatibility with:
- ✅ subconverter C++ implementation
- ✅ mihomo/Clash core
- ✅ Standard proxy share link formats
- ✅ V2Ray/Xray ecosystems

## Future Enhancements

Potential additions:
- [ ] WireGuard support (when mihomo adds it)
- [ ] HTTP/HTTPS proxies
- [ ] SOCKS5 proxies
- [ ] Custom protocol plugins
- [ ] Protocol auto-detection without scheme

## Notes

1. All parsers use mihomo for validation, ensuring compatibility
2. Fallback mechanism provides extensibility
3. Performance is optimized for high-throughput scenarios
4. IPv6 fully supported across all protocols
5. TLS/encryption options preserved during parsing
