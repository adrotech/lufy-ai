## Context

El instalador Go ya cubre install, verify, backup, restore, sync, assets embebidos y releases versionadas. La estructura actual tiene cinco riesgos base antes de ampliar release/supply chain: assets canonicos y embebidos pueden divergir, `EnsureRelativeSafe` no cubre escapes con backslash en todos los entornos, `install-state.json` y backups usan metadata hardcodeada, `SourceRootFingerprint` es fijo, y `copyFile` escribe directamente con `os.WriteFile` en install/sync/backup.

Este cambio es transversal a `assets`, `platform`, `state`, `installer`, `syncer`, `backup`, `verify` y CI. Debe preservar el contrato de wrapper estricto de `scripts/install.sh` y no cambiar defaults publicos de instalacion salvo para hacerlos mas seguros.

## Goals / Non-Goals

**Goals:**
- Evitar drift entre assets raiz y `internal/assets/embedded` con una estrategia verificable.
- Bloquear escapes relativos portables, incluyendo `..\foo` y separadores mixtos.
- Persistir metadata real de version/build en state y backup manifest.
- Calcular fingerprint estable del catalogo desde `targetRel + sourceSHA256` ordenado.
- Reusar escritura atomica para archivos gestionados en install, sync y backup.
- Agregar pruebas/validacion final para los cambios estructurales.

**Non-Goals:**
- Migrar a `goreleaser`, cosign, SLSA o SBOM.
- Cambiar el modelo de ramas, tags o releases.
- Redisenar rollback/two-phase apply completo.
- Introducir nuevas dependencias externas si una solucion stdlib es suficiente.
- Reintroducir fallback legacy en `scripts/install.sh`.

## Decisions

1. Mantener raiz como fuente canonica por ahora y validar paridad embedded.

   Rationale: es la opcion minima para resolver drift sin mover toda la estructura del repo. `go:embed` no puede referenciar fuera del modulo, por lo que el mirror embebido sigue existiendo, pero debe validarse.

   Alternative considered: mover la fuente canonica a `internal/assets/embedded` y generar la raiz. Se descarta para esta ola porque invierte el flujo editorial del repo y requiere cambios documentales mas amplios.

2. Agregar helper de fingerprint al catalogo.

   Rationale: tanto `state.New` como tests pueden consumir un hash estable derivado del catalogo real, sin depender de nombres de proposals ni strings hardcodeados.

   Alternative considered: usar commit SHA como fingerprint. Se descarta porque un binario embebido puede instalar assets sin checkout Git y el fingerprint debe describir el catalogo efectivo.

3. Usar `version.Current()` como metadata runtime para state/backup.

   Rationale: ya existe metadata inyectada por ldflags y `lufy-ai version`; el state debe reflejar la misma fuente.

   Alternative considered: mantener constantes `ToolVersion`/`SourceChangeID`. Se descarta porque impide trazabilidad post-mortem en releases reales.

4. Endurecer `EnsureRelativeSafe` normalizando separadores antes de evaluar escapes.

   Rationale: `SafeJoin` cubre muchos casos, pero `EnsureRelativeSafe` se usa directamente con datos de catalogo, state y backup manifest; debe ser seguro por si mismo.

   Alternative considered: confiar solo en `SafeJoin`. Se descarta porque deja usos directos vulnerables o inconsistentes.

5. Crear una utilidad interna para escritura atomica de archivos.

   Rationale: `state.WriteAtomic` ya demuestra el patron correcto; extraer o duplicar un helper pequeno reduce corrupcion parcial sin redisenar apply completo.

   Alternative considered: implementar two-phase apply completo. Se descarta para esta ola por mayor alcance; quedara como backlog separado.

## Risks / Trade-offs

- Riesgo: test de paridad falle inmediatamente por drift existente. Mitigacion: actualizar el mirror embebido en la misma propuesta o documentar comando de sync/generacion si se introduce.
- Riesgo: cambiar `EnsureRelativeSafe` puede rechazar paths antes aceptados. Mitigacion: los paths gestionados son relativos controlados; rechazar escapes o separadores ambiguos es el comportamiento deseado.
- Riesgo: fingerprint cambie por orden no deterministico. Mitigacion: ordenar por `TargetRel` y serializar solo campos estables.
- Riesgo: atomic write preserve permisos de forma distinta. Mitigacion: usar modo esperado actual `0o644`, temp file en el mismo directorio y rename atomico.
- Riesgo: metadata de version en tests sea `dev`. Mitigacion: aceptar `dev` para builds locales, pero verificar que proviene de `version.Current()` y no de constantes congeladas.

## Migration Plan

1. Analizar al inicio paquetes afectados y tests existentes.
2. Implementar path safety portable y tests.
3. Implementar catalog fingerprint y paridad raiz/embedded.
4. Propagar version metadata a state y backup manifest.
5. Reemplazar escrituras directas de archivos gestionados por escritura atomica.
6. Actualizar tests y CI local disponible.
7. Ejecutar validacion final agrupada desde `tools/lufy-cli-go/` y `git diff --check`.

Rollback: revertir cambios de codigo/tests y artefactos OpenSpec. Los fields existentes de state mantienen nombres actuales; no se requiere migracion destructiva de targets existentes.

## Open Questions

- Si se detecta drift raiz/embedded durante implementacion, decidir si se resuelve copiando manualmente el mirror o agregando un helper `go generate` en esta misma ola.
