#!/usr/bin/env bash
set -euo pipefail

generated_paths=(
  internal/generated
  docs/generated-operations.md
  docs/index.md
  docs/resources
  docs/data-sources
  docs/guides
  examples/provider
  examples/resources
  examples/data-sources
  examples/complete-polaris
  specs
)

snapshot_generated() {
  git diff --no-ext-diff -- "${generated_paths[@]}"
  git status --porcelain -- "${generated_paths[@]}"
  git ls-files --others --exclude-standard -- "${generated_paths[@]}" |
    while IFS= read -r file; do
      [[ -z "${file}" ]] && continue
      printf 'untracked-sha256  '
      shasum -a 256 "${file}"
    done
}

before="$(snapshot_generated)"
make generate
after="$(snapshot_generated)"

if [[ "${before}" != "${after}" ]]; then
  echo "Generated Polaris artifacts are stale. Run make generate." >&2
  git diff --stat -- "${generated_paths[@]}" >&2 || true
  exit 1
fi
