#!/usr/bin/env bash
set -euo pipefail

agent_default_codex_command() {
  local agent_model="$1"
  local root_dir="$2"
  local last_message="$3"

  if [[ -n "${OPENAI_API_KEY:-}" ]]; then
    printf 'npx --prefix tools/agent-runtime codex exec --dangerously-bypass-approvals-and-sandbox -m %q -C %q -o %q -' \
      "${agent_model}" "${root_dir}" "${last_message}"
  fi
}
