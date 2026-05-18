package main

import (
	"fmt"
	"os"

	"github.com/kaihendry/asaguard/internal/mcps"
	"github.com/kaihendry/asaguard/internal/perms"
	"github.com/kaihendry/asaguard/internal/scorer"
	"github.com/kaihendry/asaguard/internal/secrets"
	"github.com/kaihendry/asaguard/internal/settings"
	"github.com/kaihendry/asaguard/internal/transcripts"
	"github.com/kaihendry/asaguard/internal/updater"
)

var version = "dev"

func main() {
	updater.CheckAndUpdate(version)

	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "version", "--version", "-v":
		fmt.Println("asaguard", version)
	case "settings":
		settings.Run(args)
	case "mcps":
		mcps.Run(args)
	case "perms":
		perms.Run(args)
	case "tokens":
		transcripts.RunTokens(args)
	case "network":
		transcripts.RunNetwork(args)
	case "bypass":
		transcripts.RunBypass(args)
	case "sandbox":
		transcripts.RunSandbox(args)
	case "secrets":
		secrets.Run(args)
	case "score":
		scorer.Run(args, version)
	case "check":
		runCheck(args)
	case "install-hooks":
		runInstallHooks(args)
	case "uninstall-hooks":
		runUninstallHooks(args)
	case "help", "--help", "-h":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand: %s\n", cmd)
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Print(`asaguard - Claude Code security guardrail CLI

Usage:
  asaguard <subcommand> [flags]

Audit subcommands:
  settings    Verify settings.json against policy baseline
  mcps        Audit installed MCPs for policy compliance
  perms       Validate permission allow/deny configuration
  tokens      Detect anomalous token usage in transcripts
  network     Report external URLs accessed by agents
  bypass      Detect permission-bypass flags in transcripts
  sandbox     Verify file access stays within authorized paths
  secrets     Check pre-commit secret-scanning hooks
  score       Compute weighted compliance score
  check       Run all checks and print summary

Hook management:
  install-hooks    Install enforcement hooks into settings.json
  uninstall-hooks  Remove installed hooks from settings.json

Flags for install-hooks / uninstall-hooks:
  --evals    Adversarial-review and PII-detection hooks
  --hitl     Human-in-the-loop confirmation hooks
  --banner   Session-start policy banner hook
  --url      Policy URL for banner hook

Global flags (most subcommands):
  --json     Machine-readable JSON output
  --since    Restrict transcript analysis (ISO 8601 date)

SIEM reporting (check subcommand):
  AI_GUARDRAILS_SIEM_ENDPOINT  POST findings to this URL after each run
  AI_GUARDRAILS_SIEM_TOKEN     Bearer token for SIEM endpoint (optional)
  Config fallback: ~/.config/ai-check-guardrails/config.json → "siem_endpoint"

  version    Print version
`)
}

func runCheck(args []string) {
	jsonOut := false
	for _, a := range args {
		if a == "--json" {
			jsonOut = true
		}
	}
	scorer.RunCheck(jsonOut, version)
}

func runInstallHooks(args []string) {
	var evals, hitl, banner bool
	var url string
	for i, a := range args {
		switch a {
		case "--evals":
			evals = true
		case "--hitl":
			hitl = true
		case "--banner":
			banner = true
		case "--url":
			if i+1 < len(args) {
				url = args[i+1]
			}
		}
	}
	if !evals && !hitl && !banner {
		fmt.Fprintln(os.Stderr, "specify at least one of --evals, --hitl, --banner")
		os.Exit(1)
	}
	if evals {
		if err := installEvalsHooks(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	if hitl {
		if err := installHITLHook(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	if banner {
		if err := installBannerHook(url); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}

func runUninstallHooks(args []string) {
	var evals, hitl, banner bool
	for _, a := range args {
		switch a {
		case "--evals":
			evals = true
		case "--hitl":
			hitl = true
		case "--banner":
			banner = true
		}
	}
	if !evals && !hitl && !banner {
		fmt.Fprintln(os.Stderr, "specify at least one of --evals, --hitl, --banner")
		os.Exit(1)
	}
	if evals {
		if err := uninstallEvalsHooks(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	if hitl {
		if err := uninstallHITLHook(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	if banner {
		if err := uninstallBannerHook(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}
