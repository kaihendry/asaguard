## Context

`asaguard check` runs eight security checks and computes a weighted compliance score (0–100). All output goes to stdout/stderr only; there is no mechanism to forward findings to a central platform. The `AI_GUARDRAILS_SIEM_ENDPOINT` v1.0 contract defines a well-known HTTP POST interface used by the companion tool `ai-check-guardrails`. We extend `asaguard` to honour the same contract so organisations can use a single SIEM backend for both tools.

The live endpoint (`https://wnniwdexyj.execute-api.eu-west-2.amazonaws.com/`) was probed and returns HTTP 201 with `{"run_id":"…","sk":"…"}` for valid payloads — both clean runs and runs with findings.

## Goals / Non-Goals

**Goals:**
- POST a schema v1.0 payload to `AI_GUARDRAILS_SIEM_ENDPOINT` after every `asaguard check` run when the var is set.
- Support optional bearer token via `AI_GUARDRAILS_SIEM_TOKEN`.
- Fall back to `siem_endpoint` in `~/.config/ai-check-guardrails/config.json` when env var is absent.
- Map existing `result.Finding` fields to SIEM payload fields without breaking the existing `result.Finding` struct.
- Fail silently (log to stderr, do not alter exit code) on SIEM POST failure.

**Non-Goals:**
- Retry logic (contract explicitly says none in v1).
- Modifying `--json` output shape.
- Adding `remediation` or `resource` fields to internal findings (out of scope; map to empty string).
- Changing `asaguard score` subcommand (only `check` gains SIEM side-effect).

## Decisions

**Decision: new `internal/siem` package, not inlined in scorer**

Rationale: SIEM payload construction involves UUID generation, RFC3339 timestamps, hostname/username lookup, and HTTP transport — orthogonal to scoring logic. A dedicated package keeps scorer.go focused and makes the SIEM layer independently testable.

Alternatives considered: adding a `postToSIEM()` helper inside scorer.go — rejected because it would grow scorer.go by ~80 lines of unrelated HTTP code.

**Decision: derive SIEM `severity` from `result.Level`**

Mapping:
| asaguard Level | SIEM severity |
|----------------|---------------|
| PASS           | INFO          |
| WARN           | WARN          |
| FAIL           | HIGH          |

CRITICAL severity is not emitted by current checks; reserved for future use.

Rationale: clean 1:1 mapping, no information loss, reversible.

**Decision: `type` field = `strings.ToUpper(finding.Check) + "_FINDING"`**

e.g. `settings` → `SETTINGS_FINDING`. Provides namespacing without requiring changes to `result.Finding`.

**Decision: `mode` field = `"monitor"` always**

`asaguard check` does not currently block operations; it reports. Using `enforce` would misrepresent the tool's behaviour. This can be revisited if `--enforce` flag is added later.

**Decision: config file path matches `ai-check-guardrails` exactly**

Path: `~/.config/ai-check-guardrails/config.json`, key `siem_endpoint`. This ensures both tools share the same config without user duplication.

## Risks / Trade-offs

[Slow endpoint] → Mitigation: hard 10 s timeout (per contract); SIEM POST runs after all checks complete so it cannot block findings output.

[Sensitive findings data leaving the machine] → Mitigation: env var / config is opt-in; no endpoint is contacted unless explicitly configured.

[UUID dependency] → Mitigation: generate a pseudo-UUID using `crypto/rand` (stdlib); no new dependencies.

[Findings lack `resource` and `remediation`] → accepted gap; fields sent as empty string. Future work can enrich `result.Finding`.

## Migration Plan

No migration needed — purely additive. Existing users without `AI_GUARDRAILS_SIEM_ENDPOINT` set see no behaviour change.
