## Why

La instalación actual trata `AGENTS.md` como asset gestionado con contenido Lufy embebido, lo que aumenta el riesgo de sobrescribir convenciones propias del proyecto destino y dificulta la coexistencia con OpenCode/OpenSpec ya instalados. Este cambio separa el harness gestionado de Lufy en `lufy-ia.harness.md` y deja `AGENTS.md` como archivo del usuario que solo referencia ese harness.

## What Changes

- Agregar `lufy-ia.harness.md` como asset gestionado raíz, registrado por SHA-256 en `.lufy-ai/install-state.json`, con backup, sync, restore, verify e idempotencia como cualquier asset completo gestionado.
- Cambiar la semántica de `AGENTS.md`: deja de ser contenido completo gestionado por hash y pasa a ser user-owned con una referencia mínima `@lufy-ia.harness.md`.
- En install inicial, crear un `AGENTS.md` mínimo cuando no exista; si existe, agregar solo la referencia con backup y confirmación/`--yes`; si la referencia ya existe, omitir sin duplicarla.
- En sync/upgrade, actualizar solo `lufy-ia.harness.md` y no mutar `AGENTS.md`; si falta la referencia, reportar warning o acción explícita en vez de auto-fix silencioso.
- Migrar instalaciones legacy donde `AGENTS.md` aparece como asset gestionado sin sobrescribir ni borrar el archivo existente: retirar o marcar su estado legacy, instalar el harness y reportar bloque/acción legado cuando aplique.
- Preservar OpenCode/OpenSpec propios del target, archivos extra y configuraciones de usuario; bloquear conflictos exactos en rutas gestionadas en vez de pisarlos.
- Ajustar verify para exigir harness gestionado y manifest, y validar la referencia en `AGENTS.md` como requisito o warning accionable sin comparar hash completo del archivo.

## Capabilities

### New Capabilities
- `install-harness-reference-and-coexistence`: instalación, sync, migración y verificación del harness Lufy gestionado por referencia desde `AGENTS.md` user-owned, incluyendo coexistencia segura con OpenCode/OpenSpec propios.

### Modified Capabilities
- `managed-assets-install`: cambia el catálogo y la semántica de `AGENTS.md`, agregando `lufy-ia.harness.md` como asset gestionado y definiendo reglas de install/sync/verify/migración/coexistencia.
- `go-cli-installer`: cambia el contrato observable de install, sync y verify de la CLI Go para tratar `AGENTS.md` como archivo user-owned con referencia mínima en vez de asset completo crítico con hash.

## Impact

- Código esperado: `tools/lufy-cli-go/internal/assets/catalog.go` y catálogo embebido, planner/apply de install/sync, verificación, manifest/state migration y pruebas Go de paridad root/embedded.
- Assets esperados: nuevo `lufy-ia.harness.md` en raíz del repo fuente y copia embebida correspondiente; `AGENTS.md.template` deja de ser la fuente completa gestionada para targets.
- Comportamiento de usuario: upgrades más seguros en proyectos con `AGENTS.md`, OpenCode u OpenSpec propios; posibles warnings/bloqueos accionables durante migración legacy.
- No cambia `scripts/install.sh`: permanece wrapper estricto hacia la CLI Go sin fallback legacy.
