## ADDED Requirements

### Requirement: Detect --dangerously-skip-permissions in transcripts
The CLI SHALL scan JSONL transcripts for any invocation of Claude Code with the `--dangerously-skip-permissions` flag and report each occurrence.

#### Scenario: Flag found in transcript metadata
- **WHEN** a transcript entry records `--dangerously-skip-permissions` in the session launch arguments
- **THEN** the check fails and reports the session ID, date, and user

#### Scenario: No bypass flags in transcripts
- **WHEN** no transcript entries reference permission-bypass flags
- **THEN** the check passes

### Requirement: Detect --no-verify and hook-bypass patterns in Bash calls
The CLI SHALL inspect Bash tool call transcripts for `git commit --no-verify`, `git push --no-verify`, and similar hook-bypass patterns.

#### Scenario: git commit --no-verify detected
- **WHEN** a Bash tool call contains `git commit --no-verify`
- **THEN** the check emits a FAIL with the session ID and raw command

#### Scenario: Normal git commit without bypass
- **WHEN** a Bash tool call contains `git commit` without `--no-verify`
- **THEN** no violation is recorded

### Requirement: Emit machine-readable bypass report
The CLI SHALL output bypass detections as JSON when `--json` is passed, suitable for ingestion by SIEM or CI pipelines.

#### Scenario: JSON output requested
- **WHEN** `asaguard bypass --json` is run and violations are found
- **THEN** stdout is a JSON array of objects with fields: sessionId, date, flag, command

#### Scenario: No violations with JSON flag
- **WHEN** `asaguard bypass --json` finds no violations
- **THEN** stdout is an empty JSON array `[]` and exit code is 0
