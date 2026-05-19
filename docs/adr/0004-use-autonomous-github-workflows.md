# ADR 0004: Use Autonomous GitHub Workflows for Update, Repair, and Release

Date: 2026-05-18

Status: Accepted

## Context

The repository should be as self-maintaining as possible after setting `OPENAI_API_KEY` once. It must track new Polaris releases, keep dependencies current, test provider behavior, and publish release artifacts without requiring manual weekly work.

## Decision

Use GitHub Actions for:

- Weekly Polaris spec update and agentic repair.
- Weekly self-improvement for tooling, tests, and hardening.
- Daily and post-workflow autonomous PR hygiene for obsolete bot-created PRs.
- CI checks on push and pull requests.
- Real Polaris test-catalog checks.
- CodeQL and OSSF Scorecard security checks.
- Dependabot updates for Go modules, GitHub Actions, and the pinned Codex CLI runtime.
- Release ZIPs and SHA256SUMS for Terraform provider distribution.

Agentic workflows create pull requests instead of pushing directly to protected branches.

The autonomous PR hygiene workflow may close only known bot-managed PR families: `agentic/*`, `dependabot/*`, and `release-please--branches--*`, or bot-authored PRs carrying the explicit automation labels used by those systems. Human feature branches are intentionally outside the closure scope, even if they use a label such as `dependencies`.

## Consequences

- A single repository secret, `OPENAI_API_KEY`, can enable autonomous repair loops.
- Workflows need write permissions only where they create PRs, enable auto-merge, or publish releases.
- Auto-merge can be disabled with `DISABLE_AUTOMERGE=true`.
- Any new workflow that mutates the repo must preserve least-privilege permissions and avoid logging secrets.
- Stale or failed autonomous PRs are removed without manual cleanup, while fresh release and update PRs remain available for the normal merge/release train.
