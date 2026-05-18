#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT_DIR}"

RELEASES=("$@")
if [[ ${#RELEASES[@]} -eq 0 ]]; then
  RELEASES=(
    apache-polaris-0.9.0-incubating
    apache-polaris-1.0.0-incubating
    apache-polaris-1.2.0-incubating
    apache-polaris-1.3.0-incubating
    apache-polaris-1.4.1
    apache-polaris-1.5.0
  )
fi

WORK_DIR=".agent/release-matrix"
BACKUP_DIR="${WORK_DIR}/backup"
mkdir -p "${BACKUP_DIR}"

cp internal/generated/operations_gen.go "${BACKUP_DIR}/operations_gen.go"
cp docs/generated-operations.md "${BACKUP_DIR}/generated-operations.md"

cleanup() {
  cp "${BACKUP_DIR}/operations_gen.go" internal/generated/operations_gen.go
  cp "${BACKUP_DIR}/generated-operations.md" docs/generated-operations.md
  rm -rf "${WORK_DIR}/spec-cache"
  go fmt ./internal/generated >/dev/null
}
trap cleanup EXIT

for release in "${RELEASES[@]}"; do
  echo "== Polaris release matrix: ${release} =="
  POLARIS_RELEASE="${release}" SPEC_CACHE_DIR="${WORK_DIR}/spec-cache" make generate
  make fmt
  make test
  make build
done

echo "Polaris release matrix smoke test passed for: ${RELEASES[*]}"
