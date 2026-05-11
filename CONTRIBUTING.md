# Contribuir a lufy-ai

Gracias por contribuir. Este repositorio usa un flujo OpenSpec/SDD para cambios funcionales y una validación agrupada antes de entregar.

## Flujo recomendado

1. Crea una rama desde `develop`.
2. Para cambios de comportamiento o arquitectura, abre una propuesta OpenSpec antes de implementar.
3. Mantén los cambios enfocados y mínimos.
4. Ejecuta la validación agrupada antes de abrir PR:

```sh
scripts/validate.sh
```

## CLI Go

La implementación actual de la CLI vive en `tools/lufy-cli-go`.

Comandos útiles:

```sh
cd tools/lufy-cli-go
go test ./...
go build ./cmd/lufy-ai
```

## Reglas importantes

- No reintroducir fallback legacy en `scripts/install.sh`.
- No asumir tooling Node/TypeScript en la raíz del repo.
- Reportar evidencia real de validación; no afirmar tests exitosos sin haberlos ejecutado.
- No mezclar cambios no relacionados en el mismo PR.
