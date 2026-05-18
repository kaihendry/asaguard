## ADDED Requirements

### Requirement: Run tests on every push
The CI pipeline SHALL run `go test ./...` on every push to any branch to detect regressions before merge.

#### Scenario: Tests pass on push to main
- **WHEN** a commit is pushed to `main`
- **THEN** the `test` job runs `go test ./...` and succeeds before the release job proceeds

#### Scenario: Tests fail on any branch
- **WHEN** a push triggers the workflow and `go test ./...` exits non-zero
- **THEN** the workflow fails and no release is created

### Requirement: Publish a versioned release binary on every commit to main
The CI pipeline SHALL create a Git tag (`v0.0.<run_number>`) and publish cross-compiled release binaries to GitHub Releases on every successful commit to `main`.

#### Scenario: Release created after green tests
- **WHEN** all tests pass on a push to `main`
- **THEN** goreleaser creates a GitHub Release with binaries for `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`, and `windows/amd64`

#### Scenario: No release on branch push
- **WHEN** a commit is pushed to a branch other than `main`
- **THEN** tests run but no release is published

### Requirement: Embed build-time version string
The release workflow SHALL inject the Git tag into the binary at compile time so `asaguard version` reports the real semver tag.

#### Scenario: Release binary reports its version
- **WHEN** a release binary runs `asaguard version`
- **THEN** it prints `asaguard v0.0.<N>` matching the tag used to build it

#### Scenario: Dev build reports "dev"
- **WHEN** a binary built without `-ldflags -X main.version=...` runs `asaguard version`
- **THEN** it prints `asaguard dev`
