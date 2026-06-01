#!/usr/bin/env bash
# Format one changed file using .lufy/project.yaml stack metadata.

set -euo pipefail

ROOT="${LUFY_PROJECT_ROOT:-$(pwd)}"
CONFIG="${LUFY_PROJECT_CONFIG:-$ROOT/.lufy/project.yaml}"
VERBOSE="${LUFY_FORMAT_DISPATCH_VERBOSE:-}"

log() {
  if [ -n "$VERBOSE" ]; then
    printf 'format-dispatch: %s\n' "$*" >&2
  fi
}

shell_quote() {
  printf "'%s'" "$(printf '%s' "$1" | sed "s/'/'\\\\''/g")"
}

trim_yaml_value() {
  sed -E 's/^[[:space:]]+//; s/[[:space:]]+$//; s/^"//; s/"$//; s/^'\''//; s/'\''$//'
}

extract_path_from_json() {
  sed -nE 's/.*"(file_path|path|filePath)"[[:space:]]*:[[:space:]]*"([^"]+)".*/\2/p' | head -n 1
}

input_path=""
while [ "$#" -gt 0 ]; do
  case "$1" in
    --file)
      shift
      input_path="${1:-}"
      ;;
    --file=*)
      input_path="${1#--file=}"
      ;;
    -*)
      ;;
    *)
      input_path="$1"
      ;;
  esac
  shift || true
done

if [ -z "$input_path" ]; then
  input_path="${LUFY_FORMAT_FILE:-${OPENCODE_FILE:-${CLAUDE_TOOL_FILE:-}}}"
fi

if [ -z "$input_path" ] && [ ! -t 0 ]; then
  input_path="$(extract_path_from_json || true)"
fi

if [ -z "$input_path" ] || [ ! -f "$CONFIG" ]; then
  log "sin archivo o config; omitiendo"
  exit 0
fi

ROOT="$(cd "$ROOT" && pwd -P)"
case "$input_path" in
  /*) abs_path="$input_path" ;;
  *) abs_path="$ROOT/$input_path" ;;
esac

if [ ! -f "$abs_path" ]; then
  log "archivo inexistente; omitiendo: $input_path"
  exit 0
fi

abs_dir="$(cd "$(dirname "$abs_path")" && pwd -P)"
abs_file="$abs_dir/$(basename "$abs_path")"
case "$abs_file" in
  "$ROOT"/*) rel_file="${abs_file#"$ROOT"/}" ;;
  *) log "archivo fuera del root; omitiendo: $input_path"; exit 0 ;;
esac

ext=".${rel_file##*.}"
if [ "$ext" = ".$rel_file" ]; then
  exit 0
fi

selection="$(
  awk -v ext="$ext" '
    function trim(s) {
      sub(/^[[:space:]]+/, "", s)
      sub(/[[:space:]]+$/, "", s)
      sub(/^"/, "", s)
      sub(/"$/, "", s)
      sub(/^'\''/, "", s)
      sub(/'\''$/, "", s)
      return s
    }
    function emit_if_match() {
      if (in_stack && supported == "true" && has_ext && formatter != "") {
        done = 1
        print formatter "\t" autofix
        exit
      }
    }
    /^    - id:/ {
      emit_if_match()
      in_stack = 1
      supported = "false"
      formatter = ""
      autofix = ""
      has_ext = 0
      in_formatter = 0
      in_linter = 0
      in_exts = 0
      next
    }
    !in_stack { next }
    /^ci:/ { emit_if_match(); in_stack = 0; next }
    /^      supported:/ { supported = trim($0); sub(/^supported:[[:space:]]*/, "", supported); supported = trim(supported); next }
    /^      formatter:/ { in_formatter = 1; in_linter = 0; in_exts = 0; next }
    /^      linter:/ { in_linter = 1; in_formatter = 0; in_exts = 0; next }
    /^      [a-z_]+:/ { in_formatter = 0; in_linter = 0; in_exts = 0; next }
    in_formatter && /^        command:/ {
      formatter = trim($0)
      sub(/^command:[[:space:]]*/, "", formatter)
      formatter = trim(formatter)
      next
    }
    in_formatter && /^        file_extensions:/ { in_exts = 1; next }
    in_exts && /^            - / {
      value = trim($0)
      sub(/^- /, "", value)
      value = trim(value)
      if (value == ext) has_ext = 1
      next
    }
    in_linter && /^        auto_fix:/ {
      autofix = trim($0)
      sub(/^auto_fix:[[:space:]]*/, "", autofix)
      autofix = trim(autofix)
      next
    }
    END { if (!done) emit_if_match() }
  ' "$CONFIG"
)"

if [ -z "$selection" ]; then
  log "sin formatter para $ext"
  exit 0
fi

formatter="$(printf '%s' "$selection" | cut -f1 | trim_yaml_value)"
autofix="$(printf '%s' "$selection" | cut -f2- | trim_yaml_value)"

is_usable_command() {
  [ -n "$1" ] || return 1
  case "$1" in
    TODO|TODO:*) return 1 ;;
    *) return 0 ;;
  esac
}

command_available() {
  first="${1%% *}"
  command -v "$first" >/dev/null 2>&1
}

with_file_arg() {
  cmd="$1"
  file_q="$(shell_quote "$rel_file")"
  printf '%s\n' "$cmd" | awk -v file="$file_q" '
    {
      out = ""
      replaced = 0
      for (i = 1; i <= NF; i++) {
        token = $i
        if (!replaced && token == ".") {
          token = file
          replaced = 1
        }
        out = out (out == "" ? "" : " ") token
      }
      if (!replaced) {
        out = out (out == "" ? "" : " ") file
      }
      print out
    }
  '
}

PATH="$ROOT/node_modules/.bin:$ROOT/.opencode/node_modules/.bin:$PATH"
cd "$ROOT"

if is_usable_command "$formatter" && command_available "$formatter"; then
  eval "$(with_file_arg "$formatter")"
else
  log "formatter no disponible: $formatter"
fi

if is_usable_command "$autofix" && command_available "$autofix"; then
  eval "$(with_file_arg "$autofix")"
fi
