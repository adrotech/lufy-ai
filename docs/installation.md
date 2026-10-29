# Instalación, sincronización y recuperación

El modelo inline está integrado en `develop` y todavía no tiene release publicada. La referencia estable `v0.6.25` corresponde al modelo anterior. Compilá un checkout actualizado de `develop`; no descargues ese stable para reemplazar una instalación inline.

## Compilar e instalar

Requiere Go según `tools/lufy-cli-go/go.mod` (actualmente 1.24.2). Desde la raíz, en un shell POSIX:

```sh
mkdir -p "$HOME/.local/bin"
go -C tools/lufy-cli-go build -o "$HOME/.local/bin/lufy-ai" ./cmd/lufy-ai
export PATH="$HOME/.local/bin:$PATH"
lufy-ai version --runtime-contract
lufy-ai install --target /ruta/al/proyecto --tool codex --delegation off --dry-run
lufy-ai install --target /ruta/al/proyecto --tool codex --delegation off --yes
lufy-ai verify --target /ruta/al/proyecto --deep
```

El ejemplo supone un ejecutable de destino libre o un reemplazo previamente respaldado de la misma generación. Para Windows, compilá `./cmd/lufy-ai` a una ruta persistente `lufy-ai.exe` y agregá su directorio al PATH del host. Go se comprueba en Linux, macOS y Windows; los smokes Bash de instalación corren en Ubuntu.

Conservá el PATH también en el entorno desde el que se inicia Codex: el hook busca `lufy-ai`. El ejemplo fija delegación `off`; el default es `auto`, con `on_request` como tercera opción. Ver [configuración](configuration.md).

La superficie soportada es el proyecto con Codex. Se distribuyen 16 assets y se genera el bloque gestionado de `AGENTS.md`; una guía existente se preserva fuera de ese bloque. `scripts/install.sh` es un wrapper estricto de la CLI Go local/PATH y no descarga ni ejecuta motores alternativos.

`init` y `scan` detectan configuración; `setup` orquesta además memoria, grafo y verificación. Mientras stable pertenece al modelo anterior, para probar el checkout local con `setup` usá `--skip-version-check` y revisá primero `--dry-run`.

## Sincronizar y resolver drift

```sh
lufy-ai sync --target /ruta/al/proyecto --dry-run
lufy-ai conflicts plan --target /ruta/al/proyecto --json
lufy-ai sync --target /ruta/al/proyecto --yes
lufy-ai verify --target /ruta/al/proyecto --deep
```

Ownership, hashes y policies determinan qué actualizar. Los archivos ajenos se preservan. Una edición local puede generar conflicto y un candidato `.lufy-new`; inspeccioná ambos antes de resolver con `merge`. `--accept-ours` conserva el local y `--accept-theirs` adopta el candidato; ninguna elección es automática. `pin <path>` congela un asset gestionado y `unpin <path>` permite volver a sincronizarlo.

`sync` no agrega automáticamente assets nuevos ausentes que nunca estuvieron en el manifest. Una ruta nueva del catálogo que colisione con un archivo local se trata como conflicto. Ver [ADR 0004](adr/0004-sync-new-catalog-assets.md). Los extras no adquieren ownership por estar en directorios gestionados.

## Formatos y backups

| Dato | Formato actual |
| --- | --- |
| `.lufy/config/project.yaml` | Schema 2, `codex`, ejecución `inline` |
| `.lufy/managed-state/install-state.json` | Schema 4; ownership, hashes y ancestors |
| `.lufy/managed-state/backups/*/manifest.json` | Schema 2, `runtimeStateSchema: 4` |

```sh
lufy-ai backup --target /ruta/al/proyecto
lufy-ai restore --target /ruta/al/proyecto --list
lufy-ai restore --target /ruta/al/proyecto --backup /ruta/al/backup --dry-run
lufy-ai restore --target /ruta/al/proyecto --backup /ruta/al/backup --yes
```

`backup` devuelve la ruta creada. `restore --list` sólo enumera `.lufy/managed-state/backups`; no busca ubicaciones retiradas. El restore público valida generación, hashes y payload de configuración/estado antes de escribir. La retención conserva carpetas desconocidas y generaciones anteriores. El rollback interno puede recuperar una preimagen dañada si una mutación respaldada falla; si no lo logra, conservá el backup indicado y diagnosticá el error.

Un backup mínimo de archivos gestionados no incluye todo el proyecto ni reemplaza el respaldo del ejecutable, memoria y datos propios. Antes de reemplazar una generación se necesita un snapshot conjunto de ejecutable, configuración, manifest, ancestors y datos/proyecciones que deban recuperarse.

## Instalaciones del modelo retirado

Los formatos anteriores o mezclados se rechazan. No hay migrador genérico, aliases ni lectores de compatibilidad. La sustitución personal del [bloque 6](propuestas/lufy-inline-2026-10-24/implementacion/bloque-6.md) fue una operación puntual con inventario, backup conjunto, journal y verificación; no es un comando reutilizable del producto. Para volver a otra generación se requiere su ejecutable y snapshot conjunto.

`upgrade --to <version>` comprueba el contrato de generación antes de reemplazar el ejecutable y verifica SHA-256. Header y checksum no son firmas ni prueban calidad del código. La [seguridad de releases](release-security.md) describe verificaciones adicionales. Esta guía no publica una release ni modifica una instalación personal por sí sola.
