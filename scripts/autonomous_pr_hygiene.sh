#!/usr/bin/env bash
set -euo pipefail

repo="${PR_HYGIENE_REPO:-${GITHUB_REPOSITORY:-}}"
stale_days="${PR_HYGIENE_STALE_DAYS:-14}"
failed_days="${PR_HYGIENE_FAILED_DAYS:-3}"
dry_run="${PR_HYGIENE_DRY_RUN:-false}"
now_epoch="${PR_HYGIENE_NOW_EPOCH:-$(date -u +%s)}"

if [[ -z "${repo}" ]]; then
  echo "PR_HYGIENE_REPO or GITHUB_REPOSITORY is required." >&2
  exit 1
fi

if ! command -v jq >/dev/null 2>&1; then
  echo "jq is required." >&2
  exit 1
fi

if [[ -n "${PR_HYGIENE_FIXTURE:-}" ]]; then
  prs_json="$(cat "${PR_HYGIENE_FIXTURE}")"
else
  if ! command -v gh >/dev/null 2>&1; then
    echo "gh is required when PR_HYGIENE_FIXTURE is not set." >&2
    exit 1
  fi
  if [[ -z "${GH_TOKEN:-${GITHUB_TOKEN:-}}" ]]; then
    echo "GH_TOKEN or GITHUB_TOKEN is required for GitHub API access." >&2
    exit 1
  fi
  prs_json="$(
    gh pr list \
      --repo "${repo}" \
      --state open \
      --limit 100 \
      --json number,title,headRefName,updatedAt,createdAt,labels,author,isDraft
  )"
fi

candidate_json="$(
  jq -c \
    --argjson now "${now_epoch}" \
    --argjson stale_days "${stale_days}" \
    '
    def bot_author:
      ((.author.login // "") | test("\\[bot\\]$"));

    def autonomous:
      (.headRefName | test("^(agentic/|dependabot/|release-please--branches--)"))
      or (bot_author and ([.labels[]?.name] | any(. == "agentic" or . == "dependencies" or . == "autorelease: pending")));

    def family:
      if (.headRefName | startswith("release-please--branches--")) then "release-please"
      elif (.headRefName | startswith("agentic/")) then .headRefName
      elif (.headRefName | startswith("dependabot/")) then .headRefName
      else .headRefName
      end;

    def prepared:
      map(select(autonomous) | . + {
        family: family,
        updatedEpoch: (.updatedAt | fromdateiso8601)
      });

    def stale_candidates:
      prepared
      | map(select(($now - .updatedEpoch) >= ($stale_days * 86400))
      | {
          number,
          headRefName,
          reason: ("stale for at least " + ($stale_days | tostring) + " days")
        });

    def duplicate_candidates:
      prepared
      | group_by(.family)
      | map(sort_by(.updatedEpoch) | reverse | .[1:][])
      | flatten
      | map({
          number,
          headRefName,
          reason: ("superseded by a newer autonomous PR in family " + .family)
        });

    [stale_candidates, duplicate_candidates]
    | flatten
    | unique_by(.number)
    ' <<<"${prs_json}"
)"

failed_json="[]"
if [[ -z "${PR_HYGIENE_FIXTURE:-}" ]]; then
  autonomous_numbers="$(
    jq -r '
      def bot_author:
        ((.author.login // "") | test("\\[bot\\]$"));

      .[]
      | select((.headRefName | test("^(agentic/|dependabot/|release-please--branches--)"))
        or (bot_author and ([.labels[]?.name] | any(. == "agentic" or . == "dependencies" or . == "autorelease: pending"))))
      | [.number, .updatedAt, .headRefName] | @tsv
    ' <<<"${prs_json}"
  )"

  failed_lines=()
  while IFS=$'\t' read -r number updated_at head_ref; do
    [[ -z "${number}" ]] && continue
    updated_epoch="$(jq -nr --arg ts "${updated_at}" '$ts | fromdateiso8601')"
    if (( now_epoch - updated_epoch < failed_days * 86400 )); then
      continue
    fi

    checks_json="$(gh pr checks "${number}" --repo "${repo}" --json name,state,bucket 2>/dev/null || true)"
    if [[ -z "${checks_json}" ]]; then
      continue
    fi
    failed_count="$(jq '[.[] | select(.bucket == "fail" or .state == "FAILURE" or .state == "CANCELLED" or .state == "TIMED_OUT")] | length' <<<"${checks_json}")"
    if (( failed_count > 0 )); then
      failed_lines+=("$(jq -nc --argjson number "${number}" --arg head "${head_ref}" --argjson days "${failed_days}" '{number: $number, headRefName: $head, reason: ("failed checks for at least " + ($days | tostring) + " days")}')")
    fi
  done <<<"${autonomous_numbers}"

  if (( ${#failed_lines[@]} > 0 )); then
    failed_json="$(printf '%s\n' "${failed_lines[@]}" | jq -s 'unique_by(.number)')"
  fi
fi

to_close_json="$(jq -s 'add | unique_by(.number) | sort_by(.number)' <<<"${candidate_json}"$'\n'"${failed_json}")"
count="$(jq 'length' <<<"${to_close_json}")"

if (( count == 0 )); then
  echo "No autonomous PRs need closing."
  exit 0
fi

echo "Autonomous PR hygiene selected ${count} PR(s) for closing."
jq -r '.[] | "#\(.number) \(.headRefName): \(.reason)"' <<<"${to_close_json}"

if [[ "${dry_run}" == "true" ]]; then
  echo "PR_HYGIENE_DRY_RUN=true; not closing PRs."
  exit 0
fi

while IFS=$'\t' read -r number head_ref reason; do
  [[ -z "${number}" ]] && continue
  comment="Autonomous PR hygiene is closing this bot-managed PR because it is ${reason}. Matched head branch: ${head_ref}."
  if gh pr close "${number}" --repo "${repo}" --delete-branch --comment "${comment}"; then
    continue
  fi
  gh pr close "${number}" --repo "${repo}" --comment "${comment}"
done < <(jq -r '.[] | [.number, .headRefName, .reason] | @tsv' <<<"${to_close_json}")
