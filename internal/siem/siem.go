package siem

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/hendry/asaguard/internal/result"
)

type siemFinding struct {
	Type        string `json:"type"`
	Severity    string `json:"severity"`
	Module      string `json:"module"`
	Resource    string `json:"resource"`
	Description string `json:"description"`
	Remediation string `json:"remediation"`
}

type payload struct {
	SchemaVersion string        `json:"schema_version"`
	RunID         string        `json:"run_id"`
	Timestamp     string        `json:"timestamp"`
	Host          string        `json:"host"`
	User          string        `json:"user"`
	Mode          string        `json:"mode"`
	Version       string        `json:"version"`
	Findings      []siemFinding `json:"findings"`
	Score         int           `json:"score"`
	ExitCode      int           `json:"exit_code"`
	DurationMS    int64         `json:"duration_ms"`
}

type configFile struct {
	SIEMEndpoint string `json:"siem_endpoint"`
}

func resolveEndpoint() string {
	if ep := os.Getenv("AI_GUARDRAILS_SIEM_ENDPOINT"); ep != "" {
		return ep
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	cfgPath := filepath.Join(home, ".config", "ai-check-guardrails", "config.json")
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return ""
	}
	var cfg configFile
	if err := json.Unmarshal(data, &cfg); err != nil {
		return ""
	}
	return cfg.SIEMEndpoint
}

func newUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func severityFor(level result.Level) string {
	switch level {
	case result.Fail:
		return "HIGH"
	case result.Warn:
		return "WARN"
	default:
		return "INFO"
	}
}

func buildPayload(version string, findings []result.Finding, score, exitCode int, start time.Time) payload {
	host, _ := os.Hostname()
	u, _ := user.Current()
	username := ""
	if u != nil {
		username = u.Username
	}

	sf := make([]siemFinding, 0, len(findings))
	for _, f := range findings {
		sf = append(sf, siemFinding{
			Type:        strings.ToUpper(f.Check) + "_FINDING",
			Severity:    severityFor(f.Level),
			Module:      f.Check,
			Description: f.Message,
		})
	}

	return payload{
		SchemaVersion: "1.0",
		RunID:         newUUID(),
		Timestamp:     start.UTC().Format(time.RFC3339),
		Host:          host,
		User:          username,
		Mode:          "monitor",
		Version:       version,
		Findings:      sf,
		Score:         score,
		ExitCode:      exitCode,
		DurationMS:    time.Since(start).Milliseconds(),
	}
}

func post(endpoint, version string, p payload) {
	body, err := json.Marshal(p)
	if err != nil {
		fmt.Fprintln(os.Stderr, "siem: marshal error:", err)
		return
	}

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		fmt.Fprintln(os.Stderr, "siem: request error:", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "asaguard/"+version)
	if tok := os.Getenv("AI_GUARDRAILS_SIEM_TOKEN"); tok != "" {
		req.Header.Set("Authorization", "Bearer "+tok)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintln(os.Stderr, "siem: POST error:", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		fmt.Fprintf(os.Stderr, "siem: unexpected status %d\n", resp.StatusCode)
	}
}

// Report sends audit results to the configured SIEM endpoint, if any.
// Failures are logged to stderr; the caller's exit code is unaffected.
func Report(version string, findings []result.Finding, score, exitCode int, start time.Time) {
	endpoint := resolveEndpoint()
	if endpoint == "" {
		return
	}
	p := buildPayload(version, findings, score, exitCode, start)
	post(endpoint, version, p)
}
