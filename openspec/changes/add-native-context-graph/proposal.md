## Why

Los agentes LUFY dependen hoy de exploracion textual ad hoc para entender relaciones entre archivos, specs, skills y comandos. Eso incrementa costo, repeticion y riesgo de omitir impacto cuando un cambio cruza Go CLI, OpenSpec, `.opencode`, `.agents` y memoria local.

Un grafo de contexto nativo y deterministico permite construir un indice portable del repositorio, consultarlo desde la CLI Go y exponer hints compactos a `explorer`, `sdd-router` y `reviewer` sin depender de servicios externos ni de LLM por defecto.

## What Changes

- Agregar capacidad `native-context-graph` con schema estable `lufy-context-graph/v1`.
- Extender la CLI Go en `tools/lufy-cli-go` con `lufy-ai context scan/status/build/query/path/explain/diff`.
- Persistir artefactos gestionados en `.lufy/context/`:
  - `graph.json` como grafo deterministico machine-readable;
  - `graph-summary.md` como resumen humano compacto;
  - manifest/cache cuando aplique para idempotencia, incrementalidad y verificacion estructural.
- Implementar extractores iniciales deterministas para Go (`go/parser`/`go/ast`), Markdown, YAML y JSON.
- Agregar skill `lufy.context-search` para OpenCode y skill equivalente bajo `.agents` cuando el catalogo Codex lo requiera.
- Integrar hints de grafo en `explorer`, `sdd-router` y `reviewer`, degradando a `not_available` si `.lufy/context/graph.json` no existe, esta obsoleto o falla la lectura.
- Agregar `context diff --base <ref>` para estimar impacto por diff antes de implementar o revisar.
- Mantener semantica/LLM como fase futura opcional, nunca como default del grafo inicial.

## Non-Goals

- No reemplazar OpenSpec, Obsidian memory ni validacion real con el grafo.
- No introducir runtime Node/TS en la raiz ni asumir `package.json` global.
- No reintroducir fallback legacy en `scripts/install.sh`; debe seguir como wrapper estricto de la CLI Go.
- No cambiar contratos publicos existentes fuera de los nuevos comandos `lufy-ai context ...`.
- No usar embeddings, servicios remotos, LLM o ranking semantico como requisito de la fase inicial.
- No persistir secretos ni contenido completo innecesario de archivos sensibles en `graph.json`.

## Review Slices

### Slice 1: Schema y almacenamiento local

- Objetivo: definir `lufy-context-graph/v1`, nodos, edges, manifest/cache y reglas de escritura atomica en `.lufy/context/`.
- Archivos esperados: `tools/lufy-cli-go/internal/contextgraph/*`, tests unitarios y fixtures.
- Criterios:
  - WHEN se ejecuta `lufy-ai context build`, THEN se escribe un `graph.json` validable con `schema: lufy-context-graph/v1`.
  - WHEN no hay cambios de entrada, THEN el manifest/cache permite salida idempotente sin churn innecesario.
- Riesgo: schema demasiado rigido; mantener `version`, `metadata`, `nodes`, `edges` y `extensions` compatibles con evolucion.

### Slice 2: Extractores deterministas

- Objetivo: cubrir Go, Markdown, YAML y JSON con parseo deterministicamente ordenado.
- Archivos esperados: extractores por formato, fixtures en `tools/lufy-cli-go`.
- Criterios:
  - WHEN un archivo Go contiene packages, tipos, funcs, imports o tests, THEN el extractor emite nodos/edges estables desde `go/parser`/`go/ast`.
  - WHEN Markdown/YAML/JSON contienen headings, keys o referencias a rutas, THEN el extractor emite nodos/edges normalizados sin heuristicas LLM.
- Riesgo: ruido por referencias falsas; preferir edges conservadores y explicables.

### Slice 3: CLI `context` y diff de impacto

- Objetivo: exponer `scan`, `status`, `build`, `query`, `path`, `explain` y `diff --base` desde la CLI Go.
- Archivos esperados: comandos CLI, servicios de aplicacion y tests de contrato.
- Criterios:
  - WHEN el usuario corre `lufy-ai context diff --base origin/develop`, THEN recibe un resumen de nodos afectados, vecinos relevantes y rutas explicables.
  - WHEN no existe grafo, THEN `status` y comandos consumidores reportan `not_available` con accion de recuperacion.
- Riesgo: comandos lentos en repos grandes; usar caches y limites de salida desde el inicio.

### Slice 4: Skills e integracion de agentes

- Objetivo: instalar skill OpenCode `lufy.context-search`, skill equivalente Codex `.agents` si aplica, e integrar hints en `explorer`, `sdd-router` y `reviewer`.
- Archivos esperados: `.opencode/skills/lufy.context-search/`, `.agents/skills/lufy-context-search/` o equivalente, definiciones de agentes afectadas y assets gestionados si corresponde.
- Criterios:
  - WHEN existe grafo valido, THEN los agentes pueden solicitar hints compactos por query/path/diff.
  - WHEN el grafo falta, THEN los agentes continuan con exploracion normal y registran `not_available` sin bloquear.
- Riesgo: acoplar agentes al grafo como fuente canonica; documentar que es hint secundario y no reemplaza archivos/comandos.

## Validation

- `openspec validate "add-native-context-graph" --strict`
- `go test ./...` desde `tools/lufy-cli-go` o el subconjunto Go real que cubra `contextgraph` y CLI.
- `scripts/validate.sh` desde la raiz cuando el cambio toque CLI/assets instalables.
- Revision estatica de `scripts/install.sh` para confirmar que sigue siendo wrapper estricto.
- Revision estatica de skills/agentes para confirmar degradacion `not_available` y que semantica/LLM queda como fase futura opcional.
