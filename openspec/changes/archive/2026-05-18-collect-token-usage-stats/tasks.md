## 1. Extend token struct and parser

- [x] 1.1 Add `InputTokens`, `CacheWrite`, `CacheRead`, `OutputTokens` separate fields to `sessionTokens` struct in `tokens.go`, removing the collapsed `Input` field
- [x] 1.2 Update `collectTokens` to populate all four fields from `e.Message.Usage` without collapsing cache into input
- [x] 1.3 Update all references to the old `Input` field (spike detection, output formatting)

## 2. Add pricing to policy

- [x] 2.1 Add `TokenPrices` struct to `internal/policy/policy.go` with fields `InputPerMillion`, `OutputPerMillion`, `CacheWritePerMillion`, `CacheReadPerMillion` (float64)
- [x] 2.2 Set hardcoded defaults ($3.00, $15.00, $3.75, $0.30) in `policy.Load()` when fields are zero
- [x] 2.3 Add `token_prices` YAML key to policy schema so overrides are loadable from `policy.yaml`

## 3. Implement cost calculation

- [x] 3.1 Add `CostUSD()` method on `sessionTokens` that computes total cost using a `*policy.TokenPrices` argument
- [x] 3.2 Add `CacheSavingsUSD()` method on `sessionTokens` computing savings = cache_read × (input_price − cache_read_price) / 1e6
- [x] 3.3 Write unit tests for both methods covering zero-token and mixed-token cases

## 4. Add --stats flag and output

- [x] 4.1 Add `--stats` bool flag to `RunTokens` flag set
- [x] 4.2 When `--stats` is set, group sessions by project path prefix and print a table: project | input | output | cache-write | cache-read | cost | savings | cache-hit%
- [x] 4.3 When `--json` is set, include `cost_usd` and `cache_savings_usd` numeric fields in each finding's JSON

## 5. Tests and docs

- [x] 5.1 Update `tokens_test.go` to cover the new struct fields and spike detection with separated token categories
- [x] 5.2 Add integration-style test for `--stats` output using a fixture JSONL file with known token counts
- [x] 5.3 Update `README.md` token-tracker section to document `--stats` flag and default pricing
