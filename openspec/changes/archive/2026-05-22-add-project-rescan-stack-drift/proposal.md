## Why

`lufy-ai init --rescan` ya existe como contrato base para preservar overrides y sumar stacks, pero el siguiente slice necesita hacerlo operable en repositorios que cambian con el tiempo. Este cambio especifica validación de drift, detección de evidencia stale y un reporte estructurado accionable sin convertir el rescan en cleanup destructivo.

## What Changes

- Profundiza la semántica de `lufy-ai init --rescan` para comparar `.opencode/project.yaml` existente contra la evidencia actual del target.
- Agrega detección explícita de drift por stacks nuevos, stacks sin marcadores actuales, tooling/CI cambiado y campos generados desactualizados.
- Define un reporte estructurado y accionable con severidad, categoría, path/campo afectado, acción sugerida y estado de aplicación.
- Mantiene idempotencia cuando no hay drift: el rescan no reescribe archivos ni genera backups innecesarios.
- Distingue detección/reporte de modificación: el rescan puede actualizar o marcar metadata segura, pero MUST NOT borrar stacks, overrides ni archivos del usuario de forma implícita.
- No redefine la detección stack-aware básica de A1 ni agrega nuevos stacks v1 fuera del contrato existente.

## Capabilities

### New Capabilities


### Modified Capabilities

- `project-stack-config`: especifica drift validation, stale detection, reporte accionable e idempotencia avanzada para `lufy-ai init --rescan` sobre `.opencode/project.yaml` existente.
- `go-cli-installer`: especifica la superficie CLI/reporting de `lufy-ai init --rescan` sin cambiar comandos base ni el wrapper de instalación.

## Impact

- Código afectado futuro: `tools/lufy-cli-go/cmd/lufy-ai`, paquetes internos de CLI/init/scan/config/reporting bajo `tools/lufy-cli-go/internal/**`.
- Contrato de usuario afectado: salida de `lufy-ai init --rescan`, exit codes para drift bloqueante/no bloqueante y contenido actualizado de `.opencode/project.yaml`.
- Archivos del target: limitado a `.opencode/project.yaml` y metadata segura asociada al rescan; no borra stacks, overrides, managed assets, backups ni archivos desconocidos.
- Validación futura: fixtures Go para no-drift idempotente, drift de stack nuevo, stack stale, tooling/CI cambiado, YAML inválido y preservación de overrides.
