## ADDED Requirements

### Requirement: Enumerate installed MCPs and skills
The CLI SHALL read the Claude Code configuration to produce a complete list of active MCP servers and registered skills.

#### Scenario: MCP servers enumerated
- **WHEN** the user runs `asaguard mcps`
- **THEN** all MCP entries from `settings.json` are printed with their name and transport

#### Scenario: No MCPs installed
- **WHEN** no MCP entries are present in settings
- **THEN** the output reports zero MCPs and exits successfully

### Requirement: Flag non-shortlisted MCPs
The CLI SHALL compare the active MCP list against a policy-maintained allowlist and flag any MCP not present on the list.

#### Scenario: Unapproved MCP detected
- **WHEN** an active MCP is absent from the policy allowlist
- **THEN** the check fails and identifies the unapproved MCP by name and transport URI

#### Scenario: All MCPs approved
- **WHEN** every active MCP is on the allowlist
- **THEN** the check passes

### Requirement: Classify exfiltration-risk tools
The CLI SHALL assign an exfiltration-risk level (high / medium / low) to each detected MCP based on policy-defined categories (e.g., tools with outbound HTTP, filesystem write, or credential access).

#### Scenario: High-risk MCP detected
- **WHEN** an active MCP is classified as high-risk by policy
- **THEN** the report includes a WARN entry with the MCP name and risk reason

#### Scenario: Low-risk MCPs only
- **WHEN** all active MCPs are classified as low-risk
- **THEN** no exfiltration warnings are emitted
