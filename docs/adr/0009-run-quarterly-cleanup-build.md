# ADR 0009: Run a Quarterly Cleanup Build

Date: 2026-05-18

Status: Accepted

## Context

Weekly update loops keep the provider current, but they are intentionally scoped. Over time, public autonomous repositories can accumulate stale workflow glue, redundant scripts, outdated lockfiles, old assumptions, and ADR drift.

The repository needs a slower, more thorough maintenance pass that is allowed to clean up aggressively while still preserving tests, ADRs, static coverage, and public-repo safety.

## Decision

Add a quarterly GitHub Actions workflow, `.github/workflows/quarterly-cleanup.yml`, scheduled for the first day of every third month.

The workflow runs:

- `scripts/quarterly_cleanup.sh`
- dependency and lockfile cleanup checks
- generator, format, tests, and build
- shell syntax checks
- ADR update guard
- static coverage guard
- `scripts/agentic_infra_loop.sh` against a real Polaris service container

The quarterly cleanup loop may use the configured agent to simplify, remove stale code, update dependencies, and harden workflows. It creates a pull request instead of pushing directly.

## Consequences

- The repo gets a deliberate deep-clean cadence separate from weekly release tracking.
- Any durable cleanup decision must be reflected in ADRs.
- The cleanup is still bounded by checks and real Polaris validation.
- The workflow is intentionally heavier and has a longer timeout than the weekly loops.
