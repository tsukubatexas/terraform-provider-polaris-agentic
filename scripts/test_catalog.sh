#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT_DIR}"

ENDPOINT="${POLARIS_ENDPOINT:-http://localhost:8181/api/management/v1}"
DEFAULT_CATALOG_ENDPOINT="${ENDPOINT/\/api\/management/\/api\/catalog}"
CATALOG_ENDPOINT="${POLARIS_CATALOG_ENDPOINT:-${DEFAULT_CATALOG_ENDPOINT}}"
REALM="${POLARIS_REALM:-POLARIS}"
ROOT_CLIENT_ID="${POLARIS_ROOT_CLIENT_ID:-root}"
ROOT_CLIENT_SECRET="${POLARIS_ROOT_CLIENT_SECRET:-s3cr3t}"
TEST_DIR="${ROOT_DIR}/examples/test-catalog"
PLUGIN_DIR="${HOME}/.terraform.d/plugins/local/polaris/polaris/0.0.0/$(go env GOOS)_$(go env GOARCH)"

need() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "Missing required command: $1" >&2
    exit 1
  fi
}

need curl
need jq
need terraform
need go

make build
mkdir -p "${PLUGIN_DIR}"
cp dist/terraform-provider-polaris "${PLUGIN_DIR}/terraform-provider-polaris"

echo "Waiting for Polaris at ${CATALOG_ENDPOINT}."
for _ in $(seq 1 90); do
  if curl -fsS "${CATALOG_ENDPOINT}/config" >/dev/null 2>&1; then
    break
  fi
  sleep 2
done

TOKEN="$(
  curl -fsS \
    -H "Polaris-Realm: ${REALM}" \
    -u "${ROOT_CLIENT_ID}:${ROOT_CLIENT_SECRET}" \
    -d grant_type=client_credentials \
    -d scope=PRINCIPAL_ROLE:ALL \
    "${CATALOG_ENDPOINT}/oauth/tokens" | jq -er .access_token
)"

terraform -chdir="${TEST_DIR}" init -input=false
terraform -chdir="${TEST_DIR}" apply -input=false -auto-approve \
  -var "endpoint=${ENDPOINT}" \
  -var "realm=${REALM}" \
  -var "token=${TOKEN}"
terraform -chdir="${TEST_DIR}" destroy -input=false -auto-approve \
  -var "endpoint=${ENDPOINT}" \
  -var "realm=${REALM}" \
  -var "token=${TOKEN}"
