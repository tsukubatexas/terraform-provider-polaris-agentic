#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
tmp_dir="$(mktemp -d)"
trap 'rm -rf "${tmp_dir}"' EXIT

fake_curl="${tmp_dir}/fake-curl"
negative_log="${tmp_dir}/negative.log"
cat >"${fake_curl}" <<'FAKE'
#!/usr/bin/env bash
set -euo pipefail

url="${*: -1}"

case "${url}" in
  */pulls/42)
    cat <<'JSON'
{
  "state": "open",
  "draft": false,
  "title": "chore: release 1.2.3",
  "head": {
    "ref": "release-please--branches--main",
    "repo": { "full_name": "owner/repo" }
  },
  "base": { "ref": "main" }
}
JSON
    ;;
  */issues/42)
    cat <<'JSON'
{
  "labels": [
    { "name": "autorelease: pending" }
  ]
}
JSON
    ;;
  */pulls/42/files?per_page=100)
    cat <<'JSON'
[
  { "filename": "CHANGELOG.md" },
  { "filename": ".release-please-manifest.json" }
]
JSON
    ;;
  */pulls/43)
    cat <<'JSON'
{
  "state": "open",
  "draft": false,
  "title": "chore: release 1.2.3",
  "head": {
    "ref": "release-please--branches--main",
    "repo": { "full_name": "owner/repo" }
  },
  "base": { "ref": "main" }
}
JSON
    ;;
  */issues/43)
    cat <<'JSON'
{
  "labels": [
    { "name": "autorelease: pending" }
  ]
}
JSON
    ;;
  */pulls/43/files?per_page=100)
    cat <<'JSON'
[
  { "filename": "CHANGELOG.md" },
  { "filename": ".release-please-manifest.json" },
  { "filename": ".github/workflows/release.yml" }
]
JSON
    ;;
  *)
    echo "unexpected fake curl url: ${url}" >&2
    exit 1
    ;;
esac
FAKE
chmod +x "${fake_curl}"

GITHUB_REPOSITORY=owner/repo \
GITHUB_TOKEN=test-token \
CURL_BIN="${fake_curl}" \
  "${ROOT_DIR}/scripts/validate_release_please_pr.sh" 42

if GITHUB_REPOSITORY=owner/repo \
  GITHUB_TOKEN=test-token \
  CURL_BIN="${fake_curl}" \
  "${ROOT_DIR}/scripts/validate_release_please_pr.sh" 43 >"${negative_log}" 2>&1; then
  echo "Expected release PR validation to reject workflow-file changes." >&2
  exit 1
fi

grep -q '.github/workflows/release.yml' "${negative_log}"
echo "Release PR auto-merge smoke test passed."
