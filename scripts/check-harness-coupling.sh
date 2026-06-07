#!/usr/bin/env bash
# Audita acoplamientos tool/metodologia y falla solo para superficies neutrales.

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
STATUS=0

PATTERN='OpenCode|\.opencode|opencode\.json|OpenSpec|openspec/|/opsx|\.claude|\.codex|CLAUDE\.md'
HTML_OVERVIEW_COMMAND='lufy-ai opsx render --change <change> --format html --theme notion-dark'
HTML_OVERVIEW_CTA='¿Quieres que genere ahora el reporte HTML offline de los artifacts con tema Notion dark?'

NEUTRAL_PATHS=(
  "tools/lufy-cli-go/internal/core"
  "tools/lufy-cli-go/internal/instructions/roles"
  "tools/lufy-cli-go/internal/instructions/render"
  "tools/lufy-cli-go/internal/skills/registry"
  ".lufy/instructions/roles"
)

ADAPTER_RULES=(
  "tools/lufy-cli-go/internal/adapters/methodology/none:OpenSpec|openspec/|/opsx"
  "tools/lufy-cli-go/internal/adapters/tool/codex:OpenCode|\\.opencode|opencode\\.json"
  "tools/lufy-cli-go/internal/adapters/tool/claude-code:OpenCode|\\.opencode|opencode\\.json"
)

HTML_OVERVIEW_CONTRACT_PATHS=(
  ".opencode/commands/opsx-propose.md"
  ".opencode/skills/sdd-workflow/openspec-propose/SKILL.md"
  ".opencode/agents/orchestrator.md"
  "tools/lufy-cli-go/internal/assets/embedded/.opencode/commands/opsx-propose.md"
  "tools/lufy-cli-go/internal/assets/embedded/.opencode/skills/sdd-workflow/openspec-propose/SKILL.md"
  "tools/lufy-cli-go/internal/assets/embedded/.opencode/agents/orchestrator.md"
)

check_neutral_path() {
  local rel="$1"
  local path="$ROOT/$rel"

  [ -e "$path" ] || return 0

  if rg -n --glob '!**/node_modules/**' --glob '!**/.git/**' "$PATTERN" "$path"; then
    printf 'Error: superficie neutral contiene acoplamiento tool/metodologia: %s\n' "$rel" >&2
    STATUS=1
  fi
}

check_adapter_rule() {
  local rule="$1"
  local rel pattern path

  rel="${rule%%:*}"
  pattern="${rule#*:}"
  path="$ROOT/$rel"

  [ -e "$path" ] || return 0

  if rg -n --glob '!**/node_modules/**' --glob '!**/.git/**' "$pattern" "$path"; then
    printf 'Error: adapter contiene referencia prohibida por su contrato: %s\n' "$rel" >&2
    STATUS=1
  fi
}

check_html_overview_contract() {
  local rel="$1"
  local path="$ROOT/$rel"
  local required
  local missing=0

  [ -f "$path" ] || return 0

  for required in "offered_pending" "skipped_by_user" "$HTML_OVERVIEW_COMMAND" "$HTML_OVERVIEW_CTA"; do
    if ! rg -F -q -- "$required" "$path"; then
      printf 'Error: contrato HTML overview incompleto en %s; falta: %s\n' "$rel" "$required" >&2
      missing=1
    fi
  done

  if rg -n 'HTML overview opcional: skipped|record `skipped`|report `skipped`|reports `skipped`' "$path"; then
    printf 'Error: contrato HTML overview usa skipped ambiguo en %s; usar skipped_by_user o legacy explícito\n' "$rel" >&2
    missing=1
  fi

  if [ "$missing" -ne 0 ]; then
    STATUS=1
  fi
}

print_current_inventory() {
  local scope="$1"

  if [ -e "$ROOT/$scope" ]; then
    local count
    count="$(rg -n --glob '!**/node_modules/**' "$PATTERN" "$ROOT/$scope" | wc -l | tr -d ' ')"
    printf 'info: %s referencias actuales=%s\n' "$scope" "$count"
  fi
}

for rel in "${NEUTRAL_PATHS[@]}"; do
  check_neutral_path "$rel"
done

for rule in "${ADAPTER_RULES[@]}"; do
  check_adapter_rule "$rule"
done

for rel in "${HTML_OVERVIEW_CONTRACT_PATHS[@]}"; do
  check_html_overview_contract "$rel"
done

print_current_inventory ".opencode/agents"
print_current_inventory ".opencode/skills"
print_current_inventory ".opencode/commands"
print_current_inventory ".opencode/templates"
print_current_inventory ".opencode/policies"
print_current_inventory "AGENTS.md.template"
print_current_inventory "lufy-ia.harness.md"

exit "$STATUS"
