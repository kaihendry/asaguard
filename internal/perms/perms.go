// Package perms validates that Claude Code's permission posture is intentional.
//
// Claude Code ships permissive by default: with no allow or deny lists configured,
// every tool — including unrestricted Bash execution — runs without a confirmation
// prompt. This guard rail runs access probes to confirm that restrictions are
// actually in place and working as intended. It catches the common mistake of
// deploying Claude Code to an engineering team before locking down which tools
// require approval, ensuring the permission model is an explicit security decision
// rather than an accident of default configuration.
package perms

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kaihendry/asaguard/internal/result"
)

type settingsPerms struct {
	Permissions struct {
		Allow []string `json:"allow"`
		Deny  []string `json:"deny"`
	} `json:"permissions"`
}

type probe struct {
	Name          string
	Tool          string
	ExpectDenied  bool
}

var builtinProbes = []probe{
	{Name: "Bash not denied when allow-listed", Tool: "Bash", ExpectDenied: false},
	{Name: "WebSearch not in deny list", Tool: "WebSearch", ExpectDenied: false},
}

func loadPerms() (*settingsPerms, error) {
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, ".claude", "settings.json")
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &settingsPerms{}, nil
	}
	if err != nil {
		return nil, err
	}
	var sp settingsPerms
	json.Unmarshal(data, &sp)
	return &sp, nil
}

func isDenied(tool string, sp *settingsPerms) bool {
	for _, d := range sp.Permissions.Deny {
		if strings.EqualFold(d, tool) || strings.HasPrefix(strings.ToLower(d), strings.ToLower(tool)+":") {
			return true
		}
	}
	return false
}

func isAllowed(tool string, sp *settingsPerms) bool {
	if len(sp.Permissions.Allow) == 0 {
		return true // open by default
	}
	for _, a := range sp.Permissions.Allow {
		if strings.EqualFold(a, tool) || strings.HasPrefix(strings.ToLower(a), strings.ToLower(tool)+":") {
			return true
		}
	}
	return false
}

// Check runs permission probes and returns findings.
func Check() []result.Finding {
	sp, err := loadPerms()
	if err != nil {
		return []result.Finding{{Check: "perms", Level: result.Fail, Message: err.Error()}}
	}

	var findings []result.Finding

	if len(sp.Permissions.Allow) == 0 && len(sp.Permissions.Deny) == 0 {
		findings = append(findings, result.Finding{
			Check:   "perms",
			Level:   result.Warn,
			Message: "no allow or deny list configured — installation is permissive (open-by-default)",
		})
	}

	for _, p := range builtinProbes {
		denied := isDenied(p.Tool, sp)
		allowed := isAllowed(p.Tool, sp)

		var level result.Level
		var msg string

		if p.ExpectDenied {
			if denied {
				level = result.Pass
				msg = fmt.Sprintf("PASS  %s: %s correctly denied", p.Name, p.Tool)
			} else {
				level = result.Fail
				msg = fmt.Sprintf("FAIL  %s: %s expected to be denied but is allowed", p.Name, p.Tool)
			}
		} else {
			if allowed && !denied {
				level = result.Pass
				msg = fmt.Sprintf("PASS  %s: %s accessible as expected", p.Name, p.Tool)
			} else {
				level = result.Warn
				msg = fmt.Sprintf("WARN  %s: %s unexpectedly restricted", p.Name, p.Tool)
			}
		}
		findings = append(findings, result.Finding{Check: "perms", Level: level, Message: msg})
	}

	return findings
}

func Run(args []string) {
	fs := flag.NewFlagSet("perms", flag.ExitOnError)
	jsonOut := fs.Bool("json", false, "JSON output")
	fs.Parse(args)

	findings := Check()
	result.Print(findings, *jsonOut)
	if result.HasFail(findings) {
		os.Exit(1)
	}
}
