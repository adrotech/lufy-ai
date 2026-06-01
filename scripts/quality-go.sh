#!/usr/bin/env bash
# Quality gate Go: tests con coverage y go vet para tools/lufy-cli-go.

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CLI_ROOT="$ROOT/tools/lufy-cli-go"
COVERAGE_MIN="${LUFY_AI_COVERAGE_MIN:-80.0}"

cd "$CLI_ROOT"

coverage_file="$(mktemp)"
trap 'rm -f "$coverage_file"' EXIT

go test ./... -coverprofile="$coverage_file"
coverage_pct="$(go tool cover -func="$coverage_file" | awk '/^total:/ { gsub(/%/, "", $3); print $3 }')"
if [ -z "$coverage_pct" ]; then
  echo "coverage total not found" >&2
  exit 1
fi

awk -v actual="$coverage_pct" -v minimum="$COVERAGE_MIN" 'BEGIN {
  if ((actual + 0) < (minimum + 0)) {
    printf "coverage %.1f%% below threshold %.1f%%\n", actual, minimum > "/dev/stderr"
    exit 1
  }
  printf "coverage %.1f%% >= threshold %.1f%%\n", actual, minimum
}'

go vet ./...
