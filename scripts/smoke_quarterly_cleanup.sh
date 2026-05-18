#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT_DIR}"

SMOKE_MARKER=".agent/quarterly-cleanup-smoke-fail"
SMOKE_LOG=".agent/smoke-quarterly-cleanup.log"

cleanup() {
  rm -f "${SMOKE_MARKER}"
}
trap cleanup EXIT

mkdir -p .agent
touch "${SMOKE_MARKER}"

AGENT_MAX_ROUNDS=1 \
QUARTERLY_CHECK_COMMAND="test ! -f ${SMOKE_MARKER}" \
AGENT_REPAIR_COMMAND="rm -f ${SMOKE_MARKER} && echo fake-quarterly-cleanup-agent-ran > .agent/fake-quarterly-cleanup-agent.txt" \
  scripts/quarterly_cleanup.sh | tee "${SMOKE_LOG}"

grep -q "running cleanup agent" "${SMOKE_LOG}"
grep -q "Quarterly cleanup checks green." "${SMOKE_LOG}"
grep -q "fake-quarterly-cleanup-agent-ran" .agent/fake-quarterly-cleanup-agent.txt

echo "Quarterly cleanup loop smoke test passed."
