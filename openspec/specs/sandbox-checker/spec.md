## ADDED Requirements

### Requirement: Verify Claude only reads from authorised directories
The CLI SHALL inspect transcript Bash and Read tool calls and confirm that all file paths accessed fall within the policy-defined authorised directory set.

#### Scenario: All accesses within authorised paths
- **WHEN** all file paths in transcript tool calls are under authorised roots
- **THEN** the check passes with no warnings

#### Scenario: Access outside authorised paths detected
- **WHEN** a transcript tool call references a path outside the authorised set
- **THEN** the check fails and reports the path, tool name, and session ID

### Requirement: Verify Claude only writes to authorised directories
The CLI SHALL inspect transcript Write and Edit tool calls and confirm target paths are within policy-authorised write roots.

#### Scenario: Write within authorised path
- **WHEN** a Write or Edit tool call targets a path inside the authorised write root
- **THEN** no violation is recorded

#### Scenario: Write outside authorised path
- **WHEN** a Write or Edit tool call targets a path outside the authorised write root
- **THEN** the check fails and reports the offending path and session ID

### Requirement: Accept authorised path list from policy config
The CLI SHALL read authorised read and write roots from the policy JSON (`~/.config/asaguard/policy.json`), with sensible defaults (project directory, home directory).

#### Scenario: Custom authorised paths in policy
- **WHEN** `policy.json` specifies `sandbox.read_roots` and `sandbox.write_roots`
- **THEN** those values override the defaults for the check

#### Scenario: No policy file present
- **WHEN** `policy.json` is absent
- **THEN** default roots are used and the check proceeds normally
