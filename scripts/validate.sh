#!/usr/bin/env bash
# Validación agrupada local para cambios destinados a PR.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
CLI_ROOT="$REPO_ROOT/tools/lufy-cli-go"
BASE_REF="${LUFY_AI_VALIDATE_BASE:-develop}"

log() {
  printf '==> %s\n' "$1"
}

ensure_base_ref() {
  if git -C "$REPO_ROOT" rev-parse --verify --quiet "origin/${BASE_REF}" >/dev/null; then
    return 0
  fi
  if git -C "$REPO_ROOT" remote get-url origin >/dev/null 2>&1; then
    git -C "$REPO_ROOT" fetch --no-tags --prune origin "$BASE_REF"
  fi
}

whitespace_check() {
  ensure_base_ref

  if git -C "$REPO_ROOT" rev-parse --verify --quiet "origin/${BASE_REF}" >/dev/null; then
    if [ -n "$(git -C "$REPO_ROOT" status --porcelain)" ]; then
      log "Whitespace contra origin/${BASE_REF} incluyendo worktree"
      git -C "$REPO_ROOT" diff --check "origin/${BASE_REF}"
      return 0
    fi

    log "Whitespace del rango PR origin/${BASE_REF}...HEAD"
    git -C "$REPO_ROOT" diff --check "origin/${BASE_REF}...HEAD"
    return 0
  fi

  log "Whitespace local sin origin/${BASE_REF} disponible"
  git -C "$REPO_ROOT" diff --check
}

main() {
  whitespace_check

  log "Action pinning"
  "$REPO_ROOT/scripts/check-actions-pinned.sh"

  log "Doc release version"
  "$REPO_ROOT/scripts/check-doc-release-version.sh"

  log "Workflow YAML"
  (cd "$CLI_ROOT" && go run ./cmd/check-workflows-yaml --root "$REPO_ROOT")

  log "Harness coupling"
  "$REPO_ROOT/scripts/check-harness-coupling.sh"

  log "Shell lint"
  "$REPO_ROOT/scripts/check-shell.sh"

  log "Format dispatch hook smoke"
  "$REPO_ROOT/tools/lufy-cli-go/scripts/smoke-format-dispatch.sh"

  log "Go quality"
  "$REPO_ROOT/scripts/quality-go.sh"

  log "Go build"
  (cd "$CLI_ROOT" && go build ./cmd/lufy-ai)
}

main "$@"
