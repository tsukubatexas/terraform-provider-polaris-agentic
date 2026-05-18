# ADR 0004: Use Autonomous GitHub Workflows for Update, Repair, and Release

Date: 2026-05-18

Status: Accepted

## Context

The repository should be as self-maintaining as possible after setting `OPENAI_API_KEY` once. It must track new Polaris releases, keep dependencies current, test provider behavior, and publish release artifacts without requiring manual weekly work.

## Decision

Use GitHub Actions for:

- Weekly Polaris spec update and agentic repair.
- Weekly self-improvement for tooling, tests, and hardening.
- CI checks on push and pull requests.
- Real Polaris test-catalog checks.
- CodeQL and OSSF Scorecard security checks.
- Dependabot updates for Go modules, GitHub Actions, and the pinned Codex CLI runtime.
- Release ZIPs and SHA256SUMS for Terraform provider distribution.

Agentic workflows create pull requests instead of pushing directly to protected branches.

## Consequences

- A single repository secret, `OPENAI_API_KEY`, can enable autonomous repair loops.
- Workflows need write permissions only where they create PRs, enable auto-merge, or publish releases.
- Auto-merge can be disabled with `DISABLE_AUTOMERGE=true`.
- Any new workflow that mutates the repo must preserve least-privilege permissions and avoid logging secrets.

