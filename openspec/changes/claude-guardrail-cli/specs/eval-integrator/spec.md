## ADDED Requirements

### Requirement: Install Anthropic-recommended adversarial-review hooks
The CLI SHALL write `PreToolUse` and `PostToolUse` hook entries into `~/.claude/settings.json` that invoke the local eval scripts recommended by Anthropic for adversarial prompt detection.

#### Scenario: Hooks not yet installed
- **WHEN** the user runs `asaguard install-hooks --evals` and no eval hooks exist
- **THEN** the CLI shows a diff of the proposed settings change and writes it after confirmation

#### Scenario: Hooks already present
- **WHEN** eval hooks are already in `settings.json`
- **THEN** the CLI reports "already installed" and makes no changes

### Requirement: Install private-data-protection hooks
The CLI SHALL install hooks that scan tool call arguments for patterns matching PII (email, phone, SSN, credit card) and emit a WARN before the tool executes.

#### Scenario: PII pattern detected in tool argument
- **WHEN** a tool call argument matches a PII regex and the hook is active
- **THEN** the hook script emits a warning to stderr with the matched field name (not the value)

#### Scenario: No PII in tool arguments
- **WHEN** tool call arguments contain no PII patterns
- **THEN** the hook exits silently with code 0

### Requirement: Uninstall eval hooks cleanly
The CLI SHALL provide `asaguard uninstall-hooks --evals` to remove all installed eval hooks from `settings.json`.

#### Scenario: Uninstall removes eval hooks
- **WHEN** the user runs `asaguard uninstall-hooks --evals`
- **THEN** eval hook entries are removed from `settings.json` and a success message is printed

#### Scenario: Nothing to uninstall
- **WHEN** no eval hooks are present in `settings.json`
- **THEN** the CLI reports "no eval hooks found" and exits 0
