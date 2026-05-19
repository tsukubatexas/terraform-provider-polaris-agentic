# ADR 0014: Validate OpenAPI Registry Completeness

Date: 2026-05-19

Status: Accepted

## Context

The provider is generated from Apache Polaris OpenAPI specs but exposes those endpoints through generic Terraform primitives. A green build is not enough if the generated operation registry silently misses an OpenAPI operation, points an operation at the wrong method/path, or contains operations that the generic client cannot call.

The release matrix also regenerates the provider against several Polaris releases using a temporary spec cache, so OpenAPI completeness checks must work both for the normal `specs` cache and for the matrix cache.

## Decision

Add durable tests for the OpenAPI layer:

- The generator test reparses the cached OpenAPI specs for the generated release and compares every operation against `internal/generated.Operations`.
- Duplicate OpenAPI `operationId` fallback behavior is checked with the same stable ID logic used by the generator.
- The provider test iterates over every generated operation and proves the generic client can expand path parameters and construct an HTTP request for it.
- The spec-cache lookup supports the normal repo cache and the release-matrix cache.

## Consequences

- Missing, stale, or incorrectly generated OpenAPI operations fail `go test ./...`.
- All generated operations are proven technically callable through the generic Terraform provider surface.
- These tests do not prove every Polaris endpoint has safe Terraform lifecycle semantics; typed resource semantics and real Polaris workflow coverage still need explicit tests.
