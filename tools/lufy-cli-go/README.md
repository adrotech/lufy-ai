# CLI de Lufy

Implementación Go del flujo Codex inline. Esta rama contiene cambios incompatibles aún no publicados; `v0.6.25` es la referencia estable del modelo anterior, no el formato actual.

```sh
go test ./...
go vet ./...
go build -o bin/lufy-ai ./cmd/lufy-ai
```

Desde la raíz se usa `scripts/validate.sh` para agrupar las comprobaciones disponibles.

## Superficie

- install/setup/sync/verify: assets gestionados, hashes y configuración actual.
- backup/restore/uninstall: recuperación y ownership, conservando archivos propios.
- init/scan/plan: información del proyecto, superficies y validaciones pertinentes.
- memory/context: capacidades de contexto opcionales, sin preflight metodológico obligatorio.
- evidence/candidate/review: observaciones y vinculación al código para integraciones; no aprueban por sí solas calidad ni delivery.
- run/adaptive: operaciones explícitas de registro y asignación, sin selección de metodología ni autorización implícita.
- status/info/doctor/menu/version/upgrade: diagnóstico y gestión del binario.

`lufy-ai --help` es la referencia del ejecutable. Sólo hay adapter escribible Codex. Configuración schema 2; manifest schema 4. No se distribuyen motores SDD/OpenSpec, parsers de artifacts ni adapters históricos. Los datos/documentos ajenos no se convierten en instrucciones activas por su ubicación.

El payload embedded contiene sólo los 16 assets seleccionados y la plantilla del generador de guía, con paridad de fuente. Los snapshots para ensayos históricos viven fuera del producto.
