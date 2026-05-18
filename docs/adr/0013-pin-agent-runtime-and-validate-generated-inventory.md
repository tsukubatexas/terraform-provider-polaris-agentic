# ADR 0013: Pin Agent Runtime Codex and Validate Generated Inventory

Date: 2026-05-18

Status: Accepted

## Context

This repository uses autonomous GitHub Actions repair/update loops that run the OpenAI Codex CLI from `tools/agent-runtime`.
To keep agent behavior reproducible and reviewable, the agent runtime must remain pinned (including the transitive platform packages).

Separately, the Polaris OpenAPI operation inventory is generated into both Go (`internal/generated/operations_gen.go`) and documentation (`docs/generated-operations.md`).
Without a guardrail, it is easy for these two generated outputs to drift, especially when the generator changes.

The repo currently targets Go 1.24.x; some upstream dependency upgrades now require Go 1.25+, so dependency bumps must avoid unintentionally raising the minimum toolchain version.

## Decision

- Keep `tools/agent-runtime` pinned by updating both `package.json` and `package-lock.json` together whenever `@openai/codex` is updated.
- Add a unit test that verifies `docs/generated-operations.md` exactly matches `internal/generated.Operations` and `internal/generated.ReleaseTag`.
- Permit low-risk patch upgrades that remain compatible with Go 1.24.x; defer upgrades that require Go 1.25+ to an explicit Go toolchain bump ADR/PR.

## Consequences

- Agent runtime upgrades stay deterministic across platforms and are less likely to break agent workflows unexpectedly.
- Generator changes that forget to update the docs inventory (or change ordering/contents unintentionally) are caught by `go test ./...`.
- Some dependency updates will be deferred until a deliberate Go version bump is made.

