## Why

`project_profile.surfaces` ya describe frontend, backend, fullstack y otras superficies, pero los agentes deben interpretar manualmente roots, conexiones y listas textuales de validación. Eso vuelve implícito el cambio de contexto entre capas y permite planes inconsistentes ante el mismo diff.

Lufy necesita convertir el perfil declarado en un plan read-only, determinístico y explicable que seleccione superficies, componga flujos cross-surface y traduzca expectativas en gates tipados sin ejecutar comandos automáticamente.

## What Changes

- Añadir un motor `surfaceplan` con modelos neutrales, puertos pequeños, Strategy de resolución y Factory de validaciones.
- Resolver la superficie activa desde `--surface`, archivos explícitos o `git diff`, usando precedencia estable y decisiones explicables.
- Componer una superficie `fullstack` cuando el cambio alcance superficies conectadas o un contrato compartido.
- Producir reglas de validación tipadas con evidencia esperada y comandos descubiertos, sin ejecutar comandos ni saltarse allowlists.
- Incorporar capacidades opcionales por superficie para políticas reutilizables como `realtime`, `rendering`, `offline`, `persistent-state` y `desktop-shell`.
- Exponer `lufy-ai plan` con salida humana y JSON, e integrarlo al command palette.
- Documentar el plan como contrato consumible por agentes y automatización futura.

## Non-Goals

- Ejecutar las validaciones generadas.
- Reemplazar `context graph` ni inferir semántica con un LLM.
- Implementar todavía telemetría del Agent Observatory.
- Agregar lógica hardcodeada para videojuegos o frameworks concretos al dominio.
- Cambiar tiers, autorización de delivery o metodología efectiva.

## Review Slices

### Slice 1: Contrato y resolución

- Objetivo: modelos, puertos y resolución determinística de superficies.
- Archivos esperados: `internal/surfaceplan/*`, pruebas unitarias.
- WHEN un diff toca una única root específica, THEN el plan selecciona esa superficie.
- WHEN toca frontend y backend conectados, THEN compone el flujo fullstack.
- Riesgo: roots amplias como `.`; se priorizan coincidencias específicas y composición segura ante empate.

### Slice 2: Validación y capacidades

- Objetivo: transformar expectativas y señales de riesgo en reglas tipadas y deduplicadas.
- Archivos esperados: `internal/surfaceplan/*`, `internal/projectconfig/*`, pruebas.
- WHEN cambia un contrato, THEN aparecen gates de contrato y E2E además de los de cada superficie.
- WHEN una superficie declara capacidades realtime/offline/persistencia, THEN aparecen evidencias específicas sin acoplarse a un juego.
- Riesgo: sobrevalidación; cada regla conserva trigger y razón para revisión humana.

### Slice 3: CLI, palette y documentación

- Objetivo: comando estable humano/JSON y consumo interactivo.
- Archivos esperados: `internal/cli/*`, `internal/tui/commandpalette/*`, documentación.
- WHEN se ejecuta `lufy-ai plan`, THEN no modifica el repositorio target.
- WHEN la entrada es inválida o ambigua sin composición, THEN retorna un error accionable.
- Riesgo: contrato público nuevo; cubrir ayuda, JSON y códigos de salida.

## Validation

- `go test ./internal/surfaceplan ./internal/projectconfig ./internal/tui/commandpalette ./internal/cli`
- `go test ./...`
- `scripts/validate.sh`
- `openspec validate "add-surface-aware-execution-plans" --strict`
- Verificación manual de salida humana y JSON sobre fixtures frontend, backend y fullstack.

