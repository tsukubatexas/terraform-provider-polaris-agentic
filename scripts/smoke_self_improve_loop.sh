#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT_DIR}"

SMOKE_LOG=".agent/smoke-self-improve-loop.log"
mkdir -p .agent

AGENT_MAX_ROUNDS=1 \
AGENT_REPAIR_COMMAND="echo fake-self-improve-agent-ran > .agent/fake-self-improve-agent.txt" \
  scripts/self_improve.sh | tee "${SMOKE_LOG}"

grep -q "fake-self-improve-agent-ran" .agent/fake-self-improve-agent.txt
grep -q "go test ./..." "${SMOKE_LOG}"

echo "Self-improvement loop smoke test passed."
