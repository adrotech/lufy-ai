# CLI Go de Lufy

Implementación del flujo Codex inline integrado en `develop`. Los cambios incompatibles todavía no están publicados: `v0.6.25` es la referencia estable del modelo anterior.

Desde `tools/lufy-cli-go`, con Go según `go.mod`:

```sh
go test ./...
go vet ./...
go build -o bin/lufy-ai ./cmd/lufy-ai
```

Desde la raíz, `scripts/validate.sh` agrupa los gates pertinentes y exige coverage Go de al menos 80%. Las pruebas Python del evaluador son offline. No hay toolchain Node/TypeScript raíz.

## Diseño y distribución

- Tool escribible: Codex project-local; ejecución inline y delegación configurable.
- Configuración schema 2, manifest 4 y backup 2 ligado a estado 4.
- Catálogo: 16 assets seleccionados y plantilla del generador de `AGENTS.md`; cinco skills.
- Fuentes editoriales en la raíz; mirror exacto en `internal/assets/embedded`. Ver [ADR de fuente canónica](../../docs/adr/0001-managed-asset-source-of-truth.md).
- Contexto, memoria y registros explícitos son capacidades opcionales. No hay motores SDD/OpenSpec, lectores de formatos retirados ni router por tiers.

La [arquitectura](../../docs/architecture.md) describe responsabilidades internas. La [referencia de comandos](../../docs/cli.md) cubre todas las operaciones; `lufy-ai --help` y el help de subcomandos son la interfaz del ejecutable. [Instalación](../../docs/installation.md) explica PATH, wrapper y recuperación. [Releases](../../docs/releasing.md) distingue compilación local de publicación estable.
