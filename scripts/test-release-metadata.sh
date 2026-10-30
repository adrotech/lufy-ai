#!/usr/bin/env bash
# El encabezado de checksums no es un subject del provenance.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GENERATOR="$ROOT/tools/lufy-cli-go/scripts/generate-release-metadata.sh"
FIXTURE="$(mktemp -d)"
trap 'rm -rf "$FIXTURE"' EXIT

mkdir -p "$FIXTURE/bin" "$FIXTURE/dist"
cat > "$FIXTURE/bin/go" <<'GO'
#!/usr/bin/env bash
set -euo pipefail
if [[ "${1:-} ${2:-}" != "list -m" ]]; then
  exit 2
fi
printf 'example.invalid/fixture v0.0.0\n'
GO
chmod +x "$FIXTURE/bin/go"

cat > "$FIXTURE/dist/lufy-ai_v0.7.0_checksums.txt" <<'CHECKSUMS'
# lufy-runtime: execution=inline config=2 state=4 backup=2
# Comentario de fixture: tampoco representa un artifact.
aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa  lufy-ai_v0.7.0_linux_amd64.tar.gz
bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb  lufy-ai_v0.7.0_windows_arm64.zip
CHECKSUMS

PATH="$FIXTURE/bin:$PATH" \
  LUFY_AI_DIST_DIR="$FIXTURE/dist" \
  LUFY_AI_VERSION="v0.7.0" \
  LUFY_AI_COMMIT="fixture" \
  LUFY_AI_BUILD_DATE="1970-01-01T00:00:00Z" \
  bash "$GENERATOR" v0.7.0

python3 - "$FIXTURE/dist" <<'PY'
import json
import re
import sys
from pathlib import Path
root = Path(sys.argv[1])
provenance = json.loads((root / 'lufy-ai_v0.7.0_provenance.intoto.jsonl').read_text())
subjects = provenance['subject']
expected = {
    'lufy-ai_v0.7.0_linux_amd64.tar.gz': 'a' * 64,
    'lufy-ai_v0.7.0_windows_arm64.zip': 'b' * 64,
}
observed = {item['name']: item['digest']['sha256'] for item in subjects}
assert len(subjects) == len(expected), f'Subjects inesperados: {observed}'
assert observed == expected, observed
assert all(re.fullmatch('[0-9a-f]{64}', value) for value in observed.values())
assert provenance['predicate']['buildDefinition']['externalParameters']['version'] == 'v0.7.0'
sbom = json.loads((root / 'lufy-ai_v0.7.0_sbom.spdx.json').read_text())
assert sbom['packages'][0]['versionInfo'] == 'v0.7.0'
assert any(item['name'] == 'example.invalid/fixture' for item in sbom['packages'])
print('Release metadata: encabezados excluidos, subjects/digests y SBOM correctos')
PY
