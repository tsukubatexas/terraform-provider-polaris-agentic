#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT_DIR}"

diff_base_files() {
  local base_ref="${CHECK_BASE_REF:-}"
  if [[ -z "${base_ref}" && -n "${GITHUB_BASE_REF:-}" ]]; then
    base_ref="origin/${GITHUB_BASE_REF}"
  fi

  if [[ -n "${base_ref}" ]] && git rev-parse --verify "${base_ref}" >/dev/null 2>&1; then
    local merge_base
    merge_base="$(git merge-base HEAD "${base_ref}")"
    git diff --name-only "${merge_base}...HEAD" -- "$@"
  fi
}

generated_changes="$(
  {
    diff_base_files internal/generated/operations_gen.go docs/generated-operations.md
    git diff --name-only -- internal/generated/operations_gen.go docs/generated-operations.md
  } | sort -u
)"

if [[ -z "${generated_changes}" ]]; then
  exit 0
fi

static_coverage_changes="$(
  {
    diff_base_files scripts/test_catalog.sh examples/test-catalog
    git diff --name-only -- scripts/test_catalog.sh examples/test-catalog
    git ls-files --others --exclude-standard -- scripts/test_catalog.sh examples/test-catalog
  } | sort -u
)"

if [[ -n "${static_coverage_changes}" ]]; then
  exit 0
fi

cat >&2 <<'MSG'
Generated Polaris operations changed, but the static real-Polaris infra checks did not.

When a new Polaris release changes the generated operation registry, the final gate must
grow with it. Update scripts/test_catalog.sh and/or examples/test-catalog so the new
release capability is covered by a durable real-Polaris Terraform smoke test.
MSG

echo >&2
echo "Changed generated operation lines:" >&2
git diff --unified=0 -- internal/generated/operations_gen.go docs/generated-operations.md 2>/dev/null |
  sed -n '/^[+][^+].*/p' |
  head -80 >&2

exit 1
