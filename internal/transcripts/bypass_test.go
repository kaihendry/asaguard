package transcripts

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kaihendry/asaguard/internal/result"
)

func bypassRoot(t *testing.T) string {
	home := t.TempDir()
	t.Setenv("HOME", home)
	root := filepath.Join(home, ".claude", "projects", "proj1")
	os.MkdirAll(root, 0700)
	return root
}

func TestCheckBypassClean(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	findings := CheckBypass(time.Time{})
	if len(findings) != 1 || findings[0].Level != result.Pass {
		t.Errorf("expected single PASS for empty transcripts, got %v", findings)
	}
}

func TestCheckBypassDangerousFlag(t *testing.T) {
	root := bypassRoot(t)
	line := `{"type":"system","sessionId":"s1","timestamp":"2026-01-01T00:00:00Z","cliArgs":["--dangerously-skip-permissions"]}` + "\n"
	os.WriteFile(filepath.Join(root, "s1.jsonl"), []byte(line), 0600)

	findings := CheckBypass(time.Time{})
	if !result.HasFail(findings) {
		t.Error("expected FAIL for --dangerously-skip-permissions")
	}
}

func TestCheckBypassNoVerify(t *testing.T) {
	root := bypassRoot(t)
	line := `{"type":"tool_use","sessionId":"s1","timestamp":"2026-01-01T00:00:00Z","toolName":"Bash","toolInput":{"command":"git commit --no-verify -m 'bad'"}}` + "\n"
	os.WriteFile(filepath.Join(root, "s1.jsonl"), []byte(line), 0600)

	findings := CheckBypass(time.Time{})
	if !result.HasFail(findings) {
		t.Error("expected FAIL for git commit --no-verify")
	}
}

func TestCheckBashBypassPatterns(t *testing.T) {
	var violations []bypassViolation
	checkBashBypass("s1", "2026-01-01", "git push --no-verify origin main", &violations)
	if len(violations) != 1 {
		t.Errorf("expected 1 violation, got %d", len(violations))
	}
	if violations[0].Flag != "--no-verify" {
		t.Errorf("expected flag --no-verify, got %s", violations[0].Flag)
	}
}
