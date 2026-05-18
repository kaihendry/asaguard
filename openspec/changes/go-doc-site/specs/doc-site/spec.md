## ADDED Requirements

### Requirement: Doc generator command exists
The project SHALL provide a `cmd/docgen/main.go` program that, when run via `go run ./cmd/docgen`, generates a static HTML documentation site in the `docs/` directory using only Go stdlib.

#### Scenario: Generator runs successfully
- **WHEN** `go run ./cmd/docgen` is executed from the repository root
- **THEN** a `docs/` directory is created containing `index.html` and one HTML file per internal package

### Requirement: Per-package HTML pages
For each package under `internal/` the generator SHALL produce an HTML file at `docs/<package>.html` whose main content is the full output of `go doc -all github.com/kaihendry/asaguard/internal/<package>` rendered inside a `<pre>` block.

#### Scenario: Package page contains go doc output
- **WHEN** `docs/hooks.html` is opened in a browser
- **THEN** the page displays the exported symbols and doc comments for `internal/hooks` as plain preformatted text

#### Scenario: All packages are covered
- **WHEN** the generator finishes
- **THEN** HTML files exist for hooks, mcps, perms, policy, result, scorer, secrets, settings, siem, transcripts, and updater

### Requirement: Index page lists all guard rails
The generator SHALL produce `docs/index.html` with a linked list of all guard-rail package pages.

#### Scenario: Index links to each package
- **WHEN** `docs/index.html` is opened
- **THEN** each internal package name appears as a clickable link to its corresponding `.html` file

### Requirement: Makefile docs target
The `Makefile` SHALL contain a `docs` target that runs `go run ./cmd/docgen` to regenerate `docs/`.

#### Scenario: Make docs regenerates site
- **WHEN** `make docs` is run
- **THEN** the `docs/` directory is created or updated with current go doc output

### Requirement: No external dependencies
The doc generator MUST use only the Go standard library (`os/exec`, `html/template`, `os`, `strings`, etc.) and MUST NOT introduce new entries in `go.mod`.

#### Scenario: go.mod unchanged
- **WHEN** the doc generator is implemented
- **THEN** `go.mod` and `go.sum` are identical to their pre-change state
