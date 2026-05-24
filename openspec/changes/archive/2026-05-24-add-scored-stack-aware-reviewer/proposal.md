## Why

`LUFY-2` busca que el reviewer deje de ser un checklist genérico y produzca evaluación consistente, ponderada y adaptada al stack detectado. Hoy `.opencode/agents/reviewer.md` no consume `.opencode/project.yaml`, no emite score L1-L5 y no tiene gate objetivo de aprobación.

## What Changes

- Extender `.opencode/agents/reviewer.md` con scoring ponderado por categorías: Architecture, Code Quality, Simplicity, Testing, Observability y PR Template gate.
- Definir severidades L1-L5 con bloqueo para L1/L2 y aprobación solo con score >=80%.
- Hacer que reviewer lea `.opencode/project.yaml` cuando exista para stack, coverage, anti-patrones y librerías de observabilidad.
- Exigir desk-check de al menos 8 escenarios para T1/T2 relevantes.
- Mantener reviewer read-only y con Result Contract envelope v1.
- Documentar el output de review de modo capability-aware; el HTML autocontenido queda como opcional si se implementa skill en el mismo slice.

## Capabilities

### New Capabilities

- `scored-stack-aware-reviewer`: contrato de revisión ponderada, severidades L1-L5, consumo stack-aware y resultado capability-aware.

### Modified Capabilities

- `systemic-workflow`: agrega gate de revisión ponderada para cambios T1/T2 relevantes y separación entre review cualitativo, validación de comandos y delivery.

## Impact

- Archivos esperados: `.opencode/agents/reviewer.md`, copia embebida en `tools/lufy-cli-go/internal/assets/embedded/.opencode/agents/reviewer.md`, `.opencode/README.md` y `AGENTS.md` si se ajustan roles/contrato.
- Specs OpenSpec: nueva spec `scored-stack-aware-reviewer` y delta sobre `systemic-workflow`.
- Validación: OpenSpec strict, paridad de assets embebidos y `scripts/validate.sh`.
