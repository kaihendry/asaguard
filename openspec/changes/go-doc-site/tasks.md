## 1. Generator scaffolding

- [x] 1.1 Create `cmd/docgen/` directory and `cmd/docgen/main.go` with a `main()` stub
- [x] 1.2 Define the list of internal packages to document (`hooks`, `mcps`, `perms`, `policy`, `result`, `scorer`, `secrets`, `settings`, `siem`, `transcripts`, `updater`)

## 2. go doc invocation

- [x] 2.1 Implement a helper that runs `go doc -all <pkg>` via `os/exec` and returns its stdout as a string
- [x] 2.2 Add error handling so a missing package prints a warning and continues (does not abort the whole run)

## 3. HTML template

- [x] 3.1 Write an `html/template` for per-package pages: title, minimal inline CSS (system font, max-width, `<pre>` styling), and a `<pre>` block for go doc output
- [x] 3.2 Write an `html/template` for `index.html`: page title, brief intro sentence, unordered list of linked package names

## 4. File generation

- [x] 4.1 Implement `ensureDocsDir()` that creates `docs/` if it does not exist
- [x] 4.2 For each package, render the per-package template and write `docs/<package>.html`
- [x] 4.3 Render the index template (passing the package list with relative hrefs) and write `docs/index.html`

## 5. Makefile integration

- [x] 5.1 Add a `docs` target to `Makefile` that runs `go run ./cmd/docgen`
- [x] 5.2 Add `docs/` to `.gitignore` (or document in README that it is a build artifact)

## 6. Smoke test

- [x] 6.1 Run `make docs` and verify `docs/index.html` and all 11 package pages exist
- [x] 6.2 Open `docs/index.html` in a browser (or `cat`) and confirm links and go doc content are present
