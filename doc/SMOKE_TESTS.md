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

The suite covers a wide range of functionality:

1.  **`version`**: Verifies the `/version` endpoint returns the correct service name.
2.  **`sub`**: Tests basic subscription conversion (e.g., SS to Clash) with flags like `udp` and `tfo`.
3.  **`render`**: Validates the `/render` endpoint for template rendering.
4.  **`profile`**: Tests loading preset profiles via `/getprofile`.
5.  **`ruleset_remote`**: Checks fetching and formatting of remote rulesets.
6.  **`ruleset_compare`**: Compares ruleset output directly against the C++ subconverter.
7.  **`filters_regex`**: Verifies `include`/`exclude` filters using regex.
8.  **`emoji_rule`**: Tests emoji addition based on regex rules.
9.  **`rename_node`**: Tests node renaming functionality.
10. **`sub_with_external_config`**: Validates merging of external configuration files.
11. **`clash_only_config`**: Tests parsing of local Clash config files as subscriptions.
12. **`settings_comparison`**: Runs a matrix of settings (UDP, TFO, SCV, etc.) against both implementations to ensure identical behavior.
13. **`e2e_matrix`**: A comprehensive end-to-end test matrix that converts every source format (SS, SSR, VMess, etc.) to every target format (Clash, Surge, sing-box, etc.) and compares the output with the C++ version.

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
