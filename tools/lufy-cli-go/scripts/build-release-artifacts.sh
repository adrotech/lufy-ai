#!/usr/bin/env bash
set -euo pipefail

VERSION="${LUFY_AI_VERSION:-${1:-dev}}"
COMMIT="${LUFY_AI_COMMIT:-$(git rev-parse --short=12 HEAD 2>/dev/null || printf 'unknown')}"
BUILD_DATE="${LUFY_AI_BUILD_DATE:-$(date -u +%Y-%m-%dT%H:%M:%SZ)}"
OUT_DIR="${LUFY_AI_DIST_DIR:-dist}"
PKG="github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/version"

mkdir -p "$OUT_DIR"
rm -f "$OUT_DIR"/lufy-ai_* "$OUT_DIR"/checksums.txt "$OUT_DIR"/*_checksums.txt

platforms=(
  darwin/amd64
  darwin/arm64
  linux/amd64
  linux/arm64
  windows/amd64
)

checksum_file="$OUT_DIR/lufy-ai_${VERSION}_checksums.txt"
: > "$checksum_file"

for platform in "${platforms[@]}"; do
  goos="${platform%/*}"
  goarch="${platform#*/}"
  name="lufy-ai_${VERSION}_${goos}_${goarch}"
  work="$OUT_DIR/$name"
  mkdir -p "$work"
  bin="$work/lufy-ai"
  archive="$OUT_DIR/${name}.tar.gz"
  if [[ "$goos" == "windows" ]]; then
    bin="$work/lufy-ai.exe"
    archive="$OUT_DIR/${name}.zip"
  fi

  GOOS="$goos" GOARCH="$goarch" CGO_ENABLED=0 go build \
    -trimpath \
    -ldflags "-s -w -X ${PKG}.Version=${VERSION} -X ${PKG}.Commit=${COMMIT} -X ${PKG}.BuildDate=${BUILD_DATE}" \
    -o "$bin" ./cmd/lufy-ai

  cp README.md "$work/README.cli.md"
  if [[ "$goos" == "windows" ]]; then
    (cd "$OUT_DIR" && zip -qr "$(basename "$archive")" "$(basename "$work")")
  else
    tar -C "$OUT_DIR" -czf "$archive" "$(basename "$work")"
  fi
  rm -rf "$work"
  hash="$(shasum -a 256 "$archive" | awk '{print $1}')"
  printf '%s  %s\n' "$hash" "$(basename "$archive")" >> "$checksum_file"
done

(cd "$OUT_DIR" && shasum -a 256 -c "$(basename "$checksum_file")")
printf 'Artifacts escritos en %s\n' "$OUT_DIR"
