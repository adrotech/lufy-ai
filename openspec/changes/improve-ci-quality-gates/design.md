## Context

`lufy-ai` tiene una CLI Go en `tools/lufy-cli-go`, workflows de instalación/release y validación agrupada local mediante `scripts/validate.sh`. La raíz no tiene `package.json` ni toolchain Node/TS de producto, por lo que los nuevos gates deben basarse en comandos reales del alcance Go/shell y fallar de forma accionable.

La CI actual debe seguir siendo rápida y estable para PRs, pero puede separar gates obligatorios de checks opcionales/post-release para no bloquear desarrollo diario por dependencias externas o releases aún no publicados.

## Goals / Non-Goals

**Goals:**
- Medir cobertura Go con threshold inicial defendible.
- Ejecutar lint Go mínimo reproducible.
- Validar scripts shell críticos con ShellCheck en CI.
- Cubrir portabilidad Go en Linux, macOS y Windows sin correr smokes incompatibles en plataformas no soportadas.
- Definir E2E post-release contra GitHub Releases como gate separado y no como requisito de cada PR.
- Añadir tests de regresión para output de plan y runtime CLI cuando sea acotado.
- Mantener `scripts/validate.sh` como entrypoint local sin inventar comandos raíz.

**Non-Goals:**
- No introducir Node/TS global para validar este repo.
- No exigir release publicado para validar cada PR.
- No convertir ShellCheck/golangci-lint local en requisito si la herramienta no está instalada; CI puede instalar/proveer herramientas.
- No cambiar el contrato público del CLI salvo tests de regresión que capturen comportamiento existente.

## Decisions

1. Separar gates por criticidad.
   - PR gate: whitespace PR-aware, action pinning, Go tests/build, coverage, lint básico y shellcheck.
   - Platform gate: matriz OS para tests/build, con smokes completos solo donde el entorno lo soporte.
   - Release/E2E gate: workflow manual o post-release para artifacts publicados.

2. Threshold de coverage inicial conservador.
   - Rationale: iniciar con un umbral alcanzable evita un bloqueo artificial y crea una línea base para mejorar gradualmente.
   - El threshold debe vivir en script/config para ajustar sin tocar lógica de tests.

3. Lint Go acotado.
   - Rationale: `golangci-lint` puede ser ruidoso si se activa con demasiadas reglas. Empezar con configuración mínima reduce churn.
   - Si `golangci-lint` no está disponible localmente, `scripts/validate.sh` puede reportarlo como omitido o usar `go vet`, mientras CI lo instala explícitamente.

4. ShellCheck solo para scripts shell versionados.
   - Rationale: evita analizar snippets dentro de Markdown o YAML y mantiene el alcance claro.

5. E2E post-release separado.
   - Rationale: descargar artifacts desde GitHub Releases depende de un tag publicado y red externa; debe vivir en workflow separado/manual/post-release, no en PR normal.

## Risks / Trade-offs

- [Lint/coverage pueden introducir ruido inicial] -> Mitigar con reglas mínimas y threshold bajo documentado.
- [Matrix CI aumenta tiempo/costo] -> Mitigar separando smokes completos de tests/build portables.
- [ShellCheck puede no existir localmente] -> CI lo provee; local reporta limitación si no está instalado.
- [E2E post-release depende de GitHub Releases/red] -> Mantenerlo fuera del gate PR normal.

## Migration Plan

1. Crear scripts/config mínima para coverage, lint y shellcheck.
2. Actualizar `go-cli-install.yml` para ejecutar gates nuevos y matriz multi-OS donde aplique.
3. Añadir tests golden/runtime acotados.
4. Agregar workflow o documentación para E2E post-release.
5. Ejecutar validación agrupada local y evidencia CI antes de delivery.

## Open Questions

- Threshold inicial exacto de coverage tras medir la línea base actual.
- Si `golangci-lint` se instala como action pineada o vía `go run`/binary descargado en CI.
- Qué smokes deben correr en macOS/Windows sin depender de shell Bash específico.
