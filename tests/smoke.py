#!/usr/bin/env python3
"""Minimal docker-based smoke test runner for subconvergo."""

import base64
import json
import subprocess
import time
from pathlib import Path

import requests
import yaml

TESTS_DIR = Path(__file__).resolve().parents[0]
BASE_DIR = TESTS_DIR / "base"
PREF_PATH = BASE_DIR / "pref.yml"
RESULTS_DIR = TESTS_DIR / "results"
RESULTS_DIR.mkdir(parents=True, exist_ok=True)
RESULTS_FILE = RESULTS_DIR / "smoke_summary.json"
COMPOSE_FILE = TESTS_DIR / "docker-compose.test.yml"
BASE_URL = "http://127.0.0.1:25500"
MOCK_BASE = "http://mock-subscription"
TOKEN = "password"
TEMPLATE_PATH = "/app/base/test_template.tpl"

ORIGINAL_PREF_EXISTS = PREF_PATH.exists()
ORIGINAL_PREF = PREF_PATH.read_text() if ORIGINAL_PREF_EXISTS else ""


def base_pref() -> dict:
    return {
        "common": {
            "api_mode": True,
            "api_access_token": TOKEN,
            "default_url": [f"{MOCK_BASE}/subscription-ss.txt"],
            "base_path": "base",
            "clash_rule_base": "base/all_base.tpl",
            "singbox_rule_base": "base/all_base.tpl",
            "proxy_subscription": "NONE",
            "proxy_config": "NONE",
            "proxy_ruleset": "NONE",
            "include_remarks": [],
            "exclude_remarks": [],
        },
        "node_pref": {
            "append_sub_userinfo": False,
            "clash_use_new_field_name": True,
            "clash_proxies_style": "flow",
            "clash_proxy_groups_style": "block",
            "singbox_add_clash_modes": True,
        },
        "rulesets": {
            "enabled": True,
            "overwrite_original_rules": False,
            "update_ruleset_on_request": False,
            "rulesets": [
                {"ruleset": "rules/custom_test_rules.list", "group": "Auto"},
                {"rule": "MATCH,Auto", "group": "Auto"},
            ],
        },
        "proxy_groups": {
            "custom_proxy_group": [
                {"name": "Auto", "type": "select", "rule": [".*"]},
            ]
        },
        "template": {
            "template_path": "base/base",
            "globals": [
                {"key": "clash.http_port", "value": 7890},
                {"key": "clash.socks_port", "value": 7891},
                {"key": "clash.allow_lan", "value": True},
            ],
        },
        "managed_config": {
            "write_managed_config": False,
            "managed_config_prefix": "",
            "config_update_interval": 86400,
            "config_update_strict": False,
        },
        "server": {"listen": "0.0.0.0", "port": 25500},
        "advanced": {"log_level": "info"},
    }


def pref_variant(case: str) -> dict:
    pref = base_pref()
    if case == "version":
        pref["advanced"]["log_level"] = "warn"
    elif case == "sub":
        pref["node_pref"]["clash_proxies_style"] = "list"
    elif case == "render":
        pref["template"]["globals"].append({"key": "clash.render_case", "value": True})
    elif case == "profile":
        pref["common"]["include_remarks"] = ["HK"]
    elif case == "ruleset":
        pref["rulesets"]["update_ruleset_on_request"] = True
    else:
        raise ValueError(f"unknown pref case: {case}")
    return pref


def write_pref(case: str) -> None:
    config = pref_variant(case)
    PREF_PATH.parent.mkdir(parents=True, exist_ok=True)
    PREF_PATH.write_text(yaml.safe_dump(config, sort_keys=False))


def restore_pref() -> None:
    if ORIGINAL_PREF_EXISTS:
        PREF_PATH.write_text(ORIGINAL_PREF)
    elif PREF_PATH.exists():
        PREF_PATH.unlink()


def compose_up() -> None:
    cmd = ["docker", "compose", "-f", str(COMPOSE_FILE), "up", "--build", "-d"]
    subprocess.run(cmd, cwd=TESTS_DIR, check=True)


def compose_down() -> None:
    cmd = ["docker", "compose", "-f", str(COMPOSE_FILE), "down", "--volumes", "--remove-orphans"]
    subprocess.run(cmd, cwd=TESTS_DIR, check=True)


def wait_for_service(timeout: int = 120) -> None:
    deadline = time.time() + timeout
    while time.time() < deadline:
        try:
            resp = requests.get(f"{BASE_URL}/version", timeout=5)
            if resp.status_code == 200:
                return
        except requests.RequestException:
            pass
        time.sleep(2)
    raise RuntimeError("service did not become ready in time")


def api_get(path: str, *, params=None, expected=200) -> requests.Response:
    resp = requests.get(f"{BASE_URL}{path}", params=params, timeout=30)
    codes = expected if isinstance(expected, (list, tuple)) else (expected,)
    if resp.status_code not in codes:
        raise AssertionError(f"GET {path} returned {resp.status_code}: {resp.text[:200]}")
    return resp


def reload_config(case: str) -> str:
    write_pref(case)
    resp = api_get("/readconf", params={"token": TOKEN})
    return resp.text.strip()


def assert_proxies(struct: dict, names) -> None:
    proxies = {p.get("name") for p in struct.get("proxies", [])}
    missing = [name for name in names if name not in proxies]
    if missing:
        raise AssertionError(f"missing proxies: {missing}")


def assert_group_contains(struct: dict, group: str, member: str) -> None:
    for grp in struct.get("proxy-groups", []):
        if grp.get("name") == group and member in grp.get("proxies", []):
            return
    raise AssertionError(f"proxy group '{group}' missing member '{member}'")


def assert_rules(struct: dict) -> None:
    rules = struct.get("rules", [])
    if not rules or not any("MATCH" in rule for rule in rules):
        raise AssertionError("rules list missing MATCH entries")


def test_version() -> dict:
    resp = api_get("/version")
    body = resp.text.strip()
    if not body.startswith("subconvergo"):
        raise AssertionError(f"unexpected version payload: {body}")
    return {"version": body}


def test_sub() -> dict:
    params = {
        "target": "clash",
        "url": f"{MOCK_BASE}/subscription-ss.txt",
        "udp": "true",
        "tfo": "true",
    }
    resp = api_get("/sub", params=params)
    data = yaml.safe_load(resp.text)
    assert_proxies(data, ["HK-Server-01", "US-Server-01", "JP-Server-01"])
    assert_group_contains(data, "Auto", "HK-Server-01")
    assert_rules(data)
    return {"proxy_count": len(data.get("proxies", [])), "rule_count": len(data.get("rules", []))}


def test_render() -> dict:
    resp = api_get(
        "/render",
        params={"path": TEMPLATE_PATH, "token": TOKEN},
    )
    data = yaml.safe_load(resp.text)
    assert_proxies(data, ["TestProxy"])
    assert_group_contains(data, "Auto", "TestProxy")
    assert_rules(data)
    return {"lines": len(resp.text.splitlines())}


def test_profile() -> dict:
    resp = api_get(
        "/getprofile",
        params={"name": "example_profile", "token": TOKEN},
    )
    data = yaml.safe_load(resp.text)
    assert_proxies(data, ["HK-Server-01"])
    assert_group_contains(data, "Auto", "HK-Server-01")
    assert_rules(data)
    return {"bytes": len(resp.content)}


def test_ruleset() -> dict:
    encoded = base64.urlsafe_b64encode(f"{MOCK_BASE}/test_rules.list".encode()).decode()
    resp = api_get("/getruleset", params={"url": encoded, "type": "clash"})
    body = resp.text.strip()
    if "not yet" not in body:
        raise AssertionError("ruleset endpoint response unexpected")
    return {"message": body}


def main() -> None:
    cases = [
        ("version", test_version),
        ("sub", test_sub),
        ("render", test_render),
        ("profile", test_profile),
        ("ruleset", test_ruleset),
    ]

    write_pref(cases[0][0])
    compose_up()
    try:
        wait_for_service()
        results = {}
        for case, func in cases:
            reload_note = reload_config(case)
            results[f"{case}_reload"] = reload_note
            results[case] = func()
        RESULTS_FILE.write_text(json.dumps(results, indent=2))
        print(f"Smoke tests passed. Summary: {RESULTS_FILE}")
    finally:
        compose_down()
        restore_pref()


if __name__ == "__main__":
    main()
