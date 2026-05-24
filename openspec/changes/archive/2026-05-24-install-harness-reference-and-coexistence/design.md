## Context

La CLI Go instala hoy el kit OpenCode/OpenSpec como assets gestionados con manifest SHA-256. Dentro de ese catálogo, `AGENTS.md` se genera desde `AGENTS.md.template` con policy `merge-block`, y tanto `install` como `sync` pueden mutarlo. `verify` lo considera archivo crítico requerido y espera que los assets completos estén presentes en `.lufy-ai/install-state.json`.

Ese modelo es seguro para targets vacíos, pero es invasivo para proyectos que ya tienen su propio `AGENTS.md`, OpenCode u OpenSpec. La nueva frontera de ownership debe ser: Lufy gestiona `lufy-ia.harness.md`; el proyecto destino conserva ownership de `AGENTS.md` y solo contiene una referencia mínima `@lufy-ia.harness.md`.

Restricciones relevantes:

- `scripts/install.sh` permanece como wrapper estricto hacia `tools/lufy-cli-go`; no se reintroduce fallback legacy.
- El catálogo root y el catálogo embebido deben conservar paridad.
- `opencode.json` sigue siendo `merge-json` y preserva claves desconocidas del usuario.
- `openspec/changes` continúa excluido del catálogo.
- No se deben tocar specs principales hasta sync/archive de OpenSpec.

## Goals / Non-Goals

**Goals:**

- Instalar y sincronizar `lufy-ia.harness.md` como asset gestionado por hash, con manifest, backup, restore e idempotencia.
- Mantener `AGENTS.md` como archivo user-owned: install puede crear un archivo mínimo o agregar una referencia única con backup/confirmación; sync no lo corrige silenciosamente.
- Migrar instalaciones legacy donde `AGENTS.md` estaba gestionado sin borrar ni sobrescribir contenido existente.
- Preservar configuraciones OpenCode/OpenSpec propias del proyecto y bloquear conflictos exactos en rutas gestionadas.
- Ajustar verify para validar harness/manifest y referencia en `AGENTS.md` sin exigir hash completo del `AGENTS.md` del usuario.

**Non-Goals:**

- Rediseñar todo el catálogo de assets o las policies existentes fuera de los paths afectados.
- Cambiar puertos, defaults de auth, schema de bases de datos o contratos ajenos al instalador.
- Implementar descarga remota o fallback legacy en `scripts/install.sh`.
- Sobrescribir `AGENTS.md`, `.opencode/` u `openspec/` propios sin una policy segura explícita.

## Decisions

### 1. `lufy-ia.harness.md` será el único contenido Lufy gestionado para instrucciones de agentes

El catálogo incluirá `lufy-ia.harness.md` como asset completo gestionado y versionado por SHA-256. El contenido puede derivarse inicialmente de `AGENTS.md.template`, pero su target y su estado serán independientes de `AGENTS.md`.

**Alternativa considerada:** seguir usando `AGENTS.md` con `merge-block`. Se descarta porque mantiene a Lufy como mutador recurrente del archivo de convenciones del proyecto y complica la coexistencia.

### 2. `AGENTS.md` será user-owned con referencia idempotente

Install manejará `AGENTS.md` mediante una acción especial de referencia, no como asset completo. Si falta, creará un archivo mínimo con `@lufy-ia.harness.md`. Si existe y no contiene la referencia, planificará backup más inserción mínima solo con confirmación/`--yes`. Si ya contiene la referencia exacta, reportará `skip` y no duplicará líneas.

**Alternativa considerada:** exigir al usuario editar manualmente siempre. Se descarta para targets nuevos porque reduce la instalación guiada; se mantiene como acción explícita para sync/upgrade cuando falta la referencia.

### 3. Sync/upgrade no mutará `AGENTS.md` por defecto

`sync` actualizará `lufy-ia.harness.md` cuando cambie upstream y preservará `AGENTS.md` byte-for-byte. Si detecta que falta la referencia, reportará warning o acción explícita sugerida, sin auto-fix silencioso.

**Alternativa considerada:** reutilizar la acción de install para reparar `AGENTS.md` en sync. Se descarta para evitar cambios inesperados en un archivo user-owned durante upgrades.

### 4. Migración legacy será no destructiva y trazable

Si `.lufy-ai/install-state.json` registra `AGENTS.md` como asset gestionado legacy, la CLI no lo reemplazará ni borrará. La migración instalará el harness, preservará el archivo legacy, retirará o marcará la entrada legacy de estado solo si puede hacerlo sin perder trazabilidad, y reportará bloque/acción manual si el contenido requiere revisión.

**Alternativa considerada:** convertir automáticamente el bloque gestionado de `AGENTS.md` en referencia. Se descarta como regla general porque puede eliminar contexto que el usuario ya adoptó o modificó.

### 5. Coexistencia se resuelve por ownership de paths

Archivos propios fuera del catálogo no se tocan. Paths exactos del catálogo con contenido no gestionado o drift no resoluble se bloquean según las policies actuales. `opencode.json` conserva su tratamiento `merge-json`; OpenSpec propio se preserva cuando no coincide con paths gestionados, y los conflictos exactos se reportan antes de escribir.

### 6. Verify distinguirá entre checks gestionados y checks de integración

`verify` fallará si falta `lufy-ia.harness.md`, si no aparece en manifest o si su hash no coincide. Para `AGENTS.md`, verificará existencia y referencia como check de integración accionable, sin exigir manifest ni hash completo para ese archivo.

## Risks / Trade-offs

- **Riesgo: proyectos sin `AGENTS.md` podrían no cargar el harness si la referencia mínima no sigue el formato esperado por OpenCode.** → Mitigar con pruebas/smokes de install y documentación del formato exacto `@lufy-ia.harness.md`.
- **Riesgo: migraciones legacy pueden dejar contenido Lufy duplicado en `AGENTS.md` y harness.** → Mitigar reportando estado legacy y recomendación de limpieza manual; no borrar automáticamente.
- **Riesgo: sync no auto-repara referencia ausente y el usuario podría creer que Lufy quedó activo.** → Mitigar con warning claro, salida JSON accionable y verify que destaque la acción requerida.
- **Riesgo: paridad root/embedded puede romperse al agregar el nuevo asset.** → Mitigar actualizando assets embebidos y tests de paridad en la implementación.
- **Riesgo: cambios en verify pueden relajar demasiado controles sobre `AGENTS.md`.** → Mitigar separando controles: hash estricto para harness, check estructural/referencia para `AGENTS.md`.

## Migration Plan

1. Agregar `lufy-ia.harness.md` al repo fuente y al catálogo embebido con paridad validada.
2. Cambiar el catálogo para retirar `AGENTS.md` como asset completo gestionado y registrar la acción especial de referencia.
3. Implementar planificación/apply de install para crear o insertar referencia en `AGENTS.md` con backup/confirmación.
4. Implementar sync para actualizar harness y reportar referencia ausente sin mutar `AGENTS.md`.
5. Implementar lectura/migración de manifest legacy para `AGENTS.md` gestionado, preservando contenido y emitiendo estado accionable.
6. Actualizar verify/status para checks de harness gestionado y referencia user-owned.
7. Añadir pruebas Go de targets vacío, `AGENTS.md` propio, referencia duplicada, sync, legacy, coexistencia y verify.

Rollback: al ser no destructivo, los backups existentes permiten restaurar `AGENTS.md` si install insertó la referencia. Si el harness nuevo causa problemas, sync/restore puede volver al backup y la referencia puede retirarse manualmente del archivo user-owned.

## Open Questions

- ¿La ausencia de referencia en `AGENTS.md` durante `verify` debe ser `fail` por defecto o `warn` con exit code cero? La propuesta acepta requisito/warning accionable; la implementación debe fijar severidad consistente para salida humana y JSON.
- ¿Se agregará un flag explícito para reparar la referencia durante sync, por ejemplo `--repair-agents-reference`, o se documentará como ejecutar install nuevamente con confirmación?
