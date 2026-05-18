package transcripts

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/hendry/asaguard/internal/policy"
	"github.com/hendry/asaguard/internal/result"
)

type sandboxViolation struct {
	SessionID string
	Tool      string
	Path      string
	Op        string // read|write
}

func detectSandboxViolations(pol *policy.Policy, since time.Time) ([]sandboxViolation, error) {
	var violations []sandboxViolation

	err := Walk(since, func(s *Session) error {
		for _, e := range s.Entries {
			switch e.ToolName {
			case "Read":
				var inp struct {
					FilePath string `json:"file_path"`
				}
				if json.Unmarshal(e.ToolInput, &inp) == nil && inp.FilePath != "" {
					if !pathAllowed(inp.FilePath, pol.SandboxReadRoots) {
						violations = append(violations, sandboxViolation{s.ID, "Read", inp.FilePath, "read"})
					}
				}
			case "Write", "Edit":
				var inp struct {
					FilePath string `json:"file_path"`
				}
				if json.Unmarshal(e.ToolInput, &inp) == nil && inp.FilePath != "" {
					if !pathAllowed(inp.FilePath, pol.SandboxWriteRoots) {
						violations = append(violations, sandboxViolation{s.ID, e.ToolName, inp.FilePath, "write"})
					}
				}
			case "Bash":
				// Best-effort: look for obvious absolute paths in commands
				var inp struct {
					Command string `json:"command"`
				}
				if json.Unmarshal(e.ToolInput, &inp) != nil {
					continue
				}
				for _, tok := range strings.Fields(inp.Command) {
					if strings.HasPrefix(tok, "/") && !pathAllowed(tok, pol.SandboxReadRoots) {
						violations = append(violations, sandboxViolation{s.ID, "Bash", tok, "read"})
						break
					}
				}
			}
		}
		return nil
	})
	return violations, err
}

func pathAllowed(path string, roots []string) bool {
	for _, r := range roots {
		if strings.HasPrefix(path, r) {
			return true
		}
	}
	return false
}

// CheckSandbox returns findings for out-of-sandbox file accesses.
func CheckSandbox(pol *policy.Policy, since time.Time) []result.Finding {
	violations, err := detectSandboxViolations(pol, since)
	if err != nil {
		return []result.Finding{{Check: "sandbox", Level: result.Warn, Message: "transcript walk error: " + err.Error()}}
	}
	if len(violations) == 0 {
		return []result.Finding{{Check: "sandbox", Level: result.Pass, Message: "all file accesses within authorised paths"}}
	}

	var findings []result.Finding
	for _, v := range violations {
		findings = append(findings, result.Finding{
			Check:   "sandbox",
			Level:   result.Fail,
			Message: fmt.Sprintf("out-of-sandbox %s by %s in session %s: %s", v.Op, v.Tool, v.SessionID, v.Path),
		})
	}
	return findings
}

func RunSandbox(args []string) {
	fs := flag.NewFlagSet("sandbox", flag.ExitOnError)
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

	pol, err := policy.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "policy load error:", err)
		os.Exit(1)
	}

	findings := CheckSandbox(pol, sinceTime)
	result.Print(findings, *jsonOut)
	if result.HasFail(findings) {
		os.Exit(1)
	}
}
