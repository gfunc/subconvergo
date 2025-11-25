import base64
import requests
import yaml
import json
from typing import Tuple, List, Any
from . import infra
from . import utils
from .test_standalone import (
    setup_exclude_remarks,
    setup_include_remarks,
    setup_emoji_rule,
    setup_rename_node,
    setup_userinfo,
)

def detect_content_issue(content: str):
    if not content or not content.strip():
        return "EMPTY"
    if "doesn't contain any valid node info" in content:
        return "ERR_NO_NODES"
    return None

def setup_ruleset_compare(pref):
    pref["rulesets"]["update_ruleset_on_request"] = True

def fetch_ruleset_subconvergo() -> Tuple[str, int]:
    url_plain = f"{infra.MOCK_BASE}/test_rules.list"
    encoded = base64.urlsafe_b64encode(url_plain.encode()).decode()
    r = requests.get(f"{infra.BASE_URL}/getruleset", params={"url": encoded, "type": "clash"}, timeout=30)
    return r.text, r.status_code

def fetch_ruleset_subconverter() -> Tuple[str, int]:
    url_plain = f"{infra.MOCK_BASE}/test_rules.list"
    encoded = base64.urlsafe_b64encode(url_plain.encode()).decode()
    r = infra.api_get_subconverter("/getruleset", params={"url": url_plain, "type": "clash"}, expected=[200, 400])
    if r.status_code != 200:
        r = infra.api_get_subconverter("/getruleset", params={"url": encoded, "type": "clash"}, expected=[200, 400])
    return r.text, r.status_code

def validate_ruleset(res_go: Tuple[str, int], res_cv: Tuple[str, int]) -> dict:
    text_go, status_go = res_go
    text_cv, status_cv = res_cv
    
    infra.save_result("ruleset_compare", text_go, "subconvergo.txt")
    infra.save_result("ruleset_compare", text_cv, "subconverter.txt")

    if status_go != 200:
        raise AssertionError(f"subconvergo failed: {status_go}")
    if status_cv != 200:
        return {"skipped": True, "status": status_cv}
    
    b1 = text_go.strip().splitlines()
    b2 = text_cv.strip().splitlines()
    if not b1 or not b2 or len(b1) != len(b2) or b1[0] != b2[0] or b1[-1] != b2[-1]:
        raise AssertionError("ruleset output differs from subconverter")
    return {"lines": len(b1)}

def generate_settings_cases() -> List[Tuple[str, dict, str]]:
    base_params = {
        "target": "clash",
        "url": f"{infra.MOCK_BASE}/ss-subscription.txt",
    }
    scenarios = [
        {"emoji": "true"},
        {"emoji": "false"},
        {"list": "true"},
        {"udp": "true"},
        {"tfo": "true"},
        {"scv": "true"},
        {"fdn": "true"},
        {"sort": "true"},
    ]
    cases = []
    for i, settings in enumerate(scenarios):
        case_id = f"settings_{i}_{list(settings.keys())[0]}"
        params = base_params.copy()
        params.update(settings)
        cases.append((case_id, params, "clash"))
    return cases

def generate_matrix_cases() -> List[Tuple[str, dict, str]]:
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
    cases = []
    for src_name, src_file in sources.items():
        for target in targets:
            case_id = f"{src_name}->{target}"
            url = f"{infra.MOCK_BASE}/{src_file}"
            params = {"target": target, "url": url}
            cases.append((case_id, params, target))
    return cases

def fetch_cases_subconvergo(cases: List[Tuple[str, dict, str]]) -> dict:
    results = {}
    for case_id, params, _ in cases:
        try:
            r = infra.api_get_subconvergo("/sub", params=params, expected=[200, 400])
            results[case_id] = (r.text, r.status_code, None)
        except Exception as e:
            results[case_id] = ("", 0, str(e))
    return results

def fetch_cases_subconverter(cases: List[Tuple[str, dict, str]]) -> dict:
    results = {}
    for case_id, params, _ in cases:
        try:
            r = infra.api_get_subconverter("/sub", params=params, expected=[200, 400])
            results[case_id] = (r.text, r.status_code, None)
        except Exception as e:
            results[case_id] = ("", 0, str(e))
    return results

def validate_cases(cases: List[Tuple[str, dict, str]], res_go: dict, res_cv: dict, output_subdir: str) -> dict:
    results = {}
    failures = []
    out_dir = infra.RESULTS_DIR / output_subdir
    out_dir.mkdir(parents=True, exist_ok=True)

    for case_id, _, target in cases:
        content1, status1, err1 = res_go.get(case_id, ("", 0, "MISSING"))
        content2, status2, err2 = res_cv.get(case_id, ("", 0, "MISSING"))

        # Save artifacts
        safe_case_id = case_id.replace("->", "_to_").replace("/", "_")
        (out_dir / f"{safe_case_id}_subconvergo.txt").write_text(content1, encoding="utf-8")
        (out_dir / f"{safe_case_id}_subconverter.txt").write_text(content2, encoding="utf-8")

        # Validation logic (adapted from compare_services)
        status = "OK"
        cand_issue = detect_content_issue(content1)
        if err1:
            status = f"FAIL_CANDIDATE({err1})"
        elif status1 != 200:
            if status1 == 400 and ("No valid proxies found" in content1 or "doesn't contain any valid node info" in content1):
                cand_issue = "ERR_NO_NODES"
            else:
                cand_issue = f"HTTP_{status1}"
        
        if cand_issue:
            status = f"CAND_{cand_issue}"

        # Basic validation
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
                if "[Proxy]" not in content1 and "[server_local]" not in content1 and "shadowsocks=" not in content1:
                    if target == "quanx" and "server_remote" not in content1 and "shadowsocks" not in content1:
                        status = "SUSPICIOUS_INI"
                    elif target == "surge" and "[Proxy]" not in content1:
                        status = "SUSPICIOUS_INI"

        # Reference analysis
        ref_desc = "OK"
        ref_issue = detect_content_issue(content2)
        if err2:
            ref_desc = f"ERR({err2})"
        elif status2 != 200:
            ref_desc = f"HTTP_{status2}"
            if ref_issue == "ERR_NO_NODES":
                ref_desc += "(NO_NODES)"
        elif ref_issue:
            ref_desc = ref_issue

        if status != "OK":
            if status == "CAND_ERR_NO_NODES" and (ref_desc == "ERR_NO_NODES" or "NO_NODES" in ref_desc or ref_desc == "EMPTY" or "HTTP_400" in ref_desc):
                status = "OK"
                comp = "MATCH_NO_NODES"
            else:
                failures.append(f"{case_id}: {status} (Ref: {ref_desc})")
                comp = "FAIL"
        else:
            comp = "MATCH"
            cand_proxies = utils.extract_proxies(content1, target)
            ref_proxies = utils.extract_proxies(content2, target)
            count1 = len(cand_proxies)
            count2 = len(ref_proxies)

            if count1 > 0 and count2 > 0:
                comp = utils.compare_proxy_lists(cand_proxies, ref_proxies)
                if "MISMATCH" in comp or "MISSING" in comp or "EXTRA" in comp:
                    status = "FAIL_COMPARE"
            else:
                if ref_desc != "OK":
                    comp = f"REF_{ref_desc}"
                    if ("NO_NODES" in ref_desc or "HTTP_400" in ref_desc):
                        if count1 > 0:
                            status = "FAIL_REF"
                            failures.append(f"{case_id}: Reference failed: {ref_desc} but Candidate has {count1} nodes")
                        elif cand_issue == "ERR_NO_NODES" or status1 == 400:
                            status = "OK"
                            comp = "MATCH_NO_NODES"
                elif len(content1) == 0:
                    comp = "EMPTY"
                else:
                    diff = abs(len(content1) - len(content2))
                    if diff > len(content1) * 0.5:
                        comp = f"SIZE_MISMATCH({len(content1)}vs{len(content2)})"
        
        if status == "FAIL_COMPARE":
            failures.append(f"{case_id}: Comparison failed: {comp}")
        
        results[case_id] = f"{status} | {comp}"

    if failures:
        results["_failures"] = failures
    return results

CASES = [
    infra.ComparisonTestCase(
        name="ruleset_compare",
        subconvergo_func=fetch_ruleset_subconvergo,
        subconverter_func=fetch_ruleset_subconverter,
        validate_func=validate_ruleset,
        pref_modifier=setup_ruleset_compare
    ),
    infra.ComparisonTestCase(
        name="settings_comparison",
        subconvergo_func=lambda: fetch_cases_subconvergo(generate_settings_cases()),
        subconverter_func=lambda: fetch_cases_subconverter(generate_settings_cases()),
        validate_func=lambda r1, r2: validate_cases(generate_settings_cases(), r1, r2, "settings_comparison"),
        pref_modifier=None
    ),
    infra.ComparisonTestCase(
        name="e2e_matrix",
        subconvergo_func=lambda: fetch_cases_subconvergo(generate_matrix_cases()),
        subconverter_func=lambda: fetch_cases_subconverter(generate_matrix_cases()),
        validate_func=lambda r1, r2: validate_cases(generate_matrix_cases(), r1, r2, "matrix/e2e_matrix"),
        pref_modifier=None
    ),
    infra.ComparisonTestCase(
        name="e2e_matrix_exclude",
        subconvergo_func=lambda: fetch_cases_subconvergo(generate_matrix_cases()),
        subconverter_func=lambda: fetch_cases_subconverter(generate_matrix_cases()),
        validate_func=lambda r1, r2: validate_cases(generate_matrix_cases(), r1, r2, "matrix/e2e_matrix_exclude"),
        pref_modifier=setup_exclude_remarks
    ),
    infra.ComparisonTestCase(
        name="e2e_matrix_include",
        subconvergo_func=lambda: fetch_cases_subconvergo(generate_matrix_cases()),
        subconverter_func=lambda: fetch_cases_subconverter(generate_matrix_cases()),
        validate_func=lambda r1, r2: validate_cases(generate_matrix_cases(), r1, r2, "matrix/e2e_matrix_include"),
        pref_modifier=setup_include_remarks
    ),
    infra.ComparisonTestCase(
        name="e2e_matrix_emoji",
        subconvergo_func=lambda: fetch_cases_subconvergo(generate_matrix_cases()),
        subconverter_func=lambda: fetch_cases_subconverter(generate_matrix_cases()),
        validate_func=lambda r1, r2: validate_cases(generate_matrix_cases(), r1, r2, "matrix/e2e_matrix_emoji"),
        pref_modifier=setup_emoji_rule
    ),
    infra.ComparisonTestCase(
        name="e2e_matrix_rename",
        subconvergo_func=lambda: fetch_cases_subconvergo(generate_matrix_cases()),
        subconverter_func=lambda: fetch_cases_subconverter(generate_matrix_cases()),
        validate_func=lambda r1, r2: validate_cases(generate_matrix_cases(), r1, r2, "matrix/e2e_matrix_rename"),
        pref_modifier=setup_rename_node
    ),
    infra.ComparisonTestCase(
        name="e2e_matrix_userinfo",
        subconvergo_func=lambda: fetch_cases_subconvergo(generate_matrix_cases()),
        subconverter_func=lambda: fetch_cases_subconverter(generate_matrix_cases()),
        validate_func=lambda r1, r2: validate_cases(generate_matrix_cases(), r1, r2, "matrix/e2e_matrix_userinfo"),
        pref_modifier=setup_userinfo
    ),
]
