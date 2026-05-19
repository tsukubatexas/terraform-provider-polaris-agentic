#!/usr/bin/env bash
set -euo pipefail

pr_number="${1:-}"

if [[ -z "${pr_number}" ]]; then
  echo "Release Please PR number is required." >&2
  exit 1
fi

if [[ -z "${GITHUB_REPOSITORY:-}" ]]; then
  echo "GITHUB_REPOSITORY is required." >&2
  exit 1
fi

token="${GITHUB_TOKEN:-${GH_TOKEN:-}}"
if [[ -z "${token}" ]]; then
  echo "GITHUB_TOKEN or GH_TOKEN is required." >&2
  exit 1
fi

curl_bin="${CURL_BIN:-curl}"
api_root="https://api.github.com/repos/${GITHUB_REPOSITORY}"

api_get() {
  local url="$1"
  "${curl_bin}" -fsSL \
    -H "Authorization: Bearer ${token}" \
    -H "Accept: application/vnd.github+json" \
    -H "X-GitHub-Api-Version: 2022-11-28" \
    "${url}"
}

pr_json="$(api_get "${api_root}/pulls/${pr_number}")"

state="$(jq -r '.state // empty' <<<"${pr_json}")"
draft="$(jq -r '.draft // empty' <<<"${pr_json}")"
head_ref="$(jq -r '.head.ref // empty' <<<"${pr_json}")"
head_repo="$(jq -r '.head.repo.full_name // empty' <<<"${pr_json}")"
base_ref="$(jq -r '.base.ref // empty' <<<"${pr_json}")"
title="$(jq -r '.title // empty' <<<"${pr_json}")"

if [[ "${state}" != "open" ]]; then
  echo "PR #${pr_number} is not open." >&2
  exit 1
fi

if [[ "${draft}" == "true" ]]; then
  echo "PR #${pr_number} is draft and must not be auto-merged." >&2
  exit 1
fi

if [[ "${head_repo}" != "${GITHUB_REPOSITORY}" ]]; then
  echo "PR #${pr_number} head repo must be ${GITHUB_REPOSITORY}, got ${head_repo}." >&2
  exit 1
fi

if [[ "${head_ref}" != "release-please--branches--main" ]]; then
  echo "PR #${pr_number} is not the managed Release Please branch: ${head_ref}." >&2
  exit 1
fi

if [[ "${base_ref}" != "main" && "${base_ref}" != "master" ]]; then
  echo "PR #${pr_number} targets unexpected base branch: ${base_ref}." >&2
  exit 1
fi

if [[ ! "${title}" =~ ^chore:\ release\ [0-9]+\.[0-9]+\.[0-9]+ ]]; then
  echo "PR #${pr_number} title does not look like a Release Please release PR: ${title}" >&2
  exit 1
fi

labels_json="$(api_get "${api_root}/issues/${pr_number}")"
if ! jq -e 'any(.labels[].name; . == "autorelease: pending")' >/dev/null <<<"${labels_json}"; then
  echo "PR #${pr_number} does not have the autorelease: pending label." >&2
  exit 1
fi

files_json="$(api_get "${api_root}/pulls/${pr_number}/files?per_page=100")"
file_count="$(jq 'length' <<<"${files_json}")"
if [[ "${file_count}" -ge 100 ]]; then
  echo "PR #${pr_number} touches too many files for the hardened Release Please auto-merge path." >&2
  exit 1
fi

bad_files="$(
  jq -r '
    .[].filename
    | select(. != "CHANGELOG.md" and . != ".release-please-manifest.json")
  ' <<<"${files_json}"
)"

if [[ -n "${bad_files}" ]]; then
  echo "PR #${pr_number} changes files outside the Release Please allowlist:" >&2
  printf '%s\n' "${bad_files}" >&2
  exit 1
fi

required_files="$(jq -r '.[].filename' <<<"${files_json}")"
if ! grep -qx 'CHANGELOG.md' <<<"${required_files}" || ! grep -qx '.release-please-manifest.json' <<<"${required_files}"; then
  echo "PR #${pr_number} must update CHANGELOG.md and .release-please-manifest.json." >&2
  exit 1
fi

echo "Release Please PR #${pr_number} is valid for auto-merge."
