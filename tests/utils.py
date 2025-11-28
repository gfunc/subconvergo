import yaml
import json
import re
from typing import List, Dict, Any

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

def count_proxies(content: str, target: str) -> int:
    if not content:
        return 0
    try:
        proxies = extract_proxies(content, target)
        return len(proxies)
    except Exception:
        pass
    return -1

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
