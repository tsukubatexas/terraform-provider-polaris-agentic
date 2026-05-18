# ADR 0012: Harden Public Open Source Repository

Date: 2026-05-18

Status: Accepted

## Context

The repository is now public. It contains autonomous workflows that can generate code, open pull requests, merge changes, and create releases. A public autonomous repository needs stronger defaults than a private prototype: clear contribution rules, private security reporting, immutable workflow dependencies, branch protection, and visible security signals.

GitHub supports CODEOWNERS for automatic ownership review, branch protection for required reviews/checks, security policies and advisories for vulnerability handling, Dependabot for vulnerable dependencies, and CodeQL/Scorecard for automated security analysis.

## Decision

Add open source and security hardening:

- Add `LICENSE`, `SECURITY.md`, `CONTRIBUTING.md`, `CODE_OF_CONDUCT.md`, `SUPPORT.md`, and `MAINTAINERS.md`.
- Add issue templates, pull request template, and CODEOWNERS.
- Pin external GitHub Actions to full commit SHAs.
- Add `scripts/check_actions_pinned.sh` and run it in CI and release gates.
- Add Dependency Review for pull requests.
- Add `.editorconfig` and `.gitattributes` for consistent public contributions.
- Document the hardening model in `docs/security-hardening.md`.
- Upgrade the vulnerable transitive `google.golang.org/grpc` dependency to a patched release and move CI containers to Go 1.24.
- Keep the CI-installed Actionlint version compatible with the Go toolchain used in container jobs.
- Keep workflow-level permissions read-only so OpenSSF Scorecard can verify and publish results.
- Install container package prerequisites before actions that need them, such as Terraform setup requiring `unzip`.
- Suppress Shellcheck's indirect-trap unreachable warning only for the trap cleanup function.
- Disable Go VCS stamping for provider binaries so container and release builds stay reproducible without depending on a writable Git checkout.
- Mark GitHub container workspaces as safe Git directories before running repository-diff guards.
- Verify pinned GitHub Actions against the referenced action repository so annotated tag object SHAs or unrelated commits cannot pass the local guard.
- Configure the GitHub repository for Dependabot alerts, automatic security fixes, secret scanning, push protection, squash merges, branch deletion after merge, topics, and branch protection.

## Consequences

- Public contributors get a clearer path for issues, pull requests, and security reports.
- Workflow supply-chain risk is reduced by immutable action references.
- Maintainers must intentionally update pinned actions.
- The action-pin guard now needs `curl` and network access in CI and release validation jobs.
- Branch protection means direct pushes to `main` should become exceptional.
