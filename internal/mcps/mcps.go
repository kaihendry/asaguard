package mcps

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hendry/asaguard/internal/policy"
	"github.com/hendry/asaguard/internal/result"
)

type mcpEntry struct {
	Name      string `json:"name"`
	Transport string `json:"transport"`
	Command   string `json:"command"`
	URL       string `json:"url"`
}

func readMCPs() ([]mcpEntry, error) {
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, ".claude", "settings.json")
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var raw struct {
		MCPServers map[string]struct {
			Command string   `json:"command"`
			Args    []string `json:"args"`
			URL     string   `json:"url"`
			Type    string   `json:"type"`
		} `json:"mcpServers"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	var entries []mcpEntry
	for name, s := range raw.MCPServers {
		transport := s.Type
		if transport == "" {
			if s.URL != "" {
				transport = "http"
			} else {
				transport = "stdio"
			}
		}
		entries = append(entries, mcpEntry{
			Name:      name,
			Transport: transport,
			Command:   s.Command,
			URL:       s.URL,
		})
	}
	return entries, nil
}

func riskLevel(name string, pol *policy.Policy) string {
	if lvl, ok := pol.MCPRiskCategories[name]; ok {
		return lvl
	}
	return "unknown"
}

func isApproved(name string, pol *policy.Policy) bool {
	if len(pol.ApprovedMCPs) == 0 {
		return true // no allowlist = pass-through
	}
	for _, a := range pol.ApprovedMCPs {
		if a == name {
			return true
		}
	}
	return false
}

// Check audits installed MCPs and returns findings.
func Check(pol *policy.Policy) []result.Finding {
	entries, err := readMCPs()
	if err != nil {
		return []result.Finding{{Check: "mcps", Level: result.Fail, Message: err.Error()}}
	}
	if len(entries) == 0 {
		return []result.Finding{{Check: "mcps", Level: result.Pass, Message: "no MCPs installed"}}
	}

	var findings []result.Finding
	for _, e := range entries {
		if !isApproved(e.Name, pol) {
			findings = append(findings, result.Finding{
				Check:   "mcps",
				Level:   result.Fail,
				Message: fmt.Sprintf("unapproved MCP: %s (transport: %s)", e.Name, e.Transport),
			})
		}
		risk := riskLevel(e.Name, pol)
		if risk == "high" {
			findings = append(findings, result.Finding{
				Check:   "mcps",
				Level:   result.Warn,
				Message: fmt.Sprintf("high-exfiltration-risk MCP: %s", e.Name),
			})
		}
	}

	if len(findings) == 0 {
		findings = append(findings, result.Finding{
			Check:   "mcps",
			Level:   result.Pass,
			Message: fmt.Sprintf("%d MCP(s) installed, all approved", len(entries)),
		})
	}
	return findings
}

func Run(args []string) {
	fs := flag.NewFlagSet("mcps", flag.ExitOnError)
	jsonOut := fs.Bool("json", false, "JSON output")
	fs.Parse(args)

	pol, err := policy.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "policy load error:", err)
		os.Exit(1)
	}

	findings := Check(pol)
	result.Print(findings, *jsonOut)
	if result.HasFail(findings) {
		os.Exit(1)
	}
}
