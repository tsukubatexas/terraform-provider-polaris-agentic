# ADR 0013: Auto-Merge Release Please PRs Through a Hardened Queue

Date: 2026-05-19

Status: Accepted

## Context

Release Please can create the release pull request automatically, but a public protected repository still needs the pull request to pass the normal CI, security, dependency review and catalog checks before merging. A release PR created with GitHub's default `GITHUB_TOKEN` can be written to the repository, but GitHub does not trigger follow-up pull-request workflows from events created by that token. That leaves protected-branch release PRs stuck without required checks.

The repository also allows autonomous workflows to create pull requests. Release auto-merge therefore needs a narrow allowlist so a write-capable workflow cannot become a generic bypass path.

## Decision

Add a dedicated Release Please auto-merge path:

- Enable repository auto-merge.
- Require `RELEASE_PLEASE_TOKEN` for Release Please PR creation and fail fast when it is missing.
- Add `.github/workflows/release-pr-automerge.yml` to enable squash auto-merge when the managed release PR is opened, updated, labeled or manually selected.
- Add a `queue-release-pr` job to `.github/workflows/release.yml` so the weekly/manual Release Please run also queues the pending release PR.
- Validate the release PR before enabling auto-merge with `scripts/validate_release_please_pr.sh`.
- In the monthly release train, validate the Release Please PR metadata from the trusted default branch before checking out and testing the release branch.
- Allow only the same-repository branch `release-please--branches--main`, an open non-draft PR targeting `main` or `master`, the `autorelease: pending` label, a `chore: release x.y.z` title, and changes limited to `CHANGELOG.md` plus `.release-please-manifest.json`.
- Add a smoke test for the validation guard and run it in CI.

## Consequences

- Release Please PRs can merge without a human click once branch protection is green.
- Full automation requires a fine-grained `RELEASE_PLEASE_TOKEN`; without it, release workflows stop before creating an unmergeable protected release PR.
- Release auto-merge is intentionally narrower than general autonomous PR auto-merge.
- If Release Please configuration later updates additional version files, the allowlist must be expanded deliberately with an ADR update.
