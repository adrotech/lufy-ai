## 1. Preparacion OpenSpec

- [x] Validar esta propuesta con `openspec validate "add-native-context-graph" --strict`.
- [x] Confirmar alcance de runtime antes de editar `tools/lufy-cli-go`.
- [x] Identificar si el catalogo Codex vigente requiere skill bajo `.agents/skills/` ademas de `.opencode/skills/`.

## 2. Schema y artefactos persistidos

- [x] Definir tipos Go para `lufy-context-graph/v1` con version, metadata, sources, nodes, edges, manifest y extensions.
- [x] Implementar escritura atomica de `.lufy/context/graph.json`.
- [x] Implementar `.lufy/context/graph-summary.md` con conteos y hints humanos compactos.
- [x] Implementar manifest/cache cuando aplique para detectar `ready`, `stale` y evitar churn idempotente.
- [x] Agregar fixtures de schema y tests de orden deterministico.

## 3. Extractores deterministicos iniciales

- [x] Implementar extractor Go con `go/parser`/`go/ast` para packages, imports, tipos, funciones, metodos y tests.
- [x] Implementar extractor Markdown para headings, links relativos y markers OpenSpec.
- [x] Implementar extractor YAML para claves, estructura y referencias conservadoras.
- [x] Implementar extractor JSON con `encoding/json` para claves y estructura.
- [x] Asegurar que errores por archivo se reporten sin abortar otros extractores, salvo corrupcion estructural final.

## 4. CLI `lufy-ai context`

- [x] Agregar `lufy-ai context scan` para inspeccion deterministica sin persistir grafo completo por defecto.
- [x] Agregar `lufy-ai context build` para generar `graph.json`, `graph-summary.md` y manifest/cache aplicable.
- [x] Agregar `lufy-ai context status` con estados `ready`, `stale` y `not_available`.
- [x] Agregar `lufy-ai context query` con busqueda lexical deterministica.
- [x] Agregar `lufy-ai context path` para caminos explicables entre nodos.
- [x] Agregar `lufy-ai context explain` con fuente y razon de nodos/edges.
- [x] Agregar `lufy-ai context diff --base <ref>` para impacto por diff Git.
- [x] Cubrir degradacion `not_available` cuando falte o falle `.lufy/context/graph.json`.

## 5. Skills e integracion de agentes

- [x] Crear skill OpenCode `.opencode/skills/lufy.context-search/SKILL.md`.
- [x] Crear skill equivalente Codex bajo `.agents/skills/` si el catalogo vigente lo requiere.
- [x] Integrar hints opcionales en `explorer` con fallback `not_available`.
- [x] Integrar hints opcionales en `sdd-router` sin romper su modo read-only/no-shell salvo comandos permitidos por su contrato futuro.
- [x] Integrar hints opcionales en `reviewer` para orientar impacto, sin sustituir diff/tests/evidencia.
- [x] Documentar que semantica/LLM es fase futura opcional y no default.

## 6. Validacion agrupada

- [x] Ejecutar `openspec validate "add-native-context-graph" --strict`.
- [x] Ejecutar tests Go aplicables en `tools/lufy-cli-go` para context graph y CLI.
- [x] Ejecutar `scripts/validate.sh` cuando el bloque toque CLI/assets instalables.
- [x] Revisar estaticamente que `scripts/install.sh` siga siendo wrapper estricto y sin fallback legacy.
- [x] Revisar estaticamente que agentes/skills degraden a `not_available` cuando el grafo no exista.
