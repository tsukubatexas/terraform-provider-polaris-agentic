#!/usr/bin/env bash
set -euo pipefail

before="$(git diff --name-only)"
make generate
after="$(git diff --name-only)"

if [[ "${before}" != "${after}" || -n "$(git diff -- internal/generated docs specs)" ]]; then
  echo "Generated Polaris artifacts are stale. Run make generate." >&2
  git diff --stat -- internal/generated docs specs >&2 || true
  exit 1
fi
