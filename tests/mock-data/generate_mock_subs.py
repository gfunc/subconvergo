import base64
import json
import urllib.parse

def b64_encode(s):
    return base64.urlsafe_b64encode(s.encode()).decode().rstrip('=')

def generate_ss():
    # ss://user:pass@host:port
    userinfo = b64_encode("aes-256-gcm:password123")
    return f"ss://{userinfo}@example.com:8388#SS-Basic"

def generate_ssr():
    # ssr://host:port:protocol:method:obfs:password_b64/?params_b64
    host = "example.com"
    port = "8388"
    protocol = "origin"
    method = "aes-256-cfb"
    obfs = "plain"
    password = b64_encode("password123")
    base = f"{host}:{port}:{protocol}:{method}:{obfs}:{password}"
    
    params = {
        "remarks": b64_encode("SSR-Basic"),
        "protoparam": "",
        "obfsparam": ""
    }
    query = "&".join([f"{k}={v}" for k, v in params.items()])
    return "ssr://" + b64_encode(f"{base}/?{query}")

def generate_vmess():
    config = {
        "v": "2",
        "ps": "VMess-WS-TLS",
        "add": "example.com",
        "port": "443",
        "id": "23ad6b10-8d1a-40f7-8ad0-e3e35cd38297",
        "aid": "0",
        "scy": "auto",
        "net": "ws",
        "type": "none",
        "host": "example.com",
        "path": "/ws",
        "tls": "tls",
        "sni": "example.com",
        "alpn": ""
    }
    return "vmess://" + b64_encode(json.dumps(config))

def generate_trojan():
    return "trojan://password123@example.com:443?sni=example.com&allowInsecure=0#Trojan-Basic"

def generate_vless():
    return "vless://23ad6b10-8d1a-40f7-8ad0-e3e35cd38297@example.com:443?security=tls&type=ws&path=/ws&sni=example.com#VLESS-WS-TLS"

def generate_hysteria2():
    return "hysteria2://password123@example.com:443?sni=example.com&obfs=salamander&obfs-password=secret#Hysteria2-Basic"

def generate_tuic():
    return "tuic://23ad6b10-8d1a-40f7-8ad0-e3e35cd38297:password123@example.com:443?congestion_control=bbr&alpn=h3&sni=example.com#TUIC-Basic"

def generate_anytls():
    return "anytls://password123@example.com:443?sni=example.com&alpn=h2#AnyTLS-Basic"

def generate_ssd():
    config = {
        "airport": "SSD-Airport",
        "port": 8388,
        "encryption": "aes-256-gcm",
        "password": "password123",
        "servers": [
            {
                "server": "example.com",
                "remarks": "SSD-Node-1"
            },
            {
                "server": "example2.com",
                "port": 8389,
                "encryption": "chacha20-ietf-poly1305",
                "password": "password456",
                "remarks": "SSD-Node-2"
            }
        ]
    }
    return "ssd://" + b64_encode(json.dumps(config))

def generate_ss_android():
    config = [
        {
            "server": "example.com",
            "server_port": 8388,
            "password": "password123",
            "method": "aes-256-gcm",
            "remarks": "SS-Android-1",
            "plugin": "",
            "plugin_opts": ""
        },
        {
            "server": "example2.com",
            "server_port": 8389,
            "password": "password456",
            "method": "chacha20-ietf-poly1305",
            "remarks": "SS-Android-2",
            "plugin": "obfs-local",
            "plugin_opts": "obfs=http;obfs-host=example.com"
        }
    ]
    return json.dumps(config)

proxies = [
    generate_ss(),
    generate_ssr(),
    generate_vmess(),
    generate_trojan(),
    generate_vless(),
    generate_hysteria2(),
    generate_tuic(),
    generate_anytls()
]

content = "\n".join(proxies)
encoded_content = base64.b64encode(content.encode()).decode()

with open("mixed-subscription.txt", "w") as f:
    f.write(encoded_content)

print("Generated mixed-subscription.txt")

# Generate individual files for specific parsers if needed
with open("ss-subscription.txt", "w") as f:
    f.write(base64.b64encode(generate_ss().encode()).decode())

with open("ssr-subscription.txt", "w") as f:
    f.write(base64.b64encode(generate_ssr().encode()).decode())

with open("v2ray-subscription.txt", "w") as f:
    # V2Ray subscription usually contains vmess/vless/trojan/ss
    v2ray_proxies = [generate_vmess(), generate_vless(), generate_trojan(), generate_ss()]
    f.write(base64.b64encode("\n".join(v2ray_proxies).encode()).decode())

with open("ssd-subscription.txt", "w") as f:
    f.write(generate_ssd())

with open("ss-android-subscription.json", "w") as f:
    f.write(generate_ss_android())

print("Generated individual subscription files")
