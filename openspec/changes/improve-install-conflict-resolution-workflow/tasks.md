## 1. OpenSpec

- [x] Crear propuesta, tareas y deltas de spec para el issue 173.
- [x] Validar propuesta con `openspec validate "improve-install-conflict-resolution-workflow" --strict`.

## 2. Conflict plan

- [x] Crear modelo read-only de plan de conflictos por archivo/categoría.
- [x] Clasificar conflictos de `.opencode/agents`, `.opencode/commands`, `.opencode/skills`, `.opencode/templates`, `openspec/specs` y root/config.
- [x] Detectar layout legacy `.lufy-ai/*` y reportarlo como deprecated/no-mutating.
- [x] Añadir tests unitarios.

## 3. CLI

- [x] Agregar `lufy-ai conflicts plan --target <dir> [--json] [--scope <scope>] [--tool <tool>]`.
- [x] Agregar salida humana agrupada y JSON parseable.
- [x] Agregar entrada en Command Palette.
- [x] Añadir tests CLI.

## 4. Setup/status/verify

- [x] Hacer que setup apunte al conflict plan cuando install está bloqueado por conflictos.
- [x] Agregar vista Bubble Tea read-only para revisar grupos de conflicto desde setup interactivo.
- [x] Mejorar status cuando falta manifest pero hay assets LUFY existentes.
- [x] Mejorar verify cuando falta manifest con recovery específico.

## 5. Harness paralelo

- [x] Documentar que los grupos independientes del conflict plan pueden convertirse en tareas paralelas del harness.
- [x] Mantener validación agrupada después del join.

## 6. Validación final

- [x] Ejecutar tests Go aplicables.
- [x] Ejecutar `scripts/validate.sh`.
- [x] Ejecutar OpenSpec validate.
- [x] Revisar diff final y evidenciar limitaciones.
