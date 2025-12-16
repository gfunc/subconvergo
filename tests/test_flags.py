
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
