#!/usr/bin/env bash
# E2E contra artifacts publicados en GitHub Releases para un tag v*.

set -euo pipefail

TAG="${LUFY_AI_E2E_TAG:-${1:-}}"
REPO="${LUFY_AI_REPO:-adrotech/lufy-ai}"

fail() {
  printf 'Error: %s\n' "$1" >&2
  exit 1
}

if [[ ! "$TAG" =~ ^v[0-9]+\.[0-9]+\.[0-9]+([-.][0-9A-Za-z.-]+)?$ ]]; then
  fail "debes indicar un tag v* válido"
fi
if ! command -v gh >/dev/null 2>&1; then
  fail "gh CLI no disponible para descargar release artifacts"
fi

os="$(go env GOOS)"
arch="$(go env GOARCH)"
case "$os" in
  darwin|linux) archive="lufy-ai_${TAG}_${os}_${arch}.tar.gz" ;;
  windows) archive="lufy-ai_${TAG}_${os}_${arch}.zip" ;;
  *) fail "GOOS no soportado para E2E: $os" ;;
esac

work="$(mktemp -d)"
target="$(mktemp -d)"
trap 'rm -rf "$work" "$target"' EXIT

gh release download "$TAG" --repo "$REPO" --dir "$work" --pattern "$archive" --pattern "lufy-ai_${TAG}_checksums.txt"
(cd "$work" && shasum -a 256 -c "lufy-ai_${TAG}_checksums.txt" --ignore-missing)

case "$archive" in
  *.tar.gz)
    tar -C "$work" -xzf "$work/$archive"
    bin="$work/${archive%.tar.gz}/lufy-ai"
    ;;
  *.zip)
    unzip -q "$work/$archive" -d "$work"
    bin="$work/${archive%.zip}/lufy-ai.exe"
    ;;
  *) fail "artifact no soportado: $archive" ;;
esac

"$bin" version
"$bin" install --target "$target" --dry-run --yes --no-engram
"$bin" install --target "$target" --yes --no-engram
"$bin" verify --target "$target" --no-engram
