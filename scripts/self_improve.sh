#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT_DIR}"

MAX_ROUNDS="${AGENT_MAX_ROUNDS:-5}"
AGENT_MODEL="${AGENT_MODEL:-gpt-5.2}"
WORK_DIR="${ROOT_DIR}/.agent"
PROMPT_FILE="${WORK_DIR}/self-improve-prompt.md"
FAILURE_LOG="${WORK_DIR}/self-improve-checks.log"
LAST_MESSAGE="${WORK_DIR}/self-improve-last-message.md"
mkdir -p "${WORK_DIR}"

default_codex_command() {
  if [[ -n "${OPENAI_API_KEY:-}" ]]; then
    printf 'npx --prefix tools/agent-runtime codex exec --dangerously-bypass-approvals-and-sandbox -a never --search -m %q -C %q -o %q -' "${AGENT_MODEL}" "${ROOT_DIR}" "${LAST_MESSAGE}"
  fi
}

AGENT_REPAIR_COMMAND="${AGENT_REPAIR_COMMAND:-$(default_codex_command)}"

run_checks() {
  make generate &&
    make fmt &&
    make test &&
    make build &&
    bash -n scripts/*.sh
}

write_prompt() {
  local round="$1"
  {
    echo "# Self-Improving Public Repo Maintenance"
    echo
    echo "You are maintaining a public autonomous Terraform provider repo for Apache Polaris."
    echo "Make the repo more self-maintaining and secure without breaking CI."
    echo
    echo "Round: ${round}/${MAX_ROUNDS}"
    echo
    echo "Goals:"
    echo "- Keep OpenAI Codex CLI pinned in tools/agent-runtime and compatible with the workflows."
    echo "- Keep Go dependencies modern but stable."
    echo "- Improve tests for generator/provider behavior."
    echo "- Harden GitHub Actions permissions and avoid secret exposure."
    echo "- Keep external GitHub Actions pinned to full commit SHAs and keep scripts/check_actions_pinned.sh green."
    echo "- Preserve weekly Polaris update, test catalog, auto PR, auto merge, Release Please and monthly release train behavior."
    echo "- Keep PR commit messages compatible with Conventional Commits so Release Please can calculate SemVer."
    echo "- Run make generate fmt test build."
    echo
    echo "Do not:"
    echo "- Print secrets."
    echo "- Disable tests or security workflows."
    echo "- Make broad write permissions global when job-level permissions are enough."
    echo "- Hand-edit generated provider output except by changing generator inputs/code."
    echo
    echo "Current checks:"
    echo '```text'
    cat "${FAILURE_LOG}" 2>/dev/null || true
    echo '```'
    echo
    echo "Current dependency signals:"
    echo '```text'
    go list -m -u all 2>/dev/null || true
    npm --prefix tools/agent-runtime outdated 2>/dev/null || true
    echo '```'
  } >"${PROMPT_FILE}"
}

for round in $(seq 1 "${MAX_ROUNDS}"); do
  if run_checks >"${FAILURE_LOG}" 2>&1; then
    write_prompt "${round}"
  else
    write_prompt "${round}"
  fi

  if [[ -z "${AGENT_REPAIR_COMMAND}" ]]; then
    echo "No AGENT_REPAIR_COMMAND configured and OPENAI_API_KEY is not available." >&2
    echo "Set AGENT_REPAIR_COMMAND or OPENAI_API_KEY for autonomous self-improvement." >&2
    exit 1
  fi

  # shellcheck disable=SC2086
  ${SHELL:-bash} -c "${AGENT_REPAIR_COMMAND} < '${PROMPT_FILE}'"

  if run_checks >"${FAILURE_LOG}" 2>&1; then
    cat "${FAILURE_LOG}"
    exit 0
  fi
done

cat "${FAILURE_LOG}" >&2
exit 1
