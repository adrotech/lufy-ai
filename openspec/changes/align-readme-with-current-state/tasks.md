## 1. README Root

- [x] 1.1 Reestructurar `README.md` como landing page breve con banner preservado, descripción del estado real, quickstart y navegación local vigente.
- [x] 1.2 Eliminar o mover del README secciones extensas de templates, stacks y subagentes futuros para que no parezcan capacidades instalables actuales.
- [x] 1.3 Mantener referencias claras a `docs/getting-started.md`, `docs/roadmap.md`, `openspec/README.md` y `tools/lufy-cli-go/README.md`.

## 2. Docs Auxiliares

- [x] 2.1 Actualizar `docs/getting-started.md` a español y sincronizarlo con el quickstart actual de CLI Go y wrapper estricto.
- [x] 2.2 Ajustar `docs/roadmap.md` para conservar el contexto futuro de templates/subagentes y marcarlo explícitamente como roadmap, no estado instalable.
- [x] 2.3 Actualizar `tools/lufy-cli-go/README.md` para describir con precisión `install`, `verify`, `backup`, `restore`, `sync`, assets gestionados, SHA-256, idempotencia, backup/restore y límites de CI/sync en curso.

## 3. Consistencia y Validación

- [x] 3.1 Revisar que README, getting started, roadmap y README de CLI no contradigan el estado actual de CLI Go, sync, CI proposal/implementation y assets gestionados.
- [x] 3.2 Revisar enlaces locales, anchors relevantes y rutas mencionadas tras la reestructura documental.
- [x] 3.3 Ejecutar validación estática disponible, incluyendo `git diff --check`, y reportar comandos/resultados reales sin inventar toolchains de raíz.
