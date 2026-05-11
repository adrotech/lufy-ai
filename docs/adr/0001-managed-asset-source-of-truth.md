# ADR 0001: Fuente de verdad de assets gestionados

## Estado

Aceptada.

## Contexto

La CLI Go distribuye assets gestionados desde un catálogo que puede construirse desde el checkout local o desde assets embebidos en el binario. Mantener ambos árboles sin verificación puede producir drift entre lo probado en desarrollo y lo publicado en releases.

## Decisión

La raíz del repositorio sigue siendo la fuente canónica humana para `.opencode/`, `openspec/`, `AGENTS.md.template` y `tui.json`.

El árbol embebido en `tools/lufy-cli-go/internal/assets/embedded/` es un mirror de distribución y debe mantenerse en paridad con la fuente canónica mediante tests/validación.

## Consecuencias

- Los cambios en assets gestionados deben actualizar el mirror embebido.
- `TestEmbeddedCatalogMatchesRepositoryAssets` debe fallar si hay drift de rutas, policies o hashes.
- No se usa `//go:embed` contra paths fuera del módulo Go porque Go no lo permite.
