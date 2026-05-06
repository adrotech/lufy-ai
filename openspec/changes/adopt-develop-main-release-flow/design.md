## Context

El repositorio ya tiene workflows de CI y release, documentación de bootstrap/release y política de delivery. La convención previa mezclaba `development`, `develop`, `main` y `master` como bases posibles; eso deja ambigua la ruta de publicación estable y puede permitir releases desde commits no promovidos a producción.

## Goals / Non-Goals

**Goals:**

- Declarar `develop` como rama principal de trabajo e integración.
- Declarar `main` como rama productiva/estable, usada solo para promoción/release y hotfix autorizado.
- Establecer PR normal `feature/*` → `develop` y promoción `develop` → `main`.
- Mantener releases estables solo por tags `v*` y bloquear publicación si el tag no apunta a un commit alcanzable desde `main`.
- Alinear políticas, guías de agentes, workflows y documentación pública.

**Non-Goals:**

- No cambiar configuración real de GitHub en este rol ni crear commits, pushes, PRs o tags.
- No implementar un canal snapshot/main ni releases desde `develop`.
- No modificar el cambio activo `route-orchestrator-to-domain-agents`.
- No cambiar runtime del instalador, puertos, auth, esquemas o contratos de la CLI.

## Decisions

1. **`develop` como base por defecto.** Los cambios normales abren PR desde ramas `feature/*`, `fix/*`, `chore/*` o equivalentes hacia `develop`. Alternativa considerada: mantener `development`; se descarta para evitar doble rama de integración.
2. **`main` productiva y no fuente de PR.** `main` solo recibe promoción desde `develop` o hotfix/release explícito. Alternativa considerada: permitir PRs desde `main`; se descarta porque mezcla producción como rama fuente y contradice el flujo de promoción.
3. **Release por tag `v*` alcanzable desde `main`.** El workflow conserva trigger por tag `v*`, pero antes de publicar verifica que `origin/main` contenga el commit taggeado. Alternativa considerada: confiar solo en protecciones de branch; se descarta porque un tag puede crearse sobre cualquier commit si no hay guard en CI.
4. **Configuración GitHub documentada, no aplicada.** Los pasos para default branch y branch protection quedan en docs para que `delivery` o un maintainer los aplique luego con autorización. Alternativa considerada: automatizar con `gh` ahora; se descarta por límite del rol implementer y porque el usuario pidió no hacer Git delivery.

## Risks / Trade-offs

- **Riesgo:** tags existentes sobre commits no alcanzables desde `main` fallarán al publicar. → **Mitigación:** documentar que el tag `v*` debe crearse después de promover a `main`.
- **Riesgo:** repositorios remotos aún pueden tener default branch distinta hasta que se configure GitHub. → **Mitigación:** doc dedicada con pasos manuales/automatizables y estado de follow-up.
- **Riesgo:** `actions/checkout` con profundidad limitada puede no conocer `main`. → **Mitigación:** usar `fetch-depth: 0` en release y fetch explícito de `main` antes del guard.
- **Riesgo:** documentación de bootstrap puede sonar disponible sin release. → **Mitigación:** reforzar que el bootstrap estable consume releases publicadas y que snapshot/main queda fuera de alcance.
