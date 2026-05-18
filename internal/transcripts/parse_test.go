package transcripts

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestParseFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.jsonl")
	content := `{"type":"assistant","sessionId":"abc123","timestamp":"2026-01-15T10:00:00Z","message":{"usage":{"input_tokens":100,"output_tokens":50}}}` + "\n"
	os.WriteFile(path, []byte(content), 0600)

	sess, err := parseFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if sess.ID != "abc123" {
		t.Errorf("expected session ID abc123, got %s", sess.ID)
	}
	if len(sess.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(sess.Entries))
	}
	e := sess.Entries[0]
	if e.Message == nil || e.Message.Usage == nil {
		t.Fatal("expected usage data")
	}
	if e.Message.Usage.InputTokens != 100 {
		t.Errorf("expected 100 input tokens, got %d", e.Message.Usage.InputTokens)
	}
}

func TestParseFileSince(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "old.jsonl")
	os.WriteFile(path, []byte(`{"sessionId":"old"}`+"\n"), 0600)

	// Make the file old
	old := time.Now().Add(-48 * time.Hour)
	os.Chtimes(path, old, old)

	visited := 0
	// Override walk root for test — we test Walk indirectly via WalkDir
	// This is a smoke test for parseFile only
	sess, _ := parseFile(path)
	if sess != nil {
		visited++
	}
	// Just verifying parse succeeds on a minimal file
	if visited != 1 {
		t.Error("expected parseFile to succeed")
	}
}
