package mcps

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/kaihendry/asaguard/internal/policy"
	"github.com/kaihendry/asaguard/internal/result"
)

func writeSettings(t *testing.T, home string, v any) {
	t.Helper()
	dir := filepath.Join(home, ".claude")
	os.MkdirAll(dir, 0700)
	data, _ := json.Marshal(v)
	os.WriteFile(filepath.Join(dir, "settings.json"), data, 0600)
}

func TestCheckNoMCPs(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	pol := &policy.Policy{}
	findings := Check(pol)
	if len(findings) != 1 || findings[0].Level != result.Pass {
		t.Errorf("expected single PASS, got %v", findings)
	}
}

func TestCheckApprovedMCP(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	writeSettings(t, home, map[string]any{
		"mcpServers": map[string]any{
			"filesystem": map[string]any{"command": "npx", "type": "stdio"},
		},
	})

	pol := &policy.Policy{ApprovedMCPs: []string{"filesystem"}}
	findings := Check(pol)

	if result.HasFail(findings) {
		t.Errorf("expected no failures for approved MCP, got %v", findings)
	}
}

func TestCheckUnapprovedMCP(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	writeSettings(t, home, map[string]any{
		"mcpServers": map[string]any{
			"suspicious-mcp": map[string]any{"command": "bad", "type": "stdio"},
		},
	})

	pol := &policy.Policy{ApprovedMCPs: []string{"filesystem"}}
	findings := Check(pol)

	if !result.HasFail(findings) {
		t.Error("expected failure for unapproved MCP")
	}
}

func TestCheckHighRiskMCP(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	writeSettings(t, home, map[string]any{
		"mcpServers": map[string]any{
			"risky-mcp": map[string]any{"command": "risky", "type": "stdio"},
		},
	})

	pol := &policy.Policy{
		MCPRiskCategories: map[string]string{"risky-mcp": "high"},
	}
	findings := Check(pol)

	hasWarn := false
	for _, f := range findings {
		if f.Level == result.Warn {
			hasWarn = true
		}
	}
	if !hasWarn {
		t.Error("expected WARN for high-risk MCP")
	}
}
