#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT_DIR}"

if ! command -v jq >/dev/null 2>&1; then
  echo "Missing required command: jq" >&2
  exit 1
fi

jq empty release-please-config.json .release-please-manifest.json

root_release_type="$(jq -r '.packages["."]."release-type" // ."release-type" // empty' release-please-config.json)"
if [[ "${root_release_type}" != "go" ]]; then
  echo "release-please root package must use release-type=go" >&2
  exit 1
fi

include_component="$(jq -r '."include-component-in-tag"' release-please-config.json)"
if [[ "${include_component}" != "false" ]]; then
  echo "release-please must publish plain vX.Y.Z tags, without component prefix" >&2
  exit 1
fi

manifest_version="$(jq -r '."."' .release-please-manifest.json)"
if [[ ! "${manifest_version}" =~ ^[0-9]+\.[0-9]+\.[0-9]+([-.][0-9A-Za-z.-]+)?$ ]]; then
  echo "invalid root release-please manifest version: ${manifest_version}" >&2
  exit 1
fi
