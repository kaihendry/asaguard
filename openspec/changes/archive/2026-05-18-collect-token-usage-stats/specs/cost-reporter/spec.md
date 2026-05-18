## ADDED Requirements

### Requirement: Compute per-session USD cost from token counts
The CLI SHALL calculate an estimated USD cost for each session using configurable per-million-token prices for input, output, cache-write, and cache-read token categories.

#### Scenario: Session with all token types
- **WHEN** a session has input, output, cache-write, and cache-read tokens
- **THEN** the cost is computed as: (input × $3.00 + output × $15.00 + cache_write × $3.75 + cache_read × $0.30) / 1_000_000

#### Scenario: Session with no tokens
- **WHEN** a session has zero tokens across all categories
- **THEN** the reported cost is $0.00

### Requirement: Report cache savings per session
The CLI SHALL compute cache savings as the difference between what cache-read tokens would have cost at full input price versus the cache-read price.

#### Scenario: Session with cache hits
- **WHEN** a session has cache-read tokens > 0
- **THEN** cache_savings_usd = cache_read_tokens × (input_price − cache_read_price) / 1_000_000

#### Scenario: Session with no cache hits
- **WHEN** a session has zero cache-read tokens
- **THEN** cache_savings_usd is 0.00

### Requirement: Display stats table with --stats flag
The CLI SHALL accept a `--stats` flag on the `tokens` subcommand and print a human-readable table of per-project cost, token breakdown, and cache efficiency when it is set.

#### Scenario: --stats with multiple projects
- **WHEN** `--stats` is passed and transcripts span multiple projects
- **THEN** the output includes one row per project with: total tokens (by category), total cost, cache hit ratio, and cache savings

#### Scenario: --stats with no transcripts
- **WHEN** `--stats` is passed but no transcripts are found
- **THEN** the CLI reports zero projects and exits successfully

### Requirement: Include cost fields in JSON output
The CLI SHALL include `cost_usd` and `cache_savings_usd` as numeric fields in each finding's JSON output when `--json` is also set.

#### Scenario: JSON output with cost fields
- **WHEN** `--json` is passed
- **THEN** each session finding contains `cost_usd` and `cache_savings_usd` numeric fields

### Requirement: Support configurable token prices
The policy configuration SHALL accept optional per-token price overrides for all four categories; when absent, hardcoded defaults are used.

#### Scenario: Custom price in policy YAML
- **WHEN** `policy.yaml` specifies `token_prices.output_per_million: 20.00`
- **THEN** cost calculations use $20.00/M for output tokens

#### Scenario: No price overrides in policy YAML
- **WHEN** `policy.yaml` has no `token_prices` section
- **THEN** cost calculations use default prices ($3.00 input, $15.00 output, $3.75 cache-write, $0.30 cache-read)
