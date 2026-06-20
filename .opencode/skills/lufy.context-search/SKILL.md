---
name: lufy.context-search
description: Usa lufy-ai context como índice local secundario para buscar hints compactos de arquitectura, impacto y rutas sin sustituir archivos, diff o comandos.
license: MIT
compatibility: OpenCode skill autocontenido; usa lufy-ai context.
metadata:
  author: lufy-ai
  version: "1.0"
---

# Skill: lufy.context-search

Usar cuando una tarea necesita orientación rápida sobre relaciones de archivos, símbolos, docs u OpenSpec y el repositorio puede tener `.lufy/context/graph.json` generado.

## Flujo

1. Consultar disponibilidad:

```bash
lufy-ai context status --target <repo> --json
```

2. Si el estado es `not_available` o `stale`, reportar `context_graph_hints.status: not_available` o `stale` con `recovery: lufy-ai context build`; continuar con inspección normal del repositorio.
3. Para búsqueda lexical de hints:

```bash
lufy-ai context query --target <repo> --json "<term>"
```

4. Para impacto por diff cuando aplique:

```bash
lufy-ai context diff --target <repo> --json --base <ref>
```

5. Para explicar una relación antes de usarla como pista:

```bash
lufy-ai context explain --target <repo> --json <node-or-edge>
```

## Reglas

- Devolver solo hints compactos: `node`, `path`, `kind`, `reason`, `status`, `relevance`.
- Tratar el grafo como índice secundario; no es evidencia superior a archivos actuales, diff, tests, logs o comandos de validación.
- No inferir comportamiento runtime solo por edges del grafo; verificar con lectura directa cuando afecte decisiones.
- No ejecutar `context build` salvo que el usuario/rol lo autorice o el flujo lo pida explícitamente; construir el grafo muta `.lufy/context/`.
- Si la CLI no existe, falla o falta el grafo, degradar a `not_available` sin bloquear el trabajo.

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
      relevance: <why it matters>
```
