#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT_DIR}"

changed_files="$(
  {
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
