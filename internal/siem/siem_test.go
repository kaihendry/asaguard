package siem

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/kaihendry/asaguard/internal/result"
)

func TestBuildPayload(t *testing.T) {
	findings := []result.Finding{
		{Check: "settings", Level: result.Fail, Message: "dangerous setting enabled"},
		{Check: "bypass", Level: result.Warn, Message: "bypass flag used"},
		{Check: "secrets", Level: result.Pass, Message: "ok"},
	}
	start := time.Now().Add(-50 * time.Millisecond)
	p := buildPayload("0.1.0", findings, 42, 1, start)

	if p.SchemaVersion != "1.0" {
		t.Errorf("schema_version: got %q", p.SchemaVersion)
	}
	if p.Version != "0.1.0" {
		t.Errorf("version: got %q", p.Version)
	}
	if p.Mode != "monitor" {
		t.Errorf("mode: got %q", p.Mode)
	}
	if p.Score != 42 {
		t.Errorf("score: got %d", p.Score)
	}
	if p.ExitCode != 1 {
		t.Errorf("exit_code: got %d", p.ExitCode)
	}
	if p.DurationMS < 0 {
		t.Errorf("duration_ms negative: %d", p.DurationMS)
	}
	if len(p.RunID) != 36 {
		t.Errorf("run_id length: got %d", len(p.RunID))
	}
	if p.Timestamp == "" {
		t.Error("timestamp empty")
	}
	if len(p.Findings) != 3 {
		t.Fatalf("findings count: got %d", len(p.Findings))
	}

	// severity mapping
	for _, sf := range p.Findings {
		switch sf.Module {
		case "settings":
			if sf.Severity != "HIGH" {
				t.Errorf("settings severity: got %q", sf.Severity)
			}
			if sf.Type != "SETTINGS_FINDING" {
				t.Errorf("settings type: got %q", sf.Type)
			}
		case "bypass":
			if sf.Severity != "WARN" {
				t.Errorf("bypass severity: got %q", sf.Severity)
			}
			if sf.Type != "BYPASS_FINDING" {
				t.Errorf("bypass type: got %q", sf.Type)
			}
		case "secrets":
			if sf.Severity != "INFO" {
				t.Errorf("secrets severity: got %q", sf.Severity)
			}
		}
	}
}

func TestResolveEndpoint_EnvVar(t *testing.T) {
	t.Setenv("AI_GUARDRAILS_SIEM_ENDPOINT", "https://env.example.com/")
	if got := resolveEndpoint(); got != "https://env.example.com/" {
		t.Errorf("got %q", got)
	}
}

func TestResolveEndpoint_ConfigFile(t *testing.T) {
	t.Setenv("AI_GUARDRAILS_SIEM_ENDPOINT", "")

	dir := t.TempDir()
	cfgDir := filepath.Join(dir, ".config", "ai-check-guardrails")
	os.MkdirAll(cfgDir, 0755)
	cfgPath := filepath.Join(cfgDir, "config.json")
	os.WriteFile(cfgPath, []byte(`{"siem_endpoint":"https://cfg.example.com/"}`), 0600)

	// override home via user lookup — patch via env HOME
	t.Setenv("HOME", dir)

	if got := resolveEndpoint(); got != "https://cfg.example.com/" {
		t.Errorf("got %q", got)
	}
}

func TestResolveEndpoint_EnvVarPrecedence(t *testing.T) {
	t.Setenv("AI_GUARDRAILS_SIEM_ENDPOINT", "https://env.example.com/")

	dir := t.TempDir()
	cfgDir := filepath.Join(dir, ".config", "ai-check-guardrails")
	os.MkdirAll(cfgDir, 0755)
	os.WriteFile(filepath.Join(cfgDir, "config.json"),
		[]byte(`{"siem_endpoint":"https://cfg.example.com/"}`), 0600)
	t.Setenv("HOME", dir)

	if got := resolveEndpoint(); got != "https://env.example.com/" {
		t.Errorf("env var should win, got %q", got)
	}
}

func TestResolveEndpoint_NoneConfigured(t *testing.T) {
	t.Setenv("AI_GUARDRAILS_SIEM_ENDPOINT", "")
	t.Setenv("HOME", t.TempDir()) // no config file
	if got := resolveEndpoint(); got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

func TestReport_Integration(t *testing.T) {
	var gotContentType, gotUserAgent string
	var gotBody payload

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotContentType = r.Header.Get("Content-Type")
		gotUserAgent = r.Header.Get("User-Agent")
		json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	t.Setenv("AI_GUARDRAILS_SIEM_ENDPOINT", srv.URL)
	t.Setenv("AI_GUARDRAILS_SIEM_TOKEN", "")

	findings := []result.Finding{
		{Check: "mcps", Level: result.Fail, Message: "unapproved MCP"},
	}
	Report("0.1.0", findings, 30, 1, time.Now())

	if !strings.HasPrefix(gotContentType, "application/json") {
		t.Errorf("Content-Type: got %q", gotContentType)
	}
	if gotUserAgent != "asaguard/0.1.0" {
		t.Errorf("User-Agent: got %q", gotUserAgent)
	}
	if gotBody.SchemaVersion != "1.0" {
		t.Errorf("schema_version in body: got %q", gotBody.SchemaVersion)
	}
	if len(gotBody.Findings) != 1 {
		t.Errorf("findings count: got %d", len(gotBody.Findings))
	}
}
