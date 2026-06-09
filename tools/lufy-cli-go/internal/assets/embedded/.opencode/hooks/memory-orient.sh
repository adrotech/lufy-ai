#!/usr/bin/env bash
set -euo pipefail

ROOT="${LUFY_PROJECT_ROOT:-$(pwd)}"

if ! command -v lufy-ai >/dev/null 2>&1; then
  exit 0
fi

if [ ! -f "$ROOT/.lufy/project.yaml" ]; then
  exit 0
fi

lufy-ai memory status --target "$ROOT" >/dev/null 2>&1 || true
