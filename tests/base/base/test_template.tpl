# Test template for smoke render validation
proxies:
  - name: TestProxy
    type: ss
    server: 1.1.1.1
    port: 8388
    cipher: aes-256-gcm
    password: example-password
proxy-groups:
  - name: Auto
    type: select
    proxies:
      - TestProxy
rules:
  - MATCH,Auto
