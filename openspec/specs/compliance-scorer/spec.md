## ADDED Requirements

### Requirement: Aggregate check results into a weighted compliance score
The CLI SHALL run all checks, apply policy-defined weights to each, produce a total score from 0–100, and — when `AI_GUARDRAILS_SIEM_ENDPOINT` is configured — POST the run result to the SIEM endpoint after all checks complete.

#### Scenario: All checks pass
- **WHEN** every check returns PASS
- **THEN** the total score is 100 and the status is GOOD

#### Scenario: Some checks fail
- **WHEN** one or more checks return FAIL or WARN
- **THEN** the score is reduced proportionally by the failed check weights and a breakdown is shown

#### Scenario: Critical check fails
- **WHEN** a check marked `critical: true` in policy fails
- **THEN** the overall status is CRITICAL regardless of the numeric score

#### Scenario: SIEM endpoint configured
- **WHEN** `AI_GUARDRAILS_SIEM_ENDPOINT` is set and `asaguard check` completes
- **THEN** the SIEM reporter is invoked with the completed score report after output is printed

### Requirement: Display per-check score breakdown
The CLI SHALL print a table showing each check name, its weight, its status (PASS / WARN / FAIL), and its weighted contribution to the total score.

#### Scenario: Score table rendered
- **WHEN** `asaguard score` is run
- **THEN** a table is printed with columns: Check, Weight, Status, Score

#### Scenario: Score table in JSON mode
- **WHEN** `asaguard score --json` is run
- **THEN** output is a JSON object with `total`, `status`, and `checks` array

### Requirement: Support configurable check weights via policy JSON
The CLI SHALL allow overriding the default weight of any check via `~/.config/asaguard/policy.json`.

#### Scenario: Weight overridden in policy
- **WHEN** `policy.json` sets `weights.bypass_detector: 30`
- **THEN** the bypass-detector check contributes up to 30 points to the total

#### Scenario: Default weights used when no policy
- **WHEN** no policy JSON is present
- **THEN** equal weights are distributed across all checks

### Requirement: Assign performance tier labels
The CLI SHALL map score ranges to human-readable tiers: CRITICAL (0–49), AT RISK (50–69), ACCEPTABLE (70–84), GOOD (85–100).

#### Scenario: Score in GOOD tier
- **WHEN** total score is 90
- **THEN** the CLI prints "GOOD (90/100)" in green

#### Scenario: Score in CRITICAL tier
- **WHEN** total score is 40
- **THEN** the CLI prints "CRITICAL (40/100)" in red and exits non-zero
