# Security Hardening

This repository is intentionally public and autonomous, so the default posture is conservative.

## Repository Controls

- Branch protection on `main`.
- Required CI, Security, and Test Catalog checks before merging.
- Code-owner review for protected paths.
- Squash merge preferred.
- Delete branches after merge.
- Secret scanning and push protection enabled.
- Dependabot alerts and security updates enabled.

## Workflow Controls

- Workflows request write permissions only where needed.
- External GitHub Actions are pinned to full commit SHAs.
- `scripts/check_actions_pinned.sh` blocks tag-pinned actions.
- `shellcheck` and `actionlint` run in CI.
- CodeQL scans Go code.
- OpenSSF Scorecard publishes SARIF results.
- Dependency review blocks high-severity vulnerable dependency additions in pull requests.

## Provider Controls

- Provider-owned auth and realm headers cannot be overridden by user-supplied custom headers.
- Endpoints must use `http` or `https`.
- Error response bodies are truncated before being returned in diagnostics.
- Real Polaris Terraform apply/destroy is part of the final release gate.

## Release Controls

- Weekly updates prepare release content.
- Release Please maintains the release PR and changelog.
- Monthly release train validates the exact release PR head before merging.
- Release artifacts are built only after tests and generated-output checks are clean.

## Residual Risk

This provider is still a proof of concept. The generic REST resource can call broad Polaris operations, so production use should wrap it with policy, review, and typed resource promotion over time.
