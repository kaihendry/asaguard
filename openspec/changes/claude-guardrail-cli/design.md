## Context

Claude Code installations on developer workstations have no centralized enforcement surface. Security teams must manually audit each machine for misconfigured permissions, unapproved MCPs, and missing pre-commit hooks. This CLI fills that gap by running a structured set of checks and optionally installing remediation hooks, all from a single `asaguard` binary written in Go.

The tool reads Claude Code's well-defined file layout (`~/.claude/settings.json`, `~/.claude/projects/*/` transcripts, `.git/hooks/`) and the Claude Code hooks system to gather evidence without requiring elevated privileges.

## Goals / Non-Goals

**Goals:**
- Offline, zero-network auditing of a local Claude Code install
- Hook installation for real-time enforcement (with explicit user consent)
- Weighted compliance score surfaceable in CI/CD or dashboards
- Single static Go binary — no runtime dependencies beyond the OS

**Non-Goals:**
- Centralized telemetry or data collection to a remote server
- Replacing Claude Code's built-in permission system
- Auditing non-Claude AI tools (ChatGPT, Copilot, etc.)
- Windows support in v1

## Decisions

### 1. Single binary, subcommand-per-check architecture

Each check (`settings`, `mcps`, `perms`, `tokens`, `network`, `hooks`, `sandbox`, `hitl`, `bypass`, `banner`, `secrets`, `score`) is a subcommand. A top-level `asaguard check` runs all of them and prints a summary.

**Rationale**: Operators can run individual checks in CI steps or wire them to alerting without pulling the whole suite. Alternatives: plugin model (too heavy), single monolithic command (not composable).

### 2. Transcript parsing via JSONL, not SQLite

Claude Code stores transcripts as JSONL files under `~/.claude/projects/`. We parse these directly.

**Rationale**: No additional dependency, transcripts are human-readable, format is stable. Risk: large transcript directories may be slow — mitigated by streaming parse and a `--since` flag.

### 3. Hook injection via `settings.json` patch

`asaguard install-hooks` edits the user's `~/.claude/settings.json` to add `PreToolUse`/`PostToolUse` hooks. It shows a diff and requires confirmation before writing.

**Rationale**: The hooks system is the only sanctioned extension point. We never modify Claude Code binaries. Rollback is `asaguard uninstall-hooks`.

### 4. Scoring: weighted sum, JSON-configurable weights

Default weights ship in the binary; an optional `~/.config/asaguard/policy.json` overrides them.

**Rationale**: Different orgs have different risk tolerances. A hard-coded score would need a release to tune. Config file is optional — sane defaults work out of the box.

### 5. Go stdlib only

No HTTP client libraries, no GUI frameworks, no external scan engines.

**Rationale**: Minimise supply-chain surface; the tool audits supply-chain risk, it should model good practices itself.

## Risks / Trade-offs

- **Transcript schema changes** → Mitigation: version-detect transcript format; emit a warning rather than crashing on unknown fields.
- **settings.json hot-reload by Claude Code** → Mitigation: write atomically (write temp file, rename); document that Claude Code must be restarted after hook install.
- **False positives in token anomaly detection** → Mitigation: baseline window is configurable; default threshold is conservative (3× rolling average).
- **Pre-commit hook fragility** → Mitigation: we check hook existence and SHA, not execution; we don't run the hook ourselves.
- **Privilege escalation via hook scripts** → Mitigation: hooks we install are read-only reporters; we never inject `shell: true` commands with user-supplied data.

## Migration Plan

1. Ship binary via `go install`
2. `asaguard check` is always read-only; no state modified without explicit subcommand
3. Rollback: `asaguard uninstall-hooks` removes all injected hooks; binary can be deleted with no side-effects

## Open Questions

- Should `asaguard score` emit SARIF for GitHub Advanced Security integration? (post-v1 candidate)
- Policy JSON: ship a default policy from a well-known URL, or keep it fully local? (lean local for v1)
