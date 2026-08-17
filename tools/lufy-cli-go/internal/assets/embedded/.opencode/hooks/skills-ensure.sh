#!/usr/bin/env bash
set -u

ROOT="${LUFY_PROJECT_ROOT:-$(pwd)}"

if ! command -v lufy-ai >/dev/null 2>&1; then
  exit 0
fi

if [ ! -f "$ROOT/.lufy/config/project.yaml" ]; then
  exit 0
fi

lufy-ai skills ensure --target "$ROOT" --tool opencode >/dev/null 2>&1 || true
