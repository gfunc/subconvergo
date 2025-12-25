import pytest
from .utils import get_sub, verify_clash_yaml

def test_config_isolation(subconvergo_client, mock_server):
    """
    Verify that external configuration overrides do not persist to subsequent requests.
    This covers a previous bug where global config was modified by request-specific external config.
    """
    
    # 1. Create a custom clash rule base on the mock server
    custom_base_content = """
port: 7890
socks-port: 7891
allow-lan: true
mode: Rule
log-level: debug
external-controller: 127.0.0.1:9090
custom-key: "overridden-base"
"""
    mock_server.add_file("custom_base.tpl", custom_base_content)
    
    # 2. Create an external config that points to this base
    external_config_content = f"""
[common]
clash_rule_base = "{mock_server.base_url}/custom_base.tpl"
"""
    mock_server.add_file("override.ini", external_config_content)
    
    # 3. Make a request using the external config
    params = {
        "target": "clash",
        "url": f"{mock_server.base_url}/sub",
        "config": f"{mock_server.base_url}/override.ini"
    }
    resp = subconvergo_client.get("/sub", params=params)
    assert resp.status_code == 200
    data = verify_clash_yaml(resp.text)
    assert data.get("custom-key") == "overridden-base"
    
    # 4. Make a subsequent request WITHOUT the external config
    # It should use the default base (which doesn't have custom-key="overridden-base")
    params_default = {
        "target": "clash",
        "url": f"{mock_server.base_url}/sub"
    }
    resp_default = subconvergo_client.get("/sub", params=params_default)
    assert resp_default.status_code == 200
    data_default = verify_clash_yaml(resp_default.text)
    
    # The default base should NOT have the custom key from the overridden base
    assert data_default.get("custom-key") != "overridden-base"
