---
name: lufy.mem-capture
description: Captura aprendizajes durables en .lufy/memory/knowledge usando Obsidian como memoria canónica portable.
license: MIT
compatibility: OpenCode skill autocontenido; requiere lufy-ai memory para validar estructura.
metadata:
  author: lufy-ai
  version: "1.0"
---

# Skill: lufy.mem-capture

Captura únicamente memoria durable: decisiones, reglas, flows, lessons o conceptos que reducen redescubrimiento futuro. No guardar ruido rutinario, estados temporales, logs, resultados obvios de comandos ni duplicados.

## Flujo

1. Leer `.lufy/config/project.yaml` y confirmar `memory.provider: obsidian`; esta es la única memoria de proyecto para capturas durables.
2. Si falta estructura, recomendar `lufy-ai memory init --target <repo>`, reportar `memory_provider_used: obsidian:not_available` y detener mutaciones.
3. Buscar primero notas cercanas con `lufy-ai memory search --target <repo> <query>`.
4. Si existe una nota activa suficiente, actualizarla con el menor cambio posible mediante `lufy-ai memory capture --target <repo> --title <title> --type <type> [--link <slug>] <texto>`.
5. Si hace falta una nota nueva, crearla con el mismo comando bajo `.lufy/memory/knowledge/<slug>.md` con frontmatter:

```yaml
---
name: <short-name>
description: <contexto concreto, no igual al name>
type: decision | rule | flow | lesson | concept
status: active
---
```

6. Para `type: decision`, incluir una sección `**Why:**`.
7. Si el usuario corrige a la IA, capturar la corrección como `type: rule` o `type: lesson` y conectarla con notas existentes mediante `lufy-ai memory connect` cuando aplique.
8. Ejecutar o recomendar `lufy-ai memory validate --target <repo>`.
9. No capturar memoria de proyecto en Engram/MCP cuando Obsidian está configurado; si el host usa memoria externa, etiquetarla como non-project/fallback y registrar `fallback_reason`.
