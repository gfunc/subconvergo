port: {{ default .global.clash.http_port "7890" }}
socks-port: {{ default .global.clash.socks_port "7891" }}
allow-lan: {{ default .global.clash.allow_lan "true" }}
mode: Rule
log-level: {{ default .global.clash.log_level "info" }}
external-controller: {{ default .global.clash.external_controller "127.0.0.1:9090" }}
custom-key: "custom-value-from-template"
