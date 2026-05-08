## Context

El repositorio ya define validacion agrupada en `.opencode/policies/delivery.md` y en varios agentes/skills, pero la regla queda limitada a no correr tests constantemente. La necesidad nueva es mas amplia: reducir trabajo repetido durante propuestas OpenSpec mediante un flujo sistemico que separe analisis inicial, implementacion por bloques, relectura final acotada y validacion final con tests/coverage.

El cambio afecta reglas operativas de agentes y documentacion del workflow, no runtime del producto. Debe preservar que las excepciones siguen permitidas cuando existe bloqueo, cambio riesgoso, diagnostico de falla o evidencia nueva que invalida el analisis inicial.

## Goals / Non-Goals

**Goals:**
- Formalizar el pensamiento sistemico como criterio operativo de propuestas OpenSpec.
- Evitar relecturas repetidas de archivos viejos durante implementacion normal.
- Concentrar tests, coverage y validacion completa al final de todas las tareas de una propuesta.
- Mantener excepciones claras para bloqueos, riesgos y diagnostico.
- Alinear politica, agentes y skills para que el comportamiento sea consistente.

**Non-Goals:**
- Cambiar permisos de agentes o comandos permitidos.
- Eliminar validacion temprana cuando sea necesaria para desbloquear trabajo.
- Definir una suite universal de tests para todo repositorio.
- Cambiar contratos de CLI, instalador, specs de producto o rama de delivery.

## Decisions

1. Usar `systemic-workflow` como nueva capacidad OpenSpec.

   Rationale: el comportamiento cruza agentes, skills y politica, por lo que conviene definirlo como capacidad propia en vez de esconderlo como detalle de implementacion.

   Alternative considered: modificar solo `.opencode/policies/delivery.md`. Se descarta porque no deja criterios verificables ni escenarios para apply/verify/archive.

2. Mantener `current-state-documentation` como capacidad modificada solo para documentar estado operativo.

   Rationale: la documentacion vigente debe reflejar la regla sin mezclarla con capacidades runtime.

   Alternative considered: no tocar specs existentes. Se descarta porque el cambio tambien afecta como se presenta el estado real del workflow.

3. Definir fases, no comandos rigidos.

   Rationale: el repo no tiene toolchain universal; la regla debe describir cuando validar y como reportar evidencia real, no inventar comandos.

   Alternative considered: exigir comandos fijos de tests/coverage. Se descarta por la politica existente de no asumir toolchains globales.

4. Explicitar excepciones.

   Rationale: evitar tests/relecturas constantes no debe impedir diagnosticar fallas, resolver bloqueos o revisar cambios riesgosos.

   Alternative considered: prohibicion absoluta de relecturas/tests tempranos. Se descarta porque haria el flujo fragil ante incertidumbre real.

## Risks / Trade-offs

- Riesgo: agentes interpreten la regla como no leer suficiente contexto. Mitigacion: exigir analisis inicial sistemico con dependencias e interconexiones antes de implementar.
- Riesgo: tests finales fallen tarde. Mitigacion: mantener excepciones para cambios riesgosos, bloqueo o diagnostico temprano.
- Riesgo: coverage no exista para algunos alcances. Mitigacion: reportar limitacion y evidencia real disponible sin afirmar exito.
- Riesgo: relectura final sea demasiado amplia. Mitigacion: limitarla a archivos viejos modificados/afectados o evidencia nueva.

## Migration Plan

1. Actualizar politica central y `AGENTS.md` con las fases del workflow sistemico.
2. Actualizar agentes `orchestrator`, `explorer`, `implementer` y `validator` con responsabilidades por fase.
3. Actualizar skills OpenSpec `apply` y `verify` para que tareas y validacion sigan el mismo criterio.
4. Actualizar documentacion OpenSpec si corresponde.
5. Verificar consistencia documental y ejecutar validacion final disponible al cerrar la propuesta.

Rollback: revertir los cambios documentales y de instrucciones; no hay migracion de datos ni runtime.

## Open Questions

- No hay preguntas abiertas. La implementacion debe preservar excepciones ya existentes de bloqueo, riesgo y diagnostico.
