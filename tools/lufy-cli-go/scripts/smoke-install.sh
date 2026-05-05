#!/usr/bin/env bash
# Smoke reproducible de la CLI Go para install/verify/idempotencia/backup/restore.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CLI_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
BIN="${LUFY_AI_BIN:-$CLI_ROOT/bin/lufy-ai}"

log() {
    printf '==> %s\n' "$1"
}

fail() {
    printf 'Error: %s\n' "$1" >&2
    exit 1
}

ensure_bin() {
    if [ -x "$BIN" ]; then
        return 0
    fi

    log "Compilando binario local en $BIN"
    mkdir -p "$(dirname "$BIN")"
    (cd "$CLI_ROOT" && go build -o "$BIN" ./cmd/lufy-ai)
}

sha256_file() {
    if command -v sha256sum >/dev/null 2>&1; then
        sha256sum "$1" | cut -d ' ' -f 1
        return 0
    fi
    if command -v shasum >/dev/null 2>&1; then
        shasum -a 256 "$1" | cut -d ' ' -f 1
        return 0
    fi
    fail "no se encontró sha256sum ni shasum para validar idempotencia"
}

assert_empty_dir() {
    local dir="$1"
    shopt -s nullglob dotglob
    local entries=("$dir"/*)
    shopt -u nullglob dotglob
    if [ "${#entries[@]}" -ne 0 ]; then
        fail "dry-run escribió contenido inesperado en $dir"
    fi
}

expect_failure_contains() {
    local expected="$1"
    shift
    local output=""
    local status=0

    set +e
    output="$("$@" 2>&1)"
    status=$?
    set -e

    if [ "$status" -eq 0 ]; then
        printf '%s\n' "$output" >&2
        fail "el comando debía fallar: $*"
    fi
    if [[ "$output" != *"$expected"* ]]; then
        printf '%s\n' "$output" >&2
        fail "el error no contiene '$expected'"
    fi
}

main() {
    ensure_bin

    local work_root
    work_root="$(mktemp -d)"
    trap "rm -rf '$work_root'" EXIT

    local dry_target="$work_root/dry-run-target"
    mkdir -p "$dry_target"
    log "Dry-run install sin mutaciones"
    "$BIN" install --target "$dry_target" --dry-run --yes --no-engram
    assert_empty_dir "$dry_target"

    local confirm_target="$work_root/install-needs-yes"
    mkdir -p "$confirm_target"
    log "Install real sin --yes falla de forma accionable"
    expect_failure_contains "install requiere --yes" "$BIN" install --target "$confirm_target" --no-engram
    assert_empty_dir "$confirm_target"

    local target="$work_root/install-target"
    mkdir -p "$target"
    log "Install real"
    "$BIN" install --target "$target" --yes --no-engram

    log "Verify posterior a install"
    "$BIN" verify --target "$target" --no-engram

    local state_file="$target/.lufy-ai/install-state.json"
    local asset_file="$target/AGENTS.md"
    [ -f "$state_file" ] || fail "no existe $state_file"
    [ -f "$asset_file" ] || fail "no existe $asset_file"
    local state_before asset_before state_after asset_after
    state_before="$(sha256_file "$state_file")"
    asset_before="$(sha256_file "$asset_file")"

    log "Segundo install idempotente"
    "$BIN" install --target "$target" --yes --no-engram
    state_after="$(sha256_file "$state_file")"
    asset_after="$(sha256_file "$asset_file")"
    [ "$state_before" = "$state_after" ] || fail "install idempotente reescribió install-state.json"
    [ "$asset_before" = "$asset_after" ] || fail "install idempotente cambió AGENTS.md"

    log "Backup"
    local backup_output backup_dir
    backup_output="$($BIN backup --target "$target")"
    printf '%s\n' "$backup_output"
    if [[ "$backup_output" =~ Backup\ creado:\ ([^[:space:]]+) ]]; then
        backup_dir="${BASH_REMATCH[1]}"
    else
        fail "no se pudo detectar el directorio de backup"
    fi
    [ -f "$backup_dir/manifest.json" ] || fail "backup sin manifest.json"

    log "Restore sin --yes falla de forma accionable"
    expect_failure_contains "restore requiere --yes" "$BIN" restore --target "$target" --backup "$backup_dir"

    log "Restore dry-run"
    "$BIN" restore --target "$target" --backup "$backup_dir" --dry-run --yes

    log "Restore real"
    "$BIN" restore --target "$target" --backup "$backup_dir" --yes

    log "Smoke CLI completado"
}

main "$@"
