#!/usr/bin/env bash
# Verifica consistencia entre version canonica, documentacion, changelog y binario.

set -euo pipefail

ROOT="${LUFY_AI_RELEASE_ROOT:-$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)}"
VERSION_FILE="$ROOT/RELEASE_VERSION"
expected_override=""
require_changelog="false"
binary=""

usage() {
  cat <<'EOF'
Uso: scripts/check-doc-release-version.sh [--expected <vX.Y.Z>] [--require-changelog] [--binary <path>]
EOF
}

while [ "$#" -gt 0 ]; do
  case "$1" in
    --expected)
      expected_override="${2:-}"
      shift 2
      ;;
    --require-changelog)
      require_changelog="true"
      shift
      ;;
    --binary)
      binary="${2:-}"
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Error: argumento desconocido: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
done

if [ ! -f "$VERSION_FILE" ]; then
  echo "Error: falta RELEASE_VERSION" >&2
  exit 1
fi

expected="$(tr -d '[:space:]' < "$VERSION_FILE")"
if [[ ! "$expected" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "Error: RELEASE_VERSION invalido: $expected" >&2
  exit 1
fi
if [ -n "$expected_override" ] && [ "$expected_override" != "$expected" ]; then
  echo "Error: version esperada $expected_override pero RELEASE_VERSION declara $expected" >&2
  exit 1
fi

status=0
check_files=(
  "README.md"
  "docs/installation.md"
  "tools/lufy-cli-go/README.md"
)

for rel in "${check_files[@]}"; do
  file="$ROOT/$rel"
  if [ ! -f "$file" ]; then
    echo "Error: falta archivo versionado: $rel" >&2
    status=1
    continue
  fi

  found="false"
  while IFS= read -r version; do
    found="true"
    if [ "$version" != "$expected" ]; then
      echo "Error: $rel referencia $version pero RELEASE_VERSION es $expected" >&2
      status=1
    fi
  done < <(grep -Eo 'v[0-9]+\.[0-9]+\.[0-9]+' "$file" | sort -u)

  if [ "$found" != "true" ]; then
    echo "Error: $rel no contiene la version canonica $expected" >&2
    status=1
  fi
done

if [ "$require_changelog" = "true" ]; then
  changelog="$ROOT/CHANGELOG.md"
  if [ ! -f "$changelog" ]; then
    echo "Error: falta CHANGELOG.md" >&2
    status=1
  else
    changelog_found="false"
    while IFS= read -r line; do
      if [[ "$line" == "## [$expected]"* ]]; then
        changelog_found="true"
        break
      fi
    done < "$changelog"
    if [ "$changelog_found" != "true" ]; then
      echo "Error: CHANGELOG.md no contiene heading para $expected" >&2
      status=1
    fi
  fi
fi

if [ -n "$binary" ]; then
  if [ ! -x "$binary" ]; then
    echo "Error: binario no ejecutable: $binary" >&2
    status=1
  else
    if ! binary_output="$("$binary" version 2>&1)"; then
      echo "Error: no se pudo consultar version de $binary: $binary_output" >&2
      status=1
    elif [ "${binary_output%%$'\n'*}" != "lufy-ai $expected" ]; then
      echo "Error: $binary no reporta exactamente la version canonica $expected: $binary_output" >&2
      status=1
    fi
  fi
fi

if [ "$status" -eq 0 ]; then
  echo "release version ok: $expected"
fi

exit "$status"
