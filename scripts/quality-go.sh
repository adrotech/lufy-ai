#!/usr/bin/env bash
# Quality gate Go: tests con coverage y go vet para tools/lufy-cli-go.

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CLI_ROOT="$ROOT/tools/lufy-cli-go"
COVERAGE_MIN="${LUFY_AI_COVERAGE_MIN:-80.0}"

cd "$CLI_ROOT"

coverage_file="$(mktemp)"
trap 'rm -f "$coverage_file"' EXIT

go test ./... -coverpkg=./... -coverprofile="$coverage_file"
coverage_pct="$(go tool cover -func="$coverage_file" | python3 -c 'import re, sys
for line in sys.stdin:
    if line.startswith("total:"):
        match = re.search(r"([0-9]+(?:\.[0-9]+)?)%", line)
        if match:
            print(match.group(1))
            break
else:
    sys.exit("coverage total not found")')"

python3 - "$coverage_pct" "$COVERAGE_MIN" <<'PY'
import sys

actual = float(sys.argv[1])
minimum = float(sys.argv[2])
if actual < minimum:
    print(f"coverage {actual:.1f}% below threshold {minimum:.1f}%", file=sys.stderr)
    sys.exit(1)
print(f"coverage {actual:.1f}% >= threshold {minimum:.1f}%")
PY

go vet ./...
