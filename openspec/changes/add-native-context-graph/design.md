## Contexto

LUFY ya opera sobre un repositorio mixto: CLI Go en `tools/lufy-cli-go`, instalador estricto en `scripts/install.sh`, agentes/skills en `.opencode` y `.agents`, y propuestas OpenSpec bajo `openspec/`. La exploracion manual de relaciones es posible, pero no queda materializada como artefacto reutilizable entre agentes.

El cambio propone un grafo local, deterministico y verificable. Debe acelerar contexto e impacto, pero no convertirse en autoridad superior a archivos, tests, OpenSpec o memoria Obsidian.

## Arquitectura propuesta

### Capas CLI Go

- `internal/contextgraph/domain`: tipos puros del schema `lufy-context-graph` (`Graph`, `Node`, `Edge`, `Health`, `Community`, `Manifest`, `ExtractorReport`).
- `internal/contextgraph/application`: casos de uso `Scan`, `Build`, `Status`, `Query`, `Path`, `Explain`, `DiffImpact`.
- `internal/contextgraph/extractors`: extractores deterministicos para Go, Markdown, YAML y JSON.
- `internal/contextgraph/adapters`: filesystem, git diff provider, JSON encoder, cache/manifest store.
- Wiring en la CLI existente bajo `tools/lufy-cli-go` para el comando `context`.

Mantener handlers CLI delgados: parsean flags, invocan servicios y formatean salida. Los servicios reciben dependencias por constructor cuando aplique.

### Schema `lufy-context-graph`

`graph.json` debe contener como minimo:

- `schema`: literal `lufy-context-graph`.
- `generated_at`: timestamp ISO-8601.
- `root`: metadata del workspace y version de CLI.
- `sources`: archivos escaneados con hash, parser y estado.
- `nodes`: entidades normalizadas, ordenadas por id estable.
- `edges`: relaciones normalizadas, ordenadas por `(from, type, to)`.
- `health`: estado del corpus, archivos indexados, omitidos, errores y warnings accionables.
- `communities`, `important_nodes` y `suggested_questions`: analisis deterministico para ahorrar lecturas iniciales.
- `manifest`: hashes y opciones derivados usados para detectar obsolescencia; no es fuente canonica de configuracion.
- `extensions`: espacio opcional para futuras fases sin romper el schema publico.

Tipos iniciales de nodos:

- `file`, `directory`, `go_package`, `go_type`, `go_function`, `go_method`, `go_import`.
- `markdown_document`, `markdown_heading`, `openspec_change`, `openspec_requirement`.
- `yaml_document`, `yaml_key`, `json_document`, `json_key`.
- `skill`, `agent`, `command` cuando se detecten desde rutas `.opencode` o `.agents`.

Tipos iniciales de edges:

- `contains`, `defines`, `imports`, `references`, `depends_on`, `documents`, `configures`, `implements`, `tests`, `related_to`.

Los ids deben ser estables y relativos al root; no deben incluir paths absolutos salvo metadata explicitamente marcada como local.

### Configuracion canonica y artefactos derivados

La fuente canonica para grafo, memoria y vault es `.lufy/config/project.yaml`. No se deben crear archivos de configuracion adicionales para context graph, memoria o vault.

Campos canonicos iniciales:

- `context_graph.enabled`
- `context_graph.root`
- `context_graph.cache`
- `context_graph.report`
- `context_graph.skip_sensitive`
- `context_graph.sensitive_patterns`
- `context_graph.max_query_results`
- `context_graph.max_neighbors_per_hint`
- `memory.root`
- `memory.vault`

Los artefactos bajo `context_graph.root` son derivados y regenerables:

- `graph.json`: salida canonical JSON, con orden deterministico.
- `graph-summary.md`: resumen humano con conteos, top comunidades/areas por heuristica deterministica y comandos sugeridos.
- `GRAPH_REPORT.md`: reporte accionable con health, nodos importantes, areas/comunidades y preguntas sugeridas.
- `manifest.json`: derivado; registra hashes, version CLI, schema y opciones efectivas para staleness.
- `cache/`: derivado; resultados intermedios por hash de archivo para builds incrementales.

Las escrituras deben ser atomicas: generar en temporal dentro de `.lufy/context/`, validar estructura minima, luego reemplazar.

### Comandos CLI

- `lufy-ai context scan`: inspecciona fuentes soportadas y reporta que se construiria sin persistir grafo completo, salvo cache si se habilita explicitamente.
- `lufy-ai context build`: construye y persiste `graph.json`, `graph-summary.md`, `GRAPH_REPORT.md` y manifest/cache derivado aplicable.
- `lufy-ai context status`: reporta `ready`, `stale` o `not_available` con razones accionables.
- `lufy-ai context query <term>`: devuelve hints compactos rankeados por busqueda lexical expandida contra el vocabulario del grafo y conectividad estructural.
- `lufy-ai context path <from> <to>`: calcula camino explicable entre nodos cuando existe.
- `lufy-ai context explain <node-or-path>`: explica por que un nodo o relacion existe, incluyendo source spans cuando esten disponibles.
- `lufy-ai context diff --base <ref>`: usa diff Git contra `<ref>` para mapear archivos cambiados a nodos, vecinos, comunidades y posibles agentes/specs afectados.

Todos los comandos consumidores deben degradar a `not_available` si el grafo no existe, no cumple schema, esta stale o no puede leerse. La recuperacion sugerida sera `lufy-ai context build`.

### Extractores deterministas

- Go: usar `go/parser` y `go/ast`; emitir packages, imports, tipos, funciones, metodos y archivos de test. No resolver tipos con red ni descargar modulos.
- Markdown: parseo lineal de headings, links relativos, fenced code metadata y markers OpenSpec (`Requirement`, `Scenario`, delta sections).
- YAML: parsear claves y estructura con la libreria YAML ya usada por la CLI Go si existe; si no, agregar dependencia Go justificada en el slice de implementacion.
- JSON: parsear keys y estructura con `encoding/json`.

Todos los extractores deben ordenar resultados, limitar contenido textual y reportar errores por archivo sin abortar el build completo salvo corrupcion del grafo final.

### Skills y agentes

- OpenCode: crear `.opencode/skills/lufy.context-search/SKILL.md` para consultar `lufy-ai context query/path/explain/diff` y devolver hints compactos.
- Codex/agentes locales: crear skill equivalente bajo `.agents/skills/` si el catalogo vigente lo requiere.
- `explorer`: antes de grep/glob amplio, puede consultar hints por ruta/spec/concepto y registrarlos como `context_graph_hints`; si no existe grafo, usar `not_available` y continuar.
- `sdd-router`: puede consultar `context diff --base` o query para sizing/routing, pero no debe ejecutar mutaciones ni reemplazar analisis read-only requerido.
- `reviewer`: puede consultar impacto y caminos para orientar review, manteniendo scoring basado en diff, tests y evidencia real.

### Ahorro de tokens como criterio de readiness

La capacidad no esta lista si solo produce un grafo lexical que obliga al agente a leer ampliamente. Cada consulta sustantiva debe devolver un paquete compacto con top nodos, vecinos acotados, comunidades afectadas y razon de relevancia para orientar lecturas especificas antes de `grep`/reads amplios.

### Semantica/LLM futura

La fase default sigue siendo deterministica y local. Embeddings, resúmenes LLM, ranking semantico o clustering aprendido quedan como extension opcional posterior bajo `extensions`, desactivada por defecto y sujeta a nueva propuesta si cambia privacidad, costo o dependencias.

## Estrategia de migracion

1. Definir schema, fixtures y validadores internos.
2. Implementar extractores y normalizacion estable.
3. Implementar build/status/query/path/explain/diff en CLI Go.
4. Agregar skills e integracion degradable de agentes.
5. Validar OpenSpec, Go tests y `scripts/validate.sh` cuando toque assets instalables.

## Riesgos y mitigaciones

- **Rendimiento**: usar cache incremental por SHA-256, limites de salida y status que no reconstruya todo cuando el manifest derivado alcanza.
- **Ruido del grafo**: empezar con edges conservadores y `explain` para trazabilidad.
- **Privacidad**: no persistir secretos ni dumps completos; almacenar hashes, spans y snippets acotados solo cuando sean seguros.
- **Acoplamiento de agentes**: documentar `not_available` como estado normal y mantener exploracion tradicional como fallback.
- **Churn en assets instalados**: coordinar con managed assets y mantener idempotencia SHA-256 existente.
