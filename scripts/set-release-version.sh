#!/usr/bin/env bash
# Actualiza de forma coordinada la version canonica y sus referencias copiables.

set -euo pipefail

ROOT="${LUFY_AI_RELEASE_ROOT:-$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)}"
VERSION_FILE="$ROOT/RELEASE_VERSION"
next="${1:-}"

if [[ ! "$next" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "Error: debes indicar una version estable vMAJOR.MINOR.PATCH" >&2
  exit 2
fi
if [ ! -f "$VERSION_FILE" ]; then
  echo "Error: falta RELEASE_VERSION" >&2
  exit 1
fi

current="$(tr -d '[:space:]' < "$VERSION_FILE")"
if [[ ! "$current" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "Error: RELEASE_VERSION invalido: $current" >&2
  exit 1
fi

versioned_files=(
  "README.md"
  "docs/installation.md"
  "tools/lufy-cli-go/README.md"
)

current_temp=""
cleanup() {
  if [ -n "$current_temp" ] && [ -f "$current_temp" ]; then
    rm -f "$current_temp"
  fi
}
trap cleanup EXIT

for rel in "${versioned_files[@]}"; do
  file="$ROOT/$rel"
  if [ ! -f "$file" ]; then
    echo "Error: falta archivo versionado: $rel" >&2
    exit 1
  fi
  if ! grep -Fq "$current" "$file"; then
    echo "Error: $rel no contiene la version canonica actual $current" >&2
    exit 1
  fi
done

for rel in "${versioned_files[@]}"; do
  file="$ROOT/$rel"
  current_temp="${file}.lufy-version.$$"
  cp -p "$file" "$current_temp"
  : > "$current_temp"
  while IFS= read -r line || [ -n "$line" ]; do
    printf '%s\n' "${line//$current/$next}" >> "$current_temp"
  done < "$file"
  mv "$current_temp" "$file"
  current_temp=""
done

current_temp="${VERSION_FILE}.lufy-version.$$"
printf '%s\n' "$next" > "$current_temp"
mv "$current_temp" "$VERSION_FILE"
current_temp=""

echo "Version de release actualizada: $current -> $next"
echo "Accion requerida: agrega el heading ## [$next] a CHANGELOG.md antes de promover a main."
