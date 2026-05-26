#!/usr/bin/env bash
# Smoke tests for .opencode/hooks/format-dispatch.sh.

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
HOOK="$ROOT/.opencode/hooks/format-dispatch.sh"
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

mkdir -p "$TMP/.opencode" "$TMP/bin"

cat >"$TMP/.opencode/project.yaml" <<'YAML'
schema_version: 1
detected_at: 2026-05-25T00:00:00Z
stacks:
    - id: go
      supported: true
      formatter:
        command: fakefmt-go -w
        file_extensions:
            - .go
      linter:
        auto_fix: fakefix-go .
    - id: typescript
      supported: true
      formatter:
        command: fakefmt-ts --write .
        file_extensions:
            - .ts
            - .tsx
      linter:
        auto_fix: fakefix-ts . --fix
    - id: python
      supported: true
      formatter:
        command: fakefmt-py
        file_extensions:
            - .py
    - id: rust
      supported: false
      formatter:
        command: should-not-run
        file_extensions:
            - .rs
ci:
    detected: false
    workflows: []
YAML

cat >"$TMP/bin/fakefmt-go" <<'SH'
#!/usr/bin/env bash
printf 'fakefmt-go:%s\n' "$*" >>"$LUFY_SMOKE_LOG"
SH
cat >"$TMP/bin/fakefix-go" <<'SH'
#!/usr/bin/env bash
printf 'fakefix-go:%s\n' "$*" >>"$LUFY_SMOKE_LOG"
SH
cat >"$TMP/bin/fakefmt-ts" <<'SH'
#!/usr/bin/env bash
printf 'fakefmt-ts:%s\n' "$*" >>"$LUFY_SMOKE_LOG"
SH
cat >"$TMP/bin/fakefix-ts" <<'SH'
#!/usr/bin/env bash
printf 'fakefix-ts:%s\n' "$*" >>"$LUFY_SMOKE_LOG"
SH
cat >"$TMP/bin/fakefmt-py" <<'SH'
#!/usr/bin/env bash
printf 'fakefmt-py:%s\n' "$*" >>"$LUFY_SMOKE_LOG"
SH
chmod +x "$TMP/bin"/fake*

touch "$TMP/main.go" "$TMP/app.tsx" "$TMP/script.py" "$TMP/README.md" "$TMP/lib.rs"
export PATH="$TMP/bin:$PATH"
export LUFY_SMOKE_LOG="$TMP/format.log"

LUFY_PROJECT_ROOT="$TMP" "$HOOK" "$TMP/main.go"
LUFY_PROJECT_ROOT="$TMP" "$HOOK" --file app.tsx
printf '{"tool_input":{"file_path":"script.py"}}' | LUFY_PROJECT_ROOT="$TMP" "$HOOK"
LUFY_PROJECT_ROOT="$TMP" "$HOOK" README.md
LUFY_PROJECT_ROOT="$TMP" "$HOOK" lib.rs

expected="$TMP/expected.log"
cat >"$expected" <<'EOF'
fakefmt-go:-w main.go
fakefix-go:main.go
fakefmt-ts:--write app.tsx
fakefix-ts:app.tsx --fix
fakefmt-py:script.py
EOF

if ! diff -u "$expected" "$LUFY_SMOKE_LOG"; then
  echo "format-dispatch smoke failed" >&2
  exit 1
fi

echo "format-dispatch smoke ok"
