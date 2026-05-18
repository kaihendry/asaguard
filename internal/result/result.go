// Package result defines the shared output format for every guard rail finding.
//
// Each check emits a slice of Finding values with a severity level — PASS,
// WARN, or FAIL — along with the check name and a plain-English explanation.
// The consistent format means findings from any guard rail can be aggregated
// by the scorer, forwarded to a SIEM, or printed on the terminal without any
// per-check adapters. It also means engineers see the same output shape
// whether they run a single check or the full suite.
package result

import (
	"encoding/json"
	"fmt"
	"os"
)

type Level string

const (
	Pass Level = "PASS"
	Warn Level = "WARN"
	Fail Level = "FAIL"
)

type Finding struct {
	Check   string `json:"check"`
	Level   Level  `json:"level"`
	Message string `json:"message"`
}

func HasFail(findings []Finding) bool {
	for _, f := range findings {
		if f.Level == Fail {
			return true
		}
	}
	return false
}

func Print(findings []Finding, asJSON bool) {
	if asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.Encode(findings)
		return
	}
	for _, f := range findings {
		icon := checkIcon(f.Level)
		fmt.Printf("%s [%s] %s: %s\n", icon, f.Level, f.Check, f.Message)
	}
}

func checkIcon(l Level) string {
	switch l {
	case Pass:
		return "✓"
	case Warn:
		return "!"
	case Fail:
		return "✗"
	default:
		return " "
	}
}
