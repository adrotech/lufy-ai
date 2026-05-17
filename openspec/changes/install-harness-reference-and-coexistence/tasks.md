## 1. Assets y catálogo

- [x] 1.1 Crear `lufy-ia.harness.md` raíz con el contenido gestionado de Lufy derivado de la fuente actual apropiada.
- [x] 1.2 Agregar `lufy-ia.harness.md` al catálogo root de la CLI Go con policy de asset completo gestionado, scope correcto y SHA-256 en manifest.
- [x] 1.3 Actualizar el catálogo/assets embebidos para incluir `lufy-ia.harness.md` con paridad root/embedded.
- [x] 1.4 Retirar `AGENTS.md` del tratamiento de asset completo gestionado por hash y modelarlo como integración user-owned de referencia.

## 2. Install e idempotencia de referencia

- [x] 2.1 Implementar planificación de install para target vacío: instalar harness, crear `AGENTS.md` mínimo con `@lufy-ia.harness.md`, assets restantes y manifest.
- [x] 2.2 Implementar inserción idempotente de referencia en `AGENTS.md` existente con backup y confirmación/`--yes`, sin insertar contenido completo de Lufy.
- [x] 2.3 Implementar skip cuando `AGENTS.md` ya contiene `@lufy-ia.harness.md`, sin duplicar referencia ni reescribir el archivo.
- [x] 2.4 Asegurar que `AGENTS.md` no se registre en `.lufy-ai/install-state.json` como asset completo y que `lufy-ia.harness.md` sí se registre con hashes.

## 3. Sync, upgrade y migración legacy

- [x] 3.1 Implementar sync de `lufy-ia.harness.md` usando manifest, SHA-256, backup, update-managed e idempotencia existentes.
- [x] 3.2 Garantizar que sync preserve `AGENTS.md` byte-for-byte y solo reporte warning/acción explícita si falta `@lufy-ia.harness.md`.
- [x] 3.3 Detectar manifests legacy donde `AGENTS.md` aparece como gestionado y migrar o marcar esa entrada sin sobrescribir ni borrar el archivo.
- [x] 3.4 Reportar conflicto o bloque legado accionable cuando `AGENTS.md` legacy tenga drift local o requiera revisión manual.

## 4. Coexistencia OpenCode/OpenSpec

- [x] 4.1 Preservar archivos OpenCode/OpenSpec propios fuera del catálogo sin borrarlos, modificarlos ni registrarlos.
- [x] 4.2 Bloquear conflictos exactos en paths gestionados existentes no registrados cuando la policy no permita merge/adopción segura.
- [x] 4.3 Mantener `opencode.json` como `merge-json` preservando claves desconocidas y sin registro de hash completo.
- [x] 4.4 Confirmar que `openspec/changes` sigue excluido del catálogo y no se instala en targets.

## 5. Verify, status y salida accionable

- [x] 5.1 Actualizar verify para exigir `lufy-ia.harness.md` seguro, presente en manifest y con hash correcto.
- [x] 5.2 Actualizar verify para validar `AGENTS.md` como referencia user-owned sin exigir entrada de manifest ni hash completo.
- [x] 5.3 Definir e implementar severidad consistente para referencia ausente (`fail` o `warn`) en salida humana y JSON.
- [x] 5.4 Incluir recomendaciones accionables para agregar la referencia o resolver estado legacy sin mutar el target durante verify.

## 6. Pruebas y validación

- [x] 6.1 Agregar pruebas Go para target vacío: harness, `AGENTS.md` mínimo, assets y manifest.
- [x] 6.2 Agregar pruebas Go para `AGENTS.md` propio: inserta solo referencia con backup y no copia contenido completo.
- [x] 6.3 Agregar pruebas Go para referencia ya presente: no duplica y no reescribe.
- [x] 6.4 Agregar pruebas Go para sync: actualiza harness y no modifica `AGENTS.md`.
- [x] 6.5 Agregar pruebas Go para migración legacy de `AGENTS.md` gestionado no destructiva.
- [x] 6.6 Agregar pruebas Go para coexistencia OpenCode/OpenSpec y bloqueo de conflictos exactos.
- [x] 6.7 Agregar pruebas Go para verify: harness/manifest estricto y referencia `AGENTS.md` sin hash completo.
- [x] 6.8 Ejecutar validación agrupada disponible para el scope, incluyendo `go test ./...` y `go build ./cmd/lufy-ai` desde `tools/lufy-cli-go/` cuando corresponda, más chequeos de paridad de assets.

## 7. Documentación y compatibilidad

- [x] 7.1 Actualizar documentación de instalación/sync/verify para explicar `lufy-ia.harness.md`, referencia desde `AGENTS.md` y comportamiento de warnings.
- [x] 7.2 Verificar que `scripts/install.sh` permanece wrapper estricto hacia la CLI Go sin fallback legacy ni lógica propia de assets.
