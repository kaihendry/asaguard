## Why

The guard rails implemented across `internal/` packages lack a browsable reference that stays current with the code. Developers and operators need a single URL to understand what each guard rail checks, how to configure it, and what findings it emits — without reading source files.

## What Changes

- A static doc site generated from `go doc` output and inline source comments for each guard rail package
- A `make docs` (or `go generate`) target that regenerates the site whenever source changes
- Published HTML pages for each package: hooks, mcps, perms, policy, result, scorer, secrets, settings, siem, transcripts, updater

## Capabilities

### New Capabilities

- `doc-site`: Static HTML documentation site auto-generated from Go source (`go doc`) for all internal guard-rail packages, served as plain files (no external runtime dependency)

### Modified Capabilities

<!-- none -->

## Non-goals

- Interactive API explorer or Swagger-style UI
- Hosting / CI deployment pipeline (out of scope for this change)
- Documentation for `cmd/` binaries (only `internal/` packages)

## Impact

- New top-level `docs/` directory (generated, git-ignored or committed as build artifact)
- `Makefile` gains a `docs` target
- No changes to existing Go packages or their exported APIs
