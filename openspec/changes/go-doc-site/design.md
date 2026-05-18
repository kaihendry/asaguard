## Context

`asaguard` enforces security guard rails for Claude Code sessions. Its 11 internal packages — hooks, mcps, perms, policy, result, scorer, secrets, settings, siem, transcripts, updater — each expose a `Check()` or `Run()` function. Currently the only documentation is the source code itself; there is no browsable reference.

The constraint is stdlib-only Go. No Hugo, Docusaurus, or third-party doc generators are permitted. The site must be generatable with a single `make docs` invocation and produce plain HTML files in `docs/`.

## Goals / Non-Goals

**Goals:**
- Run `go doc -all` per package and convert its plain-text output to HTML
- One HTML page per package, plus an `index.html` listing all guard rails
- Regenerate with `make docs` (no external tools required beyond the Go toolchain)
- Output lands in `docs/` (can be committed or git-ignored; `.gitignore` entry added)

**Non-Goals:**
- Hosting, CI publishing, or GitHub Pages automation
- Documenting `cmd/` binaries
- Interactive or JavaScript-heavy UI

## Decisions

### Use a Go generator script rather than a shell script

**Decision**: A small Go program at `cmd/docgen/main.go` drives the site generation.

**Why**: The project is stdlib-only Go; a Go program is more portable than a bash pipeline, handles escaping correctly, and can be run with `go run ./cmd/docgen`. Shell pipelines (e.g. `go doc | sed`) are brittle across platforms and hard to extend.

**Alternative considered**: A `Makefile` loop calling `go doc pkg | pandoc` — rejected because `pandoc` is an external dependency.

### `go doc -all` as the source

**Decision**: Invoke `go doc -all <pkg>` per package via `os/exec` and wrap the output in `<pre>` inside a minimal HTML template.

**Why**: `go doc -all` emits every exported symbol with its doc comment — exactly what operators need. Parsing AST would be overkill; plain-text output in `<pre>` is readable and accurate.

**Alternative considered**: `godoc -html` — deprecated and requires a running server; not suitable for static output.

### Minimal HTML template (no CSS framework)

**Decision**: Inline a small `<style>` block (system font stack, max-width, code block styling). No external CSS.

**Why**: Zero dependencies, loads instantly, works offline. The audience is engineers who value clarity over aesthetics.

## Risks / Trade-offs

- **Stale docs if `make docs` is not re-run** → Mitigation: note in `README.md` to run `make docs` before committing; optionally add a CI check that diffs the output.
- **`go doc` output format may change across Go versions** → Low risk; the format is stable and we only wrap it in `<pre>`.
- **Packages with no doc comments produce sparse pages** → Acceptable for now; adding comments to source packages is incremental follow-up work.
