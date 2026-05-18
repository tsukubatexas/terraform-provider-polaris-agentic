# Agentic Terraform Provider for Apache Polaris

This repo is a local scaffold for an autonomous Terraform provider workflow.

It checks Apache Polaris weekly, fetches the newest OpenAPI specs from the latest GitHub release, regenerates provider operation metadata, runs tests, and can start a GenAI repair loop until the provider is green.

## What It Builds

- A Go Terraform provider.
- A generated registry of all known Polaris REST OpenAPI operations.
- A generic `polaris_rest_call` data source for GET/read-style calls.
- A generic `polaris_rest_resource` resource for create/read/update/delete workflows backed by OpenAPI `operationId`s.
- A weekly GitHub Action that opens a pull request when Polaris changes.
- An agentic loop that can run Codex CLI or any other GenAI CLI.
- Release Please tracking with a monthly controlled release train.

## Why Generic First?

Terraform providers need stable IDs, drift semantics and lifecycle behavior. OpenAPI alone does not know which fields are identities, which update operations are safe, or how deletes should handle missing objects.

So this repo does the honest enterprise thing:

```text
OpenAPI generates the operation inventory.
The generic resource can call every endpoint.
The agent loop can promote important endpoints into typed resources over time.
Tests must stay green before a PR is opened.
```

## Local Run

```bash
make generate
make fmt
make test
make build
```

Test the autonomous loop locally without spending API calls:

```bash
scripts/smoke_agentic_loop.sh
scripts/smoke_agentic_infra_loop.sh
scripts/smoke_self_improve_loop.sh
scripts/smoke_release_matrix.sh
```

The smoke test creates an intentional failing test, injects a fake repair agent, verifies that the loop enters repair mode, removes the failure and finishes green.

The release matrix smoke test currently covers Polaris `0.9.0-incubating`, `1.0.0-incubating`, `1.2.0-incubating`, `1.3.0-incubating` and `1.4.1`. Polaris `0.9.0` does not include the later Catalog/Iceberg spec files, so the generator treats them as optional and still verifies the provider against the older management API shape.

Test the provider against a real local Polaris container:

```bash
docker run -d --rm \
  --name polaris-provider-smoke \
  -p 8181:8181 \
  -e POLARIS_BOOTSTRAP_CREDENTIALS=POLARIS,root,s3cr3t \
  apache/polaris:latest

scripts/test_catalog.sh
```

That script builds the provider, installs it into Terraform's local provider mirror as `tsukubatexas/polaris`, fetches a Polaris OAuth token, applies `examples/test-catalog`, and destroys it again. It creates an internal test catalog named `agentic_test` and uses an S3-style test location only for catalog CRUD validation.

Run the final static infra loop locally:

```bash
POLARIS_INFRA_MODE=docker scripts/agentic_infra_loop.sh
```

In GitHub Actions the weekly agentic update first runs the normal provider update loop. After that, a separate final static infra loop runs against a Polaris service container. If the final apply/destroy check fails, that final loop calls the repair agent and starts over until the basic functionality is green. When a new Polaris release changes generated operations, `scripts/check_static_coverage.sh` fails unless `scripts/test_catalog.sh` or `examples/test-catalog` were extended too.

Force a specific Polaris release:

```bash
POLARIS_RELEASE=apache-polaris-1.4.1 make generate
```

## Architecture Decisions

Durable architecture decisions and Polaris runtime findings are tracked in [docs/adr](docs/adr/README.md). New autonomous changes must add or update ADRs when they discover new Polaris behavior, change workflow policy, or change how the real-Polaris static gate works. `scripts/check_adr_updates.sh` enforces this for agentic-relevant files.

## Weekly Agentic Update

The workflow `.github/workflows/agentic-update.yml` runs every Monday in a `golang:1.23-bookworm` container.

It does:

```text
1. Fetch latest Apache Polaris release.
2. Fetch Polaris OpenAPI specs from that release tag.
3. Generate operation metadata and docs.
4. Run fmt, tests and build.
5. Run a separate final static infra loop against a real Apache Polaris service container.
6. If the final infra loop fails, call a GenAI repair agent and rerun it.
7. Repeat until green or AGENT_MAX_ROUNDS is reached.
8. Open a pull request with the generated/provider changes.
```

The generated update PR uses a Conventional Commit title and commit message:

```text
feat(polaris): update generated Terraform provider
```

That keeps weekly Polaris changes visible to Release Please without publishing a release every week.

## Releases

Releases are managed by Release Please instead of synthetic run-number tags.

```text
Weekly:
  - Agentic update PRs land as Conventional Commits.
  - Release Please opens or updates the release PR.
  - CHANGELOG.md and .release-please-manifest.json show what the next release will contain.

Monthly:
  - .github/workflows/monthly-release.yml finds the pending Release Please PR.
  - It validates the release branch with linting, generation, tests, build and a real Polaris Terraform apply/destroy gate.
  - It merges the release PR.
  - Release Please creates the GitHub release.
  - scripts/build_release_artifacts.sh uploads provider binaries and SHA256SUMS.
```

Manual release preparation is still possible by running `.github/workflows/release.yml`.

Use `feat:` for provider capability changes, `fix:` for bug fixes and `feat!:` or `fix!:` for breaking changes. Pure maintenance commits can use `chore:`, `ci:` or `docs:` and will not force a version bump by themselves.

## One-Time Setup

For agentic repair mode, set one repository secret:

```text
Secret: OPENAI_API_KEY
```

That is enough for weekly provider maintenance:

- weekly Polaris OpenAPI update
- agentic repair loop
- test-catalog check
- quarterly cleanup and hardening build
- auto PR
- auto-merge when checks pass
- weekly self-improvement pass for tooling, tests and hardening
- Dependabot for Go, GitHub Actions and Codex CLI runtime
- CodeQL and OSSF Scorecard

Recommended additional secret for full release automation:

```text
Secret: RELEASE_PLEASE_TOKEN
```

This should be a fine-grained GitHub token that can create and merge Release Please pull requests and create releases. Without it, the workflows fall back to `GITHUB_TOKEN`; that works for basic repository writes, but GitHub does not trigger follow-up workflows from events created by `GITHUB_TOKEN`, so protected-branch release automation is less smooth.

With `RELEASE_PLEASE_TOKEN`, the repo also runs:

- weekly Release Please preparation
- monthly controlled GitHub releases with provider artifacts

Optional variables:

```text
AGENT_MODEL = gpt-5.2
AGENT_MAX_ROUNDS = 5
POLARIS_INFRA_MODE = service
DISABLE_AUTOMERGE = true
AGENT_REPAIR_COMMAND = custom agent command
```

By default, if `OPENAI_API_KEY` exists and no custom command is configured, the loop runs:

```bash
npx --prefix tools/agent-runtime codex exec \
  --dangerously-bypass-approvals-and-sandbox \
  -a never \
  --search \
  -m "$AGENT_MODEL" \
  -C "$GITHUB_WORKSPACE" \
  -
```

Vendor-neutral override, if you do not want Codex CLI:

```text
Variable: AGENT_REPAIR_COMMAND
```

Example:

```bash
codex exec --dangerously-bypass-approvals-and-sandbox -a never --search -m gpt-5.2 -C "$PWD" -
```

The prompt is passed on stdin and includes the failing test log, the repo rules and the current diff summary.

## Public Repo Hardening

Recommended GitHub repository settings:

```text
Actions: allow GitHub Actions and selected marketplace actions only.
Default workflow token permissions: read-only.
Branch protection on main/master: require CI, Security and Test Catalog checks.
Allow auto-merge: enabled.
Secret scanning: enabled.
Push protection: enabled.
Dependabot alerts: enabled.
CodeQL: enabled.
Ruleset: block force-pushes and deletions on main/master.
```

The workflows request write permissions only in jobs that create PRs, enable auto-merge or publish releases.

## Provider Example

```hcl
terraform {
  required_providers {
    polaris = {
      source  = "tsukubatexas/polaris"
      version = "0.0.1"
    }
  }
}

provider "polaris" {
  endpoint        = "https://polaris.example.com/api/management/v1"
  realm           = "POLARIS"
  oauth_token_url = "https://login.microsoftonline.com/<tenant>/oauth2/v2.0/token"
  oauth_scope     = "api://<polaris-app-id>/.default"
  client_id       = var.client_id
  client_secret   = var.client_secret
}

resource "polaris_rest_resource" "catalog" {
  create_operation_id = "createCatalog"
  read_operation_id   = "getCatalog"
  delete_operation_id = "deleteCatalog"

  path_params = {
    catalogName = "risk"
  }

  body = jsonencode({
    catalog = {
      type = "INTERNAL"
      name = "risk"
      properties = {
        "default-base-location" = "abfss://warehouse@risk.dfs.core.windows.net/iceberg/risk"
      }
    }
  })

  id_attribute = "catalog.name"
}
```

## Current Polaris Source

The latest release is discovered from:

<https://api.github.com/repos/apache/polaris/releases/latest>

Specs are fetched from the release tag, including:

- `spec/polaris-management-service.yml`
- `spec/polaris-catalog-service.yaml`
- `spec/iceberg-rest-catalog-open-api.yaml`
- `spec/polaris-catalog-apis/*.yaml`

As of this scaffold, GitHub reports latest release `apache-polaris-1.4.1`.
