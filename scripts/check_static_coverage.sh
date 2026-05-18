#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT_DIR}"

if git diff --quiet -- internal/generated/operations_gen.go docs/generated-operations.md; then
  exit 0
fi

if ! git diff --quiet -- scripts/test_catalog.sh examples/test-catalog; then
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
git diff --unified=0 -- internal/generated/operations_gen.go docs/generated-operations.md |
  sed -n '/^[+][^+].*/p' |
  head -80 >&2

exit 1
