# ADR 0002: Generate Operation Registry from Apache Polaris OpenAPI Specs

Date: 2026-05-18

Status: Accepted

## Context

Apache Polaris exposes management, catalog, policy, OAuth, notification, and Iceberg REST operations through OpenAPI specs. The provider should track new Polaris releases without manually copying endpoint lists.

Older Polaris releases do not contain all spec files. During testing, `apache-polaris-0.9.0-incubating` only provided the management spec while later releases provided broader catalog and Iceberg specs.

## Decision

Generate `internal/generated/operations_gen.go`, `docs/generated-operations.md`, Terraform Registry-style provider docs, and canonical example Terraform configurations from Apache Polaris release-tagged OpenAPI specs.

The management spec is required. Catalog, Iceberg, policy, OAuth token, notification, and generic table specs are optional so older releases can still be tested in the release matrix.

Generated output must remain deterministic. The generator uses a reproducible generation marker instead of embedding wall-clock timestamps.

## Consequences

- The provider can follow new Polaris REST operations automatically.
- The release matrix can validate old and new Polaris spec layouts.
- Provider docs stay in the normal Terraform documentation layout whenever operations are regenerated.
- Generated files must not be hand-edited. Fix the generator, examples, or spec input handling instead.
- OpenAPI only tells us the operations. It does not define Terraform lifecycle semantics, identity fields, drift behavior, or safe update/delete handling.
