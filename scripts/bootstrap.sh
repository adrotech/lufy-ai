#!/usr/bin/env bash
# Bootstrap remoto seguro para instalar solo el binario lufy-ai verificado.

set -euo pipefail

REPO="${LUFY_AI_REPO:-adrotech/lufy-ai}"
VERSION="${LUFY_AI_VERSION:-}"
INSTALL_DIR="${LUFY_AI_INSTALL_DIR:-$HOME/.local/bin}"
BASE_URL="${LUFY_AI_RELEASE_BASE_URL:-}"
DRY_RUN="false"

curl_common_args=(--fail --silent --show-error --location --retry 3 --retry-all-errors --connect-timeout 10 --max-time 120)

usage() {
  cat <<'EOF'
Uso: scripts/bootstrap.sh --version <vX.Y.Z|latest> [--install-dir <dir>] [--dry-run] [--repo owner/name] [--base-url <url>]

Descarga el artifact de release de lufy-ai para tu OS/arch, verifica SHA-256 y
coloca solo el binario en el directorio elegido. No ejecuta `lufy-ai install`.

Variables equivalentes: LUFY_AI_VERSION, LUFY_AI_INSTALL_DIR, LUFY_AI_REPO,
LUFY_AI_RELEASE_BASE_URL.
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --version) VERSION="${2:-}"; shift 2 ;;
    --install-dir) INSTALL_DIR="${2:-}"; shift 2 ;;
    --repo) REPO="${2:-}"; shift 2 ;;
    --base-url) BASE_URL="${2:-}"; shift 2 ;;
    --dry-run) DRY_RUN="true"; shift ;;
    -h|--help) usage; exit 0 ;;
    *) echo "Error: argumento desconocido: $1" >&2; usage >&2; exit 2 ;;
  esac
done

if [[ -z "$VERSION" ]]; then
  cat >&2 <<'EOF'
Error: debes seleccionar versión explícita con --version vX.Y.Z o --version latest.
`latest` es una conveniencia no reproducible; para automatización usa una versión fija.
EOF
  exit 2
fi
if [[ -z "$INSTALL_DIR" ]]; then
  echo "Error: --install-dir no puede estar vacío" >&2
  exit 2
fi

os="$(uname -s | tr '[:upper:]' '[:lower:]')"
arch="$(uname -m)"
case "$os" in
  darwin|linux) ;;
  *) echo "Plataforma no soportada: $os/$arch. Soportadas: darwin/amd64, darwin/arm64, linux/amd64, linux/arm64." >&2; exit 1 ;;
esac
case "$arch" in
  x86_64|amd64) arch="amd64" ;;
  arm64|aarch64) arch="arm64" ;;
  *) echo "Arquitectura no soportada: $os/$arch. Soportadas: amd64, arm64." >&2; exit 1 ;;
esac

if [[ "$VERSION" == "latest" ]]; then
  if [[ -n "$BASE_URL" ]]; then
    echo "Error: latest no se resuelve en --base-url fixture; usa versión fija" >&2
    exit 2
  fi
  echo "Aviso: --version latest no es reproducible; para automatización usa una versión fija." >&2
  VERSION="$(curl "${curl_common_args[@]}" "https://api.github.com/repos/${REPO}/releases/latest" | sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' | head -n 1)"
  if [[ -z "$VERSION" ]]; then
    echo "Error: no se pudo resolver latest para $REPO" >&2
    exit 1
  fi
  echo "latest resuelto a $VERSION"
fi

artifact="lufy-ai_${VERSION}_${os}_${arch}.tar.gz"
checksums="lufy-ai_${VERSION}_checksums.txt"
if [[ -z "$BASE_URL" ]]; then
  BASE_URL="https://github.com/${REPO}/releases/download"
fi
release_url="${BASE_URL%/}/${VERSION}"
artifact_url="$release_url/$artifact"
checksums_url="$release_url/$checksums"

echo "Versión: $VERSION"
echo "Plataforma: $os/$arch"
echo "Artifact: $artifact_url"
echo "Destino: $INSTALL_DIR/lufy-ai"

if [[ "$DRY_RUN" == "true" ]]; then
  echo "Dry-run: no se descargó ni instaló ningún archivo"
  exit 0
fi

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

validate_tarball_entries() {
  local archive="$1"
  local package_dir="$2"
  local entries="$tmp/tar-entries.txt"
  local verbose_entries="$tmp/tar-verbose-entries.txt"
  local saw_binary="false"
  local mode=""
  local type=""

  if ! tar -tzf "$archive" > "$entries"; then
    echo "Error: artifact verificado no es un tar.gz legible" >&2
    exit 1
  fi

  while IFS= read -r entry; do
    case "$entry" in
      ""|/*|../*|*/../*|*/..|..|./*|*/./*|*/.|.|*//*)
        echo "Error: artifact contiene path inseguro: $entry" >&2
        exit 1
        ;;
    esac

    case "$entry" in
      lufy-ai|"$package_dir/lufy-ai") saw_binary="true" ;;
      "$package_dir/"|"$package_dir/README.cli.md") ;;
      *)
        echo "Error: artifact contiene entrada inesperada: $entry" >&2
        exit 1
        ;;
    esac
  done < "$entries"

  if ! tar -tvzf "$archive" > "$verbose_entries"; then
    echo "Error: artifact verificado no tiene metadata tar legible" >&2
    exit 1
  fi

  while IFS= read -r line; do
    mode="${line%% *}"
    type="${mode:0:1}"
    case "$type" in
      -|d) ;;
      *)
        echo "Error: artifact contiene entrada no permitida o symlink" >&2
        exit 1
        ;;
    esac
  done < "$verbose_entries"

  if [[ "$saw_binary" != "true" ]]; then
    echo "Error: artifact verificado no contiene binario lufy-ai en ruta permitida" >&2
    exit 1
  fi
}

curl "${curl_common_args[@]}" "$artifact_url" -o "$tmp/$artifact"
curl "${curl_common_args[@]}" "$checksums_url" -o "$tmp/$checksums"
expected="$(awk -v file="$artifact" '$2 == file {print $1}' "$tmp/$checksums")"
if [[ -z "$expected" ]]; then
  echo "Error: $checksums no contiene entrada para $artifact" >&2
  exit 1
fi
actual="$(shasum -a 256 "$tmp/$artifact" | awk '{print $1}')"
if [[ "$actual" != "$expected" ]]; then
  rm -f "$tmp/$artifact"
  echo "Error: checksum mismatch para $artifact; instalación bloqueada" >&2
  exit 1
fi

mkdir -p "$tmp/extract"
package_dir="${artifact%.tar.gz}"
validate_tarball_entries "$tmp/$artifact" "$package_dir"
tar -C "$tmp/extract" -xzf "$tmp/$artifact"
if [[ -x "$tmp/extract/lufy-ai" ]]; then
  bin="$tmp/extract/lufy-ai"
elif [[ -x "$tmp/extract/$package_dir/lufy-ai" ]]; then
  bin="$tmp/extract/$package_dir/lufy-ai"
else
  echo "Error: artifact verificado no contiene binario lufy-ai ejecutable" >&2
  exit 1
fi

if [[ -e "$INSTALL_DIR" && ! -w "$INSTALL_DIR" ]]; then
  echo "Error: destino no escribible: $INSTALL_DIR. Elige --install-dir en tu HOME o ejecuta una estrategia privilegiada explícita." >&2
  exit 1
fi
mkdir -p "$INSTALL_DIR"
install -m 0755 "$bin" "$INSTALL_DIR/lufy-ai"

echo "lufy-ai instalado en $INSTALL_DIR/lufy-ai"
case ":$PATH:" in
  *":$INSTALL_DIR:"*) ;;
  *)
    cat <<EOF

$INSTALL_DIR no está en tu PATH actual. Agrega el directorio a tu shell y abre una terminal nueva:

Bash/Zsh:
  export PATH="$INSTALL_DIR:\$PATH"

Fish:
  fish_add_path $INSTALL_DIR
  # Alternativa sin fish_add_path:
  set -gx PATH $INSTALL_DIR \$PATH

Este bootstrap no modifica automáticamente archivos de configuración del shell.
EOF
    ;;
esac
