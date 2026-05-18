package transcripts

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/hendry/asaguard/internal/policy"
	"github.com/hendry/asaguard/internal/result"
)

func networkRoot(t *testing.T) string {
	home := t.TempDir()
	t.Setenv("HOME", home)
	root := filepath.Join(home, ".claude", "projects", "proj1")
	os.MkdirAll(root, 0700)
	return root
}

func TestCheckNetworkEmpty(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	pol := &policy.Policy{}
	findings := CheckNetwork(pol, time.Time{})
	if len(findings) != 1 || findings[0].Level != result.Pass {
		t.Errorf("expected single PASS for no sessions, got %v", findings)
	}
}

func TestCheckNetworkApprovedDomain(t *testing.T) {
	root := networkRoot(t)
	line := `{"type":"tool_use","sessionId":"s1","timestamp":"2026-01-01T00:00:00Z","toolName":"WebFetch","toolInput":{"url":"https://api.anthropic.com/v1/messages"}}` + "\n"
	os.WriteFile(filepath.Join(root, "s1.jsonl"), []byte(line), 0600)

	pol := &policy.Policy{AllowedDomains: []string{"api.anthropic.com"}}
	findings := CheckNetwork(pol, time.Time{})
	if result.HasFail(findings) {
		t.Errorf("expected no failures for approved domain, got %v", findings)
	}
}

func TestCheckNetworkUnapprovedDomain(t *testing.T) {
	root := networkRoot(t)
	line := `{"type":"tool_use","sessionId":"s1","timestamp":"2026-01-01T00:00:00Z","toolName":"WebFetch","toolInput":{"url":"https://evil.example.com/exfil"}}` + "\n"
	os.WriteFile(filepath.Join(root, "s1.jsonl"), []byte(line), 0600)

	pol := &policy.Policy{AllowedDomains: []string{"api.anthropic.com"}}
	findings := CheckNetwork(pol, time.Time{})
	hasWarn := false
	for _, f := range findings {
		if f.Level == result.Warn {
			hasWarn = true
		}
	}
	if !hasWarn {
		t.Error("expected WARN for unapproved domain")
	}
}

func TestExtractBashURLs(t *testing.T) {
	var out []netAccess
	extractBashURLs("sess1", "curl https://example.com/data -o /tmp/out", &out)
	if len(out) != 1 || out[0].URL != "https://example.com/data" {
		t.Errorf("expected one extracted URL, got %v", out)
	}
}

func TestExtractBashURLsAmbiguous(t *testing.T) {
	var out []netAccess
	extractBashURLs("sess1", "curl $SOME_VAR", &out)
	if len(out) != 1 || !out[0].Ambiguous {
		t.Errorf("expected one ambiguous entry, got %v", out)
	}
}
