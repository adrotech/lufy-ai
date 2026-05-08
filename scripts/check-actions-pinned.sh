#!/usr/bin/env bash
# Falla si workflows sensibles usan third-party actions con refs flotantes.

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
status=0

check_file() {
  local file="$1"
  local line ref owner_repo owner

  while IFS= read -r line; do
    case "$line" in
      *uses:*) ;;
      *) continue ;;
    esac

    ref="${line#*uses:}"
    ref="${ref%%#*}"
    ref="${ref//[[:space:]\"\']/}"
    case "$ref" in
      ""|./*|docker://*) continue ;;
      *@*) ;;
      *)
        printf 'Error: %s usa action sin ref: %s\n' "$file" "$ref" >&2
        status=1
        continue
        ;;
    esac

    owner_repo="${ref%@*}"
    owner="${owner_repo%%/*}"
    [ -n "$owner" ] || continue

    if [[ ! "${ref##*@}" =~ ^[0-9a-f]{40}$ ]]; then
      printf 'Error: %s usa action no pineada por SHA: %s\n' "$file" "$ref" >&2
      status=1
    fi
  done < "$ROOT/$file"
}

check_file ".github/workflows/go-cli-install.yml"
check_file ".github/workflows/release.yml"
check_file ".github/workflows/auto-release-tag.yml"

exit "$status"
