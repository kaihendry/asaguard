// Package settings verifies that Claude Code is configured to organisational
// standards and has not drifted or been silently overridden.
//
// The settings.json file controls Claude Code's behaviour for every session:
// which hooks run, what banner is shown, and which permission rules apply.
// Security teams need certain keys to be present on every engineer's machine
// (for example a banner linking to the AI usage policy, or a hook that enforces
// logging) and need assurance that those keys have not been quietly removed or
// overridden in settings.local.json. This guard rail detects missing required
// keys, catches locked keys that have been overridden locally, and surfaces
// unexpected configuration drift before it becomes a compliance gap.
package settings

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/kaihendry/asaguard/internal/policy"
	"github.com/kaihendry/asaguard/internal/result"
)

func claudeDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".claude")
}

func readSettings(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return m, nil
}

// Check runs all settings checks and returns results.
func Check(pol *policy.Policy) []result.Finding {
	var findings []result.Finding

	base := filepath.Join(claudeDir(), "settings.json")
	local := filepath.Join(claudeDir(), "settings.local.json")

	baseSettings, err := readSettings(base)
	if err != nil {
		findings = append(findings, result.Finding{Check: "settings", Level: result.Fail, Message: err.Error()})
		return findings
	}
	if baseSettings == nil {
		findings = append(findings, result.Finding{Check: "settings", Level: result.Fail, Message: "settings.json not found at " + base})
		return findings
	}

	localSettings, err := readSettings(local)
	if err != nil {
		findings = append(findings, result.Finding{Check: "settings", Level: result.Warn, Message: err.Error()})
	}

	// Required keys
	for _, k := range pol.RequiredSettingsKeys {
		if _, ok := baseSettings[k]; !ok {
			findings = append(findings, result.Finding{
				Check:   "settings",
				Level:   result.Fail,
				Message: fmt.Sprintf("required key missing: %s", k),
			})
		}
	}

	// Forbidden overrides in settings.local.json
	if localSettings != nil {
		for _, k := range pol.LockedSettingsKeys {
			if _, ok := localSettings[k]; ok {
				findings = append(findings, result.Finding{
					Check:   "settings",
					Level:   result.Fail,
					Message: fmt.Sprintf("locked key overridden in settings.local.json: %s", k),
				})
			}
		}
	}

	// Drift diff — keys in base not in policy required set (info only)
	diff := diffKeys(pol.RequiredSettingsKeys, baseSettings)
	for _, k := range diff.unexpected {
		findings = append(findings, result.Finding{
			Check:   "settings",
			Level:   result.Warn,
			Message: fmt.Sprintf("unexpected key in settings.json: %s", k),
		})
	}

	if len(findings) == 0 {
		findings = append(findings, result.Finding{Check: "settings", Level: result.Pass, Message: "settings.json OK"})
	}
	return findings
}

type keyDiff struct {
	missing    []string
	unexpected []string
}

func diffKeys(required []string, actual map[string]any) keyDiff {
	req := make(map[string]bool, len(required))
	for _, k := range required {
		req[k] = true
	}
	var d keyDiff
	for _, k := range required {
		if _, ok := actual[k]; !ok {
			d.missing = append(d.missing, k)
		}
	}
	if len(required) > 0 {
		for k := range actual {
			if !req[k] {
				d.unexpected = append(d.unexpected, k)
			}
		}
		sort.Strings(d.unexpected)
	}
	return d
}

func Run(args []string) {
	fs := flag.NewFlagSet("settings", flag.ExitOnError)
	jsonOut := fs.Bool("json", false, "JSON output")
	fs.Parse(args)

	pol, err := policy.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "policy load error:", err)
		os.Exit(1)
	}

	findings := Check(pol)
	result.Print(findings, *jsonOut)
	if result.HasFail(findings) {
		os.Exit(1)
	}
}
