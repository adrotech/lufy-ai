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

    # Un wrapper de esta generación rechaza un ejecutable previo antes de instalar.
    local stale_repo="$work_root/stale-repo"
    mkdir -p "$stale_repo/scripts" "$stale_repo/tools/lufy-cli-go/bin" "$stale_repo/target"
    cp "$REPO_ROOT/scripts/install.sh" "$stale_repo/scripts/install.sh"
    cat >"$stale_repo/tools/lufy-cli-go/bin/lufy-ai" <<'SH'
#!/usr/bin/env bash
if [ "$1" = version ]; then printf '%s\n' 'v0.6.25'; exit 0; fi
printf '%s\n' 'WRONG GENERATION EXECUTED' >"$2"
exit 0
SH
    chmod +x "$stale_repo/tools/lufy-cli-go/bin/lufy-ai"
    if "$stale_repo/scripts/install.sh" --target "$stale_repo/target" --yes >"$work_root/stale.log" 2>&1; then
        fail "wrapper aceptó CLI de generación anterior"
    fi
    assert_empty_dir "$stale_repo/target"
    if ! grep -q 'contrato inline actual' "$work_root/stale.log"; then
        fail "wrapper no diagnosticó generación incompatible"
    fi
    printf 'Smoke wrapper completado\n'
}

main "$@"
