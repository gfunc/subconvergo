import yaml
import base64
from . import infra
from . import utils

import yaml
import base64
from . import infra
from . import utils

def setup_version(pref):
    pref["advanced"]["log_level"] = "warn"

def setup_sub(pref):
    pref["node_pref"]["clash_proxies_style"] = "list"

def setup_render(pref):
    pref["template"]["globals"].append({"key": "clash.render_case", "value": True})

def setup_profile(pref):
    pref["common"]["include_remarks"] = ["Example"]

def setup_ruleset_remote(pref):
    pref["rulesets"]["update_ruleset_on_request"] = True

def setup_filters_regex(pref):
    pref["common"]["include_remarks"] = ["/^HK/"]
    pref["common"]["exclude_remarks"] = []

def setup_exclude_remarks(pref):
    pref["common"]["exclude_remarks"] = ["HK"]

def setup_include_remarks(pref):
    pref["common"]["include_remarks"] = ["HK"]

def setup_emoji_rule(pref):
    pref["emojis"] = {
        "add_emoji": True,
        "remove_old_emoji": True,
        "rules": [
            {"match": "(HK|Hong Kong)", "emoji": "🇭🇰"},
            {"match": "(US|United States)", "emoji": "🇺🇸"},
        ]
    }

def setup_rename_node(pref):
    pref["node_pref"]["rename_node"] = [
        {"match": "HK", "replace": "Hong Kong"},
        {"match": "US", "replace": "United States"},
    ]

def setup_userinfo(pref):
    pref["node_pref"]["append_sub_userinfo"] = True

CASES = [
    infra.StandaloneTestCase(
        name="version",
        query=lambda: infra.api_get_subconvergo("/version"),
        validate=lambda resp: (
            lambda body=resp.text.strip(): (
                AssertionError(f"unexpected version payload: {body}") if not body.startswith("subconvergo") else {"version": body}
            )
        )(),
        pref_modifier=setup_version
    ),
    infra.StandaloneTestCase(
        name="sub",
        query=lambda: infra.api_get_subconvergo("/sub", params={
            "target": "clash",
            "url": f"{infra.MOCK_BASE}/ss-subscription.txt",
            "udp": "true",
            "tfo": "true",
        }),
        validate=lambda resp: (
            lambda data=yaml.safe_load(resp.text): (
                utils.assert_proxies(data, ["HK-Server-01"]),
                utils.assert_group_contains(data, "Auto", "HK-Server-01"),
                utils.assert_rules(data),
                {"proxy_count": len(data.get("proxies", [])), "rule_count": len(data.get("rules", []))}
            )[-1]
        )(),
        pref_modifier=setup_sub
    ),
    infra.StandaloneTestCase(
        name="surge2clash",
        query=lambda: infra.api_get_subconvergo("/surge2clash", params={"url": f"{infra.MOCK_BASE}/surge-subscription.ini"}),
        validate=lambda resp: (
            lambda data=yaml.safe_load(resp.text): (
                AssertionError("surge2clash did not return valid Clash config") if not isinstance(data, dict) or "proxies" not in data else None,
                utils.assert_proxies(data, ["SS-Example"]),
                {"proxy_count": len(data.get("proxies", []))}
            )[-1]
        )(),
        pref_modifier=None
    ),
    infra.StandaloneTestCase(
        name="render",
        query=lambda: infra.api_get_subconvergo("/render", params={"path": infra.TEMPLATE_PATH, "token": infra.TOKEN}),
        validate=lambda resp: (
            lambda data=yaml.safe_load(resp.text): (
                utils.assert_proxies(data, ["TestProxy"]),
                utils.assert_group_contains(data, "Auto", "TestProxy"),
                utils.assert_rules(data),
                {"lines": len(resp.text.splitlines())}
            )[-1]
        )(),
        pref_modifier=setup_render
    ),
    infra.StandaloneTestCase(
        name="profile",
        query=lambda: infra.api_get_subconvergo("/getprofile", params={"name": "example_profile", "token": infra.TOKEN}),
        validate=lambda resp: (
            lambda data=yaml.safe_load(resp.text): (
                utils.assert_proxies(data, ["Example"]),
                utils.assert_group_contains(data, "Auto", "Example"),
                utils.assert_rules(data),
                {"bytes": len(resp.content)}
            )[-1]
        )(),
        pref_modifier=setup_profile
    ),
    infra.StandaloneTestCase(
        name="ruleset_remote",
        query=lambda: infra.api_get_subconvergo("/getruleset", params={
            "url": base64.urlsafe_b64encode(f"{infra.MOCK_BASE}/test_rules.list".encode()).decode(),
            "type": "clash"
        }),
        validate=lambda resp: (
            lambda body=resp.text.strip(): (
                AssertionError("ruleset body missing MATCH,Auto") if "MATCH,Auto" not in body else {"lines": len(body.splitlines())}
            )
        )(),
        pref_modifier=setup_ruleset_remote
    ),
    infra.StandaloneTestCase(
        name="filters_regex",
        query=lambda: infra.api_get_subconvergo("/sub", params={
            "target": "clash",
            "url": f"{infra.MOCK_BASE}/ss-subscription.txt",
        }),
        validate=lambda resp: (
            lambda data=yaml.safe_load(resp.text): (
                lambda names=[p.get("name") for p in data.get("proxies", [])]: (
                    AssertionError(f"regex filter did not restrict to HK*: {names}") if len(names) == 0 or len(names) != len([n for n in names if n and n.startswith("HK")]) else {"proxy_count": len(names)}
                )
            )()
        )(),
        pref_modifier=setup_filters_regex
    ),
    infra.StandaloneTestCase(
        name="sub_with_external_config",
        query=lambda: infra.api_get_subconvergo("/sub", params={
            "target": "clash",
            "url": f"{infra.MOCK_BASE}/ss-subscription.txt",
            "config": "/base/resource/external.yml",
        }),
        validate=lambda resp: (
            lambda data=yaml.safe_load(resp.text): (
                utils.assert_proxies(data, ["HK-Server-01"]),
                utils.assert_group_contains(data, "Auto", "HK-Server-01"),
                utils.assert_rules(data),
                {"proxy_count": len(data.get("proxies", [])), "rule_count": len(data.get("rules", []))}
            )[-1]
        )(),
        pref_modifier=None
    ),
    infra.StandaloneTestCase(
        name="clash_only_config",
        query=lambda: infra.api_get_subconvergo("/sub", params={
            "target": "clash",
            "url": f"{infra.MOCK_BASE}/clash_only.yaml",
        }),
        validate=lambda resp: (
            lambda data=yaml.safe_load(resp.text): (
                utils.assert_proxies(data, ["sshtest"]),
                AssertionError(f"sshtest type mismatch: expected ssh, got {data.get('proxies', [])[0].get('type')}") if data.get("proxies", [])[0].get("type") != "ssh" else None,
                utils.assert_group_contains(data, "sshg", "sshtest"),
                (lambda rules=data.get("rules", []): (
                    [AssertionError(f"missing rule: {rule}") for rule in ["DOMAIN-SUFFIX,home.com,sshg", "IP-CIDR,192.168.0.0/16,sshg,no-resolve"] if rule not in rules],
                    {"proxy_count": len(data.get("proxies", [])), "rule_count": len(rules)}
                )[-1])()
            )[-1]
        )(),
        pref_modifier=None
    ),
    infra.StandaloneTestCase(
        name="exclude_remarks",
        query=lambda: infra.api_get_subconvergo("/sub", params={"target": "clash", "url": f"{infra.MOCK_BASE}/ss-subscription.txt"}),
        validate=lambda resp: (
            lambda data=yaml.safe_load(resp.text): (
                lambda names=[p.get("name") for p in data.get("proxies", [])]: (
                    AssertionError(f"exclude failed: found HK in {names}") if any("HK" in n for n in names) else {"proxy_count": len(names)}
                )
            )()
        )(),
        pref_modifier=setup_exclude_remarks
    ),
    infra.StandaloneTestCase(
        name="include_remarks",
        query=lambda: infra.api_get_subconvergo("/sub", params={"target": "clash", "url": f"{infra.MOCK_BASE}/ss-subscription.txt"}),
        validate=lambda resp: (
            lambda data=yaml.safe_load(resp.text): (
                lambda names=[p.get("name") for p in data.get("proxies", [])]: (
                    AssertionError(f"include failed: found non-HK in {names}") if not all("HK" in n for n in names) else {"proxy_count": len(names)}
                )
            )()
        )(),
        pref_modifier=setup_include_remarks
    ),
    infra.StandaloneTestCase(
        name="emoji_rule",
        query=lambda: infra.api_get_subconvergo("/sub", params={"target": "clash", "url": f"{infra.MOCK_BASE}/ss-subscription.txt"}),
        validate=lambda resp: (
            lambda data=yaml.safe_load(resp.text): (
                lambda names=[p.get("name") for p in data.get("proxies", [])]: (
                    AssertionError(f"emoji failed: no flag in {names}") if not any("🇭🇰" in n for n in names) else {"proxy_count": len(names)}
                )
            )()
        )(),
        pref_modifier=setup_emoji_rule
    ),
    infra.StandaloneTestCase(
        name="rename_node",
        query=lambda: infra.api_get_subconvergo("/sub", params={"target": "clash", "url": f"{infra.MOCK_BASE}/ss-subscription.txt"}),
        validate=lambda resp: (
            lambda data=yaml.safe_load(resp.text): (
                lambda names=[p.get("name") for p in data.get("proxies", [])]: (
                    AssertionError(f"rename failed: no Hong Kong in {names}") if not any("Hong Kong" in n for n in names) else {"proxy_count": len(names)}
                )
            )()
        )(),
        pref_modifier=setup_rename_node
    ),
    infra.StandaloneTestCase(
        name="userinfo",
        query=lambda: infra.api_get_subconvergo("/sub", params={"target": "clash", "url": f"{infra.MOCK_BASE}/ss-subscription.txt"}),
        validate=lambda resp: {"status": "ok"},
        pref_modifier=setup_userinfo
    ),
]
