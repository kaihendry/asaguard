## ADDED Requirements

### Requirement: Install human-in-the-loop confirmation hooks for sensitive tools
The CLI SHALL install `PreToolUse` hooks that pause execution and prompt the user for explicit confirmation before security-sensitive tool calls proceed.

#### Scenario: Sensitive tool call intercepted
- **WHEN** Claude attempts to call a tool on the HITL watchlist (e.g., Bash with `rm -rf`, git push, curl to external host)
- **THEN** the hook prints a warning and prompts the user to type `yes` to continue or `no` to abort

#### Scenario: Non-sensitive tool call not intercepted
- **WHEN** Claude calls a tool not on the HITL watchlist
- **THEN** the hook exits immediately with code 0, adding no latency

### Requirement: Support configurable HITL watchlist
The CLI SHALL read the list of tools and argument patterns that trigger HITL prompts from the policy JSON.

#### Scenario: Custom tool added to watchlist
- **WHEN** `policy.json` includes a custom tool name in `hitl.watchlist`
- **THEN** calls to that tool are intercepted by the hook

#### Scenario: Default watchlist used when no policy
- **WHEN** no `policy.json` is present
- **THEN** a built-in default watchlist covering high-risk Bash patterns and network tools is used

### Requirement: Log HITL decisions
The CLI hook SHALL append each HITL prompt event (tool name, user decision, timestamp) to a local audit log.

#### Scenario: User approves HITL prompt
- **WHEN** the user types `yes` at the HITL prompt
- **THEN** an APPROVED entry is appended to `~/.claude/asaguard-hitl.log`

#### Scenario: User denies HITL prompt
- **WHEN** the user types `no` at the HITL prompt
- **THEN** a DENIED entry is appended to the audit log and the hook exits non-zero to block the tool
