## Why

La configuracion inicial de LUFY esta fragmentada en varios comandos (`install`, `init`, `memory init`, `context build`, seleccion de metodologia y validacion). Eso obliga al usuario a conocer el orden correcto y hace que upgrades con nuevas capacidades no tengan un punto unico para descubrir y activar configuracion faltante.

`lufy-ai setup` debe convertirse en el flujo guiado y repetible que primero verifica si existe una version nueva de LUFY AI, sugiere actualizar cuando corresponde, y luego planifica/aplica las capacidades de instalacion y configuracion pendientes sin romper los comandos scriptables existentes.

## What Changes

- Agregar comando `lufy-ai setup` como orquestador de configuracion end-to-end.
- Ejecutar un preflight de version al inicio del setup.
- Reportar version local, ultima version estable disponible y recomendacion de upgrade antes de cualquier mutacion.
- Detectar features configurables pendientes: assets instalados, project config, memoria Obsidian, context graph y verificacion final.
- Detectar features configurables pendientes mediante registry versionado: layout `.lufy`, assets instalados, project config, stack/superficies, metodologia SDD, memoria Obsidian, context graph y verificacion final.
- Soportar `--dry-run`, `--yes`, `--json`, `--skip-version-check`, `--require-latest` y `--check-new-features`.
- Ofrecer selector interactivo Bubble Tea cuando hay TTY y el usuario no pasa `--yes`, `--dry-run` ni `--json`.
- Reutilizar servicios existentes de install, projectconfig, memory, context y verify en vez de duplicar logica.
- Hacer que `upgrade` sugiera ejecutar `setup --check-new-features` despues de actualizar el binario.

## Non-Goals

- No eliminar ni cambiar la semantica de comandos existentes.
- No actualizar automaticamente el binario durante `setup` sin comando explicito de upgrade.
- No instalar todas las futuras features por defecto si requieren decision explicita o tienen drift/conflictos.
- No cambiar defaults publicos de tool/metodologia fuera de lo que el setup planifique de forma visible.

## Review Slices

### Slice 1: Version preflight

- Objetivo: extraer un servicio de consulta/comparacion de ultima release y usarlo como primer paso del setup.
- Archivos esperados: `internal/versioncheck/*`, tests unitarios, CLI setup.
- Criterios:
  - WHEN existe una version remota mas nueva, THEN setup la reporta y recomienda `lufy-ai upgrade --to <version>` antes de continuar.
  - WHEN la version local esta al dia, THEN setup reporta que LUFY AI esta al dia y continua.
  - WHEN `--require-latest` esta activo y hay update, THEN setup falla antes de mutar.

### Slice 2: Setup plan y features

- Objetivo: construir un plan repetible de features configurables pendientes.
- Archivos esperados: `internal/setup/*`, tests unitarios.
- Criterios:
  - WHEN faltan assets instalados, THEN el plan incluye install.
  - WHEN falta memoria inicializada, THEN el plan incluye memory init.
  - WHEN falta o esta stale el context graph, THEN el plan incluye context build.
  - WHEN todo esta listo, THEN el plan reporta acciones skip/noop.
  - WHEN el usuario pide `--check-new-features`, THEN el plan expone features con metadata `since` para orientar capacidades posteriores a upgrade.

### Slice 3: CLI y aplicacion

- Objetivo: exponer `lufy-ai setup` con salida humana y JSON, dry-run sin mutaciones y aplicacion con `--yes`.
- Archivos esperados: `internal/cli/app.go`, `internal/setup/*`, tests CLI.
- Criterios:
  - WHEN setup corre con `--dry-run`, THEN no escribe archivos y muestra el plan.
  - WHEN setup corre con `--yes`, THEN aplica install/memory/context segun corresponda y ejecuta verify final.
  - WHEN setup corre con `--json`, THEN emite un reporte parseable sin logs humanos mezclados.
  - WHEN setup corre en TTY sin flags de automatizacion, THEN muestra checklist Bubble Tea para elegir acciones a aplicar.

### Slice 4: Upgrade hints

- Objetivo: despues de upgrade exitoso sugerir setup para descubrir nuevas features.
- Archivos esperados: `internal/upgrade/upgrade.go`, tests existentes/ajustados.
- Criterios:
  - WHEN upgrade reemplaza el binario, THEN la salida recomienda `lufy-ai setup --target . --check-new-features`.
  - WHEN upgrade corre en dry-run, THEN no promete configuracion aplicada.

## Validation

- `openspec validate "add-setup-wizard-and-version-preflight" --strict`
- `go test ./...` desde `tools/lufy-cli-go`
- `scripts/validate.sh`
- Smokes de `lufy-ai setup --target <temp> --dry-run --skip-version-check` y `lufy-ai setup --target <temp> --yes --skip-version-check`.
