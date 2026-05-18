## 1. Workflow file

- [x] 1.1 Create `.github/workflows/docs.yml` with a `push: branches: [main]` trigger
- [x] 1.2 Add a `build` job: checkout, `setup-go` (go-version-file: go.mod), run `make docs`, upload `docs/` with `actions/upload-pages-artifact@v3`
- [x] 1.3 Add a `deploy` job: depends on `build`, uses `actions/deploy-pages@v4`, with `pages: write` and `id-token: write` permissions and `environment: github-pages`

## 2. Verify & document

- [x] 2.1 Confirm `contents: read` (not write) is the only repo-tree permission in the workflow
- [x] 2.2 Add a comment in the workflow YAML noting the one-time manual step: Settings → Pages → Source → "GitHub Actions"
