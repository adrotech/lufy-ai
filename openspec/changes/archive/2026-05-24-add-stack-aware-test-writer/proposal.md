## Why

El backlog prioriza `LUFY-1` porque el flujo TDD actual no tiene un rol dedicado que consuma `.opencode/project.yaml` ni produzca evidencia RED/GREEN/TRIANGULATE/REFACTOR consistente por stack. Esto bloquea que T1/T2 mantengan disciplina de pruebas sin asumir Go u otro toolchain fijo.

## What Changes

- Crear el agente `.opencode/agents/test-writer.md` como subagente especializado en TDD stack-aware.
- Hacer que `implementer` delegue pruebas sustantivas de T1/T2 a `test-writer` cuando el cambio requiera ciclo TDD observable.
- Hacer que `validator` bloquee o escale T1/T2 cuando falte evidencia TDD requerida por el cambio.
- Definir que `test-writer` lee comandos, coverage y anti-patrones desde `.opencode/project.yaml` cuando exista, y reporta limitaciones cuando no exista o no aplique.
- Registrar evidencia TDD en el Result Contract envelope v1 del bloque o en `apply-progress.md` cuando el cambio OpenSpec lo use.

## Capabilities

### New Capabilities

- `stack-aware-test-writer`: rol, contrato y comportamiento del agente `test-writer` para producir pruebas y evidencia TDD parametrizada por `.opencode/project.yaml`.

### Modified Capabilities

- `systemic-workflow`: agrega reglas de delegación y validación para evidencia TDD en cambios T1/T2 relevantes.

## Impact

- Archivos esperados: `.opencode/agents/test-writer.md`, `.opencode/agents/implementer.md`, `.opencode/agents/validator.md`, assets embebidos equivalentes bajo `tools/lufy-cli-go/internal/assets/embedded/.opencode/agents/` si aplica.
- Specs OpenSpec: nueva spec `stack-aware-test-writer` y delta sobre `systemic-workflow`.
- Validación: revisión estática de agentes, validación OpenSpec estricta, paridad de assets embebidos si se sincronizan assets gestionados y validación agrupada disponible.
