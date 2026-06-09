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

1. Ejecutar `lufy-ai memory status --target <repo>`.
2. Si no está inicializada, reportar `memory_hints: not_available` y continuar.
3. Buscar con consultas cortas por issue, spec, ruta, concepto o decisión:

```bash
lufy-ai memory search --target <repo> "<query>"
```

4. Devolver solo hints compactos: `path`, `line`, `status`, `relevance`.
5. No copiar notas completas ni reemplazar inspección de archivos reales.

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
