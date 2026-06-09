---
name: lufy.mem-document
description: Documenta decisiones, reglas y flujos en Obsidian con schema validable y backlinks seguros.
license: MIT
compatibility: OpenCode skill autocontenido; no requiere servicios externos.
metadata:
  author: lufy-ai
  version: "1.0"
---

# Skill: lufy.mem-document

Convierte contexto ya validado en una nota Obsidian navegable. Usar cuando el usuario pide documentar una decisión, regla, aprendizaje, flujo o perfil estable del proyecto.

## Reglas

- Escribir en `.lufy/memory/knowledge/` salvo que sea perfil de app, que vive en `.lufy/memory/maps/_app-profile.md`.
- No versionar contenido privado; la política default de `.lufy/memory/.gitignore` lo ignora.
- Mantener notas breves, fechables por contexto cuando ayude, y conectadas con backlinks existentes.
- No usar backlinks a notas inexistentes salvo que se creen en el mismo bloque.
- Validar con `lufy-ai memory validate --target <repo>` cuando haya mutaciones.

## Estructura recomendada

```markdown
---
name: <slug-humano>
description: <contexto concreto>
type: decision
status: active
---

## Summary

<hecho durable>

**Why:** <razón>

## Links

- [[otra-nota]]
```
