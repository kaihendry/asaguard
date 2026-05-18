## ADDED Requirements

### Requirement: Run permission test suite
The CLI SHALL execute a predefined set of dry-run probes and confirm that permissions blocked in `settings.json` are actually denied by Claude Code.

#### Scenario: Blocked permission is denied
- **WHEN** a probe targeting a denied permission is run
- **THEN** the probe receives a denial and the test passes

#### Scenario: Blocked permission is unexpectedly allowed
- **WHEN** a probe targeting a denied permission succeeds
- **THEN** the test fails and reports the permission name and observed behaviour

### Requirement: Validate allowlist-only tool access
The CLI SHALL verify that only tools on the configured `allow` list are accessible, and that tools outside it are denied.

#### Scenario: Non-allowlisted tool is blocked
- **WHEN** a probe attempts to invoke a tool not on the allow list
- **THEN** the probe is denied and the check passes

#### Scenario: Allow list is empty (open-by-default risk)
- **WHEN** no allow list is configured and no deny list is present
- **THEN** the check emits a WARN indicating the installation is permissive

### Requirement: Report permission validation results in structured format
The CLI SHALL output permission test results as a table with columns: test name, expected outcome, actual outcome, and pass/fail status.

#### Scenario: All tests pass
- **WHEN** all probes return expected outcomes
- **THEN** the table shows all rows as PASS and the command exits 0

#### Scenario: One or more tests fail
- **WHEN** any probe returns an unexpected outcome
- **THEN** the table marks failing rows as FAIL and the command exits non-zero
