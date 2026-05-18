# ADR 0008: Require ADR Updates for Agentic-Relevant Changes

Date: 2026-05-18

Status: Accepted

## Context

The repository is designed to be maintained by autonomous loops. If those loops change workflows, provider behavior, generator behavior, release automation, tests, or static infrastructure checks without recording why, future runs become hard to audit.

Prompt-only instructions are not enough. The agent can forget them or optimize for a green build.

## Decision

Add `scripts/check_adr_updates.sh`.

If agentic-relevant files change without any change under `docs/adr/`, the final infra loop and CI fail. The agent must then add or update an ADR before the run can become green.

Agentic-relevant files include:

- GitHub workflows.
- `AGENTS.md`.
- Go module files.
- Generator code.
- Provider code.
- Scripts.
- Test catalog examples.
- Agent runtime tooling.

## Consequences

- Autonomous changes become reviewable and explainable.
- New release behavior, new findings, and workflow policy changes are tracked in the same pull request as the implementation.
- Small mechanical changes may require a short ADR update. This is intentional for this repository because auditability matters more than saving a few lines.
- If the guard becomes noisy, update this ADR and `scripts/check_adr_updates.sh` together.
