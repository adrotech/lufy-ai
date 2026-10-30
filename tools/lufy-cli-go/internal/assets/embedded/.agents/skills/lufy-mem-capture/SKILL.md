---
name: lufy-mem-capture
description: Captura aprendizajes durables en la memoria Obsidian portable de Lufy cuando el proyecto la utiliza.
---

# Capturar memoria

Usá esta capacidad para decisiones, reglas, flujos, conceptos y aprendizajes reutilizables; evitá resúmenes rutinarios.

- Comprobá que `.lufy/config/project.yaml` configure memoria Obsidian y buscá notas relacionadas antes de duplicarlas.
- Preferí `lufy-ai memory capture --target <repo> --title <titulo> --type <decision|rule|flow|lesson|concept> [--link <slug>] <texto>` a editar Markdown manualmente. Una decisión debe explicar su porqué.
- Conservá correcciones técnicas explícitas del usuario como reglas o aprendizajes cuando sean durables. No guardes secretos ni logs innecesarios.
- Conectá sólo notas existentes con `memory connect` y comprobá `memory validate` después de cambiar memoria.
