## 1. OpenSpec artifacts

- [x] 1.1 Crear specs delta para `release-branch-flow`, `go-cli-install-ci`, `versioned-binary-distribution` y `current-state-documentation`.
- [x] 1.2 Crear `tasks.md` con tareas trazables para políticas, workflows, documentación y validación.

## 2. Políticas y guías operativas

- [x] 2.1 Actualizar `.opencode/policies/delivery.md` con `develop` como base normal, `main` productiva, PR feature→develop, promoción develop→main y tags `v*` desde main.
- [x] 2.2 Actualizar `AGENTS.md` con las mismas reglas operativas sin tocar `route-orchestrator-to-domain-agents`.

## 3. Workflows GitHub Actions

- [x] 3.1 Ajustar `.github/workflows/go-cli-install.yml` para PR/push a `develop` y `main`.
- [x] 3.2 Ajustar `.github/workflows/release.yml` para disparar solo en tags `v*` y validar localmente que el tag/commit sea alcanzable desde `origin/main` antes de publicar.

## 4. Documentación

- [x] 4.1 Actualizar `README.md`, `docs/getting-started.md`, `tools/lufy-cli-go/README.md` y `docs/roadmap.md` con el flujo `develop`/`main` y releases estables desde tags `v*` alcanzables desde `main`.
- [x] 4.2 Crear documentación breve de configuración GitHub branch settings: default branch `develop` y protección de `main`/`develop`.

## 5. Validación y cierre

- [x] 5.1 Ejecutar `openspec status --change "adopt-develop-main-release-flow" --json` y `openspec instructions apply --change "adopt-develop-main-release-flow" --json`.
- [x] 5.2 Ejecutar `git diff --check`.
- [x] 5.3 Ejecutar parse YAML local razonable para `.github/workflows/go-cli-install.yml` y `.github/workflows/release.yml` sin descargar internet.
- [x] 5.4 Marcar tareas completas solo después de validar y reportar evidencia.
