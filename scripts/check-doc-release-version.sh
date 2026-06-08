#!/usr/bin/env bash
# Verifica que las guias copiables de instalacion usen la version estable canonica.

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
VERSION_FILE="$ROOT/RELEASE_VERSION"

if [ ! -f "$VERSION_FILE" ]; then
  echo "Error: falta RELEASE_VERSION" >&2
  exit 1
fi

expected="$(tr -d '[:space:]' < "$VERSION_FILE")"
if [[ ! "$expected" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "Error: RELEASE_VERSION invalido: $expected" >&2
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

  while IFS= read -r version; do
    if [ "$version" != "$expected" ]; then
      echo "Error: $rel referencia $version pero RELEASE_VERSION es $expected" >&2
      status=1
    fi
  done < <(grep -Eo 'v[0-9]+\.[0-9]+\.[0-9]+' "$file" | sort -u)
done

if [ "$status" -eq 0 ]; then
  echo "doc release version ok: $expected"
fi

exit "$status"
