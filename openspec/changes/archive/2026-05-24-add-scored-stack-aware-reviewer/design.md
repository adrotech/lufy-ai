## Context

El repo ya cuenta con `.opencode/project.yaml` como fuente stack-aware y Result Contract envelope v1 para handoffs. `reviewer` existe como subagente read-only, pero su salida no está normalizada como score, no distingue L1-L5 y no adapta observability, coverage ni anti-patrones por stack.

## Goals / Non-Goals

**Goals:**

- Convertir `reviewer` en evaluador ponderado y stack-aware sin perder foco en hallazgos accionables.
- Definir aprobación objetiva: score >=80% y cero hallazgos L1/L2.
- Mantener separación de roles: reviewer no ejecuta delivery ni sustituye evidencia de `validator`.
- Preservar salida en español con identificadores técnicos y Result Contract envelope v1.

**Non-Goals:**

- No crear todavía un flujo de PR/GitHub automático.
- No cambiar el schema de `.opencode/project.yaml`.
- No exigir HTML si no se implementa skill dedicada en este slice.
- No convertir reviewer en agente que edita o corre validación pesada.

## Decisions

- Usar scoring por instrucciones del agente, no por código ejecutable. Rationale: LUFY-2 es un cambio de harness/agente; mantenerlo documental evita introducir tooling nuevo para una evaluación cualitativa.
- Definir pesos fijos en el agente: Architecture 20%, Code Quality 15%, Simplicity 15%, Testing 20%, Observability 15% y PR Template gate 15%. Rationale: coincide con backlog y suma 100%.
- Modelar severidades L1-L5 dentro del contrato de salida. Rationale: permite gate claro y preserva findings first.
- Consumir `.opencode/project.yaml` como input cuando exista, pero permitir `not_available` si falta. Rationale: stack-aware sin inventar comandos o librerías.
- Exigir desk-check de 8 escenarios para T1/T2 relevantes. Rationale: aumenta calidad de revisión sin convertir T3 en burocracia.

## Risks / Trade-offs

- Score puede volverse subjetivo -> Mitigación: exigir desglose por categoría, evidencia y razones de penalización.
- Stack config ausente puede generar falsos bloqueos -> Mitigación: marcar `not_available` y revisar con evidencia estática disponible.
- Reviewer puede duplicar validator -> Mitigación: reviewer revisa calidad/riesgo; validator conserva comandos y compile/test evidence.
- HTML opcional puede ampliar scope -> Mitigación: dejarlo fuera salvo implementación explícita de skill dedicada.
