## 1. Project Scaffold

- [x] 1.1 Initialise Go module `github.com/hendry/asaguard` with `go mod init` (no external dependencies)
- [x] 1.2 Set up `cmd/asaguard/main.go` with stdlib `flag`-based subcommand dispatch and version flag
- [x] 1.3 Wire subcommands: `check`, `score`, `install-hooks`, `uninstall-hooks`, plus one per check using a simple `os.Args[1]` switch
- [x] 1.4 Create `internal/policy/` package to load and merge `~/.config/asaguard/policy.json` with built-in defaults using `encoding/json`

## 2. Settings Verifier

- [x] 2.1 Implement `internal/settings/` package to parse `settings.json` and `settings.local.json`
- [x] 2.2 Add required-key check against policy baseline
- [x] 2.3 Add forbidden-override detection for `settings.local.json`
- [x] 2.4 Add structured diff output (added / removed / changed keys)
- [x] 2.5 Wire as `asaguard settings` subcommand

## 3. MCP Auditor

- [x] 3.1 Implement `internal/mcps/` package to enumerate MCP entries from `settings.json`
- [x] 3.2 Compare active MCPs against policy allowlist; report unapproved entries
- [x] 3.3 Classify each MCP by exfiltration-risk level using policy category map
- [x] 3.4 Wire as `asaguard mcps` subcommand

## 4. Permission Validator

- [x] 4.1 Define probe test cases in `internal/perms/probes.go` (deny-list and allow-list probes)
- [x] 4.2 Implement dry-run probe runner that interprets `settings.json` allow/deny lists
- [x] 4.3 Render results as a pass/fail table; exit non-zero on any FAIL
- [x] 4.4 Wire as `asaguard perms` subcommand

## 5. Token Tracker

- [x] 5.1 Implement `internal/transcripts/` package to stream-parse JSONL transcript files
- [x] 5.2 Extract per-session token fields (input, output, cache-read, cache-write)
- [x] 5.3 Compute rolling baseline average and flag sessions exceeding N× threshold
- [x] 5.4 Add `--since` flag with ISO 8601 parsing
- [x] 5.5 Wire as `asaguard tokens` subcommand

## 6. Network Monitor

- [x] 6.1 Add WebFetch tool call URL extraction to `internal/transcripts/`
- [x] 6.2 Add Bash tool call parser for `curl`/`wget` URL extraction (best-effort, warn on ambiguous)
- [x] 6.3 Compare extracted domains against policy domain allowlist
- [x] 6.4 Wire as `asaguard network` subcommand

## 7. Bypass Detector

- [x] 7.1 Scan transcript metadata for `--dangerously-skip-permissions` session launch flags
- [x] 7.2 Scan Bash tool calls for `git commit --no-verify` and `git push --no-verify` patterns
- [x] 7.3 Add `--json` output mode (JSON array of violation objects)
- [x] 7.4 Wire as `asaguard bypass` subcommand

## 8. Sandbox Checker

- [x] 8.1 Extract Read, Write, Edit, and Bash file-path arguments from transcripts
- [x] 8.2 Validate each path against policy `sandbox.read_roots` and `sandbox.write_roots`
- [x] 8.3 Report out-of-sandbox accesses with session ID and tool name
- [x] 8.4 Wire as `asaguard sandbox` subcommand

## 9. Hook Infrastructure

- [x] 9.1 Implement `internal/hooks/` package: atomic read-modify-write for `settings.json` (write to temp, rename)
- [x] 9.2 Add diff-and-confirm helper that prints proposed changes and waits for user `yes`/`no`
- [x] 9.3 Implement `asaguard install-hooks` and `asaguard uninstall-hooks` top-level commands with `--evals`, `--banner`, `--hitl` flags

## 10. Eval Integrator

- [x] 10.1 Write adversarial-review hook script to `~/.config/asaguard/hooks/eval-review.sh`
- [x] 10.2 Write PII-detection hook script with email/phone/SSN/CC regexes
- [x] 10.3 Implement `install-hooks --evals` to register both scripts as `PreToolUse` hooks

## 11. HITL Enforcer

- [x] 11.1 Write `hitl-prompt.sh` hook script that reads watchlist from env/config and prompts for confirmation
- [x] 11.2 Implement audit log append (`~/.claude/asaguard-hitl.log`) for APPROVED/DENIED decisions
- [x] 11.3 Implement `install-hooks --hitl` to register the script as a `PreToolUse` hook

## 12. Banner Installer

- [x] 12.1 Write `banner.sh` hook script that fires once per session using a session-ID sentinel file
- [x] 12.2 Support `--url` flag and `policy.json` `banner.*` config during installation
- [x] 12.3 Implement `install-hooks --banner` and `uninstall-hooks --banner`

## 13. Secret Hook Checker

- [x] 13.1 Implement `internal/secrets/` package to check `.git/hooks/pre-commit` existence and executable bit
- [x] 13.2 Read hook script and scan for gitleaks/trufflehog/detect-secrets invocations
- [x] 13.3 Detect `.pre-commit-config.yaml` and check for secret-scanner repo entries
- [x] 13.4 Wire as `asaguard secrets` subcommand

## 14. Compliance Scorer

- [x] 14.1 Implement `internal/scorer/` package to run all checks and collect pass/warn/fail results
- [x] 14.2 Apply weighted scoring with policy-overridable weights; compute 0–100 total
- [x] 14.3 Map score to CRITICAL / AT RISK / ACCEPTABLE / GOOD tiers with colour output
- [x] 14.4 Add `--json` output mode (total, status, checks array)
- [x] 14.5 Wire as `asaguard score` subcommand; also invoked by `asaguard check` for summary

## 15. Integration and Polish

- [x] 15.1 Implement `asaguard check` to run all checks sequentially and print a summary table
- [x] 15.2 Add `--json` global flag plumbed through all subcommands for CI/SIEM ingestion
- [x] 15.3 Write unit tests for transcript parser, settings differ, and scoring logic
- [x] 15.4 Write `README.md` with install instructions, usage examples, and policy JSON reference
- [x] 15.5 Add `Makefile` targets: `build`, `test`, `install`, `lint`
