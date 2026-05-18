#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT_DIR}"

ENDPOINT="${POLARIS_ENDPOINT:-http://localhost:8181/api/management/v1}"
MANAGEMENT_SEGMENT="/api/management"
CATALOG_SEGMENT="/api/catalog"
DEFAULT_CATALOG_ENDPOINT="${ENDPOINT/${MANAGEMENT_SEGMENT}/${CATALOG_SEGMENT}}"
CATALOG_ENDPOINT="${POLARIS_CATALOG_ENDPOINT:-${DEFAULT_CATALOG_ENDPOINT}}"
REALM="${POLARIS_REALM:-POLARIS}"
ROOT_CLIENT_ID="${POLARIS_ROOT_CLIENT_ID:-root}"
ROOT_CLIENT_SECRET="${POLARIS_ROOT_CLIENT_SECRET:-s3cr3t}"
TEST_DIR="${ROOT_DIR}/examples/test-catalog"
TEST_CATALOG_NAME="${POLARIS_TEST_CATALOG_NAME:-agentic_test}"
PLUGIN_VERSION="0.0.1"
PLUGIN_DIR="${HOME}/.terraform.d/plugins/registry.terraform.io/tsukubatexas/polaris/${PLUGIN_VERSION}/$(go env GOOS)_$(go env GOARCH)"

cleanup_terraform_workdir() {
  rm -rf "${TEST_DIR}/.terraform" "${TEST_DIR}/.terraform.lock.hcl"
}

cleanup_catalog() {
  if [[ -n "${TOKEN:-}" ]]; then
    curl -fsS -o /dev/null -X DELETE \
      -H "Polaris-Realm: ${REALM}" \
      -H "Authorization: Bearer ${TOKEN}" \
      "${ENDPOINT}/catalogs/${TEST_CATALOG_NAME}" 2>/dev/null || true
  fi
}

trap 'cleanup_catalog; cleanup_terraform_workdir' EXIT

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
cp dist/terraform-provider-polaris "${PLUGIN_DIR}/terraform-provider-polaris_v${PLUGIN_VERSION}"

echo "Waiting for Polaris at ${CATALOG_ENDPOINT}."
TOKEN=""
for _ in $(seq 1 90); do
  if TOKEN="$(
    curl -fsS \
      -H "Polaris-Realm: ${REALM}" \
      -u "${ROOT_CLIENT_ID}:${ROOT_CLIENT_SECRET}" \
      -d grant_type=client_credentials \
      -d scope=PRINCIPAL_ROLE:ALL \
      "${CATALOG_ENDPOINT}/oauth/tokens" | jq -er .access_token
  )"; then
    break
  fi
  sleep 2
done

if [[ -z "${TOKEN}" ]]; then
  echo "Polaris did not return an OAuth token within the readiness timeout." >&2
  exit 1
fi

cleanup_catalog
cleanup_terraform_workdir

terraform -chdir="${TEST_DIR}" init -input=false
terraform -chdir="${TEST_DIR}" apply -input=false -auto-approve \
  -var "endpoint=${ENDPOINT}" \
  -var "realm=${REALM}" \
  -var "token=${TOKEN}"
terraform -chdir="${TEST_DIR}" destroy -input=false -auto-approve \
  -var "endpoint=${ENDPOINT}" \
  -var "realm=${REALM}" \
  -var "token=${TOKEN}"
