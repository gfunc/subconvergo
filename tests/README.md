# Test Harness

This directory contains integration resources for validating the subconvergo service inside `tests/docker-compose.test.yml`.

## Layout

- `docker-compose.test.yml` – spins up the Go service plus an nginx container that serves mock subscriptions from `tests/mock-data`.
- `base/` – mounted into the app container as `/app` so that custom pref/profiles/templates are available at runtime.
- `scripts/` – Python utilities for preparing resources and exercising every HTTP endpoint exposed in `main.go`.
- `results/` – JSON summaries produced by the Python harness.

## Python utilities

Install the lightweight dependency set:

```bash
python -m venv .venv
source .venv/bin/activate
pip install -r tests/requirements.txt
```

### Smoke script

`python -m tests.scripts.smoke` keeps things intentionally small:

1. Writes a fresh `tests/base/pref.yml` for each handler under test (the service reloads it via
	`/readconf`, so every case exercises a unique configuration).
2. Runs `docker compose -f tests/docker-compose.test.yml up --build -d` once, waits for
	`http://127.0.0.1:25500/version`, and tears everything down afterwards.
3. Hits every handler (`/version`, `/readconf`, `/sub`, `/render`, `/getprofile`, `/getruleset`) against
	`localhost:25500` while the Go server fetches data from the `mock-subscription` container
	(`http://mock-subscription/...`).
4. Parses rendered configs with PyYAML so the checks focus on proxies, proxy-groups, and rules rather than
	brittle string matches.
5. Restores the original `pref.yml` when finished and drops a concise summary in
	`tests/results/smoke_summary.json`.

Because the script now owns the whole lifecycle there are no extra knobs to tune—just run it whenever you
need to smoke-test the handlers.
