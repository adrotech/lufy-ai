## Why

La instalación actual obliga a clonar el repositorio y compilar la CLI Go localmente antes de usar `scripts/install.sh`, lo que aumenta la fricción para usuarios que solo quieren instalar el kit en un proyecto destino. Ahora que `lufy-ai` ya tiene una CLI Go con instalación, verificación, sync, backups e idempotencia, el siguiente bloque lógico es diseñar una distribución versionada y verificable que permita instalar sin clone.

## What Changes

- Añadir un contrato de distribución por releases binarios versionados publicados desde GitHub Actions/GitHub Releases, con checksums SHA-256 y comando `lufy-ai version`.
- Añadir un bootstrap installer remoto que detecte OS/arch, descargue una release versionada, verifique checksum antes de instalar y coloque el binario en un directorio de `PATH` elegido por el usuario.
- Diseñar la transición hacia instalación standalone real mediante assets gestionados embebidos en el binario o release bundles versionados con assets, recomendando una estrategia incremental que no rompa el wrapper estricto actual.
- Añadir documentación final como tarea de implementación, no como estado actual: README, `docs/getting-started.md` y README técnico de la CLI deberán actualizarse solo cuando el runtime exista.
- Permitir borrar documentación obsoleta al final de la implementación si deja de aplicar, sin mantener retrocompatibilidad documental hacia instrucciones basadas en clone/build como camino principal.
- Mantener controles de seguridad: version pinning, checksum obligatorio antes de ejecutar/instalar binarios descargados, alternativa inspeccionable a `curl | bash`, y ninguna acción destructiva automática salvo flag explícito.

## Capabilities

### New Capabilities

- `versioned-binary-distribution`: distribución de binarios `lufy-ai` versionados por plataforma, checksums, metadata de release y comando `lufy-ai version`.
- `remote-bootstrap-installer`: instalador remoto seguro para descargar, verificar e instalar una versión específica o estable del binario sin clonar el repositorio.

### Modified Capabilities

- `go-cli-installer`: la CLI deberá poder operar como binario distribuido fuera del checkout fuente y resolver sus assets gestionados desde una fuente standalone definida.
- `go-cli-install-ci`: la CI mínima deberá construir, empaquetar y validar artifacts de release, incluyendo checksums y smokes de instalación desde artifacts versionados.
- `current-state-documentation`: la documentación pública deberá migrar al flujo sin clone una vez implementado, manteniendo claro qué es estado real y eliminando instrucciones obsoletas cuando correspondan.

## Impact

- Código futuro en `tools/lufy-cli-go/` para `version`, metadata de build y resolución de assets standalone.
- Configuración futura en `.github/workflows/` para release builds multiplataforma, checksums y publicación en GitHub Releases.
- Script futuro de bootstrap en `scripts/` o endpoint documentado de GitHub raw/release, manteniendo `scripts/install.sh` como wrapper estricto local.
- Documentación futura en `README.md`, `docs/getting-started.md`, `tools/lufy-cli-go/README.md` y limpieza de docs obsoletas al cierre.
- Sin cambios runtime en esta proposal; este cambio solo define el plan, specs y tareas.
