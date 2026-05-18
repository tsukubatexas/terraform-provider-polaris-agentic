# ADR 0005: Test the Provider Against a Real Polaris Container

Date: 2026-05-18

Status: Accepted

## Context

Unit tests and generated operation checks are not enough. The provider must prove that Terraform can load the local provider binary, authenticate to Polaris, create a catalog, read it, and destroy it against a real Polaris server.

## Decision

Use `scripts/test_catalog.sh` as the durable real-Polaris smoke test.

The script:

- Builds the provider.
- Installs it into Terraform's local provider mirror as `tsukubatexas/polaris`.
- Gets a Polaris OAuth token from the real service.
- Applies `examples/test-catalog`.
- Destroys the same Terraform resource.
- Cleans Terraform local state and lock artifacts after the run.

GitHub Actions runs Polaris as an `apache/polaris:latest` service container with `POLARIS_BOOTSTRAP_CREDENTIALS=POLARIS,root,s3cr3t`.

## Findings

- The Polaris container can generate random root credentials when bootstrap credentials are not supplied, so CI must set `POLARIS_BOOTSTRAP_CREDENTIALS`.
- `/api/catalog/v1/config` returned `401` during readiness testing, so readiness should use the same OAuth token flow used by the provider test.
- The current `apache/polaris:latest` image rejected `FILE` storage for catalog creation, so the static test uses an S3-style test location for catalog CRUD validation.
- `createCatalog` returns the catalog object directly, not nested under `catalog`; the Terraform example therefore uses `id_attribute = "name"`.
- Terraform's local provider mirror expects a versioned binary name such as `terraform-provider-polaris_v0.0.1`.
- Version `0.0.0` did not resolve cleanly in the local provider mirror test, so the local smoke example uses `0.0.1`.

## Consequences

- The repository has a real end-to-end guard for provider loading, authentication, create/read/delete behavior, and cleanup.
- The smoke test intentionally validates catalog CRUD only. It does not prove data-plane Iceberg reads and writes.
- New Polaris release behavior discovered in this test must be captured in an ADR.

