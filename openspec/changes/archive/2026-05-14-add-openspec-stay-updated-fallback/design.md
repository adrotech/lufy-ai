## Context

`modernize-openspec-core-v2` cerró el Sprint 1 de `v0.3.0`: config action-based, deltas, scenarios, `/opsx-sync`, `openspec-sync`, `opsx-version` y `UPSTREAM.json`. El baseline actual es deliberadamente local y embebido; permite instalación offline, pero no resuelve actualizaciones upstream ni cachea versiones.

El Sprint 2 busca desacoplar actualizaciones OpenSpec del ciclo manual de releases de `lufy-ai`, sin perder seguridad del instalador ni soporte standalone. La CLI Go debe seguir stdlib-only salvo decisión explícita posterior.

## Goals / Non-Goals

**Goals:**
- Resolver la fuente efectiva OpenSpec en tres capas: PATH, cache local y baseline embebida.
- Persistir cache y manifiestos bajo `.lufy-ai/openspec-cache/<version>/` con escrituras atómicas.
- Validar versión mínima, integridad básica y compatibilidad de capacidades antes de usar una fuente.
- Agregar workflow automatizado que proponga bumps de baseline mediante PR y nunca haga automerge.
- Mantener modo offline funcional mediante baseline embebida.

**Non-Goals:**
- No implementar perfil `expanded` ni comandos extra de Sprint 3.
- No instalar hooks de reconciliación de Sprint 4.
- No publicar `v0.3.0` ni cambiar política de release.
- No descargar ni ejecutar binarios remotos arbitrarios durante `install` o `sync` del usuario.
- No introducir dependencias externas en la CLI Go salvo aprobación posterior.

## Decisions

1. Crear paquete interno `opsx` en la CLI Go.

   El paquete contendrá tipos de manifest, resolución de versiones, validación de fuentes y operaciones de cache. Mantenerlo separado de `assets` evita mezclar catálogo de instalación con resolución runtime OpenSpec. Alternativa descartada: extender directamente `assets` o `syncer`; eso acoplaría cache upstream con assets gestionados instalables.

2. Resolver fuentes en orden conservador.

   Orden propuesto: `openspec` en `PATH` si existe y cumple versión mínima; cache local versionada si existe y valida; baseline embebida como fallback offline. Alternativa descartada: preferir red/cache antes del PATH, porque sorprende a usuarios que ya gestionan `openspec` globalmente.

3. Cache local dentro del target instalado.

   Usar `.lufy-ai/openspec-cache/<version>/manifest.json` permite rollback, inspección y confinamiento al target. Las escrituras deben usar helpers atómicos existentes. Alternativa descartada: cache global por usuario en este sprint; agrega scope/limpieza y cruza proyectos antes de tener experiencia con el resolver.

4. Workflow de bump por PR, no automerge.

   `.github/workflows/sync-openspec.yml` puede correr manual/programado y abrir PR con cambios de baseline/assets/manifiesto cuando detecte nueva versión compatible. No debe mergear automáticamente ni taggear releases. Alternativa descartada: actualizar `develop` directo; viola branch safety y reduce trazabilidad.

5. `UPSTREAM.json` sigue siendo baseline declarativo.

   Se puede extender con campos de versión mínima, cache o checksum si hace falta, pero no debe convertirse en lockfile remoto completo ni requerir red para comandos locales. Alternativa descartada: reemplazarlo por metadata generada solamente en build; dificultaría revisión humana y paridad root/embedded.

## Risks / Trade-offs

- Resolver PATH puede producir diferencias entre máquinas -> Mitigar reportando fuente efectiva y versión en `opsx-version`/diagnóstico.
- Cache corrupta puede romper comandos -> Mitigar validando manifest y fallback a baseline embebida.
- Workflow automático puede generar PRs ruidosos -> Mitigar con trigger manual y schedule conservador, sin automerge.
- Descarga/fetch upstream puede abrir riesgo supply-chain -> Mitigar con manifiesto, checksums cuando estén disponibles y revisión por PR.
- Tests con red serían frágiles -> Mitigar usando fixtures/cache local en tests normales y dejando red para workflow explícito.

## Migration Plan

1. Agregar paquete `opsx` con resolver 3 capas y pruebas unitarias sin red.
2. Extender `UPSTREAM.json` y su copia embebida solo con campos necesarios para resolución/validación.
3. Agregar cache/manifiesto local y pruebas de escritura atómica.
4. Integrar reporte de fuente efectiva en `opsx-version` o comando equivalente instalado.
5. Agregar workflow `sync-openspec.yml` con PR automático no-merge.
6. Validar con `openspec validate --all`, `scripts/validate.sh`, `git diff --check origin/develop` y smokes sandbox offline/cache.

## Open Questions

- ¿La fuente upstream inicial será un release oficial de OpenSpec, un repo GitHub o assets versionados propios de `lufy-ai`?
- ¿El resolver debe exponer un subcomando Go nuevo en este sprint o basta integrarlo a verify/status/version y assets instalados?
- ¿Qué versión mínima exacta de `openspec` debe aceptar el resolver además del baseline `1.3.1`?
