# Agent Instructions

You are maintaining a generated Terraform provider for Apache Polaris.

The weekly GitHub Action fetches the latest Apache Polaris release OpenAPI specs, regenerates operation metadata, and runs tests. If checks fail, an agentic repair loop may invoke Codex or another configured GenAI CLI.

Rules:

- Keep the provider in Go.
- Fix the generator first when generated output is wrong.
- Do not hand-edit `internal/generated/operations_gen.go`.
- Do not delete tests to get green checks.
- Always run `make generate fmt test build`.
- Use Conventional Commits for merged changes that should appear in releases. Provider capability updates should normally use `feat(polaris): ...`; bug fixes should use `fix: ...`.
- Prefer small, reviewable provider behavior over pretending every REST endpoint maps perfectly to first-class Terraform state.
- This is intended to be a public repo. Never print secrets, never broaden workflow permissions casually, and keep generated/provider changes reviewable.
- If newer Codex CLI, GitHub Action versions or Go tooling are available, update them through pinned dependencies and keep all checks green.
- Record durable technical decisions, Polaris runtime findings, workflow constraints, and test strategy changes in `docs/adr/`. If a new release changes generated operations or real Polaris behavior, add or update an ADR in the same pull request.
- Agentic-relevant changes must satisfy `scripts/check_adr_updates.sh`; do not bypass it. Add a short ADR instead.
