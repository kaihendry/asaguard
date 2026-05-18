## Why

Security teams lack a unified CLI tool to audit and enforce Claude Code safety posture across developer workstations. Without automated guardrails, misconfigured permissions, unapproved MCPs, and secret leaks go undetected until after an incident.

## What Changes

- New standalone CLI (`asaguard`) written in Go that runs checks against a local Claude Code installation
- Hook integration with Claude Code's hooks system for real-time enforcement
- Scoring system to surface compliance posture per user/team
- Pre-commit hook verification for secret scanning (gitleaks, trufflehog)

## Capabilities

### New Capabilities

- `settings-verifier`: Inspect and validate `settings.json` / `settings.local.json` for required keys, forbidden overrides, and policy drift
- `mcp-auditor`: Enumerate installed MCPs and skills; flag any not on the approved shortlist; detect exfiltration-risk tools
- `permission-validator`: Execute a test suite that confirms specific dangerous permissions are blocked as configured
- `token-tracker`: Parse Claude Code transcripts to detect irregular token-usage spikes or suspicious patterns
- `network-monitor`: Hook into `WebFetch` / Bash `curl` calls to log and alert on external URL access
- `eval-integrator`: Install and manage Anthropic-recommended adversarial-review and data-protection hooks
- `sandbox-checker`: Verify that Claude only reads/writes directories within an authorized set
- `hitl-enforcer`: Detect security-sensitive tool calls and inject human-in-the-loop confirmation prompts
- `bypass-detector`: Scan transcripts and CLI invocations for `--dangerously-skip-permissions` and similar flags
- `banner-installer`: Install pre/post-session hooks that display policy links and training reminders
- `secret-hook-checker`: Confirm `.git/hooks/pre-commit` (gitleaks, trufflehog) is active and unmodified
- `compliance-scorer`: Aggregate check results into a weighted score with pass/warn/fail breakdown

### Modified Capabilities

## Impact

- New Go module at repo root; no changes to existing code
- Reads `~/.claude/settings.json`, `~/.claude/projects/*/` transcripts, and `.git/hooks/`
- Installs hooks into `~/.claude/settings.json` (with user confirmation)
- No external runtime dependencies beyond the Go standard library (policy config uses `encoding/json`)
