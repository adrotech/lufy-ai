# Contribuir a Lufy

La base normal es `develop`; `main` contiene la línea estable. Trabajá en una rama propia, por ejemplo `codex/<tema>`, y preservá cambios locales ajenos. La [documentación actual](docs/README.md) describe el modelo inline.

## Preparar el cambio

Entendé resultado, dependencias y alcance autorizado. Para trabajo incierto o amplio, dejá una propuesta o un plan breve con aceptación observable y validación pertinente. Un arreglo mecánico puede implementarse directamente. El producto no exige tiers, OpenSpec, envelopes ni transferencias de roles.

Las decisiones de estructura del usuario y `project_profile.surfaces` forman parte de la aceptación. Usá sólo toolchains disponibles. La CLI vive en `tools/lufy-cli-go`; `.opencode` y una suite Node/TS raíz no forman parte de este producto.

## Cambiar código y documentación

- Mantené el diff enfocado y agregá pruebas para comportamiento relevante. Agrupá la validación al terminar un bloque coherente.
- Actualizá guías cuando cambien comandos, defaults, formatos o límites.
- Si cambia un asset editorial seleccionado, actualizá su mirror exacto en `tools/lufy-cli-go/internal/assets/embedded`. No edites sólo el mirror ni agregues archivos embebidos ajenos al catálogo.
- Conservá evidencia fechada de ensayos. Una conclusión nueva va en su informe o estado actual; no reescribe resultados anteriores.
- Documentá cambios integrados todavía sin publicar en `CHANGELOG.md`, sección `Unreleased`.

## Comprobar y entregar

```sh
scripts/validate.sh
```

El script cubre los [gates disponibles](docs/ci-quality-gates.md), incluyendo Python offline, Go con coverage, vet, build y paridad de assets. Los ensayos con modelos son opt-in. No afirmes haber ejecutado una comprobación omitida.

Para PRs contra `develop`, el whitespace usa el rango real: `git diff --check origin/develop...HEAD` para commits y `git diff --check origin/develop` si hay cambios pendientes. Resumí problema, comportamiento final, evidencia y limitaciones en español.

Commit, push y PR requieren autorización explícita. La entrega se completa observando los checks remotos requeridos. Promover a `main` y publicar una versión son acciones separadas; ver [política de entrega](.lufy/contracts/delivery.md) y [releases](docs/releasing.md).
