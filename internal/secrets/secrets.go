package secrets

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/hendry/asaguard/internal/result"
)

var knownScanners = []string{"gitleaks", "trufflehog", "detect-secrets"}

// Check verifies pre-commit secret-scanning hooks are in place.
func Check() []result.Finding {
	var findings []result.Finding

	hookPath := ".git/hooks/pre-commit"
	info, err := os.Stat(hookPath)
	if os.IsNotExist(err) {
		// Try pre-commit-config.yaml as an alternative signal
		findings = append(findings, result.Finding{
			Check:   "secrets",
			Level:   result.Fail,
			Message: "pre-commit hook not installed: " + hookPath + " does not exist",
		})
		findings = append(findings, checkPreCommitConfig()...)
		return findings
	}
	if err != nil {
		findings = append(findings, result.Finding{Check: "secrets", Level: result.Fail, Message: err.Error()})
		return findings
	}

	if info.Mode()&0111 == 0 {
		findings = append(findings, result.Finding{
			Check:   "secrets",
			Level:   result.Fail,
			Message: "pre-commit hook is not executable: " + hookPath,
		})
	}

	if !hookContainsScanner(hookPath) {
		// Fall back to checking .pre-commit-config.yaml
		configFindings := checkPreCommitConfig()
		if len(configFindings) > 0 {
			findings = append(findings, configFindings...)
		} else {
			findings = append(findings, result.Finding{
				Check:   "secrets",
				Level:   result.Fail,
				Message: "no secret scanner (gitleaks/trufflehog/detect-secrets) detected in pre-commit hook",
			})
		}
	} else {
		findings = append(findings, result.Finding{
			Check:   "secrets",
			Level:   result.Pass,
			Message: "pre-commit hook present and contains a secret scanner",
		})
	}

	return findings
}

func hookContainsScanner(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.ToLower(sc.Text())
		for _, s := range knownScanners {
			if strings.Contains(line, s) {
				return true
			}
		}
	}
	return false
}

func checkPreCommitConfig() []result.Finding {
	configPath := ".pre-commit-config.yaml"
	data, err := os.ReadFile(configPath)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return []result.Finding{{Check: "secrets", Level: result.Warn, Message: "could not read .pre-commit-config.yaml: " + err.Error()}}
	}

	content := strings.ToLower(string(data))
	for _, s := range knownScanners {
		if strings.Contains(content, s) {
			return []result.Finding{{
				Check:   "secrets",
				Level:   result.Pass,
				Message: fmt.Sprintf(".pre-commit-config.yaml references %s", s),
			}}
		}
	}
	return []result.Finding{{
		Check:   "secrets",
		Level:   result.Fail,
		Message: ".pre-commit-config.yaml found but no secret scanner referenced",
	}}
}

func Run(args []string) {
	fs := flag.NewFlagSet("secrets", flag.ExitOnError)
	jsonOut := fs.Bool("json", false, "JSON output")
	fs.Parse(args)

	findings := Check()
	result.Print(findings, *jsonOut)
	if result.HasFail(findings) {
		os.Exit(1)
	}
}
