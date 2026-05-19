# Architecture Decision Records

This directory tracks durable technical decisions and operational findings for the autonomous Apache Polaris Terraform provider.

Every ADR should explain the context, the decision, and the consequences. If an agent discovers new Polaris behavior, a workflow constraint, or a change that affects how the provider is generated, tested, released, or operated, it must add a new ADR or update an existing one.

| ADR | Title | Status |
| --- | --- | --- |
| [0001](0001-use-go-terraform-plugin-sdk-provider.md) | Use Go and Terraform Plugin SDK v2 | Accepted |
| [0002](0002-generate-operation-registry-from-polaris-openapi.md) | Generate Operation Registry from Apache Polaris OpenAPI Specs | Accepted |
| [0003](0003-use-generic-rest-resource-and-data-source.md) | Start with Generic REST Terraform Primitives | Accepted |
| [0004](0004-use-autonomous-github-workflows.md) | Use Autonomous GitHub Workflows for Update, Repair, and Release | Accepted |
| [0005](0005-test-against-real-polaris-container.md) | Test the Provider Against a Real Polaris Container | Accepted |
| [0006](0006-use-separate-final-static-infra-loop.md) | Use a Separate Final Static Infra Repair Loop | Accepted |
| [0007](0007-require-static-coverage-for-new-polaris-operations.md) | Require Static Coverage Updates for New Polaris Operations | Accepted |
| [0008](0008-require-adr-updates-for-agentic-changes.md) | Require ADR Updates for Agentic-Relevant Changes | Accepted |
| [0009](0009-run-quarterly-cleanup-build.md) | Run a Quarterly Cleanup Build | Accepted |
| [0010](0010-harden-production-readiness-gates.md) | Harden Production Readiness Gates | Accepted |
| [0011](0011-use-release-please-monthly-release-train.md) | Use Release Please and a Monthly Release Train | Accepted |
| [0012](0012-harden-public-open-source-repository.md) | Harden Public Open Source Repository | Accepted |
| [0013](0013-auto-merge-release-please-prs.md) | Auto-Merge Release Please PRs Through a Hardened Queue | Accepted |
