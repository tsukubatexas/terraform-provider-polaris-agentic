#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT_DIR}"

SMOKE_TEST="internal/provider/zz_agentic_loop_smoke_test.go"
SMOKE_LOG=".agent/smoke-agentic-loop.log"

cleanup() {
  rm -f "${SMOKE_TEST}"
}
trap cleanup EXIT

mkdir -p .agent

cat >"${SMOKE_TEST}" <<'GO'
package provider

import "testing"

func TestAgenticLoopSmokeFailsUntilAgentRepairs(t *testing.T) {
	t.Fatal("intentional smoke failure; fake agent must remove this file")
}
GO

AGENT_MAX_ROUNDS=2 \
AGENT_REPAIR_COMMAND="rm -f ${SMOKE_TEST} && echo fake-agent-repaired > .agent/fake-agent-repair.txt" \
  scripts/agentic_loop.sh | tee "${SMOKE_LOG}"

grep -q "running repair agent" "${SMOKE_LOG}"
grep -q "All checks green." "${SMOKE_LOG}"
grep -q "fake-agent-repaired" .agent/fake-agent-repair.txt

echo "Agentic loop smoke test passed."
