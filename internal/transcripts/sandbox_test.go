package transcripts

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/hendry/asaguard/internal/policy"
	"github.com/hendry/asaguard/internal/result"
)

func sandboxRoot(t *testing.T) (home, projRoot string) {
	home = t.TempDir()
	t.Setenv("HOME", home)
	projRoot = filepath.Join(home, ".claude", "projects", "proj1")
	os.MkdirAll(projRoot, 0700)
	return
}

func TestCheckSandboxEmpty(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	pol := &policy.Policy{SandboxReadRoots: []string{"/home"}, SandboxWriteRoots: []string{"/home"}}
	findings := CheckSandbox(pol, time.Time{})
	if len(findings) != 1 || findings[0].Level != result.Pass {
		t.Errorf("expected single PASS for empty transcripts, got %v", findings)
	}
}

func TestCheckSandboxWithinBounds(t *testing.T) {
	home, root := sandboxRoot(t)
	line := `{"type":"tool_use","sessionId":"s1","timestamp":"2026-01-01T00:00:00Z","toolName":"Read","toolInput":{"file_path":"` + home + `/project/main.go"}}` + "\n"
	os.WriteFile(filepath.Join(root, "s1.jsonl"), []byte(line), 0600)

	pol := &policy.Policy{SandboxReadRoots: []string{home}, SandboxWriteRoots: []string{home}}
	findings := CheckSandbox(pol, time.Time{})
	if result.HasFail(findings) {
		t.Errorf("expected no failure for in-bounds read, got %v", findings)
	}
}

func TestCheckSandboxOutOfBounds(t *testing.T) {
	_, root := sandboxRoot(t)
	line := `{"type":"tool_use","sessionId":"s1","timestamp":"2026-01-01T00:00:00Z","toolName":"Read","toolInput":{"file_path":"/etc/passwd"}}` + "\n"
	os.WriteFile(filepath.Join(root, "s1.jsonl"), []byte(line), 0600)

	pol := &policy.Policy{SandboxReadRoots: []string{"/home/user/project"}, SandboxWriteRoots: []string{"/home/user/project"}}
	findings := CheckSandbox(pol, time.Time{})
	if !result.HasFail(findings) {
		t.Error("expected FAIL for /etc/passwd read outside sandbox")
	}
}

func TestPathAllowed(t *testing.T) {
	roots := []string{"/home/user/projects", "/tmp"}
	if !pathAllowed("/home/user/projects/foo/bar.go", roots) {
		t.Error("expected path to be allowed")
	}
	if !pathAllowed("/tmp/scratch", roots) {
		t.Error("expected /tmp path to be allowed")
	}
	if pathAllowed("/etc/shadow", roots) {
		t.Error("expected /etc/shadow to be denied")
	}
}
