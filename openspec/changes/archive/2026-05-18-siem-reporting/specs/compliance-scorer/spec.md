## MODIFIED Requirements

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
