#!/usr/bin/env bash
# Wrapper de compatibilidad para instalar lufy-ai mediante la CLI Go.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
LOCAL_BIN="$REPO_ROOT/tools/lufy-cli-go/bin/lufy-ai"

usage() {
    cat <<'EOF'
Uso: scripts/install.sh [target-project-dir] [--target <dir>] [--scope project|global|both] [--tool opencode] [--methodology-tier T3:none] [--dry-run] [--yes] [--no-engram] [--backup]

Este script solo delega en la CLI Go:
  lufy-ai install
EOF
}

find_lufy_ai() {
    if [ -x "$LOCAL_BIN" ]; then
        printf '%s\n' "$LOCAL_BIN"
        return 0
    fi

    if command -v lufy-ai >/dev/null 2>&1; then
        command -v lufy-ai
        return 0
    fi

    return 1
}

fail_missing_cli() {
    cat >&2 <<EOF
Error: no se encontró la CLI Go 'lufy-ai'.

Instala 'lufy-ai' en PATH o compila el binario local desde este checkout:
  cd tools/lufy-cli-go && mkdir -p bin && go build -o bin/lufy-ai ./cmd/lufy-ai

Luego vuelve a ejecutar scripts/install.sh.
EOF
    exit 1
}

main() {
    local target=""
    local has_target_flag="false"
    local args=()

    while [ "$#" -gt 0 ]; do
        case "$1" in
            -h|--help)
                usage
                exit 0
                ;;
            --target)
                if [ "$#" -lt 2 ] || [[ "$2" == --* ]]; then
                    echo "Error: falta valor para --target" >&2
                    exit 2
                fi
                has_target_flag="true"
                args+=("--target" "$2")
                shift 2
                ;;
            --scope|--tool|--methodology-tier)
                if [ "$#" -lt 2 ] || [[ "$2" == --* ]]; then
                    echo "Error: falta valor para $1" >&2
                    exit 2
                fi
                args+=("$1" "$2")
                shift 2
                ;;
            --dry-run|--yes|--no-engram|--backup)
                args+=("$1")
                shift
                ;;
            --*)
                args+=("$1")
                shift
                ;;
            *)
                if [ -n "$target" ]; then
                    echo "Error: solo se acepta un argumento posicional histórico como target" >&2
                    usage >&2
                    exit 2
                fi
                target="$1"
                shift
                ;;
        esac
    done

    if [ "$has_target_flag" = "false" ] && [ -n "$target" ]; then
        args=("--target" "$target" "${args[@]}")
    fi

    local cli
    cli="$(find_lufy_ai)" || fail_missing_cli

    exec "$cli" install "${args[@]}"
}

main "$@"
