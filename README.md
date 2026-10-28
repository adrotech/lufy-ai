# lufy-ai

Lufy ayuda al agente principal a resolver el pedido y comprobar su resultado. La ejecución es inline; contexto, memoria, revisión y delegación se usan cuando aportan valor y respetan los permisos del host.

Esta rama contiene la reestructuración inline aún sin release. La última referencia estable es `v0.6.25`; esa release pertenece al modelo anterior y no debe instalarse sobre la configuración actual. No se promete interoperabilidad entre generaciones.

## Uso local

La CLI Go vive en `tools/lufy-cli-go`. Compilá desde el checkout de la implementación actual:

```sh
go -C tools/lufy-cli-go build -o /tmp/lufy-ai ./cmd/lufy-ai
/tmp/lufy-ai install --target /ruta/al/proyecto --tool codex --delegation off --dry-run
/tmp/lufy-ai install --target /ruta/al/proyecto --tool codex --delegation off --yes
/tmp/lufy-ai verify --target /ruta/al/proyecto --deep
```

`--delegation` admite `off`, `on_request` y `auto`. Delegar exige herramientas reales y no amplía permisos. Una instrucción de trabajar inline prevalece. Commit, push, PR, publicación y operaciones externas requieren autorización.

El catálogo actual es Codex: 16 assets, cinco skills y un generador de guía. No hay router por tiers, metodología seleccionable, envelope obligatorio ni comandos SDD/OpenSpec. El instalador conserva hashes, ownership, conflictos, backups e idempotencia; `scripts/install.sh` es un wrapper estricto de la CLI Go.

## Resultado y comprobación

Una explicación puede terminar con una respuesta fundamentada; una propuesta, con alternativas y límites; un texto, con revisión factual/editorial; código y datos reciben las comprobaciones pertinentes. No se exige un archivo o una cadena de roles por etapa.

- [Instalación, formatos y recuperación](docs/installation.md).
- [Trabajo inline y capacidades](docs/harness-workflow.md).
- [Estado del trabajo y decisiones retiradas](docs/estado-del-trabajo.md).
- [CLI](tools/lufy-cli-go/README.md).
- [Propuesta, implementación y evidencia](docs/propuestas/lufy-inline-2026-10-24/README.md).

## Desarrollo

```sh
scripts/validate.sh
```

La validación agrupada comprueba whitespace contra la base, Python offline, Go, cobertura, formato, shell, workflows y payload. No invoca modelos. Las comparaciones con modelos son opt-in, con fixtures externos, recursos fijados y evidencia privada; no son gates del trabajo ordinario.

`develop` es integración y `main` estable. Un release requiere integración, validación y autorización; no se publica por terminar un bloque local.
