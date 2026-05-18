#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT_DIR}"

MAX_ROUNDS="${AGENT_MAX_ROUNDS:-3}"
AGENT_MODEL="${AGENT_MODEL:-gpt-5.2}"
POLARIS_INFRA_MODE="${POLARIS_INFRA_MODE:-auto}"
POLARIS_CONTAINER_NAME="${POLARIS_CONTAINER_NAME:-polaris-provider-agentic}"
POLARIS_IMAGE="${POLARIS_IMAGE:-apache/polaris:latest}"
WORK_DIR="${ROOT_DIR}/.agent"
PROMPT_FILE="${WORK_DIR}/infra-repair-prompt.md"
FAILURE_LOG="${WORK_DIR}/infra-failure.log"
LAST_MESSAGE="${WORK_DIR}/infra-last-message.md"
STARTED_CONTAINER="false"

mkdir -p "${WORK_DIR}"

default_codex_command() {
  if [[ -n "${OPENAI_API_KEY:-}" ]]; then
    printf 'npx --prefix tools/agent-runtime codex exec --dangerously-bypass-approvals-and-sandbox -a never --search -m %q -C %q -o %q -' "${AGENT_MODEL}" "${ROOT_DIR}" "${LAST_MESSAGE}"
  fi
}

AGENT_REPAIR_COMMAND="${AGENT_REPAIR_COMMAND:-$(default_codex_command)}"

need() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "Missing required command: $1" >&2
    exit 1
  fi
}

should_start_docker() {
  case "${POLARIS_INFRA_MODE}" in
  docker)
    return 0
    ;;
  service | external | none)
    return 1
    ;;
  auto)
    if [[ "${POLARIS_ENDPOINT:-http://localhost:8181/api/management/v1}" == http://localhost:* ]] ||
      [[ "${POLARIS_ENDPOINT:-}" == http://127.0.0.1:* ]]; then
      command -v docker >/dev/null 2>&1
      return $?
    fi
    return 1
    ;;
  *)
    echo "Unknown POLARIS_INFRA_MODE=${POLARIS_INFRA_MODE}. Use auto, docker, service, external, or none." >&2
    exit 1
    ;;
  esac
}

start_infra() {
  if should_start_docker; then
    need docker
    docker rm -f "${POLARIS_CONTAINER_NAME}" >/dev/null 2>&1 || true
    docker run -d \
      --name "${POLARIS_CONTAINER_NAME}" \
      -p 8181:8181 \
      -e POLARIS_BOOTSTRAP_CREDENTIALS=POLARIS,root,s3cr3t \
      "${POLARIS_IMAGE}" >/dev/null
    STARTED_CONTAINER="true"
  fi
}

cleanup_infra() {
  if [[ "${STARTED_CONTAINER}" == "true" ]]; then
    docker rm -f "${POLARIS_CONTAINER_NAME}" >/dev/null 2>&1 || true
  fi
}

run_static_infra_checks() {
  if [[ -n "${AGENTIC_INFRA_CHECK_COMMAND:-}" ]]; then
    ${SHELL:-bash} -lc "${AGENTIC_INFRA_CHECK_COMMAND}"
    return
  fi

  make generate &&
    make fmt &&
    make test &&
    make build &&
    bash -n scripts/*.sh &&
    scripts/check_adr_updates.sh &&
    scripts/check_static_coverage.sh &&
    scripts/test_catalog.sh || return $?

  return 0
}

append_infra_diagnostics() {
  {
    echo
    echo "Infra diagnostics:"
    echo "POLARIS_INFRA_MODE=${POLARIS_INFRA_MODE}"
    echo "POLARIS_ENDPOINT=${POLARIS_ENDPOINT:-http://localhost:8181/api/management/v1}"
    echo "POLARIS_IMAGE=${POLARIS_IMAGE}"
    if command -v docker >/dev/null 2>&1; then
      echo
      echo "Docker containers:"
      docker ps --filter "name=${POLARIS_CONTAINER_NAME}" --format '{{.Names}} {{.Image}} {{.Status}} {{.Ports}}' || true
      if [[ "${STARTED_CONTAINER}" == "true" ]]; then
        echo
        echo "Polaris logs:"
        docker logs --tail 120 "${POLARIS_CONTAINER_NAME}" 2>&1 || true
      fi
    fi
  } >>"${FAILURE_LOG}"
}

write_prompt() {
  local round="$1"
  {
    echo "# Final Static Infra Repair Task"
    echo
    echo "This is the final gate after provider generation and code repair."
    echo "Goal: prove the Terraform provider still performs the basic Polaris lifecycle against real infrastructure."
    echo
    echo "Round: ${round}/${MAX_ROUNDS}"
    echo "Infra mode: ${POLARIS_INFRA_MODE}"
    echo
    echo "Hard requirements:"
    echo "- Do not remove or weaken scripts/test_catalog.sh."
    echo "- Do not skip the real Polaris apply/destroy check."
    echo "- Keep secrets and tokens out of logs."
    echo "- Fix provider code, generator code, examples, or test scripts until the static infra check is green."
    echo "- If a new Polaris release changes generated operations, extend the static infra examples/checks so the new release capability is exercised against real Polaris before declaring success."
    echo "- Prefer evolving examples/test-catalog and scripts/test_catalog.sh as the durable final gate, not adding one-off checks only in this prompt."
    echo "- Add or update docs/adr records for any new Polaris runtime behavior, release compatibility finding, or static coverage decision."
    echo "- The success condition is: make generate fmt test build, bash -n scripts/*.sh, scripts/check_adr_updates.sh, scripts/check_static_coverage.sh, and scripts/test_catalog.sh."
    echo
    echo "Recent failure log:"
    echo
    echo '```text'
    tail -n 280 "${FAILURE_LOG}" 2>/dev/null || true
    echo '```'
    echo
    echo "Current git diff summary:"
    echo
    echo '```text'
    git diff --stat || true
    echo '```'
  } >"${PROMPT_FILE}"
}

trap cleanup_infra EXIT

echo "Final static Polaris infra loop"
echo "Infra mode: ${POLARIS_INFRA_MODE}"
echo "Max rounds: ${MAX_ROUNDS}"

start_infra

for round in $(seq 1 "${MAX_ROUNDS}"); do
  echo "== Final infra round ${round}/${MAX_ROUNDS}: static provider functionality =="
  if run_static_infra_checks >"${FAILURE_LOG}" 2>&1; then
    cat "${FAILURE_LOG}"
    echo "All final infra checks green."
    exit 0
  fi

  append_infra_diagnostics
  cat "${FAILURE_LOG}"
  write_prompt "${round}"

  if [[ -z "${AGENT_REPAIR_COMMAND}" ]]; then
    echo "No AGENT_REPAIR_COMMAND configured and OPENAI_API_KEY is not available." >&2
    echo "Set AGENT_REPAIR_COMMAND or OPENAI_API_KEY for autonomous infra repair." >&2
    exit 1
  fi

  echo "== Final infra round ${round}/${MAX_ROUNDS}: running repair agent =="
  # shellcheck disable=SC2086
  ${SHELL:-bash} -lc "${AGENT_REPAIR_COMMAND} < '${PROMPT_FILE}'"
done

echo "Final infra loop reached ${MAX_ROUNDS} rounds without green checks." >&2
exit 1
