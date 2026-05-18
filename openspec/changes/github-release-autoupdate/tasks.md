## 1. Repo and Module Preparation

- [x] 1.1 Create the public GitHub repo `kaihendry/asaguard` (public, no auto-init)
- [x] 1.2 Update `go.mod` module path from `github.com/hendry/asaguard` to `github.com/kaihendry/asaguard` and fix all internal import paths
- [x] 1.3 Push existing code to `kaihendry/asaguard` main branch

## 2. Version Embedding

- [x] 2.1 Change `const version = "0.1.0"` in `cmd/asaguard/main.go` to `var version = "dev"`
- [x] 2.2 Verify `asaguard version` prints `asaguard dev` for a plain `go build` output

## 3. Custom Self-Updater

- [x] 3.1 Create `internal/updater/updater.go` with a `CheckAndUpdate(version string)` function that:
  - Returns immediately when `version == "dev"`
  - GETs `https://api.github.com/repos/kaihendry/asaguard/releases/latest` with a 5 s timeout
  - Compares `tag_name` lexicographically against `version`; skips if not newer
  - Finds the release asset matching `asaguard_<GOOS>_<GOARCH>` in the assets list
  - Downloads the asset to a temp file, sets executable bit, renames it over `os.Executable()`
  - Re-execs via `syscall.Exec` (Unix) or `exec.Command` + `os.Exit(0)` (Windows)
- [x] 3.2 Call `updater.CheckAndUpdate(version)` at the top of `main()` before the subcommand switch
- [x] 3.3 Write a unit test confirming `CheckAndUpdate("dev")` makes no network calls

## 4. GoReleaser Configuration

- [x] 4.1 Add `.goreleaser.yaml` with `project_name: asaguard`, builds for `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`, `windows/amd64`, ldflags injecting `main.version` from the tag, and a checksum file
- [x] 4.2 Run `goreleaser check` locally to validate the config

## 5. GitHub Actions Workflow

- [x] 5.1 Create `.github/workflows/release.yml` with:
  - A `test` job running `go test ./...` on every push
  - A `release` job that runs only on `main`, depends on `test`, creates tag `v$(date -u +%Y%m%d).${{ github.run_number }}`, and runs `goreleaser release --clean`
- [x] 5.2 Confirm the workflow file passes `actionlint` (or GitHub Actions syntax check)

## 6. End-to-End Verification

- [x] 6.1 Push a commit to `main` and confirm the Actions run, a tag like `v20260518.42` is created, and binaries appear under the GitHub Release
- [x] 6.2 Download the release binary, run `asaguard version`, confirm it prints the date-based tag (not `dev`)
- [x] 6.3 Push a second commit, download the first release binary, run it, and confirm it self-updates to the second release
