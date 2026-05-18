# ADR 0011: Use Release Please and a Monthly Release Train

Date: 2026-05-18

Status: Accepted

## Context

Publishing a provider release on every push to `main` is noisy and creates unnecessary instability for downstream users. The repository still needs weekly autonomous Polaris tracking, but releases should be calmer, auditable, and based on SemVer.

Release Please supports Go repositories, manifest-driven configuration, Conventional Commits, changelog generation, GitHub Releases, and release pull requests. That matches the desired operating model: weekly preparation, monthly release, and a clear human-readable release history.

## Decision

Use Release Please as the release tracker.

- Weekly Polaris update PRs use `feat(polaris): ...` commit messages so relevant API/provider changes are included in the next SemVer release.
- `.github/workflows/release.yml` runs Release Please on pushes, on a weekly schedule, and manually. It maintains the release PR and uploads provider binaries when a release is created.
- `.github/workflows/monthly-release.yml` is the monthly controlled release train. It prepares the release PR, validates it with the static release gate, merges it, asks Release Please to publish the release, and uploads provider binaries.
- Release artifacts are built by `scripts/build_release_artifacts.sh` so the weekly/manual and monthly paths use the same packaging logic.
- `scripts/check_release_please_config.sh` keeps the manifest setup valid in CI.

## Consequences

- Users get fewer, more predictable releases.
- Weekly autonomous updates still accumulate quickly in the release PR.
- Release history is based on Conventional Commits and a generated `CHANGELOG.md`.
- Full monthly automation works best with a `RELEASE_PLEASE_TOKEN` secret because GitHub's default `GITHUB_TOKEN` has event-trigger limitations for resources it creates.
