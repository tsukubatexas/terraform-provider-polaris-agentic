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

Force a specific Polaris release:

```bash
POLARIS_RELEASE=apache-polaris-1.4.1 make generate
```

## Weekly Agentic Update

The workflow `.github/workflows/agentic-update.yml` runs every Monday in a `golang:1.23-bookworm` container.

It does:

```text
1. Fetch latest Apache Polaris release.
2. Fetch Polaris OpenAPI specs from that release tag.
3. Generate operation metadata and docs.
4. Run fmt, tests and build.
5. If something fails, call a GenAI repair agent.
6. Repeat until green or AGENT_MAX_ROUNDS is reached.
7. Open a pull request with the generated/provider changes.
```

## One-Time Setup

For the autonomous mode, set one repository secret:

```text
Secret: OPENAI_API_KEY
```

That is enough for the repo to run as a self-maintaining public project:

- weekly Polaris OpenAPI update
- agentic repair loop
- test-catalog check
- auto PR
- auto-merge when checks pass
- release artifacts after merge
- weekly self-improvement pass for tooling, tests and hardening
- Dependabot for Go, GitHub Actions and Codex CLI runtime
- CodeQL and OSSF Scorecard

Optional variables:

```text
AGENT_MODEL = gpt-5.2
AGENT_MAX_ROUNDS = 5
DISABLE_AUTOMERGE = true
AGENT_REPAIR_COMMAND = custom agent command
```

By default, if `OPENAI_API_KEY` exists and no custom command is configured, the loop runs:

```bash
npx -y @openai/codex exec \
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
