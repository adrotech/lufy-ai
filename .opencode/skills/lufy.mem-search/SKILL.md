---
name: lufy.mem-search
description: Busca hints compactos en la memoria Obsidian portable sin tratarla como evidencia superior a archivos o comandos.
license: MIT
compatibility: OpenCode skill autocontenido; usa lufy-ai memory.
metadata:
  author: lufy-ai
  version: "1.0"
---

# Skill: lufy.mem-search

Usar al inicio de trabajos T1/T2 no triviales y T3 con alta probabilidad de contexto histórico.

## Flujo

1. Leer/confirmar `.lufy/config/project.yaml`. Si declara `memory.provider: obsidian`, Obsidian es la memoria de proyecto obligatoria.
2. Ejecutar `lufy-ai memory status --target <repo> --json`.
3. Si no está inicializada, reportar `memory_provider_used: obsidian:not_available`, `memory_hints: not_available` y `fallback_reason` con recovery `lufy-ai memory init --target <repo>`; continuar con archivos del repo o memoria externa solo si se etiqueta como fallback/non-project.
4. Buscar con consultas cortas por issue, spec, ruta, concepto o decisión:

```bash
lufy-ai memory search --target <repo> "<query>"
```

5. Devolver solo hints compactos: `path`, `line`, `status`, `relevance`.
6. No copiar notas completas ni reemplazar inspección de archivos reales.
7. No presentar Engram/MCP como memoria de proyecto si Obsidian está configurado; úsalo solo como fallback explícito o memoria no-project con `fallback_reason`.
8. Si el rol también necesita discovery amplio y `context_graph.enabled=true`, coordinar con `lufy.context-search` y preservar diagnósticos `context_graph_status`, `context_graph_queries` y `generic_discovery_before_graph` junto a `memory_provider_used`.

## Resultado

```yaml
memory_hints:
  provider: obsidian
  status: available | not_available
  hits:
    - path: .lufy/memory/knowledge/example.md
      line: 12
      relevance: "decisión previa relacionada"
```
