#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT_DIR}"

version="${1:?usage: scripts/build_release_artifacts.sh <version> [output-dir]}"
out_dir="${2:-dist}"

if ! command -v zip >/dev/null 2>&1; then
  echo "Missing required command: zip" >&2
  exit 1
fi

if ! command -v sha256sum >/dev/null 2>&1; then
  echo "Missing required command: sha256sum" >&2
  exit 1
fi

rm -rf "${out_dir}"
mkdir -p "${out_dir}"

for target in linux_amd64 linux_arm64 darwin_amd64 darwin_arm64; do
  os="${target%_*}"
  arch="${target#*_}"
  work="${out_dir}/${target}"
  mkdir -p "${work}"
  GOOS="${os}" GOARCH="${arch}" go build \
    -buildvcs=false \
    -ldflags "-X main.version=${version}" \
    -o "${work}/terraform-provider-polaris_v${version}" .
  (
    cd "${work}"
    zip -q "../terraform-provider-polaris_${version}_${os}_${arch}.zip" "terraform-provider-polaris_v${version}"
  )
done

(
  cd "${out_dir}"
  sha256sum ./*.zip >"terraform-provider-polaris_${version}_SHA256SUMS"
)
