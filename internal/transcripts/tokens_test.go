package transcripts

import (
	"math"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kaihendry/asaguard/internal/policy"
	"github.com/kaihendry/asaguard/internal/result"
)

var defaultPrices = &policy.TokenPrices{
	InputPerMillion:      3.00,
	OutputPerMillion:     15.00,
	CacheWritePerMillion: 3.75,
	CacheReadPerMillion:  0.30,
}

func TestCostUSDZero(t *testing.T) {
	s := sessionTokens{}
	if got := s.CostUSD(defaultPrices); got != 0 {
		t.Errorf("expected 0, got %f", got)
	}
}

func TestCostUSDMixed(t *testing.T) {
	s := sessionTokens{
		InputTokens:  1_000_000,
		OutputTokens: 1_000_000,
		CacheWrite:   1_000_000,
		CacheRead:    1_000_000,
	}
	want := 3.00 + 15.00 + 3.75 + 0.30
	got := s.CostUSD(defaultPrices)
	if math.Abs(got-want) > 1e-9 {
		t.Errorf("expected %.4f, got %.4f", want, got)
	}
}

func TestCacheSavingsZero(t *testing.T) {
	s := sessionTokens{}
	if got := s.CacheSavingsUSD(defaultPrices); got != 0 {
		t.Errorf("expected 0, got %f", got)
	}
}

func TestCacheSavingsMixed(t *testing.T) {
	s := sessionTokens{CacheRead: 1_000_000}
	want := (3.00 - 0.30) // per million
	got := s.CacheSavingsUSD(defaultPrices)
	if math.Abs(got-want) > 1e-9 {
		t.Errorf("expected %.4f, got %.4f", want, got)
	}
}

func writeTranscript(t *testing.T, dir, name, content string) {
	t.Helper()
	os.MkdirAll(dir, 0700)
	os.WriteFile(filepath.Join(dir, name), []byte(content), 0600)
}

func transcriptsRoot(t *testing.T) string {
	home := t.TempDir()
	t.Setenv("HOME", home)
	root := filepath.Join(home, ".claude", "projects", "proj1")
	os.MkdirAll(root, 0700)
	return root
}

func TestPrintStatsTable(t *testing.T) {
	root := transcriptsRoot(t)
	// 1M input + 500K output + 200K cache-write + 100K cache-read
	line := `{"type":"assistant","sessionId":"s1","timestamp":"2026-01-01T00:00:00Z","message":{"usage":{"input_tokens":1000000,"output_tokens":500000,"cache_creation_input_tokens":200000,"cache_read_input_tokens":100000}}}` + "\n"
	writeTranscript(t, root, "s1.jsonl", line)

	sessions, err := collectTokens(time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if len(sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(sessions))
	}
	s := sessions[0]
	if s.InputTokens != 1000000 {
		t.Errorf("input: want 1000000, got %d", s.InputTokens)
	}
	if s.OutputTokens != 500000 {
		t.Errorf("output: want 500000, got %d", s.OutputTokens)
	}
	if s.CacheWrite != 200000 {
		t.Errorf("cache_write: want 200000, got %d", s.CacheWrite)
	}
	if s.CacheRead != 100000 {
		t.Errorf("cache_read: want 100000, got %d", s.CacheRead)
	}

	// Cost: (1M*3 + 0.5M*15 + 0.2M*3.75 + 0.1M*0.30) / 1M = 3 + 7.5 + 0.75 + 0.03 = 11.28
	wantCost := 11.28
	gotCost := s.CostUSD(defaultPrices)
	if math.Abs(gotCost-wantCost) > 1e-9 {
		t.Errorf("cost: want %.4f, got %.4f", wantCost, gotCost)
	}
}

func TestCheckTokensEmpty(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	pol := &policy.Policy{TokenSpikeMultiple: 3.0}
	findings := CheckTokens(pol, time.Time{})
	if len(findings) != 1 || findings[0].Level != result.Pass {
		t.Errorf("expected single PASS for no sessions, got %v", findings)
	}
}

func TestCheckTokensNoSpike(t *testing.T) {
	root := transcriptsRoot(t)
	line := `{"type":"assistant","sessionId":"s1","timestamp":"2026-01-01T00:00:00Z","message":{"usage":{"input_tokens":100,"output_tokens":50}}}` + "\n"
	writeTranscript(t, root, "s1.jsonl", line+line)

	pol := &policy.Policy{TokenSpikeMultiple: 3.0}
	findings := CheckTokens(pol, time.Time{})
	if result.HasFail(findings) {
		t.Errorf("unexpected failure: %v", findings)
	}
}

func TestCheckTokensSpike(t *testing.T) {
	root := transcriptsRoot(t)
	// 9 tiny sessions to establish a low baseline (~15 tokens each)
	tiny := `{"type":"assistant","sessionId":"%d","timestamp":"2026-01-01T00:00:00Z","message":{"usage":{"input_tokens":10,"output_tokens":5}}}` + "\n"
	for i := 0; i < 9; i++ {
		writeTranscript(t, root, "tiny"+string(rune('0'+i))+".jsonl", tiny)
	}
	// Spike session: 10000× the baseline
	big := `{"type":"assistant","sessionId":"big","timestamp":"2026-01-02T00:00:00Z","message":{"usage":{"input_tokens":10000,"output_tokens":5000}}}` + "\n"
	writeTranscript(t, root, "big.jsonl", big)

	pol := &policy.Policy{TokenSpikeMultiple: 3.0}
	findings := CheckTokens(pol, time.Time{})
	hasWarn := false
	for _, f := range findings {
		if f.Level == result.Warn {
			hasWarn = true
		}
	}
	if !hasWarn {
		t.Error("expected WARN for token spike")
	}
}
