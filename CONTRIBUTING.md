# Contributing

Thanks for helping improve the autonomous Apache Polaris Terraform provider.

## Before You Start

Please read:

- [README.md](README.md)
- [docs/release/README.md](docs/release/README.md)
- [docs/adr/README.md](docs/adr/README.md)
- [SECURITY.md](SECURITY.md)

## Local Checks

Run these before opening a pull request:

```bash
make generate
make fmt
make test
make build
bash -n scripts/*.sh
scripts/check_release_please_config.sh
scripts/check_actions_pinned.sh
scripts/check_adr_updates.sh
scripts/check_static_coverage.sh
```

For changes that affect real Polaris behavior, run:

```bash
POLARIS_INFRA_MODE=docker AGENT_MAX_ROUNDS=1 scripts/agentic_infra_loop.sh
```

## Pull Request Rules

- Keep pull requests small and reviewable.
- Do not hand-edit `internal/generated/operations_gen.go`; fix the generator or spec inputs.
- Do not remove tests or weaken release gates to make CI green.
- Add or update an ADR for durable decisions, workflow changes, generator behavior, release policy, or Polaris runtime findings.
- Do not print secrets in scripts, workflows, docs, tests, or logs.

## Commit Style

Use Conventional Commits:

```text
feat(polaris): add generated support for new Polaris capability
fix: correct provider request handling
ci: harden workflow permissions
docs: explain release train
chore: refresh dependencies
```

Release Please uses these commits for SemVer and `CHANGELOG.md`.

## Security Changes

Security hardening is welcome, but keep it practical:

- prefer least privilege over broad tokens
- prefer deterministic checks over manual review-only rules
- document policy changes in ADRs
- keep automation runnable by contributors without private cloud credentials
