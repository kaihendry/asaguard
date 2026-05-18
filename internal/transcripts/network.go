package transcripts

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/kaihendry/asaguard/internal/policy"
	"github.com/kaihendry/asaguard/internal/result"
)

var urlRe = regexp.MustCompile(`https?://[^\s'"]+`)

type netAccess struct {
	SessionID string
	Tool      string
	URL       string
	Ambiguous bool
}

func extractNetwork(since time.Time) ([]netAccess, error) {
	var accesses []netAccess
	err := Walk(since, func(s *Session) error {
		for _, e := range s.Entries {
			if e.ToolName == "WebFetch" {
				var inp struct {
					URL string `json:"url"`
				}
				if json.Unmarshal(e.ToolInput, &inp) == nil && inp.URL != "" {
					accesses = append(accesses, netAccess{SessionID: s.ID, Tool: "WebFetch", URL: inp.URL})
				}
			}
			if e.ToolName == "Bash" {
				var inp struct {
					Command string `json:"command"`
				}
				if json.Unmarshal(e.ToolInput, &inp) == nil {
					extractBashURLs(s.ID, inp.Command, &accesses)
				}
			}
		}
		return nil
	})
	return accesses, err
}

func extractBashURLs(sessionID, cmd string, out *[]netAccess) {
	lower := strings.ToLower(cmd)
	if !strings.Contains(lower, "curl ") && !strings.Contains(lower, "wget ") {
		return
	}
	matches := urlRe.FindAllString(cmd, -1)
	if len(matches) == 0 {
		*out = append(*out, netAccess{SessionID: sessionID, Tool: "Bash", URL: cmd, Ambiguous: true})
		return
	}
	for _, m := range matches {
		*out = append(*out, netAccess{SessionID: sessionID, Tool: "Bash", URL: m})
	}
}

func domainAllowed(rawURL string, pol *policy.Policy) bool {
	if len(pol.AllowedDomains) == 0 {
		return true
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	host := strings.ToLower(u.Hostname())
	for _, d := range pol.AllowedDomains {
		if host == strings.ToLower(d) || strings.HasSuffix(host, "."+strings.ToLower(d)) {
			return true
		}
	}
	return false
}

// CheckNetwork returns findings for external URL access.
func CheckNetwork(pol *policy.Policy, since time.Time) []result.Finding {
	accesses, err := extractNetwork(since)
	if err != nil {
		return []result.Finding{{Check: "network", Level: result.Warn, Message: "transcript walk error: " + err.Error()}}
	}
	if len(accesses) == 0 {
		return []result.Finding{{Check: "network", Level: result.Pass, Message: "no network requests found in transcripts"}}
	}

	var findings []result.Finding
	for _, a := range accesses {
		if a.Ambiguous {
			findings = append(findings, result.Finding{
				Check:   "network",
				Level:   result.Warn,
				Message: fmt.Sprintf("ambiguous curl/wget in session %s — could not extract URL: %s", a.SessionID, truncate(a.URL, 80)),
			})
			continue
		}
		if !domainAllowed(a.URL, pol) {
			findings = append(findings, result.Finding{
				Check:   "network",
				Level:   result.Warn,
				Message: fmt.Sprintf("unapproved domain in session %s via %s: %s", a.SessionID, a.Tool, a.URL),
			})
		}
	}

	if len(findings) == 0 {
		findings = append(findings, result.Finding{
			Check:   "network",
			Level:   result.Pass,
			Message: fmt.Sprintf("%d network request(s) found, all to approved domains", len(accesses)),
		})
	}
	return findings
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func RunNetwork(args []string) {
	fs := flag.NewFlagSet("network", flag.ExitOnError)
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

	findings := CheckNetwork(pol, sinceTime)
	result.Print(findings, *jsonOut)
}
