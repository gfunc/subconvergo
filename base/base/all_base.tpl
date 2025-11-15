{{if or (eq .request.target "clash") (eq .request.target "clashr")}}
port: {{default .global.clash.http_port "7890"}}
socks-port: {{default .global.clash.socks_port "7891"}}
allow-lan: {{default .global.clash.allow_lan "true"}}
mode: Rule
log-level: {{default .global.clash.log_level "info"}}
external-controller: {{default .global.clash.external_controller "127.0.0.1:9090"}}
{{if eq (default .request.clash.dns "") "1"}}
dns:
  enable: true
  listen: :1053
{{end}}
{{if eq .local.clash.new_field_name "true"}}
proxies: ~
proxy-groups: ~
rules: ~
{{else}}
Proxy: ~
Proxy Group: ~
Rule: ~
{{end}}
{{end}}

{{if eq .request.target "surge"}}
[General]
loglevel = notify
bypass-system = true
skip-proxy = 127.0.0.1,192.168.0.0/16,10.0.0.0/8,172.16.0.0/12,100.64.0.0/10,localhost,*.local,e.crashlytics.com,captive.apple.com,::ffff:0:0:0:0/1,::ffff:128:0:0:0/1
#DNS设置或根据自己网络情况进行相应设置
bypass-tun = 192.168.0.0/16,10.0.0.0/8,172.16.0.0/12
dns-server = 119.29.29.29,223.5.5.5

[Script]
http-request https?:\/\/.*\.iqiyi\.com\/.*authcookie= script-path=https://raw.githubusercontent.com/NobyDa/Script/master/iQIYI-DailyBonus/iQIYI.js
{{end}}

{{if eq .request.target "loon"}}
[General]
ipv6 = false
dns-server = 119.29.29.29, 223.5.5.5
doh-server = https://223.5.5.5/resolve, https://sm2.doh.pub/dns-query
allow-wifi-access = false
wifi-access-http-port = 7222
wifi-access-socks5-port = 7221
proxy-test-url = http://connectivitycheck.gstatic.com
test-timeout = 2
interface-mode = auto
sni-sniffing = true
disable-stun = true
disconnect-on-policy-change = true
switch-node-after-failure-times = 3
skip-proxy = 192.168.0.0/16, 10.0.0.0/8, 172.16.0.0/12, localhost, *.local
bypass-tun = 10.0.0.0/8, 100.64.0.0/10, 127.0.0.0/8, 169.254.0.0/16
{{end}}

{{if eq .request.target "quanx"}}
[general]
excluded_routes=192.168.0.0/16, 172.16.0.0/12, 100.64.0.0/10, 10.0.0.0/8
geo_location_checker=http://ip-api.com/json/?lang=zh-CN, https://github.com/KOP-XIAO/QuantumultX/raw/master/Scripts/IP_API.js
network_check_url=http://www.baidu.com/
server_check_url=http://www.gstatic.com/generate_204

[dns]
server=119.29.29.29
server=223.5.5.5
server=1.0.0.1
server=8.8.8.8
{{end}}

{{if eq .request.target "singbox"}}
{
    "log": {"disabled": false, "level": "info", "timestamp": true},
    "dns": {
        "servers": [
            {"tag": "dns_proxy", "address": "tls://1.1.1.1", "address_resolver": "dns_resolver"},
            {"tag": "dns_direct", "address": "h3://dns.alidns.com/dns-query", "address_resolver": "dns_resolver", "detour": "DIRECT"},
            {"tag": "dns_fakeip", "address": "fakeip"},
            {"tag": "dns_resolver", "address": "223.5.5.5", "detour": "DIRECT"},
            {"tag": "block", "address": "rcode://success"}
        ],
        "fakeip": {
            "enabled": true,
            {{if eq (default .request.singbox.ipv6 "") "1"}}"inet6_range": "fc00::/18",{{end}}
            "inet4_range": "198.18.0.0/15"
        }
    },
    "inbounds": [
        {
            "type": "mixed",
            "tag": "mixed-in",
            {{if toBool (default .global.singbox.allow_lan "")}}"listen": "0.0.0.0",{{else}}"listen": "127.0.0.1",{{end}}
            "listen_port": {{default .global.singbox.mixed_port "2080"}}
        }
    ],
    "outbounds": [],
    "route": {"rules": [], "auto_detect_interface": true}
}
{{end}}
