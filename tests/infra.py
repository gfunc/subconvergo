import time
import subprocess
import requests
import yaml
import json
from pathlib import Path
from typing import Optional, Dict, Any
from dataclasses import dataclass
from typing import Callable, Optional, Any

# Constants
TESTS_DIR = Path(__file__).resolve().parent
BASE_DIR = TESTS_DIR / "base"
PREF_PATH = BASE_DIR / "pref.yml"
RESULTS_DIR = TESTS_DIR / "results"
RESULTS_FILE = RESULTS_DIR / "smoke_summary.json"
COMPOSE_FILE = TESTS_DIR / "docker-compose.test.yml"
BASE_URL = "http://127.0.0.1:25500"
SUBCONVERTER_URL = "http://127.0.0.1:25550"
MOCK_BASE = "http://mock-subscription"
TOKEN = "password"
TEMPLATE_PATH = "/base/base/test_template.tpl"

ORIGINAL_PREF_EXISTS = PREF_PATH.exists()
ORIGINAL_PREF = PREF_PATH.read_text() if ORIGINAL_PREF_EXISTS else ""

def ensure_dirs():
    RESULTS_DIR.mkdir(parents=True, exist_ok=True)

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

def write_pref(modifier=None) -> None:
    config = base_pref()
    if modifier:
        modifier(config)
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

def _api_get(base: str, path: str, *, params=None, expected=200) -> requests.Response:
    path = path.lstrip("/")
    resp = requests.get(f"{base}/{path}", params=params, timeout=30)
    codes = expected if isinstance(expected, (list, tuple)) else (expected,)
    if resp.status_code not in codes:
        raise AssertionError(f"GET {path} returned {resp.status_code}: {resp.text[:200]}")
    return resp

def api_get_subconvergo(path: str, *, params=None, expected=200) -> requests.Response:
    return _api_get(BASE_URL, path, params=params, expected=expected)

def api_get_subconverter(path: str, *, params=None, expected=200) -> requests.Response:
    return _api_get(SUBCONVERTER_URL, path, params=params, expected=expected)

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

def reload_config(modifier=None) -> str:
    write_pref(modifier)
    resp = api_get_subconvergo("/readconf", params={"token": TOKEN})
    resp2 = api_get_subconverter("/readconf", params={"token": TOKEN})
    return resp.text.strip() + " " + resp2.text.strip()

@dataclass
class StandaloneTestCase:
    name: str
    query: Callable[[], Any]
    validate: Callable[[Any], Any]
    pref_modifier: Optional[Callable[[Any], None]] = None

@dataclass
class ComparisonTestCase:
    name: str
    subconvergo_func: Callable[[], Any]
    subconverter_func: Callable[[], Any]
    validate_func: Callable[[Any, Any], Any]
    pref_modifier: Optional[Callable[[Any], None]] = None
