## ADDED Requirements

### Requirement: POST audit findings to SIEM endpoint
The CLI SHALL POST a schema v1.0 JSON payload to `AI_GUARDRAILS_SIEM_ENDPOINT` after every `asaguard check` run when the variable is set.

#### Scenario: Endpoint configured, clean run
- **WHEN** `AI_GUARDRAILS_SIEM_ENDPOINT` is set and all checks pass
- **THEN** a POST is made with `findings: []`, `score` equal to the run score, and `exit_code: 0`

#### Scenario: Endpoint configured, findings present
- **WHEN** `AI_GUARDRAILS_SIEM_ENDPOINT` is set and one or more checks return WARN or FAIL
- **THEN** a POST is made with the findings array populated and the correct `exit_code`

#### Scenario: Endpoint not configured
- **WHEN** `AI_GUARDRAILS_SIEM_ENDPOINT` is unset and no config-file entry exists
- **THEN** no HTTP request is made and behaviour is identical to the current tool

### Requirement: Payload conforms to schema v1.0
The POST body SHALL conform to the `AI_GUARDRAILS_SIEM_ENDPOINT` schema v1.0 contract.

#### Scenario: Required top-level fields present
- **WHEN** a payload is constructed
- **THEN** it contains `schema_version:"1.0"`, `run_id` (UUID), `timestamp` (RFC3339), `host`, `user`, `mode:"monitor"`, `version`, `findings`, `score`, `exit_code`, and `duration_ms`

#### Scenario: Finding fields mapped correctly
- **WHEN** a `result.Finding` with Level FAIL is included
- **THEN** the SIEM finding has `severity:"HIGH"`, `module` equal to the check name, `description` equal to the message, and `type` equal to `<CHECK>_FINDING`

#### Scenario: WARN finding mapped
- **WHEN** a `result.Finding` with Level WARN is included
- **THEN** the SIEM finding has `severity:"WARN"`

### Requirement: Set User-Agent header to asaguard version
The CLI SHALL set the `User-Agent` HTTP header to `asaguard/<version>` on every POST.

#### Scenario: User-Agent present
- **WHEN** a POST is made to the SIEM endpoint
- **THEN** the request includes `User-Agent: asaguard/0.1.0` (or whatever the current binary version is)

### Requirement: Support bearer-token authentication
The CLI SHALL include an `Authorization: Bearer` header when `AI_GUARDRAILS_SIEM_TOKEN` is set.

#### Scenario: Token set
- **WHEN** `AI_GUARDRAILS_SIEM_TOKEN=abc123` and a POST is made
- **THEN** the request includes `Authorization: Bearer abc123`

#### Scenario: Token not set
- **WHEN** `AI_GUARDRAILS_SIEM_TOKEN` is unset
- **THEN** no `Authorization` header is added

### Requirement: Config-file fallback for endpoint
The CLI SHALL read `siem_endpoint` from `~/.config/ai-check-guardrails/config.json` when `AI_GUARDRAILS_SIEM_ENDPOINT` is unset.

#### Scenario: Config file present, env var absent
- **WHEN** `~/.config/ai-check-guardrails/config.json` contains `{"siem_endpoint":"https://example.com/"}` and no env var is set
- **THEN** the POST is made to `https://example.com/`

#### Scenario: Env var takes precedence
- **WHEN** both env var and config file are set
- **THEN** the env var endpoint is used

### Requirement: SIEM failures do not affect exit code
The CLI SHALL log SIEM POST errors to stderr but SHALL NOT change the process exit code due to a SIEM failure.

#### Scenario: Endpoint unreachable
- **WHEN** the endpoint returns a network error or non-2xx status
- **THEN** an error is printed to stderr and the exit code reflects only the check results

### Requirement: 10-second POST timeout
The HTTP POST SHALL time out after 10 seconds per the contract.

#### Scenario: Slow endpoint
- **WHEN** the endpoint does not respond within 10 seconds
- **THEN** the POST is abandoned, an error is logged to stderr, and the CLI exits normally
