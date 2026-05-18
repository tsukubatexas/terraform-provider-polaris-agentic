# ADR 0010: Harden Production Readiness Gates

Date: 2026-05-18

Status: Accepted

## Context

The repository is intended to operate autonomously in public GitHub Actions. Production readiness needs more than a green happy path. The workflows and guards must catch nondeterministic generation, missing ADRs in pull requests, weak release sequencing, provider header mistakes, unsupported HTTP methods, and generator hangs.

## Decision

Harden the production gates:

- Use full checkout history in workflows that need to compare pull request changes with the base branch.
- Make ADR and static coverage guards consider both working-tree changes and pull request diffs.
- Add deterministic generator tests and HTTP client timeouts.
- Protect provider-owned headers such as `Authorization`, `Polaris-Realm`, and `User-Agent` from accidental override by custom headers.
- Validate explicit HTTP methods before issuing requests.
- Truncate HTTP response bodies in provider error messages.
- Retry transient generator HTTP fetch failures so release-matrix validation is resilient to short GitHub or raw spec fetch interruptions.
- Treat release metadata-only generator updates separately from operation registry changes so static coverage grows only when the provider surface changes.
- Run release-matrix generation against a temporary spec cache so compatibility tests do not leave old release specs in the normal repository cache.
- Pass the job-scoped `GITHUB_TOKEN` to generator workflows so GitHub release and raw spec fetches use authenticated rate limits without adding new secrets.
- Keep the default Codex CLI command aligned with the pinned `@openai/codex` exec syntax.
- Ignore installed `tools/agent-runtime/node_modules` content in ADR guards because it is an ephemeral install artifact, not a durable repo decision.
- Authenticate the pinned Codex CLI with the repo secret before non-interactive scheduled agent runs, storing the temporary login cache under `CODEX_HOME` inside the runner.
- Keep ADR guard filtering tolerant when every discovered file is ignored, so install-only working-tree noise exits cleanly.
- Build release artifacts before pushing a release tag and require no generated diff before release.
- Run shell and GitHub Actions linters in CI.
- Run every agentic maintenance smoke test in CI, including the provider update loop, final infra repair loop, repair command PATH handling, quarterly cleanup loop, self-improvement loop, and release matrix smoke.
- Run autonomous repair commands in non-login shells so GitHub Actions container PATH updates from setup steps remain available to nested agent commands.
- Keep scheduled security and scorecard workflows manually dispatchable so their scheduled paths can be verified on demand.

## Consequences

- Pull requests are more likely to fail early when they miss ADRs or real static coverage updates.
- Release tags are not pushed until the provider has generated, tested, and built successfully.
- Some custom-header edge cases become stricter, intentionally favoring provider-managed authentication and realm headers.
- Generator runs may take a few seconds longer during transient upstream failures, but scheduled release-matrix checks are less likely to fail from a single network hiccup.
- New Polaris tags with unchanged operation surfaces can flow without fake static test edits, while new operations still require durable real-Polaris coverage.
- Release-matrix smoke runs remain deterministic and do not pollute the working tree with historical spec caches.
- Scheduled generator jobs avoid unauthenticated GitHub API limits while still using the least-privilege workflow token.
- Pinned agent runtime updates can require CLI flag updates in the maintenance scripts.
- ADR enforcement remains strict for runtime lockfiles and package metadata but no longer mistakes installed dependencies for source changes.
- Codex credentials are initialized per job and remain outside the repository checkout.
- Install-only runs no longer fail the ADR guard when no durable source files changed.
- CI takes a little longer, but failures in autonomous maintenance workflows are caught before merge instead of in scheduled jobs.
- Scheduled repair jobs keep the Node and Codex runtime installed by the workflow visible to the agent subprocess.
- Operators can manually exercise every scheduled workflow after hardening changes instead of waiting for the next cron window.
- Future hardening changes should update this ADR or create a new one if they change the policy.
