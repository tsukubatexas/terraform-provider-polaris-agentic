#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT_DIR}"

failed="false"

while IFS=: read -r file line content; do
  ref="${content##*@}"
  ref="${ref%%[[:space:]#]*}"
  action="${content#*uses: }"
  action="${action%%[[:space:]#]*}"

  if [[ "${action}" == ./* ]] || [[ "${action}" == docker://* ]]; then
    continue
  fi

  if [[ ! "${ref}" =~ ^[0-9a-f]{40}$ ]]; then
    echo "${file}:${line}: external GitHub Action is not pinned to a full commit SHA: ${content}" >&2
    failed="true"
  fi
done < <(grep -RInE '^[[:space:]]*uses:[[:space:]]+[^[:space:]]+@[^[:space:]]+' .github/workflows)

if [[ "${failed}" == "true" ]]; then
  echo >&2
  echo "Pin external GitHub Actions to immutable 40-character commit SHAs." >&2
  exit 1
fi
