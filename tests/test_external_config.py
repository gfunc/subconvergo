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

def test_overwrite_original_rules_true():
    """
    Test overwrite_original_rules=true.
    Original rules from base/rules_base.tpl should be GONE.
    Only new rules should be present.
    """
    config_path = "config/test_overwrite_true.ini"
    resp = infra.api_get_subconvergo(
        "/sub", 
        params={
            "target": "clash",
            "url": f"{infra.MOCK_BASE}/ss-subscription.txt",
            "config": config_path
        }
    )
    assert resp.status_code == 200
    data = yaml.safe_load(resp.text)
    rules = data.get("rules", [])
    
    # New rule must be present
    assert "DOMAIN,example.com,DIRECT" in rules, f"New rule not found in {rules}"
    
    # Original rules must NOT be present
    assert "DOMAIN-SUFFIX,google.com,Proxy" not in rules, "Original rule found but should be overwritten"
    print("test_overwrite_original_rules_true passed")

def test_overwrite_original_rules_false():
    """
    Test overwrite_original_rules=false.
    Original rules from base/rules_base.tpl should be PRESENT.
    New rules should be appended/prepended.
    """
    config_path = "config/test_overwrite_false.ini"
    resp = infra.api_get_subconvergo(
        "/sub", 
        params={
            "target": "clash",
            "url": f"{infra.MOCK_BASE}/ss-subscription.txt",
            "config": config_path
        }
    )
    assert resp.status_code == 200
    data = yaml.safe_load(resp.text)
    rules = data.get("rules", [])
    
    # New rule must be present
    assert "DOMAIN,example.com,DIRECT" in rules, f"New rule not found in {rules}"
    
    # Original rules must ALSO be present
    assert "DOMAIN-SUFFIX,google.com,Proxy" in rules, "Original rule not found but should be preserved"
    print("test_overwrite_original_rules_false passed")

def test_overwrite_original_rules_true_toml():
    """
    Test overwrite_original_rules=true with TOML.
    """
    config_path = "config/test_overwrite_true.toml"
    resp = infra.api_get_subconvergo(
        "/sub", 
        params={
            "target": "clash",
            "url": f"{infra.MOCK_BASE}/ss-subscription.txt",
            "config": config_path
        }
    )
    assert resp.status_code == 200
    data = yaml.safe_load(resp.text)
    rules = data.get("rules", [])
    
    # New rule must be present
    assert "DOMAIN,example.com,DIRECT" in rules, f"New rule not found in {rules}"
    
    # Original rules must NOT be present
    assert "DOMAIN-SUFFIX,google.com,Proxy" not in rules, "Original rule found but should be overwritten"
    print("test_overwrite_original_rules_true_toml passed")

def test_overwrite_original_rules_false_yaml():
    """
    Test overwrite_original_rules=false with YAML.
    """
    config_path = "config/test_overwrite_false.yaml"
    resp = infra.api_get_subconvergo(
        "/sub", 
        params={
            "target": "clash",
            "url": f"{infra.MOCK_BASE}/ss-subscription.txt",
            "config": config_path
        }
    )
    assert resp.status_code == 200
    data = yaml.safe_load(resp.text)
    rules = data.get("rules", [])
    
    # New rule must be present
    assert "DOMAIN,example.com,DIRECT" in rules, f"New rule not found in {rules}"
    
    # Original rules must ALSO be present
    assert "DOMAIN-SUFFIX,google.com,Proxy" in rules, "Original rule not found but should be preserved"
    print("test_overwrite_original_rules_false_yaml passed")

def _run_import_test(config_path):
    resp = infra.api_get_subconvergo(
        "/sub", 
        params={
            "target": "clash",
            "url": f"{infra.MOCK_BASE}/ss-subscription.txt",
            "config": config_path
        }
    )
    assert resp.status_code == 200
    data = yaml.safe_load(resp.text)
    
    # Verify imported ruleset
    rules = data.get("rules", [])
    assert "DOMAIN-SUFFIX,imported.example.com,DIRECT" in rules, f"Imported rule not found in {rules}"
    
    # Verify imported proxy group
    proxy_groups = data.get("proxy-groups", [])
    group_names = [g["name"] for g in proxy_groups]
    assert "ImportedGroup" in group_names, f"Imported group not found in {group_names}"

def test_import_keyword_ini():
    """
    Test !!import: keyword in INI config.
    """
    _run_import_test("config/test_import.ini")
    print("test_import_keyword_ini passed")

def test_import_keyword_toml():
    """
    Test !!import: keyword in TOML config.
    """
    _run_import_test("config/test_import.toml")
    print("test_import_keyword_toml passed")

def test_import_keyword_yaml():
    """
    Test !!import: keyword in YAML config.
    """
    _run_import_test("config/test_import.yml")
    print("test_import_keyword_yaml passed")

CASES = [
    (setup_external_config, test_external_config_override),
    (setup_external_config, test_external_config_override_yaml),
    (setup_external_config, test_external_config_override_toml),
    (setup_external_config, test_overwrite_original_rules_true),
    (setup_external_config, test_overwrite_original_rules_false),
    (setup_external_config, test_overwrite_original_rules_true_toml),
    (setup_external_config, test_overwrite_original_rules_false_yaml),
    (setup_external_config, test_import_keyword_ini),
    (setup_external_config, test_import_keyword_toml),
    (setup_external_config, test_import_keyword_yaml),
]

if __name__ == "__main__":
    try:
        infra.start_subconvergo()
        test_external_config_override()
        test_external_config_override_yaml()
        test_external_config_override_toml()
        test_overwrite_original_rules_true()
        test_overwrite_original_rules_false()
    finally:
        infra.stop_subconvergo()
