package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kaihendry/asaguard/internal/hooks"
	"github.com/kaihendry/asaguard/internal/policy"
)

const (
	hooksDirSuffix = ".config/asaguard/hooks"

	evalReviewCmd = "~/.config/asaguard/hooks/eval-review.sh"
	piiDetectCmd  = "~/.config/asaguard/hooks/pii-detect.sh"
	hitlCmd       = "~/.config/asaguard/hooks/hitl-prompt.sh"
	bannerCmd     = "~/.config/asaguard/hooks/banner.sh"
)

func hooksDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, hooksDirSuffix)
}

// ---- evals ----

func installEvalsHooks() error {
	dir := hooksDir()
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	writeScript(filepath.Join(dir, "eval-review.sh"), evalReviewScript)
	writeScript(filepath.Join(dir, "pii-detect.sh"), piiDetectScript)

	if err := hooks.Install("PreToolUse", ".*", expandHome(evalReviewCmd), "adversarial-review"); err != nil {
		return err
	}
	return hooks.Install("PreToolUse", ".*", expandHome(piiDetectCmd), "PII-detection")
}

func uninstallEvalsHooks() error {
	if err := hooks.Uninstall("PreToolUse", expandHome(evalReviewCmd), "adversarial-review"); err != nil {
		return err
	}
	return hooks.Uninstall("PreToolUse", expandHome(piiDetectCmd), "PII-detection")
}

// ---- hitl ----

func installHITLHook() error {
	dir := hooksDir()
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	writeScript(filepath.Join(dir, "hitl-prompt.sh"), hitlPromptScript)
	return hooks.Install("PreToolUse", ".*", expandHome(hitlCmd), "HITL-enforcer")
}

func uninstallHITLHook() error {
	return hooks.Uninstall("PreToolUse", expandHome(hitlCmd), "HITL-enforcer")
}

// ---- banner ----

func installBannerHook(policyURL string) error {
	pol, err := policy.Load()
	if err != nil {
		return err
	}
	if policyURL != "" {
		pol.Banner.PolicyURL = policyURL
	}

	dir := hooksDir()
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	script := bannerScriptFor(pol.Banner.Text, pol.Banner.PolicyURL)
	dst := filepath.Join(dir, "banner.sh")
	if err := os.WriteFile(dst, []byte(script), 0755); err != nil {
		return fmt.Errorf("writing banner.sh: %w", err)
	}
	return hooks.Install("PreToolUse", ".*", expandHome(bannerCmd), "session-banner")
}

func uninstallBannerHook() error {
	return hooks.Uninstall("PreToolUse", expandHome(bannerCmd), "session-banner")
}

// ---- helpers ----

func expandHome(p string) string {
	home, _ := os.UserHomeDir()
	if len(p) >= 2 && p[:2] == "~/" {
		return filepath.Join(home, p[2:])
	}
	return p
}

func writeScript(path, content string) {
	if err := os.WriteFile(path, []byte(content), 0755); err != nil {
		fmt.Fprintln(os.Stderr, "warning: could not write", path, ":", err)
	}
}

// ---- embedded hook scripts ----

const evalReviewScript = `#!/bin/sh
# Adversarial-review hook: logs tool calls for review.
LOGFILE="${HOME}/.claude/asaguard-eval.log"
echo "$(date -u +%FT%TZ) tool=${CLAUDE_TOOL_NAME} session=${CLAUDE_SESSION_ID}" >> "${LOGFILE}"
exit 0
`

const piiDetectScript = `#!/bin/sh
# PII-detection hook: warns if tool input contains PII patterns.
INPUT="${CLAUDE_TOOL_INPUT:-}"
if echo "${INPUT}" | grep -qE '[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}'; then
  echo "WARN [asaguard/pii] possible email address in tool input for ${CLAUDE_TOOL_NAME}" >&2
fi
if echo "${INPUT}" | grep -qE '\b[0-9]{3}-[0-9]{2}-[0-9]{4}\b'; then
  echo "WARN [asaguard/pii] possible SSN pattern in tool input for ${CLAUDE_TOOL_NAME}" >&2
fi
exit 0
`

const hitlPromptScript = `#!/bin/sh
# HITL-enforcer hook: prompts for confirmation on sensitive tool calls.
LOGFILE="${HOME}/.claude/asaguard-hitl.log"
TOOL="${CLAUDE_TOOL_NAME:-unknown}"
INPUT="${CLAUDE_TOOL_INPUT:-}"

# Default watchlist patterns
WATCHLIST="rm -rf|git push|curl |wget |ssh |scp "

if echo "${INPUT}" | grep -qE "${WATCHLIST}"; then
  echo "" >&2
  echo "⚠  HITL [asaguard]: sensitive operation detected in ${TOOL}" >&2
  echo "   Input preview: $(echo "${INPUT}" | head -c 200)" >&2
  printf "   Approve? [y/N] " >&2
  read -r ans </dev/tty
  if [ "$(echo "${ans}" | tr '[:upper:]' '[:lower:]')" = "y" ]; then
    echo "$(date -u +%FT%TZ) APPROVED tool=${TOOL} session=${CLAUDE_SESSION_ID}" >> "${LOGFILE}"
    exit 0
  else
    echo "$(date -u +%FT%TZ) DENIED tool=${TOOL} session=${CLAUDE_SESSION_ID}" >> "${LOGFILE}"
    echo "Blocked by asaguard HITL hook." >&2
    exit 1
  fi
fi
exit 0
`

func bannerScriptFor(text, policyURL string) string {
	urlLine := ""
	if policyURL != "" {
		urlLine = fmt.Sprintf(`  echo "   Policy: %s" >&2`, policyURL)
	}
	return fmt.Sprintf(`#!/bin/sh
# Session-start banner hook.
SENTINEL="${TMPDIR:-/tmp}/asaguard-banner-${CLAUDE_SESSION_ID}"
if [ ! -f "${SENTINEL}" ]; then
  touch "${SENTINEL}"
  echo "" >&2
  echo "  %s" >&2
%s
  echo "" >&2
fi
exit 0
`, text, urlLine)
}
