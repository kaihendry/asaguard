## Context

`internal/transcripts/tokens.go` currently collapses cache-read and cache-write tokens into a single `Input` field. The `CheckTokens` function only flags spikes; it discards the per-category breakdown. Claude API pricing differs significantly per token category (cache-read is ~10× cheaper than full input), so cost estimation requires keeping these fields separate. The session-report Claude plugin (reference implementation) demonstrates the correct deduplication and pricing approach.

## Goals / Non-Goals

**Goals:**
- Preserve all four token categories per session: input, output, cache-read, cache-write
- Compute estimated USD cost per session using configurable per-million-token prices
- Add `--stats` flag to `asaguard tokens` for a human-readable cost table grouped by project
- Emit `cost_usd` and `cache_savings_usd` in JSON output
- Keep existing spike-detection findings unchanged

**Non-Goals:**
- Subagent or skill-level attribution (session-level granularity is sufficient)
- Deduplication by request ID (the existing walk already yields one entry per assistant turn; full dedup adds complexity without clear benefit for the spike-detection use case)
- Real-time pricing updates (prices are constants with CLI override)

## Decisions

### D1: Pricing as policy fields with hardcoded defaults

Claude Sonnet 4.x pricing (per million tokens):
- Input: $3.00
- Output: $15.00
- Cache-write: $3.75
- Cache-read: $0.30

**Rationale**: Prices change rarely; baking defaults into `policy.Policy` with optional YAML overrides avoids an external HTTP call at check time. Alternative: fetch from API — rejected, adds latency and auth dependency.

### D2: `sessionTokens` struct gets separate fields

Replace the collapsed `Input int` with `InputTokens`, `OutputTokens`, `CacheRead`, `CacheWrite` fields. Compute `TotalCost` and `CacheSavings` on the struct.

**Rationale**: Keeps cost logic colocated with the struct, avoids recalculating on every caller. Alternative: compute cost only at output time — rejected, would scatter pricing logic.

### D3: `--stats` flag on existing `tokens` subcommand

Add `--stats` to `RunTokens` rather than a new subcommand.

**Rationale**: Avoids a new top-level entry point; stats and spike detection operate on the same data. Alternative: new `asaguard cost` subcommand — rejected, premature given current scope.

### D4: Cache savings = (cache-read tokens × input price) − (cache-read tokens × cache-read price)

Savings represent what you would have paid had the cache not been used.

## Risks / Trade-offs

- [Pricing accuracy] Hardcoded prices will drift as Anthropic updates rates → Mitigation: policy YAML overrides let operators correct without a code change; document the default prices in help text
- [Dedup gap] Without request-ID dedup, retried assistant turns may inflate counts slightly → Mitigation: acceptable for cost estimation; file a follow-up if overcounting becomes material
