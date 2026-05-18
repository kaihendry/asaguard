## Context

`asaguard` is a single Go binary (`cmd/asaguard/main.go`) with a hard-coded `const version = "0.1.0"`. There is no CI pipeline, no release process, and no mechanism to distribute updates to users. The module path is `github.com/hendry/asaguard`; the target public repo is `kaihendry/asaguard`.

## Goals / Non-Goals

**Goals:**
- Run `go test ./...` on every push via GitHub Actions
- Automatically tag and publish cross-compiled release binaries on every commit to `main`
- Embed the git tag as the version string at build time (replacing the hard-coded constant)
- Release binaries check GitHub Releases for a newer version at startup; if found, download it, replace themselves, and re-exec

**Non-Goals:**
- Interactive update prompts or `--update` subcommands (update is silent/automatic at startup)
- Update checks for dev builds (i.e., binaries built with `go build` without version injection)
- Pinning to specific versions or rollback

## Decisions

### Version embedding via `-ldflags`

The hard-coded `const version = "0.1.0"` is replaced with a `var version = "dev"` default. The release workflow passes `-ldflags "-X main.version=<tag>"` so release binaries carry the real date-based tag. Dev builds retain `"dev"`, which the self-updater uses as the sentinel to skip the update check.

**Alternative considered**: `go:generate` writing a version file — rejected, adds file churn.

### Release tooling: `goreleaser`

GoReleaser handles cross-compilation (`linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`, `windows/amd64`), asset naming, checksum files, and GitHub Release creation in one tool. The workflow triggers on every push to `main` by creating a date-based tag before invoking goreleaser.

**Alternative considered**: hand-rolled `go build` matrix in the workflow — rejected, goreleaser is standard and handles checksums/archives correctly.

### Auto-tagging strategy: date-based `v20260518.<run_number>`

Every commit to `main` gets a tag of the form `v20260518.<run_number>` (UTC date + Actions run number). The date component makes tags human-readable at a glance; the run number disambiguates multiple commits on the same day. Lexicographic string comparison (`>`) is sufficient to determine "newer" since the format sorts correctly without semver parsing.

**Alternative considered**: `v0.0.<run_number>` — rejected, opaque and gives no date context.

### Self-update implementation: custom ~60-line package, no external dependency

The updater calls `https://api.github.com/repos/kaihendry/asaguard/releases/latest`, reads the `tag_name`, compares it lexicographically against the running version, finds the asset whose name matches `asaguard_<GOOS>_<GOARCH>` (the goreleaser naming convention), streams it to a temp file, `chmod`s it executable, and atomically renames it over the current binary via `os.Rename`. On Unix it then re-execs with `syscall.Exec`; on Windows it uses `exec.Command` + `os.Exit(0)`.

For a security-focused tool, a self-contained updater with no transitive dependencies is preferable: the code is short enough to audit in one sitting, and there is no risk of a third-party library introducing its own update or network behaviour.

**Alternative considered**: `github.com/creativeprojects/go-selfupdate` — rejected: adds an external dependency with its own transitive graph to a security tool, and the problem is simple enough not to warrant it.

### Update check placement: startup, before subcommand dispatch

The update check runs synchronously at the top of `main()` only when `version != "dev"`. It has a short timeout (5 s) so it does not noticeably delay normal use. After a successful self-replace, re-exec runs the new binary with the same arguments transparently.

**Alternative considered**: background goroutine that prints a message — rejected, silent self-replace is the stated goal.

## Risks / Trade-offs

- **Every commit triggers a release** → release list grows quickly. Mitigation: goreleaser can mark releases as pre-release if desired later.
- **Startup latency** → 5 s timeout cap; on slow or offline networks the check fails gracefully and is silently skipped.
- **Binary replacement on Windows** → `os.Rename` over a running executable works on Windows only after the old handle is closed; the custom updater writes to a temp path first, then renames, which is safe. Re-exec uses `exec.Command` + `os.Exit(0)` since `syscall.Exec` is unavailable.
- **Module path mismatch** → `go.mod` currently has `github.com/hendry/asaguard`; the public repo will be `kaihendry/asaguard`. The module path must be updated to `github.com/kaihendry/asaguard` so imports resolve correctly.
