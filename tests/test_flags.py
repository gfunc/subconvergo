import yaml
import base64
from . import infra

def setup_default_flags(pref):
    # Ensure flags are not set in pref, relying on defaults
    if "node_pref" in pref:
        pref["node_pref"].pop("udp_flag", None)
        pref["node_pref"].pop("tcp_fast_open_flag", None)
        pref["node_pref"].pop("skip_cert_verify_flag", None)
        pref["node_pref"].pop("tls13_flag", None)

def validate_default_flags(resp):
    content = resp.text
    failures = []
    
    if "udp: true" in content:
        failures.append("UDP flag unexpectedly true")
    if "tfo: true" in content:
        failures.append("TFO flag unexpectedly true")
    if "skip-cert-verify: true" in content:
        failures.append("SCV flag unexpectedly true")
        
    if failures:
        return {"_failures": failures}
    return {"status": "ok"}

def setup_explicit_flags(pref):
    if "node_pref" not in pref:
        pref["node_pref"] = {}
    pref["node_pref"]["udp_flag"] = True
    pref["node_pref"]["tcp_fast_open_flag"] = True
    pref["node_pref"]["skip_cert_verify_flag"] = True

def validate_explicit_flags(resp):
    content = resp.text
    failures = []
    
    if "udp: true" not in content:
        failures.append("UDP flag missing or false")
    if "tfo: true" not in content:
        failures.append("TFO flag missing or false")
    if "skip-cert-verify: true" not in content:
        failures.append("SCV flag missing or false")
        
    if failures:
        return {"_failures": failures}
    return {"status": "ok"}

def validate_query_param_flags(resp):
    """Validate that URL params override config flags"""
    content = resp.text
    failures = []
    
    if "udp: true" not in content:
        failures.append("UDP flag missing (should be set by URL param)")
    if "tfo: true" not in content:
        failures.append("TFO flag missing (should be set by URL param)")
    if "skip-cert-verify: true" not in content:
        failures.append("SCV flag missing (should be set by URL param)")
        
    if failures:
        return {"_failures": failures}
    return {"status": "ok"}

def validate_clash_source_flags(resp):
    """Validate flags work for clash-format source"""
    data = yaml.safe_load(resp.text)
    proxies = data.get("proxies", [])
    failures = []
    
    if not proxies:
        failures.append("No proxies found")
    else:
        proxy = proxies[0]
        if proxy.get("udp") != True:
            failures.append(f"UDP flag not set: {proxy.get('udp')}")
        if proxy.get("tfo") != True:
            failures.append(f"TFO flag not set: {proxy.get('tfo')}")
        if proxy.get("skip-cert-verify") != True:
            failures.append(f"SCV flag not set: {proxy.get('skip-cert-verify')}")
    
    if failures:
        return {"_failures": failures}
    return {"status": "ok", "proxy_count": len(proxies)}

def validate_surge_format_flags(resp):
    """Validate flags appear in Surge format output"""
    content = resp.text
    failures = []
    
    if "udp-relay=true" not in content and "udp=true" not in content:
        failures.append("UDP flag missing in Surge format")
    if "tfo=true" not in content:
        failures.append("TFO flag missing in Surge format")
    if "skip-cert-verify=true" not in content:
        failures.append("SCV flag missing in Surge format")
    
    if failures:
        return {"_failures": failures}
    return {"status": "ok"}

def validate_loon_format_flags(resp):
    """Validate flags appear in Loon format output"""
    content = resp.text
    failures = []
    
    # Loon may use udp=true or udp-relay=true
    has_udp = "udp=true" in content or "udp-relay=true" in content
    if not has_udp:
        failures.append("UDP flag missing in Loon format")
    if "tfo=true" not in content:
        failures.append("TFO flag missing in Loon format")
    if "skip-cert-verify=true" not in content:
        failures.append("SCV flag missing in Loon format")
    
    if failures:
        return {"_failures": failures}
    return {"status": "ok"}

def validate_quanx_format_flags(resp):
    """Validate flags appear in QuantumultX format output"""
    content = resp.text
    failures = []
    
    if "udp-relay=true" not in content:
        failures.append("UDP flag missing in QuantumultX format")
    if "fast-open=true" not in content:
        failures.append("TFO flag missing in QuantumultX format")
    
    if failures:
        return {"_failures": failures}
    return {"status": "ok"}

# Test cases to be added to test_standalone.CASES
FLAG_TEST_CASES = [
    infra.StandaloneTestCase(
        name="flags_url_params_override",
        query=lambda: infra.api_get_subconvergo("/sub", params={
            "target": "clash",
            "url": "ss://YWVzLTI1Ni1nY206cGFzc3dvcmQ=@127.0.0.1:8388%23Example",
            "udp": "true",
            "tfo": "true",
            "scv": "true",
        }),
        validate=validate_query_param_flags,
        pref_modifier=setup_default_flags
    ),
    infra.StandaloneTestCase(
        name="flags_clash_source",
        query=lambda: infra.api_get_subconvergo("/sub", params={
            "target": "clash",
            "url": f"{infra.MOCK_BASE}/clash_only.yaml",
        }),
        validate=validate_clash_source_flags,
        pref_modifier=setup_explicit_flags
    ),
    infra.StandaloneTestCase(
        name="flags_vmess_source",
        query=lambda: infra.api_get_subconvergo("/sub", params={
            "target": "clash",
            "url": "vmess://eyJ2IjogIjIiLCAicHMiOiAiVk1lc3NUZXN0IiwgImFkZCI6ICIxMjcuMC4wLjEiLCAicG9ydCI6ICI4MCIsICJpZCI6ICIxMjM0NTY3OC0xMjM0LTEyMzQtMTIzNC0xMjM0NTY3ODkwYWIiLCAiYWlkIjogIjAiLCAibmV0IjogIndzIiwgInR5cGUiOiAibm9uZSIsICJob3N0IjogIiIsICJwYXRoIjogIi8iLCAidGxzIjogIiJ9",
        }),
        validate=validate_explicit_flags,
        pref_modifier=setup_explicit_flags
    ),
    infra.StandaloneTestCase(
        name="flags_trojan_source",
        query=lambda: infra.api_get_subconvergo("/sub", params={
            "target": "clash",
            "url": "trojan://password@127.0.0.1:443?sni=example.com%23TrojanTest",
        }),
        validate=validate_explicit_flags,
        pref_modifier=setup_explicit_flags
    ),
    infra.StandaloneTestCase(
        name="flags_surge_target",
        query=lambda: infra.api_get_subconvergo("/sub", params={
            "target": "surge",
            "url": "ss://YWVzLTI1Ni1nY206cGFzc3dvcmQ=@127.0.0.1:8388%23Example",
            "ver": "4",
        }),
        validate=validate_surge_format_flags,
        pref_modifier=setup_explicit_flags
    ),
    infra.StandaloneTestCase(
        name="flags_loon_target",
        query=lambda: infra.api_get_subconvergo("/sub", params={
            "target": "loon",
            "url": "ss://YWVzLTI1Ni1nY206cGFzc3dvcmQ=@127.0.0.1:8388%23Example",
        }),
        validate=validate_loon_format_flags,
        pref_modifier=setup_explicit_flags
    ),
    infra.StandaloneTestCase(
        name="flags_quanx_target",
        query=lambda: infra.api_get_subconvergo("/sub", params={
            "target": "quan",
            "url": "trojan://password@127.0.0.1:443?sni=example.com%23TrojanTest",
        }),
        validate=validate_quanx_format_flags,
        pref_modifier=setup_explicit_flags
    ),
    infra.StandaloneTestCase(
        name="flags_mixed_sources",
        query=lambda: infra.api_get_subconvergo("/sub", params={
            "target": "clash",
            "url": "|".join([
                "ss://YWVzLTI1Ni1nY206cGFzc3dvcmQ=@127.0.0.1:8388%23SS-Test",
                "vmess://eyJ2IjogIjIiLCAicHMiOiAiVk1lc3NUZXN0IiwgImFkZCI6ICIxMjcuMC4wLjEiLCAicG9ydCI6ICI4MCIsICJpZCI6ICIxMjM0NTY3OC0xMjM0LTEyMzQtMTIzNC0xMjM0NTY3ODkwYWIiLCAiYWlkIjogIjAiLCAibmV0IjogIndzIiwgInR5cGUiOiAibm9uZSIsICJob3N0IjogIiIsICJwYXRoIjogIi8iLCAidGxzIjogIiJ9",
                "trojan://password@127.0.0.1:443?sni=example.com%23Trojan-Test",
            ]),
        }),
        validate=lambda resp: (
            lambda data=yaml.safe_load(resp.text): (
                lambda proxies=data.get("proxies", []): (
                    lambda failures=[]: (
                        failures.extend(["No proxies found"]) if len(proxies) == 0 else None,
                        [failures.extend([f"Proxy {p.get('name')} missing UDP"]) for p in proxies if p.get("udp") != True],
                        [failures.extend([f"Proxy {p.get('name')} missing TFO"]) for p in proxies if p.get("tfo") != True],
                        [failures.extend([f"Proxy {p.get('name')} missing SCV"]) for p in proxies if p.get("skip-cert-verify") != True],
                        {"_failures": failures} if failures else {"status": "ok", "proxy_count": len(proxies)}
                    )[-1]
                )()
            )()
        )(),
        pref_modifier=setup_explicit_flags
    ),
]

