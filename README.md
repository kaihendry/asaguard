# asaguard

A Claude Code security guardrail CLI for SecEng teams. Audits local Claude Code installations for misconfigured settings, unapproved MCPs, permission bypasses, and missing secret-scanning hooks. Produces a weighted compliance score and can install real-time enforcement hooks.

## Install

```sh
go install github.com/hendry/asaguard/cmd/asaguard@latest
```

Or build from source:

```sh
make build
```

## Usage

```sh
# Run all checks and print a compliance score
asaguard check

# Run individual checks
asaguard settings
asaguard mcps
asaguard perms
asaguard tokens --since 2026-01-01
asaguard network
asaguard bypass
asaguard sandbox
asaguard secrets
asaguard score

# JSON output (for CI / SIEM ingestion)
asaguard check --json
asaguard bypass --json

# Install enforcement hooks into ~/.claude/settings.json
asaguard install-hooks --evals
asaguard install-hooks --hitl
asaguard install-hooks --banner --url https://policy.example.com

# Remove hooks
asaguard uninstall-hooks --evals --hitl --banner
```

## SIEM reporting

`asaguard check` (and `asaguard score`) can POST audit findings to any HTTP endpoint that speaks the `AI_GUARDRAILS_SIEM_ENDPOINT` v1.0 schema:

```sh
export AI_GUARDRAILS_SIEM_ENDPOINT=https://<your-endpoint>/
export AI_GUARDRAILS_SIEM_TOKEN=<bearer-token>   # optional
asaguard check
```

Each run POSTs a JSON payload containing the run ID, timestamp, host, user, all findings with severity, and the compliance score. Failures are logged to stderr and do not change the exit code.

Config-file alternative — set `siem_endpoint` in `~/.config/ai-check-guardrails/config.json`:

```json
{
  "siem_endpoint": "https://<your-endpoint>/"
}
```

The environment variable takes precedence over the config file.

## Policy configuration

Create `~/.config/asaguard/policy.json` to override defaults:

```json
{
  "required_settings_keys": ["permissions"],
  "locked_settings_keys": ["permissions"],
  "approved_mcps": ["filesystem", "brave-search"],
  "mcp_risk_categories": {
    "custom-exfil-mcp": "high"
  },
  "allowed_domains": ["api.anthropic.com", "github.com"],
  "sandbox_read_roots": ["/home/user/projects"],
  "sandbox_write_roots": ["/home/user/projects"],
  "hitl_watchlist": ["rm -rf", "git push", "curl "],
  "token_spike_multiple": 3.0,
  "weights": {
    "settings": 15,
    "mcps": 15,
    "perms": 15,
    "tokens": 10,
    "network": 10,
    "bypass": 15,
    "sandbox": 10,
    "secrets": 10
  },
  "banner": {
    "text": "Reminder: follow your AI usage policy before every session.",
    "policy_url": "https://wiki.example.com/ai-policy"
  }
}
```

All fields are optional; built-in defaults are used for any omitted fields.

## Compliance score tiers

| Score  | Tier       |
|--------|------------|
| 85–100 | GOOD       |
| 70–84  | ACCEPTABLE |
| 50–69  | AT RISK    |
| 0–49   | CRITICAL   |

## Hook scripts

`asaguard install-hooks` writes shell scripts to `~/.config/asaguard/hooks/` and registers them in `~/.claude/settings.json`. All hooks are read-only reporters or interactive prompts — they never silently modify files.

| Flag      | Script              | Event        | Behaviour                                      |
|-----------|---------------------|--------------|------------------------------------------------|
| `--evals` | `eval-review.sh`    | PreToolUse   | Logs tool calls to `~/.claude/asaguard-eval.log` |
| `--evals` | `pii-detect.sh`     | PreToolUse   | Warns on PII patterns in tool inputs           |
| `--hitl`  | `hitl-prompt.sh`    | PreToolUse   | Prompts for y/n on sensitive Bash/network calls |
| `--banner`| `banner.sh`         | PreToolUse   | Prints policy link once per session            |

## Checks reference

| Subcommand  | What it checks |
|-------------|----------------|
| `settings`  | Required keys present, locked keys not overridden in `settings.local.json` |
| `mcps`      | Active MCPs vs. approved allowlist; exfiltration-risk classification |
| `perms`     | Allow/deny list sanity; open-by-default detection |
| `tokens`    | Per-session token spikes vs. rolling average |
| `network`   | URLs accessed via WebFetch/curl vs. approved domain list |
| `bypass`    | `--dangerously-skip-permissions`, `--no-verify` in transcripts |
| `sandbox`   | File reads/writes outside authorised path roots |
| `secrets`   | `.git/hooks/pre-commit` existence, executability, and scanner invocation |
