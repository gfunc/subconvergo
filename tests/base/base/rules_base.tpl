port: 7890
socks-port: 7891
allow-lan: true
mode: Rule
log-level: info
external-controller: 127.0.0.1:9090
proxies: ~
proxy-groups: ~
rules:
  - DOMAIN-SUFFIX,google.com,Proxy
  - DOMAIN-KEYWORD,facebook,Proxy
  - GEOIP,CN,DIRECT
  - MATCH,Proxy
