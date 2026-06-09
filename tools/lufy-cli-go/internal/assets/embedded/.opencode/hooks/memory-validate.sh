#!/usr/bin/env bash
set -euo pipefail

ROOT="${LUFY_PROJECT_ROOT:-$(pwd)}"
FILE="${1:-}"

if [ -n "$FILE" ] && [[ "$FILE" != *".lufy/memory/"* ]]; then
  exit 0
fi

if ! command -v lufy-ai >/dev/null 2>&1; then
  exit 0
fi

if [ ! -d "$ROOT/.lufy/memory" ]; then
  exit 0
fi

lufy-ai memory validate --target "$ROOT"
