# Bloque 2: configuración y catálogo inline

Estado: implementado y validado en el checkout de ensayo. La instalación personal todavía no se sustituyó.

Base: `origin/develop` en `1cc18d4647afc61ea1f6b1cb726ba74128fded5f` (`v0.6.25`). Rama local `codex/lufy-inline`, checkout `/private/tmp/lufy-inline-20261026`. Ejecución inline, sin delegación, commit, push ni PR.

## Resultado

El modelo de configuración y el manifest representan ejecución inline directamente. Se eliminó `MethodologyByTier`, su resolución por tier, su registro de metodologías y sus flags. No hay aliases de tipos ni lectores de formatos anteriores para ese modelo.

```yaml
schema_version: 2
tool: codex
execution:
  mode: inline
  delegation: auto
```

El manifest usa `schemaVersion: 3` y el mismo objeto `execution`. `auto`, `on_request` y `off` controlan la política de delegación; no conceden permisos del host. El pedido explícito de trabajar inline prevalece. `install` y `sync` aceptan `--delegation`.

La precedencia es selección explícita, configuración del proyecto, manifest vigente y defaults. Un desacuerdo entre proyecto y manifest se diagnostica antes de mutar; una selección explícita permite reconciliarlo. Se conservan perfiles, superficies, comandos de validación, memoria y valores propios del proyecto.

## Cambios principales

- `core/domain`, `projectconfig`, `harnessconfig`, `state` y consumidores: un solo modelo vigente. Se rechazan schemas anteriores, campos metodológicos mezclados, políticas desconocidas, keys desconocidas de execution y YAML de varios documentos. También se rechaza un campo retirado introducido por un alias YAML. Los rechazos no reescriben la instalación.
- `harnesscatalog`, `assets`, registry y adapter Codex: selección explícita del conjunto inline. El inventario y el descubrimiento de fuentes ya no necesitan `.opencode` ni OpenSpec. El registry activo sólo ofrece Codex.
- Instrucciones del principal, guía de ejecución, contrato de entrega y mirrors: se aplica `work.md` sin router obligatorio, tier, cadena de roles ni handoffs como requisito para avanzar.
- Catálogo seleccionado: política de trabajo y entrega, configuración/hooks/rules de Codex, referencia del principal y cuatro capacidades de memoria. No instala agentes de la cadena anterior, comandos o skills metodológicos, specs ni contratos de Result Contract.
- Hooks: únicamente orientación `SessionStart`. El verificador comprueba el perfil vigente y drift de `.codex` y contratos neutrales.
- Instalación/sync: se retiraron planificación y ejecución del merge de configuración de OpenCode del camino nuevo. La prevalidación no migra automáticamente layouts anteriores.
- Tests: se reemplazaron expectativas del producto retirado por aceptación del modelo actual. Se eliminaron comprobaciones textuales que exigían el antiguo proceso. No se conservó una API sólo para compilar tests históricos.
- Infraestructura de validación: el guard reconoce `.lufy/README.md` como asset fuente versionable; el chequeo de acoplamiento distingue nombres de identificadores de menciones reales y tolera inventarios con cero coincidencias.

## Fallo encontrado y corregido

La prueba integrada encontró que cambiar sólo `execution.delegation` mediante `sync` actualizaba el proyecto pero no el manifest si todos los assets eran `skip`. Se reprodujo el fallo con una regresión Go (RED), se agregó la actualización de estado ante cambios de ejecución y se comprobó el resultado (GREEN y smoke integrado).

## Evidencia

- `GOCACHE=/private/tmp/lufy-inline-go-cache scripts/validate.sh`: PASS, exit 0. Incluye whitespace contra `origin/develop`, PR guard local, 169 tests Python offline, pinning, versión/release, YAML, acoplamiento, shell, smoke de hooks, toda la suite Go, coverage, `go vet` y build. Coverage total: **80,0%**, con mínimo conservado de 80,0%.
- Regresión `go test ./internal/syncer -run TestSyncPreservesCodexHarnessAndReconcilesExplicitMismatch -count=1`: fallo observado antes de la corrección y PASS después.
- Smoke con binario compilado actual: PASS en 13 comprobaciones. Instalación nueva, schemas vigentes, assets seleccionados, un hook de orientación, info/status/verify, idempotencia, rechazo de drift sin escrituras, persistencia de política sin cambios de assets, valores propios preservados, formatos anteriores rechazados sin escrituras, fuente sin hosts históricos, distribución embebida y desinstalación/reinstalación.
- `git diff --check origin/develop`: PASS. Archivos nuevos inspeccionados también fuera de Git, porque el diff no incluye untracked por sí solo.
- Registro de skills instalado: ready, cuatro capacidades. `quick_validate.py` no estuvo disponible por ausencia de PyYAML; la estructura se verificó mediante el parser y registro reales del CLI.
- Memoria Obsidian: 13 notas, cero backlinks rotos después de guardar la lección de persistencia de ejecución.

Las suites Go que usan `httptest` se ejecutaron con permiso para bind local en loopback. No se lanzaron modelos, casos nativos adicionales ni operaciones externas de entrega.

## Límites y continuación

Este bloque activa el catálogo nuevo para instalaciones de ensayo. El retiro general de motores, comandos explícitos y consumidores de evidencia anteriores continúa en los bloques 3 y 6; no participan en la selección de metodología o catálogo nuevo. No se afirma que se haya terminado toda la refactorización.

La sustitución de la instalación personal corresponde al bloque 4. El workspace original tiene cambios previos y una base diferente, por lo que sus fuentes solapadas no se sobrescribieron. Allí se guardan este informe, evidencia y un patch reproducible del trabajo sobre develop. La evaluación de comportamiento de varias familias de tareas sigue siendo un trabajo distinto de estas pruebas deterministas.

## Archivos de respaldo

- `bloque-2-validation.log`: salida completa de la validación final.
- `bloque-2-smoke.py` y `bloque-2-smoke-result.json`: procedimiento y resultado de la prueba integrada (el script usa rutas locales de este ensayo).
- `bloques-1-y-2.patch`: cambios de código e instrucciones sobre la base develop, incluyendo los archivos nuevos de los dos bloques. El patch no fue aplicado al checkout original.
