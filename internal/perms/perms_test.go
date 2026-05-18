package perms

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/hendry/asaguard/internal/result"
)

func writeSettings(t *testing.T, home string, v any) {
	t.Helper()
	dir := filepath.Join(home, ".claude")
	os.MkdirAll(dir, 0700)
	data, _ := json.Marshal(v)
	os.WriteFile(filepath.Join(dir, "settings.json"), data, 0600)
}

func TestCheckOpenByDefault(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeSettings(t, home, map[string]any{"permissions": map[string]any{}})

	findings := Check()
	hasWarn := false
	for _, f := range findings {
		if f.Level == result.Warn && !hasWarn {
			hasWarn = true
		}
	}
	if !hasWarn {
		t.Error("expected WARN for open-by-default configuration")
	}
}

func TestCheckDeniedTool(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeSettings(t, home, map[string]any{
		"permissions": map[string]any{
			"deny": []string{"Bash"},
		},
	})

	findings := Check()
	// isDenied("Bash") should be true, so the "Bash not denied" probe fails with WARN
	hasWarnOrFail := false
	for _, f := range findings {
		if f.Level == result.Warn || f.Level == result.Fail {
			hasWarnOrFail = true
		}
	}
	if !hasWarnOrFail {
		t.Error("expected WARN/FAIL when Bash is unexpectedly denied")
	}
}

func TestIsDenied(t *testing.T) {
	sp := &settingsPerms{}
	sp.Permissions.Deny = []string{"WebSearch", "Bash:execute"}

	if !isDenied("WebSearch", sp) {
		t.Error("WebSearch should be denied")
	}
	if !isDenied("Bash", sp) {
		t.Error("Bash should match Bash:execute prefix")
	}
	if isDenied("Read", sp) {
		t.Error("Read should not be denied")
	}
}

func TestIsAllowed(t *testing.T) {
	sp := &settingsPerms{}
	// Empty allow list = open
	if !isAllowed("Anything", sp) {
		t.Error("empty allow list should allow everything")
	}

	sp.Permissions.Allow = []string{"Read", "Write"}
	if !isAllowed("Read", sp) {
		t.Error("Read should be allowed")
	}
	if isAllowed("Bash", sp) {
		t.Error("Bash should not be allowed when not in allow list")
	}
}
