#!/usr/bin/env python3
"""Minimal docker-based smoke test runner for subconvergo."""

import argparse
import sys
import json
import subprocess
from . import infra
from . import test_standalone
from . import test_comparison

def main():
    parser = argparse.ArgumentParser(description="Run smoke tests")
    parser.add_argument("-t", "--test", help="Run specific test case (substring match)")
    parser.add_argument("-s", "--skip-build", action="store_true", default=False, help="Skip docker image build step")
    parser.add_argument("--fail-fast", action="store_true", default=True, help="Stop on first failure (default: True)")
    parser.add_argument("--no-fail-fast", dest="fail_fast", action="store_false", help="Don't stop on first failure")
    args = parser.parse_args()

    all_cases = test_standalone.CASES + test_comparison.CASES

    if args.test:
        all_cases = [c for c in all_cases if args.test in c.name]
        if not all_cases:
            print(f"No tests matched '{args.test}'")
            return

    # Initial setup
    infra.ensure_dirs()
    # Use the first case's pref if available
    first_case = all_cases[0] if all_cases else None
    if first_case:
        infra.write_pref(first_case.pref_modifier)
    
    infra.compose_up(not args.skip_build)
    
    failed = False
    results = {}
    
    try:
        for case in all_cases:
            print(f"Preparing {case.name}...")
            infra.write_pref(case.pref_modifier)
            infra.restart_services()
            infra.wait_for_service()
            
            print(f"Running {case.name}...")
            try:
                if isinstance(case, infra.StandaloneTestCase):
                    res = case.query()
                    # Optional: save artifact if resp is a Response object
                    if hasattr(res, "text"):
                        infra.save_result(case.name, res.text)
                    
                    val_res = case.validate(res)
                    results[case.name] = val_res
                elif isinstance(case, infra.ComparisonTestCase):
                    res_go = case.subconvergo_func()
                    res_cv = case.subconverter_func()
                    val_res = case.validate_func(res_go, res_cv)
                    results[case.name] = val_res
                else:
                    raise ValueError(f"Unknown test case type: {type(case)}")
                
                if isinstance(val_res, dict) and "_failures" in val_res:
                    print(f"Test {case.name} reported failures.")
                    failed = True
                    if args.fail_fast:
                        raise AssertionError(f"Test {case.name} failed: {val_res['_failures']}")

            except Exception as e:
                print(f"Test {case.name} failed with exception: {e}")
                failed = True
                if args.fail_fast:
                    raise

        infra.RESULTS_FILE.write_text(json.dumps(results, indent=2))
        
        if failed:
            print(f"Smoke tests failed. Summary: {infra.RESULTS_FILE}")
            sys.exit(1)
        else:
            print(f"Smoke tests passed. Summary: {infra.RESULTS_FILE}")

    except Exception:
        print("--- Logs ---")
        subprocess.run(["docker", "logs", "tests-subconvergo-1"], cwd=infra.TESTS_DIR)
        subprocess.run(["docker", "logs", "tests-subconverter-1"], cwd=infra.TESTS_DIR)
        raise
    finally:
        infra.compose_down()
        infra.restore_pref()

if __name__ == "__main__":
    main()
