#!/usr/bin/env bash
# Fronteras del payload inline y paridad de la única fuente editorial.
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
EMBEDDED="$ROOT/tools/lufy-cli-go/internal/assets/embedded"
FILES=(
  "AGENTS.md.template"
  "lufy-ia.harness.md"
  ".lufy/README.md"
  ".lufy/contracts/README.md"
  ".lufy/contracts/work.md"
  ".lufy/contracts/evidence.md"
  ".lufy/contracts/delivery.md"
  ".codex/README.md"
  ".codex/config.toml"
  ".codex/hooks.json"
  ".codex/lufy-agent-mapping.md"
  ".codex/rules/lufy.rules"
  ".agents/skills/lufy-check/SKILL.md"
  ".agents/skills/lufy-mem-capture/SKILL.md"
  ".agents/skills/lufy-mem-connect/SKILL.md"
  ".agents/skills/lufy-mem-document/SKILL.md"
  ".agents/skills/lufy-mem-search/SKILL.md"
)
for rel in "${FILES[@]}"; do
  if ! cmp -s "$ROOT/$rel" "$EMBEDDED/$rel"; then
    printf 'Error: fuente/mirror distintos o ausentes: %s\n' "$rel" >&2
    exit 1
  fi
done
while IFS= read -r file; do
  rel="${file#"$EMBEDDED/"}"
  found=false
  for expected in "${FILES[@]}"; do
    if [ "$rel" = "$expected" ]; then found=true; break; fi
  done
  if [ "$found" != true ]; then
    printf 'Error: payload sin capacidad seleccionada: %s\n' "$rel" >&2
    exit 1
  fi
done < <(find "$EMBEDDED" -type f)
if rg -n 'type (Methodology|WorkflowModel)|case "(sdd|opsx)"|internal/(lufysdd|opsx|adapters/methodology)' \
  "$ROOT/tools/lufy-cli-go/internal" --glob '*.go' --glob '!**/*_test.go' --glob '!**/embedded/**'; then
  printf '%s\n' 'Error: el runtime conserva un motor o selector metodológico retirado.' >&2
  exit 1
fi
printf '%s\n' 'Payload inline y paridad fuente/embedded: OK'
