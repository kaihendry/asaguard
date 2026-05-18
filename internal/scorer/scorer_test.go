package scorer

import (
	"testing"

	"github.com/kaihendry/asaguard/internal/result"
)

func TestTier(t *testing.T) {
	cases := []struct {
		score int
		want  string
	}{
		{100, "GOOD"},
		{85, "GOOD"},
		{84, "ACCEPTABLE"},
		{70, "ACCEPTABLE"},
		{69, "AT RISK"},
		{50, "AT RISK"},
		{49, "CRITICAL"},
		{0, "CRITICAL"},
	}
	for _, c := range cases {
		if got := tier(c.score); got != c.want {
			t.Errorf("tier(%d) = %s, want %s", c.score, got, c.want)
		}
	}
}

func TestStatusForFindings(t *testing.T) {
	pass := []result.Finding{{Level: result.Pass}}
	warn := []result.Finding{{Level: result.Pass}, {Level: result.Warn}}
	fail := []result.Finding{{Level: result.Warn}, {Level: result.Fail}}

	if statusForFindings(pass) != "PASS" {
		t.Error("expected PASS")
	}
	if statusForFindings(warn) != "WARN" {
		t.Error("expected WARN")
	}
	if statusForFindings(fail) != "FAIL" {
		t.Error("expected FAIL")
	}
}

func TestBuildReport(t *testing.T) {
	checks := []CheckResult{
		{Name: "settings", Weight: 50, Status: "PASS", Score: 50},
		{Name: "mcps", Weight: 50, Status: "FAIL", Score: 0},
	}
	r := buildReport(checks)
	if r.Total != 50 {
		t.Errorf("expected total 50, got %d", r.Total)
	}
	if r.Status != "AT RISK" {
		t.Errorf("expected AT RISK, got %s", r.Status)
	}
}
