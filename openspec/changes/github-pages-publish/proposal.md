## Why

The `go-doc-site` change generates a static HTML reference for all guard-rail packages, but it only works locally. Publishing it to GitHub Pages means operators can bookmark a URL instead of cloning the repo and running `make docs`.

## What Changes

- A new GitHub Actions workflow (`.github/workflows/docs.yml`) that runs `make docs` and deploys `docs/` to GitHub Pages on every push to `main`
- Remove `/docs/` from `.gitignore` so the workflow can reference the generated output (GitHub Pages deployment uses the workflow artifact, not a committed folder, so `.gitignore` stays as-is)

## Capabilities

### New Capabilities

- `pages-publish`: GitHub Actions workflow that builds and publishes the doc site to GitHub Pages automatically on push to `main`

### Modified Capabilities

<!-- none -->

## Non-goals

- Custom domain configuration
- Per-PR preview deployments
- Any changes to the doc generator itself (`cmd/docgen`)

## Impact

- New file: `.github/workflows/docs.yml`
- GitHub repository Pages setting must be set to "GitHub Actions" source (one-time manual step)
- No Go code changes; no `go.mod` changes
