#!/usr/bin/env bash
# Verifica que el artifact nativo construido reporte exactamente el tag solicitado.

set -euo pipefail

TAG="${1:-}"
DIST="${2:-dist}"

if [[ ! "$TAG" =~ ^v[0-9]+\.[0-9]+\.[0-9]+([-.][0-9A-Za-z.-]+)?$ ]]; then
  echo "Error: debes indicar un tag v* valido" >&2
  exit 2
fi

os="$(go env GOOS)"
arch="$(go env GOARCH)"
case "$os" in
  darwin|linux) archive="$DIST/lufy-ai_${TAG}_${os}_${arch}.tar.gz" ;;
  windows) archive="$DIST/lufy-ai_${TAG}_${os}_${arch}.zip" ;;
  *) echo "Error: GOOS no soportado: $os" >&2; exit 1 ;;
esac

if [ ! -f "$archive" ]; then
  echo "Error: artifact nativo faltante: $archive" >&2
  exit 1
fi

work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
case "$archive" in
  *.tar.gz)
    tar -C "$work" -xzf "$archive"
    bin="$work/lufy-ai_${TAG}_${os}_${arch}/lufy-ai"
    ;;
  *.zip)
    unzip -q "$archive" -d "$work"
    bin="$work/lufy-ai_${TAG}_${os}_${arch}/lufy-ai.exe"
    ;;
esac

if [ ! -x "$bin" ]; then
  echo "Error: binario nativo faltante o no ejecutable: $bin" >&2
  exit 1
fi
output="$("$bin" version)"
if [ "${output%%$'\n'*}" != "lufy-ai $TAG" ]; then
  echo "Error: artifact reporta una version distinta de $TAG: $output" >&2
  exit 1
fi

echo "artifact release version ok: $TAG"
