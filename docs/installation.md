# Instalación y recuperación

La reestructuración inline está en desarrollo, sin nueva release publicada. La referencia estable `v0.6.25` corresponde al modelo anterior. Compilá únicamente desde el checkout que contiene esta implementación; no reemplaces el ejecutable actual con una compilación de una rama vieja.

```sh
go -C tools/lufy-cli-go build -o /tmp/lufy-ai ./cmd/lufy-ai
/tmp/lufy-ai install --target /ruta/al/proyecto --tool codex --delegation off --dry-run
/tmp/lufy-ai install --target /ruta/al/proyecto --tool codex --delegation off --yes
/tmp/lufy-ai verify --target /ruta/al/proyecto --deep
```

`scripts/install.sh` sólo ejecuta la CLI Go disponible en el binario local o PATH. No instala herramientas Node ni busca un motor alternativo.

## Formatos actuales

Configuración: `.lufy/config/project.yaml`, schema 2, tool codex y execution inline. Manifest: `.lufy/managed-state/install-state.json`, schema 4. Los assets incluyen ownership/hashes, no metodología. Los formatos anteriores o mezclados se rechazan sin traducción silenciosa.

La sustitución de la instalación propia es un procedimiento puntual vinculado al snapshot real: inventario, hashes, preview sin escrituras, backup conjunto, journal, aplicación y verificación. No es un migrador del runtime. Con drift se detiene en el archivo concreto y conserva los datos.

## Sincronización y conflictos

```sh
lufy-ai sync --target /ruta/al/proyecto --dry-run
lufy-ai conflicts plan --target /ruta/al/proyecto --json
lufy-ai sync --target /ruta/al/proyecto --yes
lufy-ai verify --target /ruta/al/proyecto --deep
```

Una edición local no se sobrescribe por una coincidencia de nombre: ownership, hashes y políticas deciden la acción. Los archivos propios fuera del manifest se conservan. Sync no descubre ni reinstala roles o metodologías por carpetas históricas.

## Backup y restauración

`backup` y `restore` trabajan con manifest/archivos del formato actual y la raíz `.lufy/managed-state/backups`. `restore --list` enumera sólo esa raíz; no existe alias `legacy:` ni búsqueda automática en ubicaciones retiradas.

La sustitución de una generación requiere respaldo conjunto del binario, configuración, manifest, ancestors y proyecciones. El restore mínimo de archivos no sustituye ese respaldo. La recuperación puntual rechaza cambios posteriores para no sobrescribir nuevo trabajo.

`upgrade` exige que el artefacto declare el contrato actual antes de reemplazar el ejecutable. Esta declaración y checksum comprueban formato e integridad; no son una firma ni evidencia de calidad del código descargado.

La instalación personal y el código fuente se identifican por separado. El código de la refactorización permanece en el checkout aislado hasta integrarlo; no se sobrescribe trabajo local ajeno ni se publica automáticamente.

Los backups actuales usan schema 2 y `runtimeStateSchema: 4`. La retención preserva carpetas desconocidas y generaciones anteriores. Restore público comprueba también el payload de configuración y estado antes de escribir. Los backups históricos se conservan como datos; la recuperación de otra generación requiere su ejecutable y snapshot conjunto.
