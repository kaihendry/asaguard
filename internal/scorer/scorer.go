package scorer

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/kaihendry/asaguard/internal/mcps"
	"github.com/kaihendry/asaguard/internal/perms"
	"github.com/kaihendry/asaguard/internal/policy"
	"github.com/kaihendry/asaguard/internal/result"
	"github.com/kaihendry/asaguard/internal/secrets"
	"github.com/kaihendry/asaguard/internal/settings"
	"github.com/kaihendry/asaguard/internal/siem"
	"github.com/kaihendry/asaguard/internal/transcripts"
)

type CheckResult struct {
	Name   string         `json:"check"`
	Weight int            `json:"weight"`
	Status string         `json:"status"`
	Score  int            `json:"score"`
	Notes  []result.Finding `json:"findings,omitempty"`
}

type ScoreReport struct {
	Total  int           `json:"total"`
	Status string        `json:"status"`
	Checks []CheckResult `json:"checks"`
}

func statusForFindings(findings []result.Finding) string {
	for _, f := range findings {
		if f.Level == result.Fail {
			return "FAIL"
		}
	}
	for _, f := range findings {
		if f.Level == result.Warn {
			return "WARN"
		}
	}
	return "PASS"
}

func tier(score int) string {
	switch {
	case score >= 85:
		return "GOOD"
	case score >= 70:
		return "ACCEPTABLE"
	case score >= 50:
		return "AT RISK"
	default:
		return "CRITICAL"
	}
}

func runAllChecks(pol *policy.Policy) []CheckResult {
	zero := time.Time{}
	checkMap := map[string][]result.Finding{
		"settings": settings.Check(pol),
		"mcps":     mcps.Check(pol),
		"perms":    perms.Check(),
		"tokens":   transcripts.CheckTokens(pol, zero),
		"network":  transcripts.CheckNetwork(pol, zero),
		"bypass":   transcripts.CheckBypass(zero),
		"sandbox":  transcripts.CheckSandbox(pol, zero),
		"secrets":  secrets.Check(),
	}

	var results []CheckResult
	for name, findings := range checkMap {
		weight := pol.Weights[name]
		if weight == 0 {
			weight = 100 / len(checkMap)
		}
		status := statusForFindings(findings)
		score := 0
		switch status {
		case "PASS":
			score = weight
		case "WARN":
			score = weight / 2
		}
		results = append(results, CheckResult{
			Name:   name,
			Weight: weight,
			Status: status,
			Score:  score,
			Notes:  findings,
		})
	}
	return results
}

// Run executes the scorer subcommand.
func Run(args []string, version string) {
	start := time.Now()
	fs := flag.NewFlagSet("score", flag.ExitOnError)
	jsonOut := fs.Bool("json", false, "JSON output")
	fs.Parse(args)

	pol, err := policy.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "policy load error:", err)
		os.Exit(1)
	}

	checks := runAllChecks(pol)
	report := buildReport(checks)

	exitCode := 0
	if report.Total < 50 {
		exitCode = 1
	}

	if *jsonOut {
		json.NewEncoder(os.Stdout).Encode(report)
		siem.Report(version, allFindings(report), report.Total, exitCode, start)
		return
	}
	printReport(report)
	siem.Report(version, allFindings(report), report.Total, exitCode, start)
	if exitCode == 1 {
		os.Exit(1)
	}
}

// RunCheck is called by `asaguard check` to run all checks and print summary.
func RunCheck(jsonOut bool, version string) {
	start := time.Now()
	pol, err := policy.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "policy load error:", err)
		os.Exit(1)
	}

	checks := runAllChecks(pol)
	report := buildReport(checks)

	exitCode := 0
	if report.Total < 50 {
		exitCode = 1
	}

	if jsonOut {
		json.NewEncoder(os.Stdout).Encode(report)
		siem.Report(version, allFindings(report), report.Total, exitCode, start)
		return
	}

	// Print per-check findings then summary
	for _, c := range checks {
		result.Print(c.Notes, false)
	}
	fmt.Println()
	printReport(report)
	siem.Report(version, allFindings(report), report.Total, exitCode, start)
	if exitCode == 1 {
		os.Exit(1)
	}
}

func allFindings(r ScoreReport) []result.Finding {
	var out []result.Finding
	for _, c := range r.Checks {
		out = append(out, c.Notes...)
	}
	return out
}

func buildReport(checks []CheckResult) ScoreReport {
	total := 0
	for _, c := range checks {
		total += c.Score
	}
	return ScoreReport{
		Total:  total,
		Status: tier(total),
		Checks: checks,
	}
}

func printReport(r ScoreReport) {
	fmt.Printf("%-12s %-8s %-6s %s\n", "CHECK", "STATUS", "WEIGHT", "SCORE")
	fmt.Println("---------------------------------------------")
	for _, c := range r.Checks {
		fmt.Printf("%-12s %-8s %-6d %d\n", c.Name, c.Status, c.Weight, c.Score)
	}
	fmt.Println("---------------------------------------------")
	color := ""
	reset := ""
	if r.Total >= 85 {
		color = "\033[32m" // green
	} else if r.Total < 50 {
		color = "\033[31m" // red
	} else {
		color = "\033[33m" // yellow
	}
	if os.Getenv("NO_COLOR") == "" {
		reset = "\033[0m"
	} else {
		color = ""
	}
	fmt.Printf("\n%sCompliance score: %s (%d/100)%s\n", color, r.Status, r.Total, reset)
}
