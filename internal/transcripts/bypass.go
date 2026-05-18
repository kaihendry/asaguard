// Package transcripts audits recorded Claude Code sessions for four classes of
// risky behaviour, turning the session transcript log into a security signal.
//
// Policy bypass: detects sessions launched with --dangerously-skip-permissions,
// which disables all tool approval prompts, or Bash commands that used
// --no-verify or --no-gpg-sign to circumvent git safety hooks. These flags are
// the clearest sign that a user is actively evading the controls asaguard
// enforces.
//
// Network access: inspects every WebFetch call and curl/wget Bash command and
// checks each target URL against the policy domain allowlist. This surfaces what
// external services Claude is contacting — package registries, internal APIs,
// third-party endpoints — and flags anything outside the approved list.
//
// Sandbox enforcement: verifies that all file Read, Write, Edit, and Bash
// accesses stayed within the authorised path roots. Accesses to /etc, /var,
// other users' home directories, or any path outside the sandbox roots are
// flagged, confirming that Claude only touched data it was supposed to.
//
// Token usage and cost: tracks token consumption per session and flags spikes
// that exceed a configurable multiple of the rolling per-session average —
// a pattern associated with prompt injection, runaway agents, or accidental
// bulk processing. It also reports estimated cost in USD and cache-hit
// efficiency per project, giving teams visibility into Claude API spend without
// needing direct access to billing dashboards.
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
