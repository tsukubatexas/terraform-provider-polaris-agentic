#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT_DIR}"

SMOKE_MARKER=".agent/final-infra-smoke-fail"
SMOKE_LOG=".agent/smoke-agentic-infra-loop.log"

cleanup() {
  rm -f "${SMOKE_MARKER}"
}
trap cleanup EXIT

mkdir -p .agent
touch "${SMOKE_MARKER}"

AGENT_MAX_ROUNDS=2 \
POLARIS_INFRA_MODE=none \
AGENTIC_INFRA_CHECK_COMMAND="test ! -f ${SMOKE_MARKER}" \
AGENT_REPAIR_COMMAND="rm -f ${SMOKE_MARKER} && echo fake-final-infra-agent-repaired > .agent/fake-final-infra-agent-repair.txt" \
  scripts/agentic_infra_loop.sh | tee "${SMOKE_LOG}"

grep -q "running repair agent" "${SMOKE_LOG}"
grep -q "All final infra checks green." "${SMOKE_LOG}"
grep -q "fake-final-infra-agent-repaired" .agent/fake-final-infra-agent-repair.txt

echo "Final static infra loop smoke test passed."
