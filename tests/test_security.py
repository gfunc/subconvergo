"""Security regression smoke cases (black-box HTTP against the running stack).

These pin the landed security hardening end-to-end (see
.scratch/security-hardening/issues/):

- fail-closed token gate on protected endpoints (ticket 05),
- hardened outbound fetcher: loopback/private/link-local/metadata refusal
  (tickets 01/02),
- api_mode refusal of file:// and bare local paths as subscription URLs
  (ticket 03),
- CR/LF sanitization of subscription-controlled remark fields in line-oriented
  outputs (ticket 04).

All cases run with the default test pref (api_mode=true, real token set) —
no pref_modifier needed.
"""

import base64

import requests

from . import infra

WRONG_TOKEN = "definitely-not-the-token"
RULESET_URL_B64 = base64.urlsafe_b64encode(b"http://mock-subscription/test_rules.list").decode()
METADATA_URL_B64 = base64.urlsafe_b64encode(b"http://169.254.169.254/latest/meta-data/").decode()


def _get(path, params=None):
    return requests.get(f"{infra.BASE_URL}/{path.lstrip('/')}", params=params, timeout=30)


def _result(failures, **info):
    if failures:
        return {"_failures": failures}
    return info or {"status": "ok"}


# ---------------------------------------------------------------------------
# /getruleset requires the API token (fails closed)
# ---------------------------------------------------------------------------

def query_getruleset_requires_token():
    params = {"url": RULESET_URL_B64, "type": "clash"}
    return {
        "no_token": _get("/getruleset", params),
        "wrong_token": _get("/getruleset", {**params, "token": WRONG_TOKEN}),
        "valid_token": _get("/getruleset", {**params, "token": infra.TOKEN}),
    }

def validate_getruleset_requires_token(res):
    failures = []
    for label in ("no_token", "wrong_token"):
        code = res[label].status_code
        if code not in (401, 403):
            failures.append(f"/getruleset with {label} must be refused (401/403), got {code}: {res[label].text[:120]}")
    ok = res["valid_token"]
    if ok.status_code != 200:
        failures.append(f"/getruleset with valid token must succeed, got {ok.status_code}: {ok.text[:120]}")
    elif "MATCH,Auto" not in ok.text:
        failures.append("/getruleset with valid token returned unexpected body (missing MATCH,Auto)")
    return _result(failures)


# ---------------------------------------------------------------------------
# /getruleset refuses the cloud metadata address even with a valid token
# ---------------------------------------------------------------------------

def query_getruleset_blocks_metadata_ip():
    return _get("/getruleset", {"url": METADATA_URL_B64, "type": "clash", "token": infra.TOKEN})

def validate_getruleset_blocks_metadata_ip(resp):
    failures = []
    if resp.status_code == 200:
        failures.append("metadata fetch must be refused, got 200")
    # Typical cloud-metadata payload keys must never leak into the response.
    for marker in ("ami-id", "instance-id", "local-ipv4"):
        if marker in resp.text:
            failures.append(f"metadata content leaked into response (found {marker!r})")
    return _result(failures, status=resp.status_code)


# ---------------------------------------------------------------------------
# api_mode rejects file:// / local-path subscription URLs
# ---------------------------------------------------------------------------

def query_sub_rejects_file_url():
    return {
        # /etc/hostname exists in every container; /base/pref.yml holds the API token.
        "hostname": _get("/sub", {"target": "clash", "url": "file:///etc/hostname"}),
        "pref": _get("/sub", {"target": "clash", "url": "file:///base/pref.yml"}),
    }

def validate_sub_rejects_file_url(res):
    failures = []
    for label, resp in res.items():
        if resp.status_code == 200:
            failures.append(f"file:// subscription ({label}) must not succeed, got 200")
        if infra.TOKEN in resp.text:
            failures.append(f"file:// subscription ({label}) leaked the api_access_token")
    if "not allowed in api_mode" not in res["hostname"].text:
        failures.append(
            "file:// refusal should come from the api_mode guard, got: "
            + res["hostname"].text[:160]
        )
    return _result(failures)


# ---------------------------------------------------------------------------
# /sub refuses loopback destinations (self-fetch SSRF)
# ---------------------------------------------------------------------------

def query_sub_blocks_loopback():
    return _get("/sub", {"target": "clash", "url": f"{infra.BASE_URL}/version"})

def validate_sub_blocks_loopback(resp):
    failures = []
    if resp.status_code == 200:
        failures.append("loopback subscription fetch must be refused, got 200")
    if "subconvergo v" in resp.text:
        failures.append("loopback self-fetch returned the /version payload")
    return _result(failures, status=resp.status_code)


# ---------------------------------------------------------------------------
# CR/LF in node remarks cannot inject lines into Quantumult X output
# ---------------------------------------------------------------------------
# tests/mock-data/ss-newline-injection.txt is a base64 ss subscription whose
# node fragment is %0A%5Brewrite_local%5D%0A%5Ehttps%3A%2F%2Fevil — i.e. the
# remark decodes to "\n[rewrite_local]\n^https://evil". Pre-fix, that produced
# an injected [rewrite_local] INI section in quanx output (see ticket 04);
# post-fix the control chars are stripped and the remark renders as inert
# single-line text.

def query_quanx_newline_injection():
    return _get("/sub", {"target": "quanx", "url": f"{infra.MOCK_BASE}/ss-newline-injection.txt"})

def validate_quanx_newline_injection(resp):
    failures = []
    if resp.status_code != 200:
        return {"_failures": [f"expected 200, got {resp.status_code}: {resp.text[:160]}"]}
    lines = resp.text.splitlines()
    if any(l.strip() == "[rewrite_local]" for l in lines):
        failures.append("injected [rewrite_local] section present in quanx output")
    if any(l.lstrip().startswith("^https://evil") for l in lines):
        failures.append("injected rewrite rule line present in quanx output")
    # The node itself must survive sanitization (strip, don't reject).
    if "shadowsocks=1.2.3.4:8388" not in resp.text:
        failures.append("node missing from quanx output after sanitization")
    return _result(failures)


# ---------------------------------------------------------------------------
# All protected endpoints fail closed without a token
# ---------------------------------------------------------------------------

PROTECTED_ENDPOINTS = ["/flushcache", "/readconf", "/render", "/getprofile", "/getruleset"]

def query_protected_endpoints_fail_closed():
    res = {ep: _get(ep) for ep in PROTECTED_ENDPOINTS}
    res["flushcache_wrong"] = _get("/flushcache", {"token": WRONG_TOKEN})
    res["flushcache_valid"] = _get("/flushcache", {"token": infra.TOKEN})
    return res

def validate_protected_endpoints_fail_closed(res):
    failures = []
    for ep in PROTECTED_ENDPOINTS:
        code = res[ep].status_code
        if code not in (401, 403):
            failures.append(f"{ep} without token must be refused (401/403), got {code}")
    code = res["flushcache_wrong"].status_code
    if code not in (401, 403):
        failures.append(f"/flushcache with wrong token must be refused (401/403), got {code}")
    if res["flushcache_valid"].status_code != 200:
        failures.append(f"/flushcache with valid token must succeed, got {res['flushcache_valid'].status_code}")
    return _result(failures)


CASES = [
    infra.StandaloneTestCase(
        name="security_getruleset_requires_token",
        query=query_getruleset_requires_token,
        validate=validate_getruleset_requires_token,
    ),
    infra.StandaloneTestCase(
        name="security_getruleset_blocks_metadata_ip",
        query=query_getruleset_blocks_metadata_ip,
        validate=validate_getruleset_blocks_metadata_ip,
    ),
    infra.StandaloneTestCase(
        name="security_sub_rejects_file_url",
        query=query_sub_rejects_file_url,
        validate=validate_sub_rejects_file_url,
    ),
    infra.StandaloneTestCase(
        name="security_sub_blocks_loopback",
        query=query_sub_blocks_loopback,
        validate=validate_sub_blocks_loopback,
    ),
    infra.StandaloneTestCase(
        name="security_quanx_newline_injection",
        query=query_quanx_newline_injection,
        validate=validate_quanx_newline_injection,
    ),
    infra.StandaloneTestCase(
        name="security_protected_endpoints_fail_closed",
        query=query_protected_endpoints_fail_closed,
        validate=validate_protected_endpoints_fail_closed,
    ),
]
