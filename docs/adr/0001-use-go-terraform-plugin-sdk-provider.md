# ADR 0001: Use Go and Terraform Plugin SDK v2

Date: 2026-05-18

Status: Accepted

## Context

Terraform providers are normally distributed as native binaries and are most commonly implemented in Go. The provider needs to run in GitHub Actions, be cross-compiled for releases, and integrate with Terraform's provider protocol without adding a large custom runtime.

## Decision

Implement the provider in Go with `github.com/hashicorp/terraform-plugin-sdk/v2`.

Use the module path `github.com/tsukubatexas/terraform-provider-polaris-agentic` and publish the provider namespace as `tsukubatexas/polaris`.

## Consequences

- The provider can be built and tested with the standard Go toolchain.
- Release automation can produce native Terraform provider binaries.
- The repository stays close to established Terraform provider patterns.
- Future migrations to the Terraform Plugin Framework remain possible, but should be handled as an explicit ADR if the provider gains enough typed resources to justify it.

