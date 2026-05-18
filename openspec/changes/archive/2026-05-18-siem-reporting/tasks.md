## 1. New SIEM package

- [x] 1.1 Create `internal/siem/siem.go` with `Report(report scorer.ScoreReport, startTime time.Time)` function
- [x] 1.2 Implement `resolveEndpoint()`: read `AI_GUARDRAILS_SIEM_ENDPOINT`, fall back to `~/.config/ai-check-guardrails/config.json` `siem_endpoint`
- [x] 1.3 Implement `buildPayload()`: map `ScoreReport` → schema v1.0 JSON struct (uuid, RFC3339 timestamp, host, user, mode:"monitor", version, findings[], score, exit_code, duration_ms)
- [x] 1.4 Implement finding severity mapping: PASS→INFO, WARN→WARN, FAIL→HIGH; type as `<CHECK>_FINDING`
- [x] 1.5 Implement `post()`: HTTP POST with `Content-Type: application/json`, `User-Agent: asaguard/<version>`, optional `Authorization: Bearer` from `AI_GUARDRAILS_SIEM_TOKEN`, 10 s timeout; log non-2xx to stderr without error return

## 2. Wire SIEM into scorer

- [x] 2.1 Export `ScoreReport` start-time or pass `duration_ms` so `siem.Report` can compute it
- [x] 2.2 In `scorer.RunCheck()`: record start time before `runAllChecks`, call `siem.Report(report, start)` after output is printed
- [x] 2.3 In `scorer.Run()`: same — record start time, call `siem.Report` after score is printed

## 3. Tests

- [x] 3.1 Unit test `buildPayload`: verify required fields, severity mapping, type naming
- [x] 3.2 Unit test `resolveEndpoint`: env-var takes precedence, config-file fallback, both absent returns empty string
- [x] 3.3 Integration smoke test: start `httptest.NewServer`, run `siem.Report` against it, assert HTTP 200 and correct Content-Type header

## 4. Docs & polish

- [x] 4.1 Add `AI_GUARDRAILS_SIEM_ENDPOINT` / `AI_GUARDRAILS_SIEM_TOKEN` to `asaguard check` help text
- [x] 4.2 Update README with SIEM reporting section referencing the env vars and config-file fallback
