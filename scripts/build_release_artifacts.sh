#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT_DIR}"

version="${1:?usage: scripts/build_release_artifacts.sh <version> [output-dir]}"
out_dir="${2:-dist}"
provider_name="${PROVIDER_NAME:-polaris}"
checksum_file="terraform-provider-${provider_name}_${version}_SHA256SUMS"
manifest_file="terraform-provider-${provider_name}_${version}_manifest.json"

if ! command -v zip >/dev/null 2>&1; then
  echo "Missing required command: zip" >&2
  exit 1
fi

if ! command -v sha256sum >/dev/null 2>&1; then
  echo "Missing required command: sha256sum" >&2
  exit 1
fi

if [[ "${REQUIRE_TERRAFORM_REGISTRY_SIGNATURE:-false}" == "true" ]]; then
  if ! command -v gpg >/dev/null 2>&1; then
    echo "Missing required command: gpg" >&2
    exit 1
  fi
  if [[ -z "${GPG_PRIVATE_KEY:-}" ]]; then
    echo "GPG_PRIVATE_KEY is required when REQUIRE_TERRAFORM_REGISTRY_SIGNATURE=true" >&2
    exit 1
  fi
fi

rm -rf "${out_dir}"
mkdir -p "${out_dir}"

cp terraform-registry-manifest.json "${out_dir}/${manifest_file}"

targets=(
  linux_amd64
  linux_arm64
  darwin_amd64
  darwin_arm64
  windows_amd64
  windows_arm64
)

for target in "${targets[@]}"; do
  os="${target%_*}"
  arch="${target#*_}"
  work="${out_dir}/${target}"
  binary="terraform-provider-${provider_name}_v${version}"
  if [[ "${os}" == "windows" ]]; then
    binary="${binary}.exe"
  fi
  mkdir -p "${work}"
  GOOS="${os}" GOARCH="${arch}" go build \
    -buildvcs=false \
    -ldflags "-X main.version=${version}" \
    -o "${work}/${binary}" .
  (
    cd "${work}"
    zip -q "../terraform-provider-${provider_name}_${version}_${os}_${arch}.zip" "${binary}"
  )
done

(
  cd "${out_dir}"
  sha256sum -- *.zip "${manifest_file}" >"${checksum_file}"
)

if [[ "${REQUIRE_TERRAFORM_REGISTRY_SIGNATURE:-false}" == "true" ]]; then
  gpg_home="$(mktemp -d)"
  cleanup_gpg_home() {
    rm -rf "${gpg_home}"
  }
  trap cleanup_gpg_home EXIT
  chmod 700 "${gpg_home}"
  printf '%s' "${GPG_PRIVATE_KEY}" | gpg --batch --homedir "${gpg_home}" --import >/dev/null 2>&1
  sign_args=(
    --batch
    --yes
    --homedir "${gpg_home}"
    --pinentry-mode loopback
    --detach-sign
    --output "${out_dir}/${checksum_file}.sig"
  )
  if [[ -n "${GPG_PASSPHRASE:-}" ]]; then
    sign_args+=(--passphrase "${GPG_PASSPHRASE}")
  fi
  gpg "${sign_args[@]}" "${out_dir}/${checksum_file}"
fi
