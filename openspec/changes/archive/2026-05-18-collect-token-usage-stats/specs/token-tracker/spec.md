## MODIFIED Requirements

### Requirement: Parse transcript token usage
The CLI SHALL read Claude Code JSONL transcripts from `~/.claude/projects/` and extract per-session token counts as four separate categories: input (uncached), output, cache-write, and cache-read.

#### Scenario: Transcripts present and parseable
- **WHEN** transcript files exist and contain token usage fields
- **THEN** the CLI produces a summary with separate input, output, cache-write, and cache-read counts per session, plus total cost and cache savings

#### Scenario: No transcripts found
- **WHEN** no transcript files exist under the projects directory
- **THEN** the CLI reports zero usage and exits successfully

### Requirement: Detect anomalous token spikes
The CLI SHALL flag sessions whose token usage exceeds a configurable multiple of the rolling baseline average.

#### Scenario: Session within normal range
- **WHEN** a session's token count is within the threshold multiple of the baseline
- **THEN** no alert is raised for that session

#### Scenario: Session exceeds threshold
- **WHEN** a session's token count is greater than N× the rolling average (default N=3)
- **THEN** the CLI emits a WARN with the session ID, date, and token count

### Requirement: Support --since flag for time-bounded analysis
The CLI SHALL accept a `--since` flag (ISO 8601 date) and restrict analysis to transcripts newer than that date.

#### Scenario: --since filters old transcripts
- **WHEN** `--since 2026-01-01` is passed and some transcripts predate that
- **THEN** only transcripts on or after 2026-01-01 are analysed

#### Scenario: --since with no matching transcripts
- **WHEN** `--since` is set to a future date
- **THEN** the CLI reports zero sessions analysed and exits successfully
