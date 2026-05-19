#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT_DIR}"

need() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "Missing required command: $1" >&2
    exit 1
  fi
}

need gpg
need jq
need sha256sum
need unzip

tmp_dir="$(mktemp -d)"
gpg_home="${tmp_dir}/gnupg"
out_dir="${tmp_dir}/dist"
mkdir -p "${gpg_home}"
chmod 700 "${gpg_home}"
trap 'rm -rf "${tmp_dir}"' EXIT

cat >"${tmp_dir}/key.conf" <<'EOF'
%no-protection
Key-Type: RSA
Key-Length: 3072
Name-Real: Terraform Registry Smoke
Name-Email: terraform-registry-smoke@example.com
Expire-Date: 0
%commit
EOF

gpg --batch --homedir "${gpg_home}" --generate-key "${tmp_dir}/key.conf" >/dev/null 2>&1
fingerprint="$(gpg --batch --homedir "${gpg_home}" --list-secret-keys --with-colons | awk -F: '/^fpr:/ {print $10; exit}')"
if [[ -z "${fingerprint}" ]]; then
  echo "Could not create smoke-test GPG key." >&2
  exit 1
fi

private_key="$(gpg --batch --homedir "${gpg_home}" --armor --export-secret-keys "${fingerprint}")"
GPG_PRIVATE_KEY="${private_key}" \
REQUIRE_TERRAFORM_REGISTRY_SIGNATURE=true \
  scripts/build_release_artifacts.sh 9.9.9 "${out_dir}"

checksum_file="${out_dir}/terraform-provider-polaris_9.9.9_SHA256SUMS"
signature_file="${checksum_file}.sig"
manifest_file="${out_dir}/terraform-provider-polaris_9.9.9_manifest.json"

test -s "${checksum_file}"
test -s "${signature_file}"
test -s "${manifest_file}"
jq -e '.version == 1 and (.metadata.protocol_versions | index("5.0"))' "${manifest_file}" >/dev/null

(
  cd "${out_dir}"
  sha256sum -c "$(basename "${checksum_file}")"
  gpg --batch --homedir "${gpg_home}" --verify "$(basename "${signature_file}")" "$(basename "${checksum_file}")" >/dev/null 2>&1
)

for target in linux_amd64 linux_arm64 darwin_amd64 darwin_arm64 windows_amd64 windows_arm64; do
  os="${target%_*}"
  arch="${target#*_}"
  zip_file="${out_dir}/terraform-provider-polaris_9.9.9_${os}_${arch}.zip"
  binary="terraform-provider-polaris_v9.9.9"
  if [[ "${os}" == "windows" ]]; then
    binary="${binary}.exe"
  fi
  test -s "${zip_file}"
  unzip -l "${zip_file}" "${binary}" >/dev/null
done

grep -q 'terraform-provider-polaris_9.9.9_manifest.json' "${checksum_file}"
echo "Terraform Registry release artifact smoke test passed."
