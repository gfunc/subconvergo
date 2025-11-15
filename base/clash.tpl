mixed-port: {{default .global.clash.mixed_port "10808"}}
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