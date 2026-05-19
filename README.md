# Terraform Provider for Apache Polaris

[![CI](https://github.com/tsukubatexas/terraform-provider-polaris/actions/workflows/ci.yml/badge.svg)](https://github.com/tsukubatexas/terraform-provider-polaris/actions/workflows/ci.yml)
[![Security](https://github.com/tsukubatexas/terraform-provider-polaris/actions/workflows/security.yml/badge.svg)](https://github.com/tsukubatexas/terraform-provider-polaris/actions/workflows/security.yml)
[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/tsukubatexas/terraform-provider-polaris/badge)](https://scorecard.dev/viewer/?uri=github.com/tsukubatexas/terraform-provider-polaris)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

Terraform provider for managing Apache Polaris catalogs and REST API resources.

Autonomous update workflows track new Apache Polaris OpenAPI releases, refresh provider metadata, run real Polaris tests, and prepare pull requests.

Built for serious platform teams: Registry-ready releases, GPG-signed provider checksums, real Polaris apply/destroy validation, OpenAPI compatibility gates, branch protection, CodeQL, Scorecard, and autonomous maintenance PRs.

## Start Here: Terraform Usage

The provider is generic-first: you create Polaris objects by combining OpenAPI `operationId`s, `path_params`, JSON `body`, and an `id_attribute`.

Prominent user docs:

- [Terraform Registry provider docs](docs/index.md): generated provider documentation in the default Terraform docs layout.
- [AI-generated Terraform Provider Guide](docs/terraform-provider.md): provider configuration, catalog creation, principals, principal roles, catalog roles, grants, and read calls.
- [Generated Polaris Operation Registry](docs/generated-operations.md): every currently generated `operationId`, method and path.
- [Complete Polaris example](examples/complete-polaris/main.tf): realm header, catalog, principal, roles, grants, namespace and table metadata in one Terraform configuration.
- [Real Terraform smoke example](examples/test-catalog/main.tf): the catalog create/destroy example tested against a real Polaris container.

Minimal catalog example:

```hcl
provider "polaris" {
  endpoint = "https://polaris.example.com/api/management/v1"
  realm    = "POLARIS"
  token    = var.polaris_token
}

resource "polaris_rest_resource" "risk_catalog" {
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
        "default-base-location" = "abfss://warehouse@riskstore.dfs.core.windows.net/iceberg/risk"
      }
      storageConfigInfo = {
        storageType      = "AZURE"
        allowedLocations = ["abfss://warehouse@riskstore.dfs.core.windows.net/iceberg/risk"]
        tenantId         = "<tenant-id>"
      }
    }
  })

  id_attribute = "name"
}
```

## What It Builds

- A Go Terraform provider published as `tsukubatexas/polaris`.
- A generated registry of Apache Polaris REST OpenAPI operations.
- A generic `polaris_rest_call` data source for GET/read-style calls.
- A generic `polaris_rest_resource` resource for create/read/update/delete workflows backed by OpenAPI `operationId`s.
- A real Polaris catalog smoke test that applies and destroys Terraform against a running Polaris service.
- Automated OpenAPI tracking, provider regeneration, and PR preparation when Polaris changes.
- Release Please tracking with a controlled monthly release train.
- Terraform Registry-ready GitHub releases with provider ZIPs, manifest, SHA256SUMS, and detached GPG signature.

## Why Generic First?

Terraform providers need stable IDs, drift semantics and lifecycle behavior. OpenAPI alone does not know which fields are identities, which update operations are safe, or how deletes should handle missing objects.

So this repo does the honest enterprise thing:

```text
OpenAPI generates the operation inventory.
The generic resource can call every endpoint.
The maintenance loop can promote important endpoints into typed resources over time.
Tests must stay green before a PR is opened.
```

## Local Run

```bash
make generate
make fmt
make test
make build
```

Test the autonomous maintenance loop locally without spending API calls:

```bash
scripts/smoke_agentic_loop.sh
scripts/smoke_agentic_infra_loop.sh
scripts/smoke_self_improve_loop.sh
scripts/smoke_release_matrix.sh
```

The smoke test creates an intentional failing test, injects a fake repair agent, verifies that the loop enters repair mode, removes the failure and finishes green.

The release matrix smoke test currently covers Polaris `0.9.0-incubating`, `1.0.0-incubating`, `1.2.0-incubating`, `1.3.0-incubating`, `1.4.1` and `1.5.0`. Polaris `0.9.0` does not include the later Catalog/Iceberg spec files, so the generator treats them as optional and still verifies the provider against the older management API shape.

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

In GitHub Actions the weekly provider update first regenerates the provider metadata. After that, a separate final static infra loop runs against a Polaris service container. If the final apply/destroy check fails, that final loop calls the repair agent and starts over until the basic functionality is green. When a new Polaris release changes generated operations, `scripts/check_static_coverage.sh` fails unless `scripts/test_catalog.sh` or `examples/test-catalog` were extended too.

Force a specific Polaris release:

```bash
POLARIS_RELEASE=apache-polaris-1.4.1 make generate
```

## Architecture Decisions

Durable architecture decisions and Polaris runtime findings are tracked in [docs/adr](docs/adr/README.md). New maintenance changes must add or update ADRs when they discover new Polaris behavior, change workflow policy, or change how the real-Polaris static gate works. `scripts/check_adr_updates.sh` enforces this for automation-relevant files.

## Automated Provider Updates

The workflow `.github/workflows/agentic-update.yml` runs every Monday in a `golang:1.24-bookworm` container.

It does:

```text
1. Fetch latest Apache Polaris release.
2. Fetch Polaris OpenAPI specs from that release tag.
3. Generate operation metadata, Terraform Registry docs and examples.
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

The detailed operating guide is in [docs/release/README.md](docs/release/README.md).

```text
Weekly:
  - Provider update PRs land as Conventional Commits.
  - Release Please opens or updates the release PR.
  - CHANGELOG.md and .release-please-manifest.json show what the next release will contain.

Monthly:
  - .github/workflows/monthly-release.yml finds the pending Release Please PR.
  - It validates the release branch with linting, generation, tests, build and a real Polaris Terraform apply/destroy gate.
  - It merges the release PR.
  - Release Please creates the GitHub release.
  - scripts/build_release_artifacts.sh uploads provider binaries, registry manifest, SHA256SUMS and detached GPG signature.

Auto-merge:
  - .github/workflows/release.yml opens or updates the Release Please PR and queues it for auto-merge.
  - .github/workflows/release-pr-automerge.yml retries auto-merge when the release PR is opened, updated or labeled.
  - The auto-merge path accepts only the managed release branch and only CHANGELOG.md plus .release-please-manifest.json changes.
  - Branch protection still decides when the PR can merge: required checks and reviews must be green first.
```

Manual release preparation is still possible by running `.github/workflows/release.yml`.

Use `feat:` for provider capability changes, `fix:` for bug fixes and `feat!:` or `fix!:` for breaking changes. Pure maintenance commits can use `chore:`, `ci:` or `docs:` and will not force a version bump by themselves.

## Terraform Registry Publishing

The big release path is designed for the public Terraform Registry. Release assets are built in HashiCorp's expected shape:

```text
terraform-provider-polaris_<version>_<os>_<arch>.zip
terraform-provider-polaris_<version>_manifest.json
terraform-provider-polaris_<version>_SHA256SUMS
terraform-provider-polaris_<version>_SHA256SUMS.sig
```

Release platforms currently include Linux, macOS and Windows for `amd64` and `arm64`.

One-time registry setup:

```text
1. Add the provider in the Terraform Registry UI as tsukubatexas/polaris.
2. Add [the GPG public key](docs/release/terraform-registry-gpg-public-key.asc) to the Terraform Registry provider settings.
3. Keep RELEASE_PLEASE_TOKEN, TERRAFORM_REGISTRY_GPG_PRIVATE_KEY and TERRAFORM_REGISTRY_GPG_PASSPHRASE as GitHub secrets.
```

The repository name intentionally matches the Terraform Registry provider naming convention: `terraform-provider-polaris` publishes as `tsukubatexas/polaris`. After the one-time UI connection, monthly GitHub Releases are Registry-ingestable. The Registry watches GitHub releases; the workflow does not need a Terraform Registry API token.

## One-Time Setup

For autonomous repair mode, set one repository secret:

```text
Secret: OPENAI_API_KEY
```

That is enough for weekly provider maintenance:

- weekly Polaris OpenAPI update
- optional GenAI repair loop
- test-catalog check
- quarterly cleanup and hardening build
- auto PR
- auto-merge when branch protection allows it
- weekly self-improvement pass for tooling, tests and hardening
- Dependabot for Go, GitHub Actions and Codex CLI runtime
- CodeQL and OSSF Scorecard

Recommended additional secret for full release automation:

```text
Secret: RELEASE_PLEASE_TOKEN
```

This must be a fine-grained GitHub token that can create and merge Release Please pull requests and create releases. `GITHUB_TOKEN` is intentionally not used for Release Please PR creation because GitHub does not trigger follow-up workflows from events created by that token. Without `RELEASE_PLEASE_TOKEN`, release workflows fail fast instead of creating a protected release PR that cannot auto-merge.

The token-created release PR triggers the normal PR checks, `.github/workflows/release-pr-automerge.yml` enables squash auto-merge, and GitHub merges only after branch protection is satisfied.

With `RELEASE_PLEASE_TOKEN`, the repo also runs:

- weekly Release Please preparation
- monthly controlled GitHub releases with provider artifacts

Required for Terraform Registry publishing:

```text
Secret: TERRAFORM_REGISTRY_GPG_PRIVATE_KEY
Secret: TERRAFORM_REGISTRY_GPG_PASSPHRASE
```

`TERRAFORM_REGISTRY_GPG_PRIVATE_KEY` should be the ASCII-armored private key for the public GPG key registered in Terraform Registry. The release workflows fail before upload if signing is unavailable, so a "big release" cannot silently ship unsigned provider artifacts.

Current Terraform Registry signing key fingerprint:

```text
C9CEBB9BFC7B93194688356A7FADB37AD7485B8F
```

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
  -m "$AGENT_MODEL" \
  -C "$GITHUB_WORKSPACE" \
  -o "$LAST_MESSAGE" \
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

The full hardening model is documented in [docs/security-hardening.md](docs/security-hardening.md).

Recommended GitHub repository settings:

```text
Actions: allow GitHub Actions and selected marketplace actions only.
Default workflow token permissions: read-only.
Branch protection on main/master: require CI, Security, Dependency Review and Test Catalog checks.
Branch protection on main/master: require one code-owner review for protected paths.
Allow auto-merge: enabled.
Secret scanning: enabled.
Push protection: enabled.
Dependabot alerts: enabled.
CodeQL: enabled.
Ruleset: block force-pushes and deletions on main/master.
```

The workflows request write permissions only in jobs that create PRs, enable auto-merge or publish releases.

## Provider Example

For full provider usage, start with [docs/index.md](docs/index.md), [docs/terraform-provider.md](docs/terraform-provider.md), and [examples/complete-polaris/main.tf](examples/complete-polaris/main.tf). The short example below shows the generic resource shape.

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
      storageConfigInfo = {
        storageType      = "AZURE"
        allowedLocations = ["abfss://warehouse@risk.dfs.core.windows.net/iceberg/risk"]
        tenantId         = "<tenant-id>"
      }
    }
  })

  id_attribute = "name"
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

As of this provider snapshot, GitHub reports latest release `apache-polaris-1.5.0`.
