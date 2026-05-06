#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
CLI_ROOT="$ROOT/tools/lufy-cli-go"
DIST="$(mktemp -d)"
FIXTURES="$(mktemp -d)"
BIN_DIR="$(mktemp -d)"
trap 'rm -rf "$DIST" "$FIXTURES" "$BIN_DIR"' EXIT

cd "$CLI_ROOT"
LUFY_AI_DIST_DIR="$DIST" LUFY_AI_VERSION="v0.0.0-smoke" LUFY_AI_COMMIT="smoke" LUFY_AI_BUILD_DATE="1970-01-01T00:00:00Z" \
  bash scripts/build-release-artifacts.sh v0.0.0-smoke >/dev/null
mkdir -p "$FIXTURES/v0.0.0-smoke"
cp "$DIST"/* "$FIXTURES/v0.0.0-smoke/"

bash "$ROOT/scripts/bootstrap.sh" --version v0.0.0-smoke --base-url "file://$FIXTURES" --install-dir "$BIN_DIR" --dry-run | grep -q "Dry-run"
bash "$ROOT/scripts/bootstrap.sh" --version v0.0.0-smoke --base-url "file://$FIXTURES" --install-dir "$BIN_DIR"
"$BIN_DIR/lufy-ai" version | grep -q "v0.0.0-smoke"

artifact="lufy-ai_v0.0.0-smoke_$(uname -s | tr '[:upper:]' '[:lower:]')_$(uname -m).tar.gz"
case "$(uname -m)" in
  x86_64) artifact="lufy-ai_v0.0.0-smoke_$(uname -s | tr '[:upper:]' '[:lower:]')_amd64.tar.gz" ;;
  arm64|aarch64) artifact="lufy-ai_v0.0.0-smoke_$(uname -s | tr '[:upper:]' '[:lower:]')_arm64.tar.gz" ;;
esac
if [[ "$(uname -s)" == "Darwin" ]]; then artifact="${artifact/linux/darwin}"; fi
printf '0000000000000000000000000000000000000000000000000000000000000000  %s\n' "$artifact" > "$FIXTURES/v0.0.0-smoke/lufy-ai_v0.0.0-smoke_checksums.txt"
if bash "$ROOT/scripts/bootstrap.sh" --version v0.0.0-smoke --base-url "file://$FIXTURES" --install-dir "$BIN_DIR/bad"; then
  echo "checksum mismatch no bloqueó instalación" >&2
  exit 1
fi

traversal_work="$FIXTURES/traversal-work"
mkdir -p "$traversal_work/safe"
printf 'no debe extraerse\n' > "$traversal_work/evil"
tar -P -C "$traversal_work/safe" -czf "$FIXTURES/v0.0.0-smoke/$artifact" ../evil
traversal_hash="$(shasum -a 256 "$FIXTURES/v0.0.0-smoke/$artifact" | awk '{print $1}')"
printf '%s  %s\n' "$traversal_hash" "$artifact" > "$FIXTURES/v0.0.0-smoke/lufy-ai_v0.0.0-smoke_checksums.txt"
if bash "$ROOT/scripts/bootstrap.sh" --version v0.0.0-smoke --base-url "file://$FIXTURES" --install-dir "$BIN_DIR/traversal" 2>"$FIXTURES/traversal.err"; then
  echo "artifact con path traversal no bloqueó instalación" >&2
  exit 1
fi
grep -q "path inseguro" "$FIXTURES/traversal.err"
if [[ -e "$BIN_DIR/traversal/lufy-ai" ]]; then
  echo "artifact con path traversal instaló binario inesperadamente" >&2
  exit 1
fi

symlink_work="$FIXTURES/symlink-work"
mkdir -p "$symlink_work/${artifact%.tar.gz}"
printf 'placeholder\n' > "$symlink_work/${artifact%.tar.gz}/README.cli.md"
ln -s README.cli.md "$symlink_work/${artifact%.tar.gz}/lufy-ai"
tar -C "$symlink_work" -czf "$FIXTURES/v0.0.0-smoke/$artifact" "${artifact%.tar.gz}"
symlink_hash="$(shasum -a 256 "$FIXTURES/v0.0.0-smoke/$artifact" | awk '{print $1}')"
printf '%s  %s\n' "$symlink_hash" "$artifact" > "$FIXTURES/v0.0.0-smoke/lufy-ai_v0.0.0-smoke_checksums.txt"
if bash "$ROOT/scripts/bootstrap.sh" --version v0.0.0-smoke --base-url "file://$FIXTURES" --install-dir "$BIN_DIR/symlink" 2>"$FIXTURES/symlink.err"; then
  echo "artifact con symlink interno no bloqueó instalación" >&2
  exit 1
fi
grep -q "symlink" "$FIXTURES/symlink.err"
if [[ -e "$BIN_DIR/symlink/lufy-ai" ]]; then
  echo "artifact con symlink interno instaló binario inesperadamente" >&2
  exit 1
fi

bad_work="$FIXTURES/bad-work"
mkdir -p "$bad_work/${artifact%.tar.gz}"
printf 'sin binario esperado\n' > "$bad_work/${artifact%.tar.gz}/README.cli.md"
tar -C "$bad_work" -czf "$FIXTURES/v0.0.0-smoke/$artifact" "${artifact%.tar.gz}"
bad_hash="$(shasum -a 256 "$FIXTURES/v0.0.0-smoke/$artifact" | awk '{print $1}')"
printf '%s  %s\n' "$bad_hash" "$artifact" > "$FIXTURES/v0.0.0-smoke/lufy-ai_v0.0.0-smoke_checksums.txt"
if bash "$ROOT/scripts/bootstrap.sh" --version v0.0.0-smoke --base-url "file://$FIXTURES" --install-dir "$BIN_DIR/missing-bin"; then
  echo "artifact sin binario esperado no bloqueó instalación" >&2
  exit 1
fi
