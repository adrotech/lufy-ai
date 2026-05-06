## 1. Contrato OpenSpec

- [x] 1.1 Crear proposal, design, tasks y spec deltas para `define-opencode-json-merge-managed-asset`.
- [x] 1.2 Definir que `opencode.json` usa estrategia especial `merge-json`, no asset completo por hash.

## 2. Implementación CLI Go

- [x] 2.1 Completar merge conservador de `opencode.json` en `install`, preservando claves desconocidas y fallando sin overwrite ante JSON inválido.
- [x] 2.2 Completar `sync` para aplicar `merge-json` seguro cuando corresponde, sin `copy`/`update-managed` por hash para `opencode.json`.
- [x] 2.3 Completar `verify` para validar JSON y estructura merge-managed mínima sin exigir hash completo en manifest.

## 3. Tests y documentación

- [x] 3.1 Añadir/ajustar tests de install para preservación, idempotencia, exclusión de manifest y JSON inválido.
- [x] 3.2 Añadir/ajustar tests de sync para merge-json, preservación, exclusión de manifest y JSON inválido.
- [x] 3.3 Añadir/ajustar tests de verify para JSON inválido y estructura mínima de `opencode.json`.
- [x] 3.4 Actualizar `docs/roadmap.md` y documentación CLI aplicable.

## 4. Validación

- [x] 4.1 Ejecutar `gofmt` sobre archivos Go modificados.
- [x] 4.2 Ejecutar `go test ./... && go build ./cmd/lufy-ai` desde `tools/lufy-cli-go/`.
- [x] 4.3 Ejecutar smokes disponibles de CLI/wrapper.
- [x] 4.4 Ejecutar `openspec status --change define-opencode-json-merge-managed-asset --json`.
- [x] 4.5 Ejecutar `openspec instructions apply --change define-opencode-json-merge-managed-asset --json`.
- [x] 4.6 Ejecutar `git diff --check`.
