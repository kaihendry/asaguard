## ADDED Requirements

### Requirement: Docs workflow file exists
The repository SHALL contain `.github/workflows/docs.yml` that defines a GitHub Actions workflow triggered on pushes to `main`.

#### Scenario: Workflow triggers on main push
- **WHEN** a commit is pushed to the `main` branch
- **THEN** the `docs` workflow run is created in GitHub Actions

### Requirement: Workflow builds the doc site
The workflow SHALL check out the repository, set up Go using the version from `go.mod`, and run `make docs` to generate the `docs/` directory.

#### Scenario: make docs runs in CI
- **WHEN** the workflow executes the build job
- **THEN** `go run ./cmd/docgen` completes and `docs/index.html` plus all 11 package pages are present in the workspace

### Requirement: Workflow deploys to GitHub Pages
The workflow SHALL upload `docs/` as a Pages artifact and deploy it using `actions/deploy-pages`, making the site accessible at the repository's GitHub Pages URL.

#### Scenario: Successful deployment
- **WHEN** the build job completes successfully
- **THEN** the deploy job runs and the site is published to GitHub Pages

#### Scenario: No write access to repo tree
- **WHEN** the workflow runs
- **THEN** it MUST NOT push any commits or modify any branch; `contents` permission SHALL be `read`

### Requirement: Workflow uses minimal permissions
The workflow job that deploys MUST declare `pages: write` and `id-token: write` and nothing broader, following least-privilege.

#### Scenario: Permissions are scoped
- **WHEN** the workflow YAML is inspected
- **THEN** no job has `contents: write` or `packages: write` or other elevated permissions
