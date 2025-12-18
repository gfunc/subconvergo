import yaml
from . import infra
from . import utils

def setup_external_config(pref):
    # Ensure we can load external configs
    pref["common"]["default_external_config"] = ""

def test_external_config_override():
    """
    Test that external config overrides proxy groups and rulesets.
    We use test_external.ini which defines:
    - A custom proxy group 'conf_test_group'
    - Overwrites rules with GEOIP,CN,DIRECT and MATCH,conf_test_group
    """
    
    # The path is relative to base_path (tests/base)
    config_path = "config/test_external.ini"
    
    # We request a clash config using the external config
    _run_external_config_test("config/test_external.ini", "conf_test_group")

def test_external_config_override_yaml():
    _run_external_config_test("config/test_external.yaml", "conf_test_group_yaml")

def test_external_config_override_toml():
    _run_external_config_test("config/test_external.toml", "conf_test_group_toml")

def _run_external_config_test(config_path, group_name):
    resp = infra.api_get_subconvergo(
        "/sub", 
        params={
            "target": "clash",
            "url": f"{infra.MOCK_BASE}/ss-subscription.txt",
            "config": config_path
        }
    )
    
    assert resp.status_code == 200, f"Request failed: {resp.text}"
    
    data = yaml.safe_load(resp.text)
    
    # 1. Verify Proxy Groups
    # We expect group_name to be present
    proxy_groups = data.get("proxy-groups", [])
    group_names = [g["name"] for g in proxy_groups]
    
    assert group_name in group_names, f"External config group '{group_name}' not found in {group_names}"
    
    # 2. Verify Rules
    # We expect rules to be overwritten by the external config
    rules = data.get("rules", [])
    
    # Check for the specific rules
    has_geoip = any("GEOIP,CN,DIRECT" in r for r in rules)
    has_match = any(f"MATCH,{group_name}" in r for r in rules)
    
    assert has_geoip, f"External config rule 'GEOIP,CN,DIRECT' not found in {rules}"
    assert has_match, f"External config rule 'MATCH,{group_name}' not found in {rules}"
    
    print(f"External config test passed for {config_path}: Proxy group and ruleset overrides verified.")

CASES = [
    (setup_external_config, test_external_config_override),
    (setup_external_config, test_external_config_override_yaml),
    (setup_external_config, test_external_config_override_toml),
]
