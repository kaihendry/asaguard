## Context

The repo already has `.github/workflows/release.yml` using `actions/checkout@v4`, `actions/setup-go@v5`, and push-triggered jobs. GitHub Pages now supports deploying directly from a workflow artifact via `actions/upload-pages-artifact` + `actions/deploy-pages`, without needing a `gh-pages` branch or a committed `docs/` folder.

## Goals / Non-Goals

**Goals:**
- Run `make docs` in CI and publish the output to GitHub Pages on every push to `main`
- Use the official GitHub Pages Actions (`actions/upload-pages-artifact`, `actions/deploy-pages`) — no third-party publisher
- Keep the workflow isolated from the existing `release.yml` (separate file, separate job)

**Non-Goals:**
- Custom domain, PR previews, or any changes to `cmd/docgen`

## Decisions

### Use `actions/deploy-pages` (artifact-based), not `gh-pages` branch

**Decision**: Upload `docs/` as a Pages artifact and deploy via `actions/deploy-pages@v4`.

**Why**: The `gh-pages` branch approach requires force-pushing generated content into git history, polluting the log and requiring `persist-credentials`. The artifact approach is cleaner, officially supported, and aligns with GitHub's recommended path for Actions-driven Pages.

**Alternative considered**: `peaceiris/actions-gh-pages` — third-party, adds an external dependency, and the official actions now cover the same use case.

### Separate workflow file

**Decision**: `.github/workflows/docs.yml`, not merged into `release.yml`.

**Why**: Docs publishing has different permissions (`pages: write`, `id-token: write`) and a different trigger scope (only `main`). Keeping it separate means a failure in docs publishing never blocks a release, and the permissions are narrowly scoped.

### Required repository permission: `pages: write` + `id-token: write`

The `actions/deploy-pages` action requires these two permissions at the job level. `contents: read` is sufficient for checkout — no write access to the repo tree is needed.

## Risks / Trade-offs

- **GitHub Pages not enabled on the repo** → One-time manual step: Settings → Pages → Source → "GitHub Actions". The workflow will fail with a clear error until this is done.
- **`make docs` changes over time** → The workflow always runs the current generator, so docs stay in sync automatically. No risk of stale output.
- **Concurrent deployments** → GitHub serialises Pages deployments; a second push while a deploy is in progress will queue, not conflict.

## Migration Plan

1. Merge the `docs.yml` workflow to `main`
2. In repository Settings → Pages, set Source to **GitHub Actions**
3. The next push to `main` triggers the workflow and publishes the site
4. Rollback: disable Pages in Settings or delete the workflow file
