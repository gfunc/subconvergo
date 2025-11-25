#!/usr/bin/env python3
"""Minimal docker-based smoke test runner for subconvergo."""

import argparse
import base64
import json
import re
import subprocess
import sys
import time
from pathlib import Path
from typing import List, Dict, Any

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
SUBCONVERTER_URL = "http://127.0.0.1:25550"
MOCK_BASE = "http://mock-subscription"
TOKEN = "password"
TEMPLATE_PATH = "/base/base/test_template.tpl"

ORIGINAL_PREF_EXISTS = PREF_PATH.exists()
ORIGINAL_PREF = PREF_PATH.read_text() if ORIGINAL_PREF_EXISTS else ""


def save_result(case: str, content: str, filename: str = "subconvergo.txt") -> None:
    out_dir = RESULTS_DIR / case
    out_dir.mkdir(parents=True, exist_ok=True)
    (out_dir / filename).write_text(content, encoding="utf-8")


def base_pref() -> dict:
    return {
        "common": {
            "api_mode": True,
            "api_access_token": TOKEN,
            "default_url": [f"{MOCK_BASE}/ss-subscription.txt"],
            "base_path": "base",
            "clash_rule_base": "base/all_base.tpl",
            "surge_rule_base": "base/all_base.tpl",
            "surfboard_rule_base": "base/all_base.tpl",
            "mellow_rule_base": "base/all_base.tpl",
            "quan_rule_base": "base/all_base.tpl",
            "quanx_rule_base": "base/all_base.tpl",
            "loon_rule_base": "base/all_base.tpl",
            "sssub_rule_base": "base/all_base.tpl",
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
            "template_path": "base",
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
        pref["common"]["include_remarks"] = ["Example"]
    elif case in ("ruleset", "ruleset_remote", "ruleset_compare"):
        pref["rulesets"]["update_ruleset_on_request"] = True
    elif case == "sub_with_external_config":
        # keep defaults; no special pref needed beyond base
        pass
    elif case == "clash_only_config":
        # keep defaults
        pass
    elif case == "filters_regex":
        pref["common"]["include_remarks"] = ["/^HK/"]
        pref["common"]["exclude_remarks"] = []
    elif case in ("exclude_remarks", "e2e_matrix_exclude"):
        pref["common"]["exclude_remarks"] = ["HK"]
    elif case in ("include_remarks", "e2e_matrix_include"):
        pref["common"]["include_remarks"] = ["HK"]
    elif case in ("emoji_rule", "e2e_matrix_emoji"):
        pref["emojis"] = {
            "add_emoji": True,
            "remove_old_emoji": True,
            "rules": [
                {"match": "(HK|Hong Kong)", "emoji": "🇭🇰"},
                {"match": "(US|United States)", "emoji": "🇺🇸"},
            ]
        }
    elif case in ("rename_node", "e2e_matrix_rename"):
        pref["node_pref"]["rename_node"] = [
            {"match": "HK", "replace": "Hong Kong"},
            {"match": "US", "replace": "United States"},
        ]
    elif case in ("userinfo", "e2e_matrix_userinfo"):
        pref["node_pref"]["append_sub_userinfo"] = True
    elif case == "e2e_matrix":
        # Use base pref
        pass
    elif case == "settings_comparison":
        # Use base pref
        pass
    elif case == "surge2clash":
        # Use base pref
        pass
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


def compose_up(build: bool = True) -> None:
    cmd = ["docker", "compose", "-f", str(COMPOSE_FILE), "up"]
    if build:
        cmd.append("--build")
    cmd.append("-d")
    subprocess.run(cmd, cwd=TESTS_DIR, check=True)


def compose_down() -> None:
    cmd = ["docker", "compose", "-f", str(COMPOSE_FILE), "down", "--volumes", "--remove-orphans"]
    subprocess.run(cmd, cwd=TESTS_DIR, check=True)


def restart_services() -> None:
    cmd = ["docker", "compose", "-f", str(COMPOSE_FILE), "restart"]
    subprocess.run(cmd, cwd=TESTS_DIR, check=True)


def wait_for_service(timeout: int = 120) -> None:
    deadline = time.time() + timeout
    while time.time() < deadline:
        try:
            api_get_subconvergo("/version")
            api_get_subconverter("/version")
            return
        except requests.RequestException:
            pass
        time.sleep(2)
    raise RuntimeError("service did not become ready in time")


def _api_get(base: str, path: str, *, params=None, expected=200) -> requests.Response:
    path = path.lstrip("/")
    resp = requests.get(f"{base}/{path}", params=params, timeout=30)
    codes = expected if isinstance(expected, (list, tuple)) else (expected,)
    if resp.status_code not in codes:
        raise AssertionError(f"GET {path} returned {resp.status_code}: {resp.text[:200]}")
    return resp

def api_get_subconvergo(path: str, *, params=None, expected=200) -> requests.Response:
    resp = _api_get(BASE_URL, path, params=params, expected=expected)
    return resp

def api_get_subconverter(path: str, *, params=None, expected=200) -> requests.Response:
    resp = _api_get(SUBCONVERTER_URL, path, params=params, expected=expected)
    return resp

def reload_config(case: str) -> str:
    write_pref(case)
    resp = api_get_subconvergo("/readconf", params={"token": TOKEN})
    resp2 = api_get_subconverter("/readconf", params={"token": TOKEN})
    return resp.text.strip() + " " + resp2.text.strip()


def assert_proxies(struct: dict, names) -> None:
    proxies = {p.get("name") for p in struct.get("proxies", [])}
    missing = [name for name in names if name not in proxies]
    if missing:
        raise AssertionError(f"missing proxies: {missing}. Found: {proxies}")


def assert_group_contains(struct: dict, group: str, member: str) -> None:
    for grp in struct.get("proxy-groups", []):
        if grp.get("name") == group and member in grp.get("proxies", []):
            return
    raise AssertionError(f"proxy group '{group}' missing member '{member}'")


def assert_rules(struct: dict) -> None:
    rules = struct.get("rules", [])
    if not rules or not any("MATCH" in rule for rule in rules):
        raise AssertionError("rules list missing MATCH entries")


def count_proxies(content: str, target: str) -> int:
    if not content:
        return 0
    try:
        proxies = extract_proxies(content, target)
        return len(proxies)
    except Exception:
        pass
    return -1


def extract_proxies(content: str, target: str) -> List[Dict[str, Any]]:
    """Extract list of proxies with normalized keys: name, type, server, port."""
    proxies = []
    if not content:
        return proxies
    try:
        if target == "clash":
            data = yaml.safe_load(content)
            if isinstance(data, dict):
                for p in data.get("proxies", []):
                    proxies.append({
                        "name": str(p.get("name", "")),
                        "type": str(p.get("type", "")),
                        "server": str(p.get("server", "")),
                        "port": str(p.get("port", "")),
                        "original": p
                    })
        elif target == "singbox":
            data = json.loads(content)
            if isinstance(data, dict):
                for p in data.get("outbounds", []):
                    # Skip structural outbounds
                    if p.get("type") in ["selector", "urltest", "direct", "block", "dns"]:
                        continue
                    proxies.append({
                        "name": str(p.get("tag", "")),
                        "type": str(p.get("type", "")),
                        "server": str(p.get("server", "")),
                        "port": str(p.get("server_port", "")),
                        "original": p
                    })
        elif target in ["surge", "loon"]:
            match = re.search(r'\[Proxy\]\s*(.*?)\s*(\[|$)', content, re.DOTALL | re.IGNORECASE)
            if match:
                lines = [l.strip() for l in match.group(1).splitlines() if l.strip() and not l.strip().startswith(('#', ';'))]
                for line in lines:
                    if "=" in line:
                        name, rest = line.split("=", 1)
                        parts = [p.strip() for p in rest.split(",")]
                        if len(parts) >= 3:
                            proxies.append({
                                "name": name.strip(),
                                "type": parts[0],
                                "server": parts[1],
                                "port": parts[2],
                                "original": line
                            })
                        else:
                             proxies.append({"name": name.strip(), "original": line})
        elif target == "quanx":
             for section in ["server_remote", "server_local"]:
                 match = re.search(rf'\[{section}\]\s*(.*?)\s*(\[|$)', content, re.DOTALL | re.IGNORECASE)
                 if match:
                    lines = [l.strip() for l in match.group(1).splitlines() if l.strip() and not l.strip().startswith(('#', ';'))]
                    for line in lines:
                        tag_match = re.search(r'tag\s*=\s*([^,]+)', line)
                        if tag_match:
                            name = tag_match.group(1).strip()
                            # Try to extract type/server/port if possible, but QuanX format varies
                            # e.g. shadowsocks=server:port,...
                            # or vmess=server:port,...
                            # Simple heuristic for server:port
                            parts = line.split("=")
                            if len(parts) > 1:
                                type_part = parts[0].strip()
                                val_part = parts[1]
                                server_port = val_part.split(",")[0]
                                if ":" in server_port:
                                    srv, prt = server_port.split(":", 1)
                                    proxies.append({
                                        "name": name,
                                        "type": type_part,
                                        "server": srv.strip(),
                                        "port": prt.strip(),
                                        "original": line
                                    })
                                    continue
                            proxies.append({"name": name, "original": line})
    except Exception:
        pass
    return proxies


def compare_proxy_lists(cand_proxies: List[Dict[str, Any]], ref_proxies: List[Dict[str, Any]]) -> str:
    """Compare two lists of proxies and return a status string."""
    cand_map = {p["name"]: p for p in cand_proxies}
    ref_map = {p["name"]: p for p in ref_proxies}
    
    cand_names = set(cand_map.keys())
    ref_names = set(ref_map.keys())
    
    if cand_names == ref_names:
        # Check details
        mismatches = []
        for name in cand_names:
            c = cand_map[name]
            r = ref_map[name]
            # Compare core fields if they exist
            for field in ["type", "server", "port"]:
                if field in c and field in r:
                    val_c = str(c[field]).strip()
                    val_r = str(r[field]).strip()
                    if val_c != val_r:
                         mismatches.append(f"{name}.{field}({val_c}!={val_r})")
        if mismatches:
            return f"DETAIL_MISMATCH({len(mismatches)}): {','.join(mismatches[:3])}..."
        return f"MATCH({len(cand_names)})"
    
    missing = ref_names - cand_names
    extra = cand_names - ref_names
    
    msg = []
    if missing:
        msg.append(f"MISSING({len(missing)}):{list(missing)[:3]}")
    if extra:
        msg.append(f"EXTRA({len(extra)}):{list(extra)[:3]}")
        
    return " ".join(msg)



def test_version(case: str) -> dict:
    # raise AssertionError("Forced failure")
    resp = api_get_subconvergo("/version")
    save_result(case, resp.text)
    body = resp.text.strip()
    if not body.startswith("subconvergo"):
        raise AssertionError(f"unexpected version payload: {body}")
    return {"version": body}


def test_sub(case: str) -> dict:
    params = {
        "target": "clash",
        "url": f"{MOCK_BASE}/ss-subscription.txt",
        "udp": "true",
        "tfo": "true",
    }
    resp = api_get_subconvergo("/sub", params=params)
    save_result(case, resp.text)
    data = yaml.safe_load(resp.text)
    assert_proxies(data, ["HK-Server-01"])
    assert_group_contains(data, "Auto", "HK-Server-01")
    assert_rules(data)
    return {"proxy_count": len(data.get("proxies", [])), "rule_count": len(data.get("rules", []))}


def test_surge2clash(case: str) -> dict:
    # Test /surge2clash endpoint
    # It should behave like /sub?target=clash&url=...
    url = f"{MOCK_BASE}/surge-subscription.ini"
    params = {"url": url}
    resp = api_get_subconvergo("/surge2clash", params=params)
    save_result(case, resp.text)
    
    # Verify it's valid Clash YAML
    data = yaml.safe_load(resp.text)
    if not isinstance(data, dict) or "proxies" not in data:
        raise AssertionError("surge2clash did not return valid Clash config")
        
    # Verify content
    assert_proxies(data, ["SS-Example"])
    
    return {"proxy_count": len(data.get("proxies", []))}


def test_render(case: str) -> dict:
    resp = api_get_subconvergo(
        "/render",
        params={"path": TEMPLATE_PATH, "token": TOKEN},
    )
    save_result(case, resp.text)
    data = yaml.safe_load(resp.text)
    assert_proxies(data, ["TestProxy"])
    assert_group_contains(data, "Auto", "TestProxy")
    assert_rules(data)
    return {"lines": len(resp.text.splitlines())}


def test_profile(case: str) -> dict:
    resp = api_get_subconvergo(
        "/getprofile",
        params={"name": "example_profile", "token": TOKEN},
    )
    save_result(case, resp.text)
    data = yaml.safe_load(resp.text)
    assert_proxies(data, ["Example"])
    assert_group_contains(data, "Auto", "Example")
    assert_rules(data)
    return {"bytes": len(resp.content)}


def test_ruleset_remote(case: str) -> dict:
    encoded = base64.urlsafe_b64encode(f"{MOCK_BASE}/test_rules.list".encode()).decode()
    resp = api_get_subconvergo("/getruleset", params={"url": encoded, "type": "clash"})
    save_result(case, resp.text)
    body = resp.text.strip()
    if "MATCH,Auto" not in body:
        raise AssertionError("ruleset body missing MATCH,Auto")
    return {"lines": len(body.splitlines())}

def test_ruleset_compare_with_subconverter(case: str) -> dict:
    url_plain = f"{MOCK_BASE}/test_rules.list"
    encoded = base64.urlsafe_b64encode(url_plain.encode()).decode()
    # Subconvergo (encoded)
    r1 = requests.get(f"{BASE_URL}/getruleset", params={"url": encoded, "type": "clash"}, timeout=30)
    r1.raise_for_status()
    save_result(case, r1.text, "subconvergo.txt")
    # Subconverter may expect plain URL; try plain first, then encoded
    r2 = api_get_subconverter("/getruleset", params={"url": url_plain, "type": "clash"}, expected=[200, 400])
    
    if r2.status_code != 200:
        r2 = api_get_subconverter("/getruleset", params={"url": encoded, "type": "clash"}, expected=[200, 400])
    save_result(case, r2.text, "subconverter.txt")
    if r2.status_code != 200:
        # Subconverter didn't support this combination; skip strict compare
        return {"skipped": True, "status": r2.status_code}
    b1 = r1.text.strip().splitlines()
    b2 = r2.text.strip().splitlines()
    # Loose compare: same number of lines and first/last equal
    if not b1 or not b2 or len(b1) != len(b2) or b1[0] != b2[0] or b1[-1] != b2[-1]:
        raise AssertionError("ruleset output differs from subconverter")
    return {"lines": len(b1)}

def test_filters_regex(case: str) -> dict:
    # Expect only HK* proxies due to include regex
    params = {
        "target": "clash",
        "url": f"{MOCK_BASE}/ss-subscription.txt",
    }
    # The pref for this case sets include_remarks to ["/^HK/"]
    resp = api_get_subconvergo("/sub", params=params)
    save_result(case, resp.text)
    data = yaml.safe_load(resp.text)
    names = [p.get("name") for p in data.get("proxies", [])]
    hk_only = [n for n in names if n and n.startswith("HK")] 
    if len(names) != len(hk_only) or len(names) == 0:
        raise AssertionError(f"regex filter did not restrict to HK*: {names}")
    return {"proxy_count": len(names)}

def test_sub_with_external_config(case: str) -> dict:
    # Use external config mounted at /base/resource/external.yml
    params = {
        "target": "clash",
        "url": f"{MOCK_BASE}/ss-subscription.txt",
        "config": "/base/resource/external.yml",
    }
    resp = api_get_subconvergo("/sub", params=params)
    save_result(case, resp.text)
    data = yaml.safe_load(resp.text)
    # Same structural checks as base sub
    assert_proxies(data, ["HK-Server-01"])
    assert_group_contains(data, "Auto", "HK-Server-01")
    assert_rules(data)
    return {"proxy_count": len(data.get("proxies", [])), "rule_count": len(data.get("rules", []))}


def test_clash_only_config(case: str) -> dict:
    # Use external config mounted at /mock-data/clash_only.yaml
    # Treat it as a subscription URL (local file)
    params = {
        "target": "clash",
        "url": f"{MOCK_BASE}/clash_only.yaml",
    }
    resp = api_get_subconvergo("/sub", params=params)
    save_result(case, resp.text)
    data = yaml.safe_load(resp.text)
    
    # Verify Proxy
    assert_proxies(data, ["sshtest"])
    # Verify Proxy Type (ssh)
    proxies = {p.get("name"): p for p in data.get("proxies", [])}
    if proxies["sshtest"].get("type") != "ssh":
        raise AssertionError(f"sshtest type mismatch: expected ssh, got {proxies['sshtest'].get('type')}")
        
    # Verify Proxy Group
    assert_group_contains(data, "sshg", "sshtest")
    
    # Verify Rules
    rules = data.get("rules", [])
    expected_rules = [
        "DOMAIN-SUFFIX,home.com,sshg",
        "IP-CIDR,192.168.0.0/16,sshg,no-resolve"
    ]
    for rule in expected_rules:
        if rule not in rules:
             raise AssertionError(f"missing rule: {rule}")

    return {"proxy_count": len(data.get("proxies", [])), "rule_count": len(rules)}


def test_exclude_remarks(case: str) -> dict:
    # Exclude "HK"
    params = {"target": "clash", "url": f"{MOCK_BASE}/ss-subscription.txt"}
    resp = api_get_subconvergo("/sub", params=params)
    save_result(case, resp.text)
    data = yaml.safe_load(resp.text)
    names = [p.get("name") for p in data.get("proxies", [])]
    if any("HK" in n for n in names):
        raise AssertionError(f"exclude failed: found HK in {names}")
    return {"proxy_count": len(names)}


def test_include_remarks(case: str) -> dict:
    # Include only "HK"
    params = {"target": "clash", "url": f"{MOCK_BASE}/ss-subscription.txt"}
    resp = api_get_subconvergo("/sub", params=params)
    save_result(case, resp.text)
    data = yaml.safe_load(resp.text)
    names = [p.get("name") for p in data.get("proxies", [])]
    if not all("HK" in n for n in names):
        raise AssertionError(f"include failed: found non-HK in {names}")
    return {"proxy_count": len(names)}


def test_emoji_rule(case: str) -> dict:
    # Check if HK gets flag
    params = {"target": "clash", "url": f"{MOCK_BASE}/ss-subscription.txt"}
    resp = api_get_subconvergo("/sub", params=params)
    save_result(case, resp.text)
    data = yaml.safe_load(resp.text)
    names = [p.get("name") for p in data.get("proxies", [])]
    # HK-Server-01 -> 🇭🇰 HK-Server-01
    if not any("🇭🇰" in n for n in names):
        raise AssertionError(f"emoji failed: no flag in {names}")
    return {"proxy_count": len(names)}


def test_rename_node(case: str) -> dict:
    # HK -> Hong Kong
    params = {"target": "clash", "url": f"{MOCK_BASE}/ss-subscription.txt"}
    resp = api_get_subconvergo("/sub", params=params)
    save_result(case, resp.text)
    data = yaml.safe_load(resp.text)
    names = [p.get("name") for p in data.get("proxies", [])]
    if not any("Hong Kong" in n for n in names):
        raise AssertionError(f"rename failed: no Hong Kong in {names}")
    return {"proxy_count": len(names)}


def test_userinfo(case: str) -> dict:
    # Just check it runs without error for now, as mock might not have userinfo
    params = {"target": "clash", "url": f"{MOCK_BASE}/ss-subscription.txt"}
    resp = api_get_subconvergo("/sub", params=params)
    save_result(case, resp.text)
    return {"status": "ok"}


def test_settings_comparison(case: str) -> dict:
    """Compare results with various settings (emoji, rules, etc)."""
    base_params = {
        "target": "clash",
        "url": f"{MOCK_BASE}/ss-subscription.txt",
    }
    
    # Settings to permute or test individually
    scenarios = [
        {"emoji": "true"},
        {"emoji": "false"},
        {"list": "true"}, # Clash list only
        {"udp": "true"},
        {"tfo": "true"},
        {"scv": "true"},
        {"fdn": "true"},
        {"sort": "true"},
    ]
    
    results = {}
    failures = []
    
    for i, settings in enumerate(scenarios):
        case_id = f"settings_{i}_{list(settings.keys())[0]}"
        params = base_params.copy()
        params.update(settings)
        settings_compare_dir = RESULTS_DIR /  "settings_comparison"
        settings_compare_dir.mkdir(parents=True, exist_ok=True)
        # 1. Subconvergo
        try:
            r1 = api_get_subconvergo("/sub", params=params)
            c1 = count_proxies(r1.text, "clash")
            (settings_compare_dir / "subconvergo.txt").write_text(r1.text, encoding="utf-8")
        except Exception as e:
            failures.append(f"{case_id}: Subconvergo failed: {e}")
            continue
            
        # 2. Subconverter
        try:
            r2 = api_get_subconverter("/sub", params=params, expected=[200, 400])
            c2 = count_proxies(r2.text, "clash")
            (settings_compare_dir / "subconverter.txt").write_text(r2.text, encoding="utf-8")
        except Exception as e:
            failures.append(f"{case_id}: Subconverter failed: {e}")
            continue
            
        if c1 != c2:
            failures.append(f"{case_id}: Proxy count mismatch: {c1} vs {c2}")
            results[case_id] = f"MISMATCH({c1}vs{c2})"
        else:
            results[case_id] = f"MATCH({c1})"
            
    if failures:
        results["_failures"] = failures
        
    return results


def test_e2e_matrix(test_name: str) -> dict:
    sources = {
        "mixed": "mixed-subscription.txt",
        "ss": "ss-subscription.txt",
        "ssr": "ssr-subscription.txt",
        "v2ray": "v2ray-subscription.txt",
        "ssd": "ssd-subscription.txt",
        "ss-android": "ss-android-subscription.json",
        "surge": "surge-subscription.ini",
        "clash": "clash-subscription.yaml",
    }
    targets = ["clash", "surge", "quanx", "loon", "singbox"]
    
    matrix_dir = RESULTS_DIR / "matrix" / test_name
    matrix_dir.mkdir(parents=True, exist_ok=True)

    results = {}
    failures = []

    for src_name, src_file in sources.items():
        for target in targets:
            case_id = f"{src_name}->{target}"
            url = f"{MOCK_BASE}/{src_file}"
            params = {"target": target, "url": url}
            
            # 1. Request from Subconvergo (Candidate)
            try:
                r1 = api_get_subconvergo("/sub", params=params, expected=[200, 400])
                content1 = r1.text
                cand_status = r1.status_code
            except Exception as e:
                failures.append(f"{case_id}: Subconvergo failed: {e}")
                results[case_id] = "FAIL_CANDIDATE"
                continue

            # 2. Request from Subconverter (Reference)
            try:
                r2 = api_get_subconverter("/sub", params=params, expected=[200, 400])
                content2 = r2.text
                ref_status = r2.status_code
            except Exception as e:
                content2 = ""
                ref_status = f"ERR({e})"

            # Save artifacts
            safe_case_id = case_id.replace("->", "_to_")
            (matrix_dir / f"{safe_case_id}_subconvergo.txt").write_text(content1, encoding="utf-8")
            (matrix_dir / f"{safe_case_id}_subconverter.txt").write_text(content2, encoding="utf-8")

            # 3. Validation & Comparison
            status = "OK"
            
            # Helper to detect content errors
            def detect_content_issue(content: str):
                if not content or not content.strip():
                    return "EMPTY"
                if "doesn't contain any valid node info" in content:
                    return "ERR_NO_NODES"
                return None

            cand_issue = detect_content_issue(content1)
            if cand_status != 200:
                if cand_status == 400 and ("No valid proxies found" in content1 or "doesn't contain any valid node info" in content1):
                    cand_issue = "ERR_NO_NODES"
                else:
                    cand_issue = f"HTTP_{cand_status}"

            if cand_issue:
                status = f"CAND_{cand_issue}"
            
            # Basic validation of Candidate output
            if status == "OK":
                if target == "clash":
                    try:
                        y = yaml.safe_load(content1)
                        if not isinstance(y, dict) or "proxies" not in y:
                            status = "INVALID_YAML_STRUCTURE"
                    except:
                        status = "INVALID_YAML"
                elif target == "singbox":
                    try:
                        j = json.loads(content1)
                        if "outbounds" not in j:
                            status = "INVALID_JSON_STRUCTURE"
                    except:
                        status = "INVALID_JSON"
                elif target in ["surge", "quanx", "loon"]:
                    # Heuristic check for INI-like structure
                    if "[Proxy]" not in content1 and "[server_local]" not in content1 and "shadowsocks=" not in content1:
                         # QuanX might not have [Proxy] but has other sections or just lines
                         if target == "quanx" and "server_remote" not in content1 and "shadowsocks" not in content1:
                             status = "SUSPICIOUS_INI"
                         elif target == "surge" and "[Proxy]" not in content1:
                             status = "SUSPICIOUS_INI"

            # Analyze Reference Result
            ref_desc = "OK"
            ref_issue = detect_content_issue(content2)
            
            if ref_status != 200:
                ref_desc = f"HTTP_{ref_status}"
                if ref_issue == "ERR_NO_NODES":
                    ref_desc += "(NO_NODES)"
            elif ref_issue:
                ref_desc = ref_issue

            if status != "OK":
                # Allow CAND_ERR_NO_NODES if Reference also has no nodes or is empty or returns 400
                if status == "CAND_ERR_NO_NODES" and (ref_desc == "ERR_NO_NODES" or "NO_NODES" in ref_desc or ref_desc == "EMPTY" or "HTTP_400" in ref_desc):
                     status = "OK"
                     comp = "MATCH_NO_NODES"
                else:
                     failures.append(f"{case_id}: {status} (Ref: {ref_desc})")
            
            # Comparison note
            comp = "MATCH"
            
            cand_proxies = extract_proxies(content1, target)
            ref_proxies = extract_proxies(content2, target)
            
            count1 = len(cand_proxies)
            count2 = len(ref_proxies)
            
            if count1 > 0 and count2 > 0:
                comp = compare_proxy_lists(cand_proxies, ref_proxies)
                if "MISMATCH" in comp or "MISSING" in comp or "EXTRA" in comp:
                    status = "FAIL_COMPARE"
            else:
                if ref_desc != "OK":
                    comp = f"REF_{ref_desc}"
                    if ("NO_NODES" in ref_desc or "HTTP_400" in ref_desc):
                        if count1 > 0:
                            status = "FAIL_REF"
                            failures.append(f"{case_id}: Reference failed: {ref_desc} but Candidate has {count1} nodes")
                        elif cand_issue == "ERR_NO_NODES" or cand_status == 400:
                            status = "OK"
                            comp = "MATCH_NO_NODES"
                elif len(content1) == 0:
                    comp = "EMPTY"
                else:
                    # Simple size comparison as proxy for correctness
                    diff = abs(len(content1) - len(content2))
                    if diff > len(content1) * 0.5: # >50% difference
                        comp = f"SIZE_MISMATCH({len(content1)}vs{len(content2)})"
            
            results[case_id] = f"{status} | {comp}"
            if status == "FAIL_COMPARE":
                 failures.append(f"{case_id}: Comparison failed: {comp}")

    if failures:

        # Don't raise immediately, let other tests run, but mark result
        results["_failures"] = failures
        print(f"E2E Failures: {failures}")

    return results


def print_logs() -> None:
    print("--- Subconvergo Logs ---")
    subprocess.run(["docker", "logs", "tests-subconvergo-1"], cwd=TESTS_DIR)
    print("--- Subconverter Logs ---")
    subprocess.run(["docker", "logs", "tests-subconverter-1"], cwd=TESTS_DIR)
    print("------------------------")

def main() -> None:
    parser = argparse.ArgumentParser(description="Run smoke tests")
    parser.add_argument("-t", "--test", help="Run specific test case (substring match)")
    parser.add_argument("-s", "--skip-build", action="store_true", default=False, help="Skip docker image build step")
    parser.add_argument("--fail-fast", action="store_true", default=True, help="Stop on first failure (default: True)")
    parser.add_argument("--no-fail-fast", dest="fail_fast", action="store_false", help="Don't stop on first failure")
    args = parser.parse_args()

    cases = [
        ("version", test_version),
        ("sub", test_sub),
        ("surge2clash", test_surge2clash),
        ("render", test_render),
        ("profile", test_profile),
        ("ruleset_remote", test_ruleset_remote),
        ("ruleset_compare", test_ruleset_compare_with_subconverter),
        ("filters_regex", test_filters_regex),
        ("exclude_remarks", test_exclude_remarks),
        ("include_remarks", test_include_remarks),
        ("emoji_rule", test_emoji_rule),
        ("rename_node", test_rename_node),
        ("userinfo", test_userinfo),
        ("sub_with_external_config", test_sub_with_external_config),
        ("clash_only_config", test_clash_only_config),
        ("settings_comparison", test_settings_comparison),
        ("e2e_matrix", test_e2e_matrix),
        ("e2e_matrix_exclude", test_e2e_matrix),
        ("e2e_matrix_include", test_e2e_matrix),
        ("e2e_matrix_emoji", test_e2e_matrix),
        ("e2e_matrix_rename", test_e2e_matrix),
        ("e2e_matrix_userinfo", test_e2e_matrix),
    ]

    if args.test:
        cases = [c for c in cases if args.test in c[0]]
        if not cases:
            print(f"No tests matched '{args.test}'")
            return

    # Initial setup
    # Use the first case's pref if available, or base if filtered list is empty (though we return above)
    write_pref(cases[0][0] if cases else "version")
    compose_up(not args.skip_build)
    
    failed = False
    try:
        results = {}
        for case, func in cases:
            print(f"Preparing {case}...")
            write_pref(case)
            restart_services()
            wait_for_service()
            
            print(f"Running {case}...")
            try:
                res = func(case)
                
                results[case] = res
                
                # Check for soft failures (like in e2e_matrix or settings_comparison)
                if isinstance(res, dict) and "_failures" in res:
                    print(f"Test {case} reported failures.")
                    failed = True
                    if args.fail_fast:
                        raise AssertionError(f"Test {case} failed: {res['_failures']}")

            except Exception as e:
                print(f"Test {case} failed with exception: {e}")
                failed = True
                if args.fail_fast:
                    raise

        RESULTS_FILE.write_text(json.dumps(results, indent=2))
        
        if failed:
            print(f"Smoke tests failed. Summary: {RESULTS_FILE}")
            sys.exit(1)
        else:
            print(f"Smoke tests passed. Summary: {RESULTS_FILE}")

    except Exception:
        print_logs()
        raise
    finally:
        # Only print logs if we haven't already (exception handler does it)
        # But finally runs always. Let's just rely on exception handler for logs on crash, 
        # and maybe print logs on exit if failed?
        # The original code printed logs in finally.
        # If we exit via sys.exit(1), finally block runs.
        # We might want to avoid double printing.
        # Let's keep it simple and close to original but ensure cleanup.
        compose_down()
        restore_pref()


if __name__ == "__main__":
    main()
