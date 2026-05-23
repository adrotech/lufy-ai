---
description: Genera un reporte local de tiempo/ROI LUFY en HTML autocontenido.
agent: orchestrator
---

Genera `/lufy.timereport` delegando la ejecución al skill local `lufy.timereport`.

## Entradas

- `--output <ruta>` opcional: ruta explícita del HTML a generar. Si se omite, usar `/tmp/lufy-timereport-<timestamp>.html`.
- `--target-dir <ruta>` opcional: repositorio a analizar. Si se omite, usar el directorio actual de OpenCode.
- `--from <fecha>` / `--to <fecha>` opcionales: rango temporal en formato aceptado por Git/Python. Si se omiten, el skill usa la actividad disponible de OpenCode y Git.
- `--db <ruta>` opcional para validación local: SQLite de OpenCode alternativo. Por defecto usar `~/.local/share/opencode/opencode.db`.

## Comportamiento

1. Carga el skill `lufy.timereport`.
2. Ejecuta el generador local/offline documentado por el skill.
3. Reporta la ruta final del HTML o una degradación accionable si faltan fuentes.

## Privacidad por defecto

- No incluir prompts, respuestas completas, argumentos completos de tools, outputs de tools, contenidos de archivos, diffs ni `session_diff`.
- No leer JSONL ni `session_diff` salvo que un cambio futuro agregue opt-in explícito.
- Todas las fuentes son locales y read-only: OpenCode SQLite, Git y `.opencode/project.yaml`.

## Ejemplos

```bash
python3 .opencode/skills/lufy.timereport/scripts/generate_timereport.py
python3 .opencode/skills/lufy.timereport/scripts/generate_timereport.py --output /tmp/lufy.html --target-dir "$PWD"
```
