package transcripts

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/kaihendry/asaguard/internal/policy"
	"github.com/kaihendry/asaguard/internal/result"
)

type sessionTokens struct {
	ID         string
	Path       string
	InputTokens int
	OutputTokens int
	CacheWrite int
	CacheRead  int
}

func (s sessionTokens) Total() int {
	return s.InputTokens + s.OutputTokens + s.CacheWrite + s.CacheRead
}

func (s sessionTokens) CostUSD(p *policy.TokenPrices) float64 {
	return (float64(s.InputTokens)*p.InputPerMillion+
		float64(s.OutputTokens)*p.OutputPerMillion+
		float64(s.CacheWrite)*p.CacheWritePerMillion+
		float64(s.CacheRead)*p.CacheReadPerMillion) / 1e6
}

func (s sessionTokens) CacheSavingsUSD(p *policy.TokenPrices) float64 {
	return float64(s.CacheRead) * (p.InputPerMillion - p.CacheReadPerMillion) / 1e6
}

func collectTokens(since time.Time) ([]sessionTokens, error) {
	var sessions []sessionTokens
	err := Walk(since, func(s *Session) error {
		var st sessionTokens
		st.ID = s.ID
		st.Path = s.Path
		for _, e := range s.Entries {
			if e.Message != nil && e.Message.Usage != nil {
				u := e.Message.Usage
				st.InputTokens += u.InputTokens
				st.OutputTokens += u.OutputTokens
				st.CacheWrite += u.CacheWrite
				st.CacheRead += u.CacheRead
			}
		}
		if st.Total() > 0 {
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

	var total int
	for _, s := range sessions {
		total += s.Total()
	}
	avg := float64(total) / float64(len(sessions))
	threshold := avg * pol.TokenSpikeMultiple

	prices := &pol.TokenPrices
	var findings []result.Finding
	for _, s := range sessions {
		if float64(s.Total()) > threshold && !math.IsNaN(threshold) {
			findings = append(findings, result.Finding{
				Check:  "tokens",
				Level:  result.Warn,
				Message: fmt.Sprintf("spike in session %s: %d tokens, $%.4f (avg %.0f, threshold %.0f×)", s.ID, s.Total(), s.CostUSD(prices), avg, pol.TokenSpikeMultiple),
			})
		}
	}

	if len(findings) == 0 {
		var totalCost float64
		for _, s := range sessions {
			totalCost += s.CostUSD(prices)
		}
		findings = append(findings, result.Finding{
			Check:  "tokens",
			Level:  result.Pass,
			Message: fmt.Sprintf("%d sessions analysed, avg %.0f tokens, total $%.4f, no spikes detected", len(sessions), avg, totalCost),
		})
	}
	return findings
}

type projectStats struct {
	Project      string
	InputTokens  int
	OutputTokens int
	CacheWrite   int
	CacheRead    int
	CostUSD      float64
	SavingsUSD   float64
}

func printStatsTable(sessions []sessionTokens, prices *policy.TokenPrices) {
	byProject := map[string]*projectStats{}
	home, _ := os.UserHomeDir()
	root := filepath.Join(home, ".claude", "projects")

	for _, s := range sessions {
		proj := filepath.Dir(s.Path)
		if rel, err := filepath.Rel(root, proj); err == nil {
			// Use only the top-level project dir name
			parts := strings.SplitN(rel, string(filepath.Separator), 2)
			proj = parts[0]
		}
		ps := byProject[proj]
		if ps == nil {
			ps = &projectStats{Project: proj}
			byProject[proj] = ps
		}
		ps.InputTokens += s.InputTokens
		ps.OutputTokens += s.OutputTokens
		ps.CacheWrite += s.CacheWrite
		ps.CacheRead += s.CacheRead
		ps.CostUSD += s.CostUSD(prices)
		ps.SavingsUSD += s.CacheSavingsUSD(prices)
	}

	projects := make([]*projectStats, 0, len(byProject))
	for _, ps := range byProject {
		projects = append(projects, ps)
	}
	sort.Slice(projects, func(i, j int) bool { return projects[i].CostUSD > projects[j].CostUSD })

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "PROJECT\tINPUT\tOUTPUT\tCACHE-WRITE\tCACHE-READ\tCOST ($)\tSAVINGS ($)\tCACHE-HIT%")
	for _, ps := range projects {
		totalIn := ps.InputTokens + ps.CacheWrite + ps.CacheRead
		hitPct := 0.0
		if totalIn > 0 {
			hitPct = float64(ps.CacheRead) / float64(totalIn) * 100
		}
		fmt.Fprintf(w, "%s\t%d\t%d\t%d\t%d\t%.4f\t%.4f\t%.1f%%\n",
			ps.Project, ps.InputTokens, ps.OutputTokens, ps.CacheWrite, ps.CacheRead,
			ps.CostUSD, ps.SavingsUSD, hitPct)
	}
	w.Flush()
}

func RunTokens(args []string) {
	fs := flag.NewFlagSet("tokens", flag.ExitOnError)
	jsonOut := fs.Bool("json", false, "JSON output")
	since := fs.String("since", "", "ISO 8601 date (e.g. 2026-01-01)")
	stats := fs.Bool("stats", false, "print per-project cost and cache efficiency table")
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

	if *stats {
		sessions, err := collectTokens(sinceTime)
		if err != nil {
			fmt.Fprintln(os.Stderr, "transcript walk error:", err)
			os.Exit(1)
		}
		if len(sessions) == 0 {
			fmt.Println("0 projects analysed")
			return
		}
		printStatsTable(sessions, &pol.TokenPrices)
		return
	}

	findings := CheckTokens(pol, sinceTime)

	if *jsonOut {
		type findingWithCost struct {
			result.Finding
			CostUSD        float64 `json:"cost_usd,omitempty"`
			CacheSavingsUSD float64 `json:"cache_savings_usd,omitempty"`
		}
		sessions, _ := collectTokens(sinceTime)
		var totalCost, totalSavings float64
		for _, s := range sessions {
			totalCost += s.CostUSD(&pol.TokenPrices)
			totalSavings += s.CacheSavingsUSD(&pol.TokenPrices)
		}
		out := make([]findingWithCost, len(findings))
		for i, f := range findings {
			out[i] = findingWithCost{Finding: f, CostUSD: totalCost, CacheSavingsUSD: totalSavings}
		}
		enc := json.NewEncoder(os.Stdout)
		enc.Encode(out)
	} else {
		result.Print(findings, false)
	}
}
