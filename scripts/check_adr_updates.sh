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
    git diff --name-only "${merge_base}...HEAD"
  fi
}

changed_files="$(
  {
    diff_base_files
    git diff --name-only
    git ls-files --others --exclude-standard
  } | sort -u
)"
if [[ -z "${changed_files}" ]]; then
  exit 0
fi

if printf '%s\n' "${changed_files}" | grep -q '^docs/adr/'; then
  exit 0
fi

if ! printf '%s\n' "${changed_files}" | grep -Eq '^(\.github/workflows/|AGENTS\.md|Makefile|go\.mod|go\.sum|cmd/|internal/|scripts/|examples/test-catalog/|tools/agent-runtime/)'; then
  exit 0
fi

cat >&2 <<'MSG'
Agentic-relevant files changed, but no ADR was added or updated.

Durable provider, generator, workflow, test, release, runtime, or agent-loop changes
must be tracked in docs/adr so future autonomous runs stay explainable.
MSG

echo >&2
echo "Changed files requiring ADR consideration:" >&2
printf '%s\n' "${changed_files}" |
  grep -E '^(\.github/workflows/|AGENTS\.md|Makefile|go\.mod|go\.sum|cmd/|internal/|scripts/|examples/test-catalog/|tools/agent-runtime/)' >&2

exit 1
