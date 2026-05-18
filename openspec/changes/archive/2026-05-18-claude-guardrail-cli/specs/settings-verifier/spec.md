## ADDED Requirements

### Requirement: Detect missing required settings keys
The CLI SHALL read `~/.claude/settings.json` and `~/.claude/settings.local.json` and compare them against a policy-defined set of required keys, reporting any that are absent.

#### Scenario: Required key is present
- **WHEN** `settings.json` contains all keys listed in policy
- **THEN** the check passes with no warnings

#### Scenario: Required key is missing
- **WHEN** `settings.json` is missing a policy-required key
- **THEN** the check fails and names the missing key(s)

### Requirement: Detect forbidden overrides in settings.local.json
The CLI SHALL flag any key in `settings.local.json` that overrides a value locked by policy.

#### Scenario: Override of a locked key
- **WHEN** `settings.local.json` sets a key that policy marks as locked
- **THEN** the check fails and reports the key and both values (policy vs. actual)

#### Scenario: No forbidden overrides present
- **WHEN** `settings.local.json` contains only permitted keys
- **THEN** the check passes

### Requirement: Report settings drift from policy baseline
The CLI SHALL produce a structured diff between the observed settings and the policy baseline, showing added, removed, and changed keys.

#### Scenario: Settings match baseline exactly
- **WHEN** all observed settings match the policy baseline
- **THEN** the diff output is empty and the check passes

#### Scenario: Settings have unexpected additions
- **WHEN** `settings.json` contains keys not present in the policy baseline
- **THEN** the diff highlights the unexpected keys as warnings
