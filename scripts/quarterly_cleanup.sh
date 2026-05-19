#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT_DIR}"

MAX_ROUNDS="${AGENT_MAX_ROUNDS:-4}"
AGENT_MODEL="${AGENT_MODEL:-gpt-5.2}"
WORK_DIR="${ROOT_DIR}/.agent"
PROMPT_FILE="${WORK_DIR}/quarterly-cleanup-prompt.md"
FAILURE_LOG="${WORK_DIR}/quarterly-cleanup-failure.log"
LAST_MESSAGE="${WORK_DIR}/quarterly-cleanup-last-message.md"

mkdir -p "${WORK_DIR}"

default_codex_command() {
  if [[ -n "${OPENAI_API_KEY:-}" ]]; then
    printf 'npx --prefix tools/agent-runtime codex exec --dangerously-bypass-approvals-and-sandbox -m %q -C %q -o %q -' "${AGENT_MODEL}" "${ROOT_DIR}" "${LAST_MESSAGE}"
  fi
}

AGENT_REPAIR_COMMAND="${AGENT_REPAIR_COMMAND:-$(default_codex_command)}"

run_checks() {
  if [[ -n "${QUARTERLY_CHECK_COMMAND:-}" ]]; then
    ${SHELL:-bash} -c "${QUARTERLY_CHECK_COMMAND}"
    return
  fi

  go mod tidy &&
    npm --prefix tools/agent-runtime install --package-lock-only &&
    make generate &&
    make fmt &&
    make test &&
    make build &&
    bash -n scripts/*.sh &&
    scripts/smoke_autonomous_pr_hygiene.sh &&
    scripts/check_adr_updates.sh &&
    scripts/check_static_coverage.sh || return $?

  return 0
}

write_prompt() {
  local round="$1"
  local status="$2"
  {
    echo "# Quarterly Cleanup and Hardening Task"
    echo
    echo "You are running the quarterly maintenance loop for a public autonomous Terraform provider for Apache Polaris."
    echo
    echo "Round: ${round}/${MAX_ROUNDS}"
    echo "Status: ${status}"
    echo
    echo "Goals:"
    echo "- Remove stale, duplicated, or misleading code and docs."
    echo "- Simplify scripts while preserving behavior."
    echo "- Update Go and Node lockfiles only when safe and intentional."
    echo "- Keep generated files deterministic."
    echo "- Keep GitHub Actions least-privilege and public-repo safe."
    echo "- Preserve the normal agentic update loop and the separate final static infra loop."
    echo "- Keep autonomous PR hygiene working so stale bot PRs are closed without human cleanup."
    echo "- Add or update ADRs for every durable cleanup, workflow, dependency, or test strategy decision."
    echo "- Do not modify .github/workflows files with the default GitHub token; workflow-file changes require a separate reviewed maintainer PR."
    echo
    echo "Do not:"
    echo "- Remove tests, smoke checks, ADR guards, or real Polaris static coverage to get green."
    echo "- Print secrets or tokens."
    echo "- Broaden workflow permissions casually."
    echo "- Edit .github/workflows files from this autonomous cleanup loop."
    echo "- Hand-edit generated files instead of fixing the generator."
    echo
    echo "Success condition:"
    echo "- go mod tidy"
    echo "- npm --prefix tools/agent-runtime install --package-lock-only"
    echo "- make generate fmt test build"
    echo "- bash -n scripts/*.sh"
    echo "- scripts/smoke_autonomous_pr_hygiene.sh"
    echo "- scripts/check_adr_updates.sh"
    echo "- scripts/check_static_coverage.sh"
    echo "- final workflow also runs scripts/agentic_infra_loop.sh against real Polaris"
    echo
    echo "Recent failure log:"
    echo
    echo '```text'
    tail -n 280 "${FAILURE_LOG}" 2>/dev/null || true
    echo '```'
    echo
    echo "Dependency signals:"
    echo
    echo '```text'
    go list -m -u all 2>/dev/null || true
    npm --prefix tools/agent-runtime outdated 2>/dev/null || true
    echo '```'
    echo
    echo "Current git diff summary:"
    echo
    echo '```text'
    git diff --stat || true
    echo '```'
  } >"${PROMPT_FILE}"
}

if [[ -z "${AGENT_REPAIR_COMMAND}" ]]; then
  echo "No agent configured; running deterministic quarterly cleanup checks only."
  run_checks
  exit 0
fi

for round in $(seq 1 "${MAX_ROUNDS}"); do
  write_prompt "${round}" "quarterly cleanup"

  echo "== Quarterly cleanup round ${round}/${MAX_ROUNDS}: running cleanup agent =="
  # shellcheck disable=SC2086
  ${SHELL:-bash} -c "${AGENT_REPAIR_COMMAND} < '${PROMPT_FILE}'"

  echo "== Quarterly cleanup round ${round}/${MAX_ROUNDS}: checks =="
  if run_checks >"${FAILURE_LOG}" 2>&1; then
    cat "${FAILURE_LOG}"
    echo "Quarterly cleanup checks green."
    exit 0
  fi

  cat "${FAILURE_LOG}"
done

echo "Quarterly cleanup loop reached ${MAX_ROUNDS} rounds without green checks." >&2
exit 1
