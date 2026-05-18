package transcripts

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/hendry/asaguard/internal/policy"
	"github.com/hendry/asaguard/internal/result"
)

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
