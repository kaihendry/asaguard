package policy

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	// Point config home somewhere empty
	t.Setenv("HOME", t.TempDir())

	pol, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if pol.TokenSpikeMultiple != 3.0 {
		t.Errorf("expected default spike multiple 3.0, got %f", pol.TokenSpikeMultiple)
	}
	if len(pol.HITLWatchlist) == 0 {
		t.Error("expected non-empty default HITL watchlist")
	}
	if len(pol.Weights) == 0 {
		t.Error("expected non-empty default weights")
	}
}

func TestLoadMergesOverride(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	dir := filepath.Join(home, ".config", "asaguard")
	os.MkdirAll(dir, 0700)

	override := map[string]any{
		"token_spike_multiple": 5.0,
		"approved_mcps":        []string{"my-mcp"},
		"weights":              map[string]int{"settings": 99},
	}
	data, _ := json.Marshal(override)
	os.WriteFile(filepath.Join(dir, "policy.json"), data, 0600)

	pol, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if pol.TokenSpikeMultiple != 5.0 {
		t.Errorf("expected overridden spike multiple 5.0, got %f", pol.TokenSpikeMultiple)
	}
	if len(pol.ApprovedMCPs) != 1 || pol.ApprovedMCPs[0] != "my-mcp" {
		t.Errorf("expected approved_mcps=[my-mcp], got %v", pol.ApprovedMCPs)
	}
	if pol.Weights["settings"] != 99 {
		t.Errorf("expected weight for settings=99, got %d", pol.Weights["settings"])
	}
}

func TestLoadInvalidJSON(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	dir := filepath.Join(home, ".config", "asaguard")
	os.MkdirAll(dir, 0700)
	os.WriteFile(filepath.Join(dir, "policy.json"), []byte("not json"), 0600)

	_, err := Load()
	if err == nil {
		t.Error("expected error on invalid JSON")
	}
}
