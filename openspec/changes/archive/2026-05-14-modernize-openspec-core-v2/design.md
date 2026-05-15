## Context

`lufy-ai` instala hoy un kit OpenCode/OpenSpec desde assets gestionados y embebidos por la CLI Go. Después de `v0.2.0`, install/sync pueden actualizar esos assets sin destruir drift local, por lo que el siguiente riesgo se mueve al contrato OpenSpec instalado: los cambios pueden existir sin deltas claros, scenarios testables o una acción explícita para sincronizar specs principales.

El alcance de este cambio es el Sprint 1 de `v0.3.0`: cerrar el core gap con OpenSpec v1.3.1. Los sprints posteriores (`stay-updated` 3 capas, perfil expanded, reconciliation hooks y docs de release) quedan fuera para mantener el cambio mergeable.

## Goals / Non-Goals

**Goals:**
- Instalar una superficie OpenSpec core v2 con configuración action-based, baseline local y comandos/skills mínimos.
- Hacer que proposals futuras declaren deltas `ADDED`, `MODIFIED` y `REMOVED` con scenarios verificables.
- Agregar `/opsx-sync` y `openspec-sync` para sincronizar deltas validados hacia specs principales antes de archive.
- Exponer `opsx-version` como reporte local de versión/baseline/fuente del workflow OpenSpec instalado.
- Mantener compatibilidad con targets existentes usando el modelo de drift resolution de `v0.2.0`.

**Non-Goals:**
- No implementar todavía resolución runtime de OpenSpec por PATH/cache/baseline embebida.
- No añadir perfil expanded ni comandos `/opsx-new`, `/opsx-continue`, `/opsx-ff`, `/opsx-bulk-archive`, `/opsx-onboard` o `/opsx-doctor`.
- No instalar hooks de reconciliación ni modificar hooks existentes del usuario.
- No cambiar la política de release ni publicar `v0.3.0` en este proposal.
- No introducir dependencias externas en la CLI Go salvo decisión explícita posterior.

## Decisions

1. Mantener el core v2 como assets gestionados.

   La configuración, comandos, skills y `UPSTREAM.json` se versionan dentro de `openspec/`, `.opencode/commands/` y `.opencode/skills/`, y se copian también a `tools/lufy-cli-go/internal/assets/embedded/`. Esto conserva la instalación standalone y reutiliza catalog/hash/backup existentes. Alternativa descartada: descargar OpenSpec v2 en install; eso pertenece al Sprint 2 stay-updated y agregaría red/cache antes de estabilizar el contrato local.

2. Validar deltas/scenarios en el flujo `/opsx-*` antes de archive.

   `/opsx-sync` será el punto explícito para aplicar deltas a specs principales y fallará si faltan markers o scenarios testables. `opsx-archive` deberá depender de que los specs principales estén sincronizados. Alternativa descartada: sincronizar implícitamente durante archive, porque oculta cambios contractuales y hace más difícil auditar qué se modificó.

3. Tratar `UPSTREAM.json` como baseline local declarativo.

   El archivo describe la versión efectiva de OpenSpec y el perfil core cubierto por los assets instalados. No es todavía un resolver remoto ni un lockfile de cache. Alternativa descartada: usar solo documentación humana; no es machine-readable para verify/status futuros.

4. Mantener `opsx-version` como comando/script ligero del kit instalado.

   El reporte debe poder ejecutarse sin toolchain Node/TS global y sin red. Puede implementarse como slash command que lee `UPSTREAM.json` y muestra versión, perfil y fuente. Alternativa descartada: agregar subcomando Go nuevo en este sprint; no hace falta para cumplir el core workflow instalado.

5. Preservar upgrade brownfield mediante policies existentes.

   Los nuevos assets son managed; `AGENTS.md` sigue `merge-block`; archivos user-owned siguen `no-replace` o policy existente. El proposal no debe reabrir el modelo de drift resolution.

## Risks / Trade-offs

- Specs existentes con formato anterior pueden fallar validaciones nuevas -> Mitigar documentando migración mínima y aplicando enforcement solo a cambios nuevos del workflow v2.
- `/opsx-sync` puede duplicar parte de `openspec archive` -> Mitigar separando responsabilidades: sync aplica deltas, archive mueve/cierra cambios ya sincronizados.
- Baseline local puede quedar obsoleta -> Mitigar dejando `UPSTREAM.json` explícito y reservar actualización automática para Sprint 2.
- Nuevos assets embebidos pueden desincronizarse -> Mitigar manteniendo tests de paridad `TestEmbeddedCatalogMatchesRepositoryAssets` y validación `scripts/validate.sh`.
- Cambio amplio en comandos/skills puede afectar usuarios con customizaciones -> Mitigar usando `v0.2.0` policies, backups y `.lufy-new` cuando haya drift.

## Migration Plan

1. Añadir assets core v2 en fuente raíz y embebidos.
2. Actualizar catálogo/verify solo si los nuevos archivos deben ser obligatorios.
3. Actualizar comandos y skills OpenSpec para usar config action-based y deltas/scenarios.
4. Validar con `openspec validate --all`, `scripts/validate.sh`, `git diff --check origin/develop` y smokes sandbox greenfield/brownfield.
5. Si el cambio introduce regressions, revertir el PR completo antes de release; no hay migración persistente nueva fuera de assets gestionados.

## Open Questions

- ¿El formato final de `UPSTREAM.json` debe incluir checksum de templates core o solo versión/perfil/fuente?
- ¿`opsx-version` debe ser solo slash command OpenCode o también subcomando de la CLI Go en un sprint posterior?
- ¿El enforcement de scenarios debe aceptar `GIVEN` opcional o requerir siempre `GIVEN`/`WHEN`/`THEN` completo?
