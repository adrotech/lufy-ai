#!/usr/bin/env bash
# ShellCheck para scripts shell versionados.

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

if ! command -v shellcheck >/dev/null 2>&1; then
  echo "::notice::shellcheck no disponible; shell lint omitido localmente"
  exit 0
fi

shellcheck "$ROOT"/scripts/*.sh "$ROOT"/tools/lufy-cli-go/scripts/*.sh
