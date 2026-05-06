## 1. Workflow de tagging automático

- [x] 1.1 Agregar workflow `auto-release-tag.yml` para eventos `pull_request.closed` hacia `main` con permisos mínimos.
- [x] 1.2 Implementar guardas para PR mergeado, branch `main`, merge commit resoluble y alcanzabilidad desde `origin/main`.
- [x] 1.3 Implementar cálculo del siguiente tag patch desde tags `vMAJOR.MINOR.PATCH`, con fallback inicial `v0.1.0`.
- [x] 1.4 Crear y pushear tag anotado sin sobrescribir tags existentes, con no-op explícito si el tag calculado ya existe.

## 2. Documentación y trazabilidad

- [x] 2.1 Documentar el release automático, idempotencia y relación con `release.yml` en README y docs de ramas/releases.
- [x] 2.2 Validar OpenSpec, whitespace, YAML/shell estático y actualizar tasks solo tras completar la implementación.
