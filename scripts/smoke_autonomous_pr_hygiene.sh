#!/usr/bin/env bash
set -euo pipefail

tmp_dir="$(mktemp -d)"
trap 'rm -rf "${tmp_dir}"' EXIT

fixture="${tmp_dir}/prs.json"
cat >"${fixture}" <<'JSON'
[
  {
    "number": 1,
    "title": "Old agentic update",
    "headRefName": "agentic/polaris-provider-update",
    "updatedAt": "2026-04-01T00:00:00Z",
    "createdAt": "2026-04-01T00:00:00Z",
    "labels": [{"name": "agentic"}],
    "author": {"login": "github-actions[bot]"},
    "isDraft": false
  },
  {
    "number": 2,
    "title": "Human dependency work",
    "headRefName": "codex/human-feature",
    "updatedAt": "2026-04-01T00:00:00Z",
    "createdAt": "2026-04-01T00:00:00Z",
    "labels": [{"name": "dependencies"}],
    "author": {"login": "tsukubatexas"},
    "isDraft": false
  },
  {
    "number": 3,
    "title": "Old release PR",
    "headRefName": "release-please--branches--old-main",
    "updatedAt": "2026-05-01T00:00:00Z",
    "createdAt": "2026-05-01T00:00:00Z",
    "labels": [{"name": "autorelease: pending"}],
    "author": {"login": "github-actions[bot]"},
    "isDraft": false
  },
  {
    "number": 4,
    "title": "New release PR",
    "headRefName": "release-please--branches--main",
    "updatedAt": "2026-05-18T00:00:00Z",
    "createdAt": "2026-05-18T00:00:00Z",
    "labels": [{"name": "autorelease: pending"}],
    "author": {"login": "github-actions[bot]"},
    "isDraft": false
  }
]
JSON

output="$(
  PR_HYGIENE_REPO="example/repo" \
    PR_HYGIENE_FIXTURE="${fixture}" \
    PR_HYGIENE_DRY_RUN=true \
    PR_HYGIENE_NOW_EPOCH=1779148800 \
    PR_HYGIENE_STALE_DAYS=14 \
    scripts/autonomous_pr_hygiene.sh
)"

grep -q '#1 agentic/polaris-provider-update: stale for at least 14 days' <<<"${output}"
grep -q '#3 release-please--branches--old-main: stale for at least 14 days' <<<"${output}"

if grep -q '#2 codex/human-feature' <<<"${output}"; then
  echo "Human PR was selected for closure." >&2
  exit 1
fi

if grep -q '#4 release-please--branches--main' <<<"${output}"; then
  echo "Fresh release PR was selected for closure." >&2
  exit 1
fi
