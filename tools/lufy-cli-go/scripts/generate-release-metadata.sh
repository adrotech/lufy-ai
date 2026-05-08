#!/usr/bin/env bash
# Genera SBOM SPDX JSON y provenance in-toto/SLSA para artifacts de release.

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DIST="${LUFY_AI_DIST_DIR:-$ROOT/dist}"
VERSION="${LUFY_AI_VERSION:-${1:-dev}}"
COMMIT="${LUFY_AI_COMMIT:-$(git -C "$ROOT" rev-parse --short=12 HEAD 2>/dev/null || printf 'unknown')}"
BUILD_DATE="${LUFY_AI_BUILD_DATE:-$(date -u +%Y-%m-%dT%H:%M:%SZ)}"
CHECKSUM_FILE="$DIST/lufy-ai_${VERSION}_checksums.txt"
SBOM_FILE="$DIST/lufy-ai_${VERSION}_sbom.spdx.json"
PROVENANCE_FILE="$DIST/lufy-ai_${VERSION}_provenance.intoto.jsonl"

fail() {
  printf 'Error: %s\n' "$1" >&2
  exit 1
}

json_escape() {
  local value="$1"
  value="${value//\\/\\\\}"
  value="${value//\"/\\\"}"
  value="${value//$'\n'/ }"
  value="${value//$'\r'/ }"
  printf '%s' "$value"
}

[ -f "$CHECKSUM_FILE" ] || fail "no existe $CHECKSUM_FILE"

generate_sbom() {
  local first="true" path version name
  {
    printf '{\n'
    printf '  "spdxVersion": "SPDX-2.3",\n'
    printf '  "dataLicense": "CC0-1.0",\n'
    printf '  "SPDXID": "SPDXRef-DOCUMENT",\n'
    printf '  "name": "lufy-ai-%s",\n' "$(json_escape "$VERSION")"
    printf '  "documentNamespace": "https://github.com/adrotech/lufy-ai/releases/%s/sbom",\n' "$(json_escape "$VERSION")"
    printf '  "creationInfo": {"created": "%s", "creators": ["Tool: lufy-ai-release-metadata"]},\n' "$(json_escape "$BUILD_DATE")"
    printf '  "packages": [\n'
    printf '    {"name": "lufy-ai", "SPDXID": "SPDXRef-Package-lufy-ai", "downloadLocation": "NOASSERTION", "filesAnalyzed": false, "versionInfo": "%s"}' "$(json_escape "$VERSION")"
    while read -r path version _; do
      [ -n "$path" ] || continue
      name="$(json_escape "$path")"
      printf ',\n    {"name": "%s", "SPDXID": "SPDXRef-Package-%s", "downloadLocation": "NOASSERTION", "filesAnalyzed": false, "versionInfo": "%s"}' "$name" "$(printf '%s' "$path" | tr -c 'A-Za-z0-9.' '-')" "$(json_escape "${version:-unknown}")"
    done < <(cd "$ROOT" && go list -m -f '{{.Path}} {{.Version}}' all)
    printf '\n  ]\n}\n'
  } > "$SBOM_FILE"
}

generate_provenance() {
  local first="true" hash file
  {
    printf '{"_type":"https://in-toto.io/Statement/v1","predicateType":"https://slsa.dev/provenance/v1","subject":['
    while read -r hash file; do
      [ -n "$hash" ] || continue
      if [ "$first" = "true" ]; then
        first="false"
      else
        printf ','
      fi
      printf '{"name":"%s","digest":{"sha256":"%s"}}' "$(json_escape "$file")" "$(json_escape "$hash")"
    done < "$CHECKSUM_FILE"
    printf '],"predicate":{"buildDefinition":{"buildType":"https://github.com/adrotech/lufy-ai/.github/workflows/release.yml","externalParameters":{"version":"%s"},"internalParameters":{"commit":"%s"}},"runDetails":{"builder":{"id":"https://github.com/adrotech/lufy-ai/actions"},"metadata":{"invocationId":"%s","startedOn":"%s"}}}}\n' "$(json_escape "$VERSION")" "$(json_escape "$COMMIT")" "$(json_escape "${GITHUB_RUN_ID:-local}")" "$(json_escape "$BUILD_DATE")"
  } > "$PROVENANCE_FILE"
}

generate_sbom
generate_provenance
printf 'Metadata de release escrita: %s %s\n' "$SBOM_FILE" "$PROVENANCE_FILE"
