// Package transcripts analyses Claude Code session transcripts to detect
// four categories of risky behaviour after the fact.
//
// Bypass: flags sessions where --dangerously-skip-permissions was passed or
// where Bash commands used --no-verify or --no-gpg-sign, which circumvent git
// safety checks.
//
// Network: identifies WebFetch calls and curl/wget commands in Bash and checks
// each target domain against the policy allowlist, surfacing unexpected
// outbound connections.
//
// Sandbox: detects file Read, Write, Edit, and Bash accesses outside the
// configured sandbox roots, catching attempts to reach sensitive paths such as
// /etc, /var, or other users' home directories.
//
// Tokens: tracks token consumption per session and flags spikes that exceed a
// configurable multiple of the per-session average. It also reports estimated
// cost in USD and cache-hit efficiency per project, giving teams visibility into
// Claude API spend without requiring direct access to billing dashboards.
package transcripts

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/kaihendry/asaguard/internal/result"
)

type bypassViolation struct {
	SessionID string `json:"sessionId"`
	Date      string `json:"date"`
	Flag      string `json:"flag"`
	Command   string `json:"command,omitempty"`
}

func detectBypasses(since time.Time) ([]bypassViolation, error) {
	var violations []bypassViolation

	err := Walk(since, func(s *Session) error {
		for _, e := range s.Entries {
			// Check session launch args
			for _, arg := range e.CLIArgs {
				if arg == "--dangerously-skip-permissions" {
					violations = append(violations, bypassViolation{
						SessionID: s.ID,
						Date:      e.Timestamp.Format("2006-01-02"),
						Flag:      "--dangerously-skip-permissions",
					})
				}
			}

			// Check Bash tool calls
			if e.ToolName == "Bash" {
				var inp struct {
					Command string `json:"command"`
				}
				if json.Unmarshal(e.ToolInput, &inp) == nil {
					checkBashBypass(s.ID, e.Timestamp.Format("2006-01-02"), inp.Command, &violations)
				}
			}
		}
		return nil
	})
	return violations, err
}

var bashBypassPatterns = []string{
	"--no-verify",
	"--no-gpg-sign",
}

func checkBashBypass(sessionID, date, cmd string, out *[]bypassViolation) {
	lower := strings.ToLower(cmd)
	for _, pat := range bashBypassPatterns {
		if strings.Contains(lower, pat) {
			*out = append(*out, bypassViolation{
				SessionID: sessionID,
				Date:      date,
				Flag:      pat,
				Command:   truncate(cmd, 120),
			})
		}
	}
}

// CheckBypass returns findings for permission-bypass patterns.
func CheckBypass(since time.Time) []result.Finding {
	violations, err := detectBypasses(since)
	if err != nil {
		return []result.Finding{{Check: "bypass", Level: result.Warn, Message: "transcript walk error: " + err.Error()}}
	}
	if len(violations) == 0 {
		return []result.Finding{{Check: "bypass", Level: result.Pass, Message: "no bypass flags detected"}}
	}

	var findings []result.Finding
	for _, v := range violations {
		msg := fmt.Sprintf("bypass flag %q in session %s on %s", v.Flag, v.SessionID, v.Date)
		if v.Command != "" {
			msg += ": " + v.Command
		}
		findings = append(findings, result.Finding{Check: "bypass", Level: result.Fail, Message: msg})
	}
	return findings
}

func RunBypass(args []string) {
	fs := flag.NewFlagSet("bypass", flag.ExitOnError)
	jsonOut := fs.Bool("json", false, "JSON output")
	since := fs.String("since", "", "ISO 8601 date")
	fs.Parse(args)

	var sinceTime time.Time
	if *since != "" {
		t, err := time.Parse("2006-01-02", *since)
		if err != nil {
			fmt.Fprintln(os.Stderr, "invalid --since:", err)
			os.Exit(1)
		}
		sinceTime = t
	}

	findings := CheckBypass(sinceTime)

	if *jsonOut {
		// Re-run to get structured violations for JSON
		violations, _ := detectBypasses(sinceTime)
		enc := json.NewEncoder(os.Stdout)
		if violations == nil {
			violations = []bypassViolation{}
		}
		enc.Encode(violations)
		return
	}

	result.Print(findings, false)
	if result.HasFail(findings) {
		os.Exit(1)
	}
}
