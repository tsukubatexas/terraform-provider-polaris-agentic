# Security Policy

## Supported Versions

This project is an early proof-of-concept provider. Security fixes are applied to the default branch first and then shipped through the next monthly release train.

| Version | Supported |
| --- | --- |
| `main` | Yes |
| Released versions | Best effort until the next release |

## Reporting a Vulnerability

Please do not open a public issue for a suspected vulnerability.

Use GitHub private vulnerability reporting when it is available on the repository, or contact the maintainer through a private channel on GitHub.

Include:

- affected commit or release
- affected workflow, provider resource, data source, or script
- reproduction steps
- expected impact
- whether credentials, tokens, Terraform state, or cloud resources are involved

## Security Expectations

The repository is designed for public autonomous maintenance:

- no secrets in repository files, logs, examples, or generated output
- least-privilege GitHub workflow permissions
- pinned GitHub Actions
- mandatory linting for shell scripts and workflows
- CodeQL and OpenSSF Scorecard runs
- real Polaris Terraform apply/destroy gate before controlled monthly releases

## Credential Handling

Do not paste real tokens into issues, pull requests, examples, or logs. Rotate any credential that was accidentally exposed.

Provider examples should use variables or GitHub secrets for sensitive values. Terraform state can contain sensitive material, so do not commit local state files.
