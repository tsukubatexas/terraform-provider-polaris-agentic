#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT_DIR}"

failed="false"

if ! command -v curl >/dev/null 2>&1; then
  echo "curl is required to verify pinned GitHub Action commits." >&2
  exit 1
fi

github_headers=(
  -H "Accept: application/vnd.github+json"
  -H "X-GitHub-Api-Version: 2022-11-28"
)

if [[ -n "${GITHUB_TOKEN:-}" ]]; then
  github_headers+=(-H "Authorization: Bearer ${GITHUB_TOKEN}")
fi

while IFS=: read -r file line content; do
  action="${content#*uses:}"
  action="${action#"${action%%[![:space:]]*}"}"
  action="${action%%[[:space:]#]*}"
  ref="${action##*@}"
  action_without_ref="${action%@*}"

  if [[ "${action_without_ref}" == ./* ]] || [[ "${action_without_ref}" == docker://* ]]; then
    continue
  fi

  if [[ ! "${ref}" =~ ^[0-9a-f]{40}$ ]]; then
    echo "${file}:${line}: external GitHub Action is not pinned to a full commit SHA: ${content}" >&2
    failed="true"
    continue
  fi

  owner="${action_without_ref%%/*}"
  rest="${action_without_ref#*/}"
  repo="${rest%%/*}"

  if [[ -z "${owner}" ]] || [[ -z "${repo}" ]] || [[ "${owner}" == "${action_without_ref}" ]]; then
    echo "${file}:${line}: cannot determine GitHub Action repository for: ${content}" >&2
    failed="true"
    continue
  fi

  action_repo="${owner}/${repo}"
  commit_url="https://api.github.com/repos/${action_repo}/commits/${ref}"

  if ! curl -fsSL "${github_headers[@]}" "${commit_url}" >/dev/null; then
    echo "${file}:${line}: pinned SHA is not a commit in ${action_repo}: ${ref}" >&2
    failed="true"
  fi
done < <(grep -RInE '^[[:space:]]*uses:[[:space:]]+[^[:space:]]+@[^[:space:]]+' .github/workflows)

if [[ "${failed}" == "true" ]]; then
  echo >&2
  echo "Pin external GitHub Actions to immutable 40-character commit SHAs that belong to the referenced action repository." >&2
  exit 1
fi
