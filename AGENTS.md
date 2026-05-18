# Agent Instructions

You are maintaining a generated Terraform provider for Apache Polaris.

The weekly GitHub Action fetches the latest Apache Polaris release OpenAPI specs, regenerates operation metadata, and runs tests. If checks fail, an agentic repair loop may invoke Codex or another configured GenAI CLI.

Rules:

- Keep the provider in Go.
- Fix the generator first when generated output is wrong.
- Do not hand-edit `internal/generated/operations_gen.go`.
- Do not delete tests to get green checks.
- Always run `make generate fmt test build`.
- Prefer small, reviewable provider behavior over pretending every REST endpoint maps perfectly to first-class Terraform state.
- This is intended to be a public repo. Never print secrets, never broaden workflow permissions casually, and keep generated/provider changes reviewable.
- If newer Codex CLI, GitHub Action versions or Go tooling are available, update them through pinned dependencies and keep all checks green.
