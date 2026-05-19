# ADR 0015: Publish Registry-Ready Release Assets

Date: 2026-05-19

Status: Accepted

## Context

The provider should be consumable as `tsukubatexas/polaris` from the public Terraform Registry. GitHub release ZIPs alone are not enough for a professional provider release: Terraform Registry ingestion expects provider archives, a registry manifest, checksums, and a detached GPG signature created with the key registered for the provider.

The repository name must stay aligned with the public Terraform Registry address. For `tsukubatexas/polaris`, the GitHub repository used for onboarding is named `terraform-provider-polaris`.

## Decision

Make release artifacts Terraform Registry-ready:

- Add `terraform-registry-manifest.json` with provider protocol `5.0`.
- Include `terraform-provider-polaris_<version>_manifest.json` in every release build.
- Include provider ZIPs and the manifest in `terraform-provider-polaris_<version>_SHA256SUMS`.
- Sign the checksum file with a detached GPG signature.
- Require `TERRAFORM_REGISTRY_GPG_PRIVATE_KEY` and `TERRAFORM_REGISTRY_GPG_PASSPHRASE` in release workflows by setting `REQUIRE_TERRAFORM_REGISTRY_SIGNATURE=true`.
- Add a CI smoke test that creates a temporary GPG key, builds artifacts, verifies checksums, verifies the signature, validates the manifest, and inspects ZIP contents.
- Generate a dedicated Terraform Registry release-signing key, store only the private key and passphrase as GitHub secrets, and commit only the public key and fingerprint.
- Mark the registry manifest and public signing-key files as release-sensitive CODEOWNERS paths.
- Document the one-time Terraform Registry UI onboarding and repository naming requirement.
- Keep the Go module path, badges, User-Agent strings and documentation aligned with `github.com/tsukubatexas/terraform-provider-polaris`.

## Consequences

- Big releases fail before upload if Terraform Registry signing material is missing.
- GitHub Releases contain the assets required for Terraform Registry ingestion.
- Registry publication still needs a one-time Terraform Registry UI connection and the public GPG key registration.
- If the repository is ever mirrored or renamed again, Terraform Registry onboarding must continue to point at a GitHub repository named `terraform-provider-polaris`.
- The release signing key expires in 2028 and must be rotated before expiry.
