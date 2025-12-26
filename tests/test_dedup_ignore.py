import yaml
from . import infra

def setup_dedup_ignore(pref):
    pref["custom_proxy_group"] = [
        {
            "name": "DuplicateGroup",
            "type": "select",
            "rule": ["[]DuplicateProxy"]
        }
    ]

def validate_ignore_source_default(resp):
    data = yaml.safe_load(resp.text)
    groups = [g["name"] for g in data.get("proxy-groups", [])]
    
    # SourceGroup should NOT be present (default ignore_source=true)
    if "SourceGroup" in groups:
        return {"_failures": ["SourceGroup should be ignored by default"]}
    
    # DuplicateGroup should be present (from config)
    if "DuplicateGroup" not in groups:
        return {"_failures": ["DuplicateGroup from config missing"]}
        
    return {"groups": groups}

def validate_ignore_source_false(resp):
    data = yaml.safe_load(resp.text)
    groups = [g["name"] for g in data.get("proxy-groups", [])]
    
    # SourceGroup SHOULD be present
    if "SourceGroup" not in groups:
        return {"_failures": ["SourceGroup should be present when ignore_source=false"]}
        
    return {"groups": groups}

def validate_dedup_proxies(resp):
    data = yaml.safe_load(resp.text)
    proxies = [p["name"] for p in data.get("proxies", [])]
    
    # Should have DuplicateProxy and DuplicateProxy 2
    if "DuplicateProxy" not in proxies:
        return {"_failures": ["DuplicateProxy missing"]}
    if "DuplicateProxy 2" not in proxies:
        return {"_failures": ["DuplicateProxy 2 missing (deduplication failed)"]}
        
    return {"proxies": proxies}

CASES = [
    infra.StandaloneTestCase(
        name="ignore_source_default",
        query=lambda: infra.api_get_subconvergo("/sub?target=clash&url=http://mock-subscription/dedup-ignore.yaml"),
        validate=validate_ignore_source_default,
        pref_modifier=setup_dedup_ignore
    ),
    infra.StandaloneTestCase(
        name="ignore_source_false",
        query=lambda: infra.api_get_subconvergo("/sub?target=clash&url=http://mock-subscription/dedup-ignore.yaml&ignore_source=false"),
        validate=validate_ignore_source_false,
        pref_modifier=setup_dedup_ignore
    ),
    infra.StandaloneTestCase(
        name="ignore_source_external_config",
        query=lambda: infra.api_get_subconvergo("/sub?target=clash&url=http://mock-subscription/dedup-ignore.yaml&config=http://mock-subscription/ignore_source_false.yml"),
        validate=validate_ignore_source_false,
        pref_modifier=setup_dedup_ignore
    ),
    infra.StandaloneTestCase(
        name="dedup_proxies",
        query=lambda: infra.api_get_subconvergo("/sub?target=clash&url=http://mock-subscription/dedup-ignore.yaml"),
        validate=validate_dedup_proxies,
        pref_modifier=setup_dedup_ignore
    ),
]
