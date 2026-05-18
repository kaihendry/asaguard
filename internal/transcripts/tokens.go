package transcripts

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/hendry/asaguard/internal/policy"
	"github.com/hendry/asaguard/internal/result"
)

type sessionTokens struct {
	ID     string
	Input  int
	Output int
	Total  int
}

func collectTokens(since time.Time) ([]sessionTokens, error) {
	var sessions []sessionTokens
	err := Walk(since, func(s *Session) error {
		var st sessionTokens
		st.ID = s.ID
		for _, e := range s.Entries {
			if e.Message != nil && e.Message.Usage != nil {
				u := e.Message.Usage
				st.Input += u.InputTokens + u.CacheRead + u.CacheWrite
				st.Output += u.OutputTokens
			}
		}
		st.Total = st.Input + st.Output
		if st.Total > 0 {
			sessions = append(sessions, st)
		}
		return nil
	})
	return sessions, err
}

// CheckTokens returns findings for anomalous token usage.
func CheckTokens(pol *policy.Policy, since time.Time) []result.Finding {
	sessions, err := collectTokens(since)
	if err != nil {
		return []result.Finding{{Check: "tokens", Level: result.Warn, Message: "transcript walk error: " + err.Error()}}
	}
	if len(sessions) == 0 {
		return []result.Finding{{Check: "tokens", Level: result.Pass, Message: "no sessions found"}}
	}

	// Compute rolling average
	var total int
	for _, s := range sessions {
		total += s.Total
	}
	avg := float64(total) / float64(len(sessions))
	threshold := avg * pol.TokenSpikeMultiple

	var findings []result.Finding
	for _, s := range sessions {
		if float64(s.Total) > threshold && !math.IsNaN(threshold) {
			findings = append(findings, result.Finding{
				Check:   "tokens",
				Level:   result.Warn,
				Message: fmt.Sprintf("spike in session %s: %d tokens (avg %.0f, threshold %.0f×)", s.ID, s.Total, avg, pol.TokenSpikeMultiple),
			})
		}
	}

	if len(findings) == 0 {
		findings = append(findings, result.Finding{
			Check:   "tokens",
			Level:   result.Pass,
			Message: fmt.Sprintf("%d sessions analysed, avg %.0f tokens, no spikes detected", len(sessions), avg),
		})
	}
	return findings
}

func RunTokens(args []string) {
	fs := flag.NewFlagSet("tokens", flag.ExitOnError)
	jsonOut := fs.Bool("json", false, "JSON output")
	since := fs.String("since", "", "ISO 8601 date (e.g. 2026-01-01)")
	fs.Parse(args)

	var sinceTime time.Time
	if *since != "" {
		var err error
		sinceTime, err = time.Parse("2006-01-02", *since)
		if err != nil {
			fmt.Fprintln(os.Stderr, "invalid --since date:", err)
			os.Exit(1)
		}
	}

	pol, err := policy.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "policy load error:", err)
		os.Exit(1)
	}

	findings := CheckTokens(pol, sinceTime)
	if *jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.Encode(findings)
	} else {
		result.Print(findings, false)
	}
}
