## 1. Análisis local y modelo de drift

- [x] 1.1 Revisar implementación existente de `lufy-ai init`, lectura/escritura de `.opencode/project.yaml` y fixtures A1 sin cambiar comandos no relacionados.
- [x] 1.2 Definir estructuras internas para snapshot detectado, configuración existente, diff/plan de merge y items de reporte con categoría, severidad, path/campo, estado y acción sugerida.
- [x] 1.3 Delimitar campos generated/detected actualizables y campos user-managed que deben preservarse por defecto.

## 2. Rescan stack-aware y merge seguro

- [x] 2.1 Implementar comparación de drift para stacks nuevos, tooling cambiado y CI cambiado usando la detección existente de A1.
- [x] 2.2 Implementar stale detection no destructiva para stacks configurados cuyos marcadores actuales ya no existen.
- [x] 2.3 Implementar merge de `.opencode/project.yaml` que preserve overrides, campos desconocidos y preferencias user-managed.
- [x] 2.4 Garantizar que el caso no-drift no reescriba `.opencode/project.yaml`, no genere backups y no modifique install state ni assets gestionados.

## 3. CLI y reporte accionable

- [x] 3.1 Conectar `lufy-ai init --rescan` al flujo de drift/merge/reporting manteniendo `main.go` como entrada delgada.
- [x] 3.2 Actualizar ayuda de `lufy-ai init` para describir `--rescan` como refresh con preservación de overrides y sin cleanup destructivo.
- [x] 3.3 Emitir reporte humano estructurado con items de drift, applied/skipped/detected, sugerencias accionables y estado limpio cuando no hay drift.
- [x] 3.4 Manejar errores de YAML/config inválida con exit code non-zero, mensaje accionable y cero escrituras.

## 4. Validación y documentación mínima

- [x] 4.1 Agregar fixtures/tests Go para no-drift idempotente, stack nuevo, tooling drift, CI drift, stale stack, YAML inválido y preservación de campos desconocidos/overrides.
- [x] 4.2 Ejecutar validación agrupada disponible (`scripts/validate.sh` o comandos Go equivalentes si aplica) y registrar evidencia real.
- [x] 4.3 Actualizar documentación o README de CLI solo si existe una sección relevante para `lufy-ai init --rescan`.
