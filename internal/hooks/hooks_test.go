package hooks

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func claudeSettings(t *testing.T, home string, v any) string {
	t.Helper()
	dir := filepath.Join(home, ".claude")
	os.MkdirAll(dir, 0700)
	path := filepath.Join(dir, "settings.json")
	data, _ := json.Marshal(v)
	os.WriteFile(path, data, 0600)
	return path
}

func TestReadRawMissing(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	m, err := readRaw()
	if err != nil {
		t.Fatal(err)
	}
	if len(m) != 0 {
		t.Error("expected empty map for missing settings.json")
	}
}

func TestGetSetHooks(t *testing.T) {
	m := map[string]json.RawMessage{}
	entries := []HookEntry{
		{Matcher: ".*", Hooks: []Hook{{Type: "command", Command: "/bin/true"}}},
	}
	if err := setHooks(m, "PreToolUse", entries); err != nil {
		t.Fatal(err)
	}
	got, err := getHooks(m, "PreToolUse")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Hooks[0].Command != "/bin/true" {
		t.Errorf("unexpected hooks: %v", got)
	}
}

func TestWriteAtomic(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	claudeSettings(t, home, map[string]any{})

	m, _ := readRaw()
	raw, _ := json.Marshal("hello")
	m["test"] = raw

	if err := writeAtomic(m); err != nil {
		t.Fatal(err)
	}

	m2, err := readRaw()
	if err != nil {
		t.Fatal(err)
	}
	var val string
	json.Unmarshal(m2["test"], &val)
	if val != "hello" {
		t.Errorf("expected 'hello', got %q", val)
	}
}

func TestUninstallNotFound(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	claudeSettings(t, home, map[string]any{})

	// Should not error — just report "not found"
	err := Uninstall("PreToolUse", "/nonexistent", "test-hook")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
