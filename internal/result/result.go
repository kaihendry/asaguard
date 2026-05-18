// Package result defines the common finding type shared by all guard rails.
//
// Every check returns a slice of Finding values, each carrying a severity level
// (PASS, WARN, or FAIL), the name of the check that produced it, and a
// human-readable message. This uniform structure lets the scorer aggregate
// results from all checks, drives the SIEM reporter, and means any guard rail
// can be run and read in isolation with a consistent output format.
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
