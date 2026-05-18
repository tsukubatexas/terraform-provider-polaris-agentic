# ADR 0007: Require Static Coverage Updates for New Polaris Operations

Date: 2026-05-18

Status: Accepted

## Context

Polaris releases can add or change REST operations. The generator may successfully update `internal/generated/operations_gen.go` while the real-Polaris static test still only covers the old baseline. That would make the repository look green while silently missing new release behavior.

## Decision

Add `scripts/check_static_coverage.sh` to the final infra loop.

If generated operation files changed but `scripts/test_catalog.sh` or `examples/test-catalog` did not change, the final infra loop fails. The repair agent must then extend the durable static infra coverage before the workflow can become green.

## Consequences

- New Polaris operations force an explicit coverage decision.
- The final gate grows with the release surface instead of staying frozen.
- The guard is intentionally conservative. Sometimes the right update may be a small comment or an explicit no-op decision in the static test files, but the change must be visible in review.
- If this policy becomes too strict, update this ADR and the script together.

