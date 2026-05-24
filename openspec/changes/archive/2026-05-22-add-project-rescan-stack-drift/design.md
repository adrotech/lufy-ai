## Context

A1 (`add-stack-aware-project-init`) ya dejó sincronizado el contrato base de `lufy-ai init`: detección de stacks v1, creación de `.opencode/project.yaml`, preservación de overrides y semántica inicial de `--rescan`. A2 parte de ese estado y se limita al delta posterior: hacer que `--rescan` compare la configuración existente con la evidencia actual del repositorio, reporte drift accionable y mantenga idempotencia cuando no hay cambios reales.

El archivo `.opencode/project.yaml` es editable por el usuario, por lo que el rescan no puede tratarlo como asset gestionado por hash ni como archivo reemplazable completo. La implementación futura debe separar evidencia generada por detección, campos user-managed y reporte de drift para evitar pérdida de preferencias.

## Goals / Non-Goals

**Goals:**

- Definir un modelo de drift para `lufy-ai init --rescan` sobre stacks, tooling, CI y metadata generada.
- Producir un reporte estructurado y legible por humanos que distinga `detected`, `applied`, `skipped`, `stale`, `conflict` y próximos pasos.
- Mantener idempotencia: sin drift observable, el comando no reescribe `.opencode/project.yaml`, no crea backups y reporta estado limpio.
- Preservar overrides, notas y preferencias user-managed existentes.
- Tratar evidencia stale como señal accionable, no como autorización para borrar entradas.

**Non-Goals:**

- No redefine la detección stack-aware básica ni agrega stacks v1 nuevos más allá de A1.
- No cambia `install`, `sync`, `verify`, `backup`, `restore` ni `scripts/install.sh`.
- No implementa cleanup destructivo, borrado automático de stacks, eliminación de overrides ni migración obligatoria de archivos del usuario.
- No convierte `.opencode/project.yaml` en asset gestionado por manifest/hash.

## Decisions

1. **Modelo de diff declarativo antes de escribir**
   - Decisión: el rescan futuro debe construir un diff entre `existing project.yaml`, `detected snapshot` y `merged result` antes de escribir.
   - Razón: permite reportar drift y validar idempotencia sin depender de efectos laterales.
   - Alternativa descartada: reescribir siempre el YAML generado; es simple, pero rompe overrides, orden/comentarios y genera ruido.

2. **Campos user-managed por allowlist de generated fields**
   - Decisión: solo actualizar campos explícitamente generados/detectados; todo campo desconocido o preferencia editable se preserva por defecto.
   - Razón: `.opencode/project.yaml` es contrato editable y debe tolerar extensiones futuras.
   - Alternativa descartada: comparar el archivo completo contra un render canónico; produciría falsos drift y overwrites innecesarios.

3. **Stale detection no destructiva**
   - Decisión: cuando un stack configurado ya no tenga marcadores actuales, marcar/reportar estado stale/deprecated o emitir una recomendación, pero no borrar la entrada.
   - Razón: la ausencia de marcadores puede ser temporal o intencional, y la entrada puede contener overrides valiosos.
   - Alternativa descartada: eliminar stacks stale automáticamente; contradice preservación de trabajo del usuario.

4. **Reporte estructurado estable**
   - Decisión: la salida debe incluir items con categoría, severidad, stack/campo/path, estado de aplicación y acción sugerida; la forma exacta puede ser texto tabular/JSON interno, pero los datos deben ser testeables.
   - Razón: permite tests de fixtures y habilita consumidores futuros sin acoplarse a frases libres.
   - Alternativa descartada: solo mensajes narrativos; son difíciles de validar y poco accionables.

5. **Exit codes conservadores**
   - Decisión: drift reportable que puede ser aplicado/preservado exitosamente termina con exit code `0`; errores de parseo, path safety o conflictos que impiden un merge seguro terminan non-zero.
   - Razón: detectar drift no es por sí mismo un fallo si el comando completó el rescan de forma segura.
   - Alternativa descartada: non-zero ante cualquier drift; bloquearía flujos de onboarding y CI informativa.

## Risks / Trade-offs

- **Falsos positivos de drift por formato YAML u orden de campos** → Mitigar comparando estructuras normalizadas y no bytes del archivo.
- **Pérdida accidental de overrides** → Mitigar con merge por campos generados, fixtures con campos desconocidos y pruebas de round-trip.
- **Salida demasiado rígida para usuarios** → Mitigar manteniendo contrato de datos estable y permitiendo presentación humana amigable.
- **Ambigüedad entre stale y removed** → Mitigar nombrando stale/deprecated como estado no destructivo y documentando que cleanup explícito queda fuera de alcance.
- **Conflictos de YAML inválido** → Mitigar fallando non-zero sin escribir y con instrucción accionable de corrección/backup manual.

## Migration Plan

- Implementar detrás del comando existente `lufy-ai init --rescan`, sin cambiar el flujo de `init` inicial.
- Agregar fixtures que partan de archivos A1 válidos y verifiquen solamente el delta A2.
- Mantener compatibilidad con configuraciones existentes: si no hay drift, no hay mutación; si hay drift seguro, preservar campos user-managed.
- Rollback operativo: al no tocar assets gestionados ni hacer cleanup destructivo, revertir el cambio de código devuelve la semántica A1; archivos `project.yaml` preservados siguen siendo válidos.

## Open Questions

- Si se añade salida machine-readable (`--json`) en un slice posterior, debe reutilizar el mismo modelo de reporte sin ampliar este alcance.
- La nomenclatura exacta de estados internos (`stale`, `deprecated`, `missing_evidence`) puede ajustarse en implementación mientras los escenarios observables se mantengan.
