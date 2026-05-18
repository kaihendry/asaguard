## Why

The current `tokens` check only detects anomalous spikes but discards the detailed breakdown data (input vs output vs cache-read vs cache-write) needed to understand costs and cache efficiency. Adding cost calculation and richer per-project/per-session stats lets operators see what Claude usage actually costs and whether prompt caching is working effectively.

## What Changes

- Extend `collectTokens` to keep input, output, cache-read, and cache-write as separate fields (currently collapsed into one `Input` total)
- Add cost calculation using Claude API pricing tiers (input, output, cache-write, cache-read per million tokens)
- Add `--stats` mode to `asaguard tokens` that prints a detailed cost + cache efficiency table grouped by project
- Report cache hit ratio and estimated savings (cache-read cost vs full input cost) alongside total cost
- Add `total_cost_usd` and `cache_savings_usd` fields to JSON output

## Capabilities

### New Capabilities
- `cost-reporter`: Computes estimated USD cost from token counts using configurable per-token pricing, produces per-session and aggregate cost summaries with cache savings metrics

### Modified Capabilities
- `token-tracker`: Token breakdown must now preserve cache-read, cache-write, input, and output as separate fields rather than collapsing cache into input; adds cost and savings fields to output

## Impact

- `internal/transcripts/tokens.go`: extend `sessionTokens` struct, update `collectTokens`, update `CheckTokens` and `RunTokens`
- `internal/policy/policy.go`: add pricing config fields (or use hardcoded defaults with optional override)
- JSON output schema gains new fields (`cost_usd`, `cache_savings_usd`, per-token breakdown)
- No breaking changes to existing pass/warn findings format
