#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT_DIR}"

SMOKE_MARKER=".agent/agent-command-path-smoke-fail"
SMOKE_LOG=".agent/smoke-agent-command-path.log"
FAKE_BIN_DIR="$(mktemp -d)"

cleanup() {
  rm -f "${SMOKE_MARKER}"
  rm -rf "${FAKE_BIN_DIR}"
}
trap cleanup EXIT

mkdir -p .agent
touch "${SMOKE_MARKER}"

{
  printf '%s\n' '#!/usr/bin/env bash'
  printf '%s\n' 'set -euo pipefail'
  printf '%s\n' "rm -f \"\$1\""
  printf '%s\n' 'echo fake-agent-command-path-ok > .agent/fake-agent-command-path.txt'
} >"${FAKE_BIN_DIR}/fake-agent-command"
chmod +x "${FAKE_BIN_DIR}/fake-agent-command"

PATH="${FAKE_BIN_DIR}:${PATH}" \
AGENT_MAX_ROUNDS=2 \
POLARIS_INFRA_MODE=none \
AGENTIC_INFRA_CHECK_COMMAND="test ! -f ${SMOKE_MARKER}" \
AGENT_REPAIR_COMMAND="fake-agent-command ${SMOKE_MARKER}" \
  scripts/agentic_infra_loop.sh | tee "${SMOKE_LOG}"

grep -q "running repair agent" "${SMOKE_LOG}"
grep -q "All final infra checks green." "${SMOKE_LOG}"
grep -q "fake-agent-command-path-ok" .agent/fake-agent-command-path.txt

echo "Agent command PATH smoke test passed."
