# Smoke Tests

Subconvergo includes a comprehensive smoke test suite designed to validate the service's functionality in a containerized environment. These tests ensure that the Go implementation behaves correctly and maintains parity with the original C++ subconverter.

## Overview

The smoke tests are orchestrated by a Python script (`tests/smoke.py`) that manages the lifecycle of the test environment using Docker Compose.

### Key Components

- **`tests/smoke.py`**: The main test runner. It generates configurations, starts containers, executes test cases, and validates results.
- **`tests/docker-compose.test.yml`**: Defines the test environment, including:
  - `subconvergo`: The Go service under test.
  - `subconverter`: The original C++ service (for parity comparison).
  - `mock-subscription`: An Nginx container serving static subscription files from `tests/mock-data`.
- **`tests/base/`**: Mounted as `/app/base` in the container, allowing dynamic configuration injection.
- **`tests/mock-data/`**: Contains sample subscriptions (SS, SSR, VMess, Clash, etc.) used during testing.

## Running Tests

The easiest way to run the smoke tests is via the Makefile:

```bash
make test
```

This command is equivalent to running:

```bash
python3 -m tests.smoke
```

### Options

You can run specific tests or control the build process using flags:

- `-t, --test <substring>`: Run only test cases matching the substring (e.g., `-t version`).
- `-s, --skip-build`: Skip rebuilding the Docker image (useful for rapid iteration).
- `--no-fail-fast`: Continue running tests even if one fails.

## Test Cases

The suite is organized across multiple test files:

### `test_standalone.py` — Core Functionality

1.  **`version`**: Verifies the `/version` endpoint returns the correct service name.
2.  **`default_flags`**: Tests default flag behavior (UDP, TFO, SCV).
3.  **`explicit_flags`**: Tests explicitly set flag parameters.
4.  **`sub`**: Tests basic subscription conversion (e.g., SS to Clash) with flags like `udp` and `tfo`.
5.  **`surge2clash`**: Tests the `/surge2clash` endpoint alias.
6.  **`render`**: Validates the `/render` endpoint for template rendering.
7.  **`profile`**: Tests loading preset profiles via `/getprofile`.
8.  **`ruleset_remote`**: Checks fetching and formatting of remote rulesets.
9.  **`filters_regex`**: Verifies `include`/`exclude` filters using regex.
10. **`exclude_remarks`**: Tests filtering via `exclude_remarks` config.
11. **`include_remarks`**: Tests filtering via `include_remarks` config.
12. **`emoji_rule`**: Tests emoji addition based on regex rules.
13. **`rename_node`**: Tests node renaming functionality.
14. **`userinfo`**: Tests subscription-userinfo header forwarding.
15. **`relay_migration`**: Tests relay-to-dialer-proxy conversion.
16. **`sub_with_external_config`**: Validates merging of external configuration files.
17. **`clash_only_config`**: Tests parsing of local Clash config files as subscriptions.

### `test_comparison.py` — Parity Comparison

1.  **`ruleset_compare`**: Compares ruleset output directly against the C++ subconverter.
2.  **`edge_cases`**: Tests edge case handling.
3.  **`settings_comparison`**: Runs a matrix of settings (UDP, TFO, SCV, etc.) against both implementations.
4.  **`e2e_matrix`**: Converts every source format to every target format and compares with C++ version.
5.  **`e2e_matrix_exclude`**: E2E matrix with exclude filter.
6.  **`e2e_matrix_include`**: E2E matrix with include filter.
7.  **`e2e_matrix_emoji`**: E2E matrix with emoji enabled.
8.  **`e2e_matrix_rename`**: E2E matrix with rename rules.
9.  **`e2e_matrix_userinfo`**: E2E matrix with userinfo.

### `test_external_config.py` — External Config

Tests external config overrides via YAML, TOML, and INI formats, including `overwrite_original_rules`, import keywords, and template overrides.

### `test_dedup_ignore.py` — Dedup & Ignore Source

1.  **`ignore_source_default`**: Verifies default `ignore_source=true` behavior.
2.  **`ignore_source_false`**: Tests `ignore_source=false`.
3.  **`ignore_source_external_config`**: Tests ignore_source with external config.
4.  **`dedup_proxies`**: Checks that duplicate proxy names are automatically handled.

### `test_flags.py` — Flag Precedence

Tests flag precedence across URL params, Clash sources, VMess/Trojan sources, and Surge/Loon/QuanX targets.

### `test_config_isolation.py` — Config Isolation

Tests that config changes from one request don't leak to subsequent requests.

## Parity Verification

A unique feature of this test suite is the direct comparison with the C++ `subconverter`.

- The `docker-compose.test.yml` spins up the official `tindy2013/subconverter` image on port 25550.
- Tests like `e2e_matrix` and `settings_comparison` send identical requests to both services.
- The outputs are compared for:
  - **Proxy Count**: Ensuring the same number of nodes are generated.
  - **Structure**: Validating YAML/JSON structure.
  - **Content**: Checking for critical fields (server, port, type).

> **Note**: Minor differences in output (e.g., whitespace, field order) are expected and handled by the comparison logic.

## Troubleshooting

If tests fail:

1.  Check the summary file: `tests/results/smoke_summary.json`.
2.  Inspect the generated artifacts in `tests/results/<test_case>/`.
3.  View container logs (printed automatically on failure):
    ```bash
    docker logs tests-subconvergo-1
    docker logs tests-subconverter-1
    ```
