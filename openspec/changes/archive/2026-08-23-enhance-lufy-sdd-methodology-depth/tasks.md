# Tasks: enhance-lufy-sdd-methodology-depth

- [x] 1. Definir contrato OpenSpec
  - [x] 1.1 Agregar deltas para profundidad de Lufy SDD Full/Lite.
  - [x] 1.2 Validar `openspec validate enhance-lufy-sdd-methodology-depth --strict`.

- [x] 2. Profundizar assets Lufy SDD
  - [x] 2.1 Expandir `.lufy/sdd/templates/proposal.md`.
  - [x] 2.2 Expandir `.lufy/sdd/templates/design.md`.
  - [x] 2.3 Expandir `.lufy/sdd/templates/spec.md`.
  - [x] 2.4 Expandir `.lufy/sdd/templates/tasks.md`.
  - [x] 2.5 Actualizar `.lufy/sdd/README.md` y acciones lifecycle relevantes.

- [x] 3. Alinear scaffold runtime
  - [x] 3.1 Actualizar `tools/lufy-cli-go/internal/lufysdd/service.go`.
  - [x] 3.2 Agregar tests de contenido para scaffolds Full y Lite.

- [x] 4. Sincronizar assets embebidos
  - [x] 4.1 Copiar `.lufy/sdd` actualizado a `tools/lufy-cli-go/internal/assets/embedded/.lufy/sdd`.
  - [x] 4.2 Verificar que el catálogo embebido no tenga drift.

- [x] 5. Validar
  - [x] 5.1 Ejecutar `openspec validate enhance-lufy-sdd-methodology-depth --strict`.
  - [x] 5.2 Ejecutar `go test ./internal/lufysdd`.
  - [x] 5.3 Ejecutar `go test ./internal/assets`.
  - [x] 5.4 Ejecutar `scripts/validate.sh`.
