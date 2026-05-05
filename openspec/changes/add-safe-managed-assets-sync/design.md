## Context

La CLI Go en `tools/lufy-cli-go` ya es el punto de entrada previsto para instalar, verificar, respaldar y restaurar assets gestionados. Las specs existentes definen un catálogo permitido, resolución segura de source/target, manifest `.lufy-ai/install-state.json`, hashes SHA-256, backups multiasset, restore y verify estructural.

El nuevo flujo debe cubrir el caso posterior a una instalación: reaplicar cambios del source a un proyecto target existente sin tocar personalizaciones fuera de scope. El wrapper Bash debe permanecer como compatibilidad estricta de `install` y no recuperar lógica legacy.

## Goals / Non-Goals

**Goals:**
- Añadir `lufy-ai sync` como comando explícito para actualizar assets gestionados ya instalados.
- Reusar el catálogo, planner, hashing, manifest, backup, restore y verify existentes en vez de crear un mecanismo paralelo.
- Garantizar idempotencia: sync repetido sin cambios produce `skip` y no modifica contenido ni timestamps de assets idénticos.
- Proteger personalizaciones locales: drift local, archivos no gestionados y estado inválido bloquean la actualización del asset afectado.
- Mantener mutaciones confinadas al target y `.lufy-ai/` dentro del target.

**Non-Goals:**
- No implementar templates nuevos ni generación de contenido parametrizable.
- No productizar distribución remota, auto-update, descarga de binarios ni telemetry.
- No cambiar contratos públicos existentes de `install`, `verify`, `backup` o `restore` salvo el reuso necesario para `sync`.
- No extender `scripts/install.sh` para exponer sync ni añadir fallback legacy.

## Decisions

1. `sync` será un comando propio de la CLI Go.
   - Rationale: el usuario necesita distinguir instalación inicial de reaplicación segura. Un comando explícito permite mensajes, exit codes y ayuda dedicados.
   - Alternativa considerada: hacer que `install` actúe como sync cuando existe estado. Se descarta porque mezcla intención inicial y mantenimiento, y aumenta el riesgo de sobrescrituras inesperadas.

2. El planner de sync derivará acciones desde tres hashes: source actual, target actual y último target/source registrado en `.lufy-ai/install-state.json`.
   - Rationale: este modelo permite separar upstream cambiado, target sin drift y drift local sin depender de timestamps.
   - Alternativa considerada: comparar solo source contra target. Se descarta porque no distingue un cambio upstream seguro de una personalización local.

3. `sync` solo actualizará assets gestionados registrados o ausentes que pertenezcan al catálogo y tengan estado suficiente para una decisión segura.
   - Rationale: protege archivos del usuario y evita incorporar contenido desconocido como gestionado por accidente.
   - Alternativa considerada: sobrescribir cualquier path del catálogo si el usuario pasa `--yes`. Se descarta porque `--yes` confirma un plan seguro, no convierte conflictos en sobrescrituras.

4. Todo `update-managed` de sync requiere backup previo cuando el target existe.
   - Rationale: el rollback debe ser posible para cambios multiasset, y el comportamiento coincide con la política de instalación gestionada existente.
   - Alternativa considerada: backup opcional para sync. Se descarta porque el costo es bajo frente al riesgo de perder estado local.

5. `--dry-run` compartirá el mismo planner que apply real y prohibirá toda escritura, incluyendo backups y reparación de estado.
   - Rationale: dry-run debe ser evidencia confiable del plan sin efectos laterales.
   - Alternativa considerada: permitir reparación no destructiva del manifest durante dry-run. Se descarta porque viola la expectativa de cero mutaciones.

## Risks / Trade-offs

- Estado ausente en proyectos antiguos → `sync` no puede probar propiedad segura del asset. Mitigación: reportar conflicto/acción bloqueante y recomendar `install`, `verify` o restore según diagnóstico.
- Manifest corrupto bloquea más casos que una comparación directa de archivos. Mitigación: fallar de forma accionable y no asumir seguridad cuando falta evidencia.
- Reusar planner de install puede acoplar reglas de instalación inicial y sync. Mitigación: mantener una intención/modo de operación explícito para que el mismo motor clasifique acciones con políticas distintas.
- Backups obligatorios consumen espacio. Mitigación: registrar manifest portable y ubicar backups bajo `.lufy-ai/backups/<timestamp>/` para limpieza manual segura posterior.

## Migration Plan

1. Introducir parsing y ayuda del comando `sync` sin cambiar el wrapper Bash.
2. Reusar o extraer lógica común de catálogo, path safety, hashing, manifest y backup para que `install` y `sync` compartan invariantes.
3. Implementar planner de sync y apply real con `--dry-run` sin mutaciones.
4. Actualizar estado tras sync exitoso solo para assets aplicados o verificados como gestionados.
5. Validar con `go test ./...`, `go build ./cmd/lufy-ai`, dry-run en temp dir, sync real en temp dir y verify posterior cuando el toolchain Go esté disponible.

## Open Questions

- Ninguna bloqueante para apply. Si durante implementación aparece un alias existente equivalente a `sync`, debe conservarse `sync` como interfaz preferida salvo decisión explícita distinta.
