# Superficie Lufy para Codex

Esta carpeta contiene assets locales gestionados por `lufy-ai`.

- `agents/`: instrucciones de los roles Lufy; conservan límites particulares y autorización de delivery.
- `lufy-agent-mapping.md`: fuente Codex para modos native/emulated/inline, menciones `@<rol>`, recuperación de handoffs y resolución de skills. Leerla antes de delegar; evita repetir el procedimiento en cada entrada.
- `instructions/proportional-workflow.md`: wording compartido materializado inline en tres roles; la autoridad sigue en `.lufy/contracts/delivery.md`.
- `hooks.json`: lifecycle para orientación, diagnóstico de handoffs, SDD y validación best-effort de memoria; sin transcripts ni avance de gates.
- `rules/`: controles conservadores de delivery y comandos destructivos; `prompt` nunca sustituye autorización humana.
- `config.toml`: configuración local mínima con multi-agent y hooks habilitados; los flags no prueban disponibilidad efectiva.

Las capas del proyecto requieren confianza. Revisar cambios de hooks con `/hooks`; si falta `lufy-ai`, SessionStart informa recovery y conserva el workflow. Consultar [el protocolo de regresión](https://github.com/adrotech/lufy-ai/blob/develop/docs/codex-native-regression-2026-10-22.md) para distinguir configuración de callbacks/handoffs observados. Una instrucción read-only no crea por sí sola un sandbox independiente: se conserva la política efectiva del padre.

En el repositorio fuente de Lufy, antes de publicar cambios, ejecutar `python3 scripts/render-codex-instructions.py --check`. Para materializar una edición compartida autorizada, usar `--write`; después ejecutar la validación agrupada. Los mirrors deben coincidir y permanecen visibles en Git/review. No se adopta una variante de menor contexto sin el experimento controlado del bloque 7.
