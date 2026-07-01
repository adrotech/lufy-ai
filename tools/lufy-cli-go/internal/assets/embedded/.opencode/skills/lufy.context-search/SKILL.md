---
name: lufy.context-search
description: Usa lufy-ai context como preflight local obligatorio cuando context_graph.enabled=true para buscar hints compactos antes de discovery genérico amplio.
license: MIT
compatibility: OpenCode skill autocontenido; usa lufy-ai context.
metadata:
  author: lufy-ai
  version: "1.0"
---

# Skill: lufy.context-search

Usar cuando una tarea necesita orientación rápida sobre relaciones de archivos, símbolos, docs u OpenSpec y el repositorio puede tener un grafo generado bajo `context_graph.root` de `.lufy/config/project.yaml`.

## Flujo

1. Tratar `.lufy/config/project.yaml` como fuente canonica de `context_graph` y memoria/vault. `manifest.json`, cache y reportes del grafo son derivados/regenerables.
2. Consultar disponibilidad:

```bash
lufy-ai context status --target <repo> --json
```

3. Si `.lufy/config/project.yaml` declara `context_graph.enabled: true`, este status es preflight obligatorio antes de discovery genérico amplio (`glob`/`grep`/`find`/lecturas exploratorias). Son excepciones: leer configuración, paths exactos nombrados por usuario/handoff o artifacts ya seleccionados.
4. Si el estado es `not_available` o `stale`, reportar `context_graph_hints.status: not_available` o `stale`, `recovery: lufy-ai context build` y `fallback_reason`; recién entonces continuar con inspección normal del repositorio.
5. Para hints rankeados cuando el grafo está listo:

```bash
lufy-ai context query --target <repo> --json "<term>"
```

6. Para impacto por diff cuando aplique:

```bash
lufy-ai context diff --target <repo> --json --base <ref>
```

7. Para explicar una relación antes de usarla como pista:

```bash
lufy-ai context explain --target <repo> --json <node-or-edge>
```

## Reglas

- Devolver solo hints compactos: `node`, `path`, `kind`, `reason`, `status`, `rank`, `confidence`, `relevance`, `matched_signals`, `neighbors`, `noise`, `next_commands`.
- Priorizar salidas que ahorren tokens: top nodos, vecinos acotados, comunidades afectadas, preguntas sugeridas y `token_savings`.
- Tratar el grafo como preflight obligatorio para orientación inicial cuando está habilitado, no como evidencia superior a archivos actuales, diff, tests, logs o comandos de validación.
- No inferir comportamiento runtime solo por edges del grafo; verificar con lectura directa cuando afecte decisiones.
- No ejecutar `context build` salvo que el usuario/rol lo autorice o el flujo lo pida explícitamente; construir el grafo muta `.lufy/context/`.
- Si la CLI no existe, falla o falta el grafo, degradar a `not_available` sin bloquear el trabajo.
- Registrar en Result Contract `memory_provider_used`, `context_graph_status`, `context_graph_queries`, `fallback_reason` y `generic_discovery_before_graph`.

## Resultado

```yaml
context_graph_hints:
  provider: lufy-ai-context
  status: available | stale | not_available
  recovery: lufy-ai context build | not_applicable
  hits:
    - node: <id or not_available>
      path: <path or not_available>
      kind: <kind or not_available>
      reason: <short reason>
      rank: <1-based rank or not_available>
      score: <ranking score or not_available>
      confidence: <high | medium | low | not_available>
      relevance: <why it matters>
      matched_signals:
        - <lexical/path/type/degree signal or not_available>
  noise: true | false | not_available
  next_commands:
    - <focused next command or not_available>
  token_savings: <bounded hints summary>
  suggested_questions:
    - <question or not_available>
```
