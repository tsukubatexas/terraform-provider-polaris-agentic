# ADR 0006: Use a Separate Final Static Infra Repair Loop

Date: 2026-05-18

Status: Accepted

## Context

The normal agentic update loop should focus on generating and repairing provider code. A separate final gate is needed to prove the basic provider functionality against real Polaris infrastructure after the code loop has finished.

## Decision

Keep `scripts/agentic_loop.sh` focused on:

- `make generate`
- `make fmt`
- `make test`
- `make build`

Run `scripts/agentic_infra_loop.sh` after the normal loop. The final infra loop runs static checks and then `scripts/test_catalog.sh` against real Polaris. If it fails, it invokes the configured repair agent and retries.

## Consequences

- The normal code-generation loop stays fast and focused.
- The final gate is explicit and easy to reason about.
- A real Polaris failure can trigger autonomous repair without hiding inside the generator loop.
- The infra loop must not weaken or skip `scripts/test_catalog.sh` to become green.

