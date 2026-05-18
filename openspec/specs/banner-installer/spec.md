## ADDED Requirements

### Requirement: Install session-start banner hook
The CLI SHALL install a `PreToolUse` hook that fires on the first tool call of each session and prints a banner containing a link to the organisation's security policy and training resources.

#### Scenario: Banner displays on session start
- **WHEN** Claude Code starts a new session and invokes its first tool
- **THEN** the hook prints the configured banner text (including policy URL) to the terminal before the tool runs

#### Scenario: Banner only fires once per session
- **WHEN** subsequent tool calls occur in the same session
- **THEN** the hook detects the session is already active and exits silently

### Requirement: Support configurable banner text and URL
The CLI SHALL read banner text, policy URL, and training URL from the policy JSON or command-line flags during installation.

#### Scenario: Custom banner configured via policy JSON
- **WHEN** `policy.json` specifies `banner.text` and `banner.policy_url`
- **THEN** the installed hook script uses those values

#### Scenario: Banner installed with --url flag
- **WHEN** `asaguard install-hooks --banner --url https://policy.example.com` is run
- **THEN** the hook is installed using that URL as the policy link

### Requirement: Uninstall banner hook cleanly
The CLI SHALL remove the banner hook from `settings.json` when `asaguard uninstall-hooks --banner` is run.

#### Scenario: Banner hook uninstalled
- **WHEN** the user runs `asaguard uninstall-hooks --banner`
- **THEN** the banner hook entries are removed from `settings.json`

#### Scenario: No banner hook present
- **WHEN** no banner hook exists in `settings.json`
- **THEN** the CLI reports "no banner hook found" and exits 0
