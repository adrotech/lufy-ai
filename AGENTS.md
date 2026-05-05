# AGENTS.md

Guía operativa para agentes que trabajan en este repositorio `lufy-ai`.

## Snapshot del proyecto

- **Repositorio**: configuración local de OpenCode y flujo SDD/OpenSpec para `lufy-ai`.
- **Tooling raíz**: no hay `package.json` ni `tsconfig*.json` en la raíz; no asumir comandos Node/TS globales.
- **Tooling `.opencode`**: `.opencode/package.json` contiene dependencias del plugin TUI, no una suite de validación del producto.
- **Validación real**: normalmente estática/documental salvo que la tarea indique un toolchain específico. Siempre reportar comandos ejecutados y resultados reales.
- **Idioma**: respuestas, documentación humana, PRs y comentarios en español; preservar identificadores técnicos, rutas, flags y nombres de comandos.

## Estructura relevante

- `.opencode/agents/`: definiciones de agentes (`orchestrator`, `explorer`, `implementer`, `validator`, `reviewer`, `delivery`).
- `.opencode/commands/`: slash commands del flujo OpenSpec: `opsx-explore`, `opsx-propose`, `opsx-apply`, `opsx-verify`, `opsx-archive`.
- `.opencode/skills/sdd-workflow/`: skills para explorar, proponer, aplicar, verificar y archivar cambios OpenSpec.
- `.opencode/plugins/agent-observatory.tsx`: plugin TUI local Agent Observatory.
- `.opencode/policies/delivery.md`: fuente canónica para delivery, branch safety, validación y gates de cambios completos.
- `openspec/`: propuestas, especificaciones y tareas del flujo OpenSpec.
- `docs/`: documentación del proyecto cuando exista.
- `AGENTS.md.template`: plantilla genérica; este archivo es la guía real del repo.

## Comandos disponibles y límites

Ejecutar desde la raíz salvo que se indique otra ruta.

- OpenSpec/OpenCode: usar `/opsx-explore`, `/opsx-propose`, `/opsx-apply`, `/opsx-verify`, `/opsx-archive` cuando corresponda.
- Observatory TUI: `/observatory`, `/observatory-agents`, `/observatory-subagents`, `/observatory-cost`.
- Git inspección: `git status --short`, `git diff`, `git diff --check`, `git log` según permisos del rol.
- No inventar `npm test`, `npm run typecheck`, `tsc` u otros comandos si el toolchain no existe para el alcance actual.
- Si se requiere validación no disponible, reportar la limitación y la evidencia estática/manual realizada.

## Reglas de arquitectura y workflow

1. Mantener cambios enfocados y mínimos.
2. No revertir ni sobrescribir trabajo local no relacionado.
3. Mantener handlers/controllers delgados; servicios contienen reglas de negocio.
4. No exponer entidades de persistencia como contratos HTTP/API.
5. Usar inyección por constructor donde aplique.
6. Mantener scopes transaccionales estrechos.
7. No cambiar puertos, defaults de auth, esquema de base de datos ni contratos públicos salvo autorización explícita.
8. Añadir o actualizar pruebas/documentación solo cuando estén ligadas al cambio.
9. Nunca afirmar validación exitosa sin evidencia de comando o revisión manual concreta.
10. Preferir lectura y edición específicas sobre exploración amplia.
11. En handoffs y gestión de contexto, resumir decisiones, evitar dumps largos y preservar solo la evidencia mínima útil.

## Roles de agentes

- `orchestrator`: coordina y enruta; no edita ni ejecuta shell.
- `explorer`: investiga en modo read-only y produce handoff para implementación.
- `implementer`: implementa cambios acotados; no hace commit, push, PR ni sync de Projects.
- `validator`: valida y diagnostica en modo read-only; no edita.
- `reviewer`: revisa calidad, riesgos y cobertura; no edita.
- `delivery`: con autorización explícita, maneja Git/GH, PRs y trazabilidad siguiendo `.opencode/policies/delivery.md`.

## OpenSpec workflow

- Explorar idea o impacto: `opsx-explore` / skill `openspec-explore`.
- Crear propuesta completa: `opsx-propose` / skill `openspec-propose`.
- Implementar tareas: `opsx-apply` / skill `openspec-apply-change`.
- Verificar implementación contra artefactos: `opsx-verify` / skill `openspec-verify-change`.
- Archivar cambio completado: `opsx-archive` / skill `openspec-archive-change`.
- Una tarea OpenSpec solo se considera cerrada si cumple los gates de `.opencode/policies/delivery.md`.

## Política de delivery

- Consultar `.opencode/policies/delivery.md` para validación por tiers, branch safety, PRs, sync y estados `blocked` / `sync_pending`.
- No hacer commit, push, PR ni actualizar GitHub Projects sin autorización explícita del usuario y rol `delivery`.
- No crear PR desde `development`, `develop`, `main` o `master`.
- Nunca usar force push salvo solicitud explícita.

## Formato de reporte

- Incluir objetivo, cambios/evidencia, riesgos y estado listo/bloqueado.
- Mantener resúmenes concisos; usar rutas y líneas cuando ayuden.
- Si falta contexto o una decisión, pedirla o devolver el bloqueo exacto.
