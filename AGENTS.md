# AGENTS.md

Guía operativa para agentes que trabajan en este repositorio `lufy-ai`.

## Snapshot del proyecto

- **Repositorio**: configuración local de OpenCode y flujo SDD/OpenSpec para `lufy-ai`.
- **CLI del producto**: la CLI Go vive en `tools/lufy-cli-go`; no asumir una CLI legacy fuera de esa ruta.
- **Instalador**: `scripts/install.sh` es un wrapper estricto del CLI Go y no debe reintroducir fallback legacy.
- **Tooling raíz**: no hay `package.json` ni `tsconfig*.json` en la raíz; no asumir comandos Node/TS globales.
- **Tooling `.opencode`**: `.opencode/package.json` contiene dependencias del plugin TUI, no una suite de validación del producto.
- **Validación real**: normalmente estática/documental salvo que la tarea indique un toolchain específico. Siempre reportar comandos ejecutados y resultados reales.
- **Workflow sistémico**: analizar archivos existentes, dependencias e interconexiones al inicio; evitar relecturas repetidas durante implementación; releer al final solo archivos viejos modificados/afectados o casos justificados.
- **Idioma**: respuestas, documentación humana, PRs y comentarios en español; preservar identificadores técnicos, rutas, flags y nombres de comandos.
- **Ramas y releases**: `develop` es la base normal de integración; `main` es productiva/estable; los releases estables se publican solo desde tags `v*` sobre commits alcanzables desde `main`.

## Estructura relevante

- `.opencode/agents/`: definiciones de agentes (`orchestrator`, `explorer`, `implementer`, `validator`, `reviewer`, `delivery`).
- `.opencode/commands/`: slash commands del flujo OpenSpec: `opsx-explore`, `opsx-propose`, `opsx-apply`, `opsx-verify`, `opsx-archive`.
- `.opencode/skills/sdd-workflow/`: skills para explorar, proponer, aplicar, verificar y archivar cambios OpenSpec.
- `.opencode/plugins/agent-observatory.tsx`: plugin TUI local Agent Observatory.
- `.opencode/policies/delivery.md`: fuente canónica para delivery, branch safety, validación y gates de cambios completos.
- `openspec/`: propuestas, especificaciones y tareas del flujo OpenSpec.
- `tools/lufy-cli-go/`: implementación actual de la CLI Go usada por el instalador.
- `scripts/install.sh`: wrapper estricto hacia `tools/lufy-cli-go`, sin fallback legacy.
- `docs/`: documentación del proyecto cuando exista.
- `AGENTS.md.template`: plantilla genérica; este archivo es la guía real del repo.

## Comandos disponibles y límites

Ejecutar desde la raíz salvo que se indique otra ruta.

- OpenSpec/OpenCode: usar `/opsx-explore`, `/opsx-propose`, `/opsx-apply`, `/opsx-verify`, `/opsx-archive` cuando corresponda.
- Observatory TUI: `/observatory`, `/observatory-agents`, `/observatory-subagents`, `/observatory-cost`.
- Git inspección: `git status --short`, `git diff`, `git diff --check`, `git log` según permisos del rol.
- No inventar `npm test`, `npm run typecheck`, `tsc` u otros comandos si el toolchain no existe para el alcance actual.
- Respetar la preferencia de validación agrupada: no correr tests constantemente; agrupar tests, coverage y validación completa al final de todas las tareas de un bloque/proposal salvo bloqueo, cambio riesgoso o diagnóstico.
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
12. Mantener `scripts/install.sh` como wrapper estricto de `tools/lufy-cli-go`; no reintroducir rutas legacy.
13. Aplicar pensamiento sistémico: entender el todo, interconexiones, dependencias, bucles de feedback y cómo la estructura estática produce comportamiento dinámico.
14. Durante una propuesta, concentrar el análisis de código viejo al inicio y la revisión final en archivos viejos modificados/afectados; no releer archivos ya analizados salvo conflicto, bloqueo, nueva evidencia, cambio de alcance o riesgo explícito.

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
- En `opsx-apply`, completar tareas por bloque sin test loops ni relecturas rutinarias; en `opsx-verify`, correr la validación final agrupada disponible, incluyendo tests/coverage solo si existen para el alcance real.
- Foco activo actual: `install-managed-assets-with-hash-idempotency` (assets gestionados, SHA-256, manifest, idempotencia, backup/restore y verify estructural).
- No archivar `migrate-installer-to-go-cli` mientras tenga tasks incompletas; tasks incompletas implican `blocked`, no archive.

## Política de delivery

- Consultar `.opencode/policies/delivery.md` para validación por tiers, branch safety, PRs, sync y estados `blocked` / `sync_pending`.
- PR normal: ramas `feature/*`, `fix/*`, `chore/*` o equivalentes → `develop`.
- Promoción productiva: `develop` → `main` con autorización y evidencia de validación.
- `main` no es base de trabajo diario; se reserva para producción, release y hotfix explícitamente autorizado.
- Tags de release estable: `v*` creados desde commits alcanzables desde `origin/main`; no publicar releases desde `develop` sin promoción.
- No hacer commit, push, PR ni actualizar GitHub Projects sin autorización explícita del usuario y rol `delivery`.
- No crear PR desde ramas protegidas como `develop`, `main`, `master` o `development`, salvo promoción `develop` → `main` explícitamente autorizada.
- Nunca usar force push salvo solicitud explícita.

## Formato de reporte

- Incluir objetivo, cambios/evidencia, riesgos y estado listo/bloqueado.
- Mantener resúmenes concisos; usar rutas y líneas cuando ayuden.
- Si falta contexto o una decisión, pedirla o devolver el bloqueo exacto.
