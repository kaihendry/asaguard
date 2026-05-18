## Why

The `asaguard` tool has no public distribution channel — users must build from source. Publishing versioned releases to GitHub and enabling the binary to self-update removes that friction and ensures users always run the latest security guardrails without manual intervention.

## What Changes

- Create a GitHub Actions workflow that runs `go test ./...` on every push and creates a tagged release binary on every commit to `main`
- Add a `goreleaser` (or equivalent) configuration to cross-compile and attach binaries to GitHub Releases
- Embed a build-time version string so release binaries know their own version
- Add auto-update logic that runs at startup for non-dev binaries: check the latest GitHub release, and if newer, download and replace the running binary

## Capabilities

### New Capabilities

- `release-workflow`: GitHub Actions CI/CD pipeline that runs tests and publishes versioned release binaries on every commit to `main`
- `self-updater`: Runtime update check and self-replacement for release binaries — detects newer version on GitHub, downloads it, replaces the current binary, and re-executes

### Modified Capabilities

- `banner-installer`: Version string embedded at build time will be surfaced in the existing banner/version output

## Impact

- New files: `.github/workflows/release.yml`, `.goreleaser.yaml` (or `Makefile` release target)
- Modified: `main.go` or root command to embed version and call update check on startup
- Dependency: `github.com/creativeprojects/go-selfupdate` or similar for update logic
- Requires the GitHub repo `kaihendry/asaguard` to exist and have Actions enabled with a `GITHUB_TOKEN` that can create releases
