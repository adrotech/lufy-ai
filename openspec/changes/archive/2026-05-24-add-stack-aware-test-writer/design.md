## Context

El repo ya tiene `.opencode/project.yaml` como contrato stack-aware generado por `lufy-ai init`, y los agentes principales consumen Result Contract envelope v1. `implementer` y `validator` hoy documentan validación agrupada, pero no existe un subagente dedicado a escribir pruebas ni una regla explícita para evidencia RED/GREEN/TRIANGULATE/REFACTOR parametrizada por stack.

El cambio afecta configuración de OpenCode: agentes bajo `.opencode/agents/` y, si esos agentes son assets gestionados, sus copias embebidas bajo `tools/lufy-cli-go/internal/assets/embedded/.opencode/agents/`.

## Goals / Non-Goals

**Goals:**

- Introducir `test-writer` como subagente especializado en TDD para cambios T1/T2 sustantivos.
- Mantener el flujo stack-aware leyendo comandos, coverage y anti-patrones desde `.opencode/project.yaml` cuando exista.
- Evitar defaults hardcodeados a Go; ante config ausente o incompleta, reportar limitaciones en vez de inventar comandos.
- Conectar `implementer` y `validator` con el nuevo rol sin convertir cada micro-cambio en un ciclo TDD obligatorio.

**Non-Goals:**

- No implementar nuevos scanners ni modificar el schema de `.opencode/project.yaml`.
- No exigir TDD para cambios T3 triviales, documentación simple o tareas donde no aplica suite automatizada.
- No crear un runner de tests nuevo; el agente usa comandos existentes/configurados.
- No cambiar delivery, PRs ni policies fuera de la evidencia TDD requerida.

## Decisions

- Crear `test-writer` como agente de archivo en `.opencode/agents/test-writer.md`. Rationale: mantiene el patrón existente de agentes project-local y permite permisos específicos sin inflar `implementer`. Alternativa considerada: convertirlo en skill, pero el backlog pide agent y el uso requiere delegación como subagente.
- El agente debe operar con Result Contract envelope v1. Rationale: evita un contrato paralelo y permite que `implementer`, `validator` y `delivery` consuman evidencia homogénea. Alternativa considerada: checklist libre en texto, descartado porque no es parseable ni consistente.
- `test-writer` debe preferir `.opencode/project.yaml` y fallar de forma informativa cuando falten comandos. Rationale: preserva la dirección stack-aware y evita regresar a suposiciones Go. Alternativa considerada: fallback a comandos comunes por lenguaje, descartado salvo que estén declarados en project config o en el alcance del cambio.
- La delegación desde `implementer` será condicional para T1/T2 con pruebas sustantivas. Rationale: reduce ruido en cambios T3 o documentales y respeta validación proporcional. Alternativa considerada: delegación obligatoria para todo cambio, descartada por exceso de costo y falsos bloqueos.
- `validator` debe revisar presencia/calidad de evidencia TDD cuando el cambio la requiere. Rationale: hace observable el gate sin convertir al validator en escritor de tests. Alternativa considerada: confiar solo en el reporte de implementer, descartada porque no detecta omisiones de evidencia.

## Risks / Trade-offs

- Agente nuevo puede aumentar fricción en cambios pequeños -> Mitigación: limitar obligación a T1/T2 sustantivos y permitir `not_applicable` explícito.
- Config ausente en repos destino puede bloquear comandos reales -> Mitigación: reportar `blocked`/`not_available` y recomendar `lufy-ai init` o configuración manual.
- Evidencia TDD puede volverse narrativa y no verificable -> Mitigación: exigir comandos exactos, resultado, fase TDD y archivos de prueba tocados.
- Assets root y embebidos pueden quedar desalineados -> Mitigación: incluir tarea explícita de sincronización/paridad si se modifican assets gestionados.

## Migration Plan

1. Agregar `test-writer` con permisos mínimos para lectura, edición de tests y comandos de validación configurados.
2. Actualizar `implementer` para delegar pruebas T1/T2 sustantivas y preservar evidencia TDD en el Result Contract.
3. Actualizar `validator` para bloquear/escalar cuando falte evidencia TDD requerida.
4. Sincronizar assets embebidos si los agentes son parte del catálogo instalado.
5. Validar OpenSpec, coherencia documental y comandos locales aplicables.
