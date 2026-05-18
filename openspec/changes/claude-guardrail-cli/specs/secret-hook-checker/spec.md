## ADDED Requirements

### Requirement: Verify pre-commit hook file exists and is executable
The CLI SHALL check that `.git/hooks/pre-commit` exists in the current working directory repository and is marked executable.

#### Scenario: pre-commit hook present and executable
- **WHEN** `.git/hooks/pre-commit` exists and has executable permissions
- **THEN** the check passes this step

#### Scenario: pre-commit hook missing
- **WHEN** `.git/hooks/pre-commit` does not exist
- **THEN** the check fails with "pre-commit hook not installed"

#### Scenario: pre-commit hook not executable
- **WHEN** `.git/hooks/pre-commit` exists but lacks executable permission
- **THEN** the check fails with "pre-commit hook is not executable"

### Requirement: Detect gitleaks or trufflehog invocation in pre-commit hook
The CLI SHALL read the pre-commit hook script and verify it invokes at least one recognised secret-scanning tool (gitleaks, trufflehog, detect-secrets).

#### Scenario: gitleaks found in hook
- **WHEN** the hook script contains an invocation of `gitleaks`
- **THEN** the check passes secret-scanner detection

#### Scenario: No secret scanner found in hook
- **WHEN** the hook script does not reference any recognised scanner
- **THEN** the check fails with "no secret scanner detected in pre-commit hook"

### Requirement: Check for pre-commit framework configuration
The CLI SHALL detect the presence of `.pre-commit-config.yaml` and verify it references a secret-scanning hook.

#### Scenario: pre-commit-config.yaml references gitleaks
- **WHEN** `.pre-commit-config.yaml` exists and contains a gitleaks or trufflehog repo entry
- **THEN** the check passes

#### Scenario: pre-commit-config.yaml missing
- **WHEN** no `.pre-commit-config.yaml` is found
- **THEN** the check falls back to inspecting the raw hook script per the previous requirement
