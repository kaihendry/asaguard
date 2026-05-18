package result

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestHasFail(t *testing.T) {
	if HasFail([]Finding{{Level: Pass}, {Level: Warn}}) {
		t.Error("expected no fail")
	}
	if !HasFail([]Finding{{Level: Pass}, {Level: Fail}}) {
		t.Error("expected fail detected")
	}
}

func TestPrintText(t *testing.T) {
	// Redirect stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	Print([]Finding{
		{Check: "settings", Level: Pass, Message: "ok"},
		{Check: "mcps", Level: Warn, Message: "watch out"},
		{Check: "bypass", Level: Fail, Message: "bad"},
	}, false)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	out := buf.String()

	if !strings.Contains(out, "PASS") || !strings.Contains(out, "WARN") || !strings.Contains(out, "FAIL") {
		t.Errorf("unexpected output: %s", out)
	}
}

func TestPrintJSON(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	Print([]Finding{{Check: "settings", Level: Pass, Message: "ok"}}, true)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	out := buf.String()

	if !strings.Contains(out, `"level"`) {
		t.Errorf("expected JSON output, got: %s", out)
	}
}
