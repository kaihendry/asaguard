## Why

`asaguard check` surfaces compliance findings locally but has no way to ship them to a centralised security platform. Teams running multiple workstations or CI pipelines cannot aggregate or alert on guardrail findings without a SIEM integration. The `AI_GUARDRAILS_SIEM_ENDPOINT` contract (schema v1.0) is already defined and a live endpoint is available; we just need `asaguard` to POST to it.

## What Changes

- `asaguard check` reads `AI_GUARDRAILS_SIEM_ENDPOINT` (and optional `AI_GUARDRAILS_SIEM_TOKEN`) and, when set, POSTs the run result as a schema v1.0 JSON payload after completing all checks.
- A new internal package `internal/siem` owns payload construction and the HTTP POST (10 s timeout, no retry per contract, failures logged to stderr without changing exit code).
- `result.Finding` is mapped to the SIEM `findings[]` entry format: severity derived from `Level` (PASS→INFO, WARN→WARN, FAIL→HIGH), `module` from `Check`, `description` from `Message`.
- The `asaguard check` `--json` flag output is unchanged; SIEM reporting is a side-effect.
- Config-file fallback: reads `siem_endpoint` from `~/.config/ai-check-guardrails/config.json` when the env var is absent (per contract spec).

## Capabilities

### New Capabilities

- `siem-reporter`: POST audit findings to a remote SIEM endpoint using the AI_GUARDRAILS_SIEM_ENDPOINT v1.0 contract; supports bearer-token auth and config-file fallback.

### Modified Capabilities

- `compliance-scorer`: `RunCheck` and `Run` gain optional SIEM side-effect after computing the score report.

## Impact

- New file: `internal/siem/siem.go` (no external dependencies beyond stdlib)
- Modified: `internal/scorer/scorer.go` — calls `siem.Report()` when endpoint is configured
- No changes to CLI flags, output format, or exit codes
- Dependency: none (stdlib `net/http`, `encoding/json`, `os/user`)
