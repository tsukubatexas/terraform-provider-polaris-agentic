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
- Build release artifacts before pushing a release tag and require no generated diff before release.
- Run shell and GitHub Actions linters in CI.

## Consequences

- Pull requests are more likely to fail early when they miss ADRs or real static coverage updates.
- Release tags are not pushed until the provider has generated, tested, and built successfully.
- Some custom-header edge cases become stricter, intentionally favoring provider-managed authentication and realm headers.
- Future hardening changes should update this ADR or create a new one if they change the policy.
