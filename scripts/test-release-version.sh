#!/usr/bin/env bash
# Pruebas aisladas para setter y checker de version de release.

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CHECK="$ROOT/scripts/check-doc-release-version.sh"
SET="$ROOT/scripts/set-release-version.sh"
VERIFY_ARTIFACT="$ROOT/tools/lufy-cli-go/scripts/verify-release-artifact-version.sh"
FIXTURE="$(mktemp -d)"
trap 'rm -rf "$FIXTURE"' EXIT

mkdir -p "$FIXTURE/docs" "$FIXTURE/tools/lufy-cli-go" "$FIXTURE/bin"
printf 'v1.2.3\n' > "$FIXTURE/RELEASE_VERSION"
printf 'Instalar v1.2.3\n' > "$FIXTURE/README.md"
printf 'Instalar v1.2.3\n' > "$FIXTURE/docs/installation.md"
printf 'Instalar v1.2.3\n' > "$FIXTURE/tools/lufy-cli-go/README.md"
printf '# Changelog\n\n## [v1.2.4] - 2026-08-17\n' > "$FIXTURE/CHANGELOG.md"
cat > "$FIXTURE/bin/lufy-ai" <<'EOF'
#!/usr/bin/env bash
printf 'lufy-ai v1.2.4\ncommit: test\nbuildDate: fixture\n'
EOF
chmod +x "$FIXTURE/bin/lufy-ai"

LUFY_AI_RELEASE_ROOT="$FIXTURE" "$CHECK"
LUFY_AI_RELEASE_ROOT="$FIXTURE" "$SET" v1.2.4
LUFY_AI_RELEASE_ROOT="$FIXTURE" "$CHECK" --expected v1.2.4 --require-changelog --binary "$FIXTURE/bin/lufy-ai"

if LUFY_AI_RELEASE_ROOT="$FIXTURE" "$CHECK" --expected v1.2.5 >/dev/null 2>&1; then
  echo "Error: el checker acepto una version esperada divergente" >&2
  exit 1
fi

printf '\nReferencia stale v9.9.9\n' >> "$FIXTURE/README.md"
if LUFY_AI_RELEASE_ROOT="$FIXTURE" "$CHECK" >/dev/null 2>&1; then
  echo "Error: el checker acepto documentacion stale" >&2
  exit 1
fi
sed -i.bak '/v9\.9\.9/d' "$FIXTURE/README.md"
rm -f "$FIXTURE/README.md.bak"

printf 'Sin version\n' > "$FIXTURE/docs/installation.md"
if LUFY_AI_RELEASE_ROOT="$FIXTURE" "$CHECK" >/dev/null 2>&1; then
  echo "Error: el checker acepto un documento sin version canonica" >&2
  exit 1
fi
printf 'Instalar v1.2.4\n' > "$FIXTURE/docs/installation.md"

printf '# Changelog\n' > "$FIXTURE/CHANGELOG.md"
if LUFY_AI_RELEASE_ROOT="$FIXTURE" "$CHECK" --require-changelog >/dev/null 2>&1; then
  echo "Error: el checker acepto changelog sin entrada canonica" >&2
  exit 1
fi
printf '# Changelog\n\n## [v1.2.4] - 2026-08-17\n' > "$FIXTURE/CHANGELOG.md"

cat > "$FIXTURE/bin/lufy-ai" <<'EOF'
#!/usr/bin/env bash
printf 'lufy-ai v1.2.40\ncommit: test\nbuildDate: fixture\n'
EOF
chmod +x "$FIXTURE/bin/lufy-ai"
if LUFY_AI_RELEASE_ROOT="$FIXTURE" "$CHECK" --binary "$FIXTURE/bin/lufy-ai" >/dev/null 2>&1; then
  echo "Error: el checker acepto un binario con version divergente" >&2
  exit 1
fi

mkdir -p \
  "$FIXTURE/fake-bin" \
  "$FIXTURE/dist/lufy-ai_v1.2.4_linux_amd64"
cat > "$FIXTURE/fake-bin/go" <<'EOF'
#!/usr/bin/env bash
case "${1:-} ${2:-}" in
  "env GOOS") printf 'linux\n' ;;
  "env GOARCH") printf 'amd64\n' ;;
  *) exit 2 ;;
esac
EOF
cat > "$FIXTURE/dist/lufy-ai_v1.2.4_linux_amd64/lufy-ai" <<'EOF'
#!/usr/bin/env bash
printf 'lufy-ai v1.2.4\ncommit: test\nbuildDate: fixture\n'
EOF
chmod +x \
  "$FIXTURE/fake-bin/go" \
  "$FIXTURE/dist/lufy-ai_v1.2.4_linux_amd64/lufy-ai"
tar -C "$FIXTURE/dist" -czf \
  "$FIXTURE/dist/lufy-ai_v1.2.4_linux_amd64.tar.gz" \
  lufy-ai_v1.2.4_linux_amd64
PATH="$FIXTURE/fake-bin:$PATH" "$VERIFY_ARTIFACT" v1.2.4 "$FIXTURE/dist"

cat > "$FIXTURE/dist/lufy-ai_v1.2.4_linux_amd64/lufy-ai" <<'EOF'
#!/usr/bin/env bash
printf 'lufy-ai v1.2.40\ncommit: test\nbuildDate: fixture\n'
EOF
chmod +x "$FIXTURE/dist/lufy-ai_v1.2.4_linux_amd64/lufy-ai"
tar -C "$FIXTURE/dist" -czf \
  "$FIXTURE/dist/lufy-ai_v1.2.4_linux_amd64.tar.gz" \
  lufy-ai_v1.2.4_linux_amd64
if PATH="$FIXTURE/fake-bin:$PATH" "$VERIFY_ARTIFACT" v1.2.4 "$FIXTURE/dist" >/dev/null 2>&1; then
  echo "Error: el verifier acepto un artifact con version divergente" >&2
  exit 1
fi

echo "release version tests ok"
