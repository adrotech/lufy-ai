#!/usr/bin/env bash
# Smoke del wrapper estricto scripts/install.sh usando el binario Go local.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CLI_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
REPO_ROOT="$(cd "$CLI_ROOT/../.." && pwd)"
BIN="$CLI_ROOT/bin/lufy-ai"
work_root=""

fail() {
    printf 'Error: %s\n' "$1" >&2
    exit 1
}

assert_empty_dir() {
    local dir="$1"
    shopt -s nullglob dotglob
    local entries=("$dir"/*)
    shopt -u nullglob dotglob
    if [ "${#entries[@]}" -ne 0 ]; then
        fail "dry-run del wrapper escribió contenido inesperado en $dir"
    fi
}

cleanup() {
    rm -rf "$work_root"
}

main() {
    if [ ! -x "$BIN" ]; then
        mkdir -p "$(dirname "$BIN")"
        (cd "$CLI_ROOT" && go build -o "$BIN" ./cmd/lufy-ai)
    fi

    local dry_target target
    work_root="$(mktemp -d)"
    trap cleanup EXIT

    dry_target="$work_root/wrapper-dry-run"
    mkdir -p "$dry_target"
    "$REPO_ROOT/scripts/install.sh" "$dry_target" --dry-run --yes
    assert_empty_dir "$dry_target"

    target="$work_root/wrapper-install"
    mkdir -p "$target"
    "$REPO_ROOT/scripts/install.sh" --target "$target" --yes
    "$BIN" verify --target "$target"

    printf 'Smoke wrapper completado\n'
}

main "$@"
