#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DIST="$(mktemp -d)"
TARGET="$(mktemp -d)"
trap 'rm -rf "$DIST" "$TARGET"' EXIT

cd "$ROOT"
LUFY_AI_DIST_DIR="$DIST" LUFY_AI_VERSION="v0.0.0-smoke" LUFY_AI_COMMIT="smoke" LUFY_AI_BUILD_DATE="1970-01-01T00:00:00Z" \
  bash scripts/build-release-artifacts.sh v0.0.0-smoke
LUFY_AI_DIST_DIR="$DIST" LUFY_AI_VERSION="v0.0.0-smoke" LUFY_AI_COMMIT="smoke" LUFY_AI_BUILD_DATE="1970-01-01T00:00:00Z" \
  bash scripts/generate-release-metadata.sh v0.0.0-smoke

(cd "$DIST" && shasum -a 256 -c lufy-ai_v0.0.0-smoke_checksums.txt)
[ -s "$DIST/lufy-ai_v0.0.0-smoke_sbom.spdx.json" ] || { echo "SBOM faltante" >&2; exit 1; }
[ -s "$DIST/lufy-ai_v0.0.0-smoke_provenance.intoto.jsonl" ] || { echo "provenance faltante" >&2; exit 1; }
if [[ "${LUFY_AI_REQUIRE_SIGNATURES:-false}" == "true" ]]; then
  shopt -s nullglob
  bundles=("$DIST"/*.bundle)
  [ "${#bundles[@]}" -gt 0 ] || { echo "firmas cosign faltantes" >&2; exit 1; }
fi

artifact="$DIST/lufy-ai_v0.0.0-smoke_$(go env GOOS)_$(go env GOARCH).tar.gz"
if [[ ! -f "$artifact" ]]; then
  echo "Smoke omitido: artifact no ejecutable para $(go env GOOS)/$(go env GOARCH)" >&2
  exit 0
fi

extract="$(mktemp -d)"
trap 'rm -rf "$DIST" "$TARGET" "$extract"' EXIT
tar -C "$extract" -xzf "$artifact"
bin="$extract/lufy-ai_v0.0.0-smoke_$(go env GOOS)_$(go env GOARCH)/lufy-ai"
"$bin" version | grep -q "v0.0.0-smoke"

outside="$(mktemp -d)"
trap 'rm -rf "$DIST" "$TARGET" "$extract" "$outside"' EXIT
(cd "$outside" && "$bin" install --target "$TARGET" --dry-run --yes --no-engram)
(cd "$outside" && "$bin" install --target "$TARGET" --yes --no-engram)
(cd "$outside" && "$bin" verify --target "$TARGET" --no-engram)
