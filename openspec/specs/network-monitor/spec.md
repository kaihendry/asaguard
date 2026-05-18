## ADDED Requirements

### Requirement: Extract URLs from WebFetch tool calls in transcripts
The CLI SHALL scan transcripts for `WebFetch` tool invocations and collect the URLs accessed.

#### Scenario: WebFetch calls present
- **WHEN** transcripts contain WebFetch tool call entries with URL parameters
- **THEN** the CLI lists all accessed URLs with session ID and timestamp

#### Scenario: No WebFetch calls
- **WHEN** no WebFetch entries exist in the transcripts
- **THEN** the CLI reports zero network requests and exits successfully

### Requirement: Extract URLs from Bash curl/wget calls in transcripts
The CLI SHALL parse Bash tool call entries for `curl`, `wget`, and similar HTTP invocations and extract the target URLs.

#### Scenario: curl invocation detected
- **WHEN** a Bash tool call contains `curl <url>` or `curl -X ... <url>`
- **THEN** the URL is included in the network access report

#### Scenario: Ambiguous or multi-URL curl command
- **WHEN** a Bash tool call contains a curl invocation that cannot be statically parsed
- **THEN** the raw command is included in the report with a WARN tag

### Requirement: Flag access to non-allowlisted domains
The CLI SHALL compare extracted URLs against a policy domain allowlist and flag requests to unlisted domains.

#### Scenario: Approved domain accessed
- **WHEN** a URL's domain is on the policy allowlist
- **THEN** it appears in the report without a warning

#### Scenario: Unapproved domain accessed
- **WHEN** a URL's domain is absent from the policy allowlist
- **THEN** the CLI flags it as WARN with the domain, session ID, and tool name
