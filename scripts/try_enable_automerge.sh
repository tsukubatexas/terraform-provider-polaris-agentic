#!/usr/bin/env bash
set -euo pipefail

pr_number="${1:-}"
merge_method="${2:-SQUASH}"

if [[ -z "${pr_number}" ]]; then
  echo "No pull request number provided; skipping auto-merge."
  exit 0
fi

if [[ "${DISABLE_AUTOMERGE:-}" == "true" ]]; then
  echo "DISABLE_AUTOMERGE=true; skipping auto-merge for PR #${pr_number}."
  exit 0
fi

if [[ -z "${GITHUB_TOKEN:-}" ]]; then
  echo "GITHUB_TOKEN is not available; skipping auto-merge for PR #${pr_number}."
  exit 0
fi

if [[ -z "${GITHUB_REPOSITORY:-}" ]]; then
  echo "GITHUB_REPOSITORY is not available; skipping auto-merge for PR #${pr_number}."
  exit 0
fi

merge_method="${merge_method^^}"

echo "Trying to enable ${merge_method} auto-merge for ${GITHUB_REPOSITORY}#${pr_number}."

pr_json="$(
  curl -fsSL \
    -H "Authorization: Bearer ${GITHUB_TOKEN}" \
    -H "Accept: application/vnd.github+json" \
    -H "X-GitHub-Api-Version: 2022-11-28" \
    "https://api.github.com/repos/${GITHUB_REPOSITORY}/pulls/${pr_number}" 2>/dev/null || true
)"

node_id="$(jq -r '.node_id // empty' <<<"${pr_json}")"
if [[ -z "${node_id}" ]]; then
  echo "Could not resolve PR node id; skipping auto-merge."
  exit 0
fi

payload="$(
  jq -nc \
    --arg id "${node_id}" \
    --arg method "${merge_method}" \
    '{
      query: "mutation($id: ID!, $method: PullRequestMergeMethod!) { enablePullRequestAutoMerge(input: { pullRequestId: $id, mergeMethod: $method }) { pullRequest { number } } }",
      variables: { id: $id, method: $method }
    }'
)"

response="$(
  curl -fsS \
    -H "Authorization: Bearer ${GITHUB_TOKEN}" \
    -H "Accept: application/vnd.github+json" \
    -H "X-GitHub-Api-Version: 2022-11-28" \
    -H "Content-Type: application/json" \
    -d "${payload}" \
    "https://api.github.com/graphql" 2>/dev/null || true
)"

errors="$(jq -r '.errors[]?.message' <<<"${response}")"
if [[ -n "${errors}" ]]; then
  echo "Auto-merge was not enabled:"
  printf '%s\n' "${errors}"
  echo "Leaving the pull request open for normal review."
  exit 0
fi

echo "Auto-merge enabled for PR #${pr_number}."
