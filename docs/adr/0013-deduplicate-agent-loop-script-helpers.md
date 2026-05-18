# ADR 0013: Deduplicate Agent Loop Script Helpers

Date: 2026-05-18

Status: Accepted

## Context

The repository has multiple autonomous maintenance loops (`scripts/agentic_loop.sh`, `scripts/agentic_infra_loop.sh`, `scripts/self_improve.sh`, `scripts/quarterly_cleanup.sh`) that:

- build a default Codex CLI command string when `OPENAI_API_KEY` is present
- run the same `make generate fmt test build` check suite (plus loop-specific extras)

Historically this logic was copy/pasted per script, which increases drift risk and makes quarterly cleanup harder (the same fix must be applied in multiple places).

## Decision

Introduce a small shared helper library, `scripts/agent_loop_lib.sh`, and use it from the agent-loop scripts to generate the default Codex CLI command.

Also simplify loop check pipelines to prefer a single `make generate fmt test build` invocation (while keeping loop-specific checks intact).

## Consequences

- Script changes that affect agent-loop behavior are easier to review and keep consistent.
- Future hardening (e.g., Codex CLI flag changes) can be made in one place.
- The helper must remain stable and side-effect free when sourced.

