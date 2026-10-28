# Lufy inline en Codex

El principal trabaja con `execution.mode: inline` y la política de delegación del proyecto. No se instala una cadena de roles.

- `config.toml` habilita capacidades del host; no concede permisos ni prueba que una delegación haya ocurrido.
- `hooks.json` sólo orienta `SessionStart`; no procesa reportes ni registra mensajes o sesiones.
- `rules/lufy.rules` conserva controles de entrega y comandos destructivos.
- `lufy-agent-mapping.md` describe la ejecución y las capacidades reales.

La política de trabajo y comprobación vive en `.lufy/contracts/work.md` y `.lufy/contracts/evidence.md`. Los datos y decisiones explícitas del proyecto viven en `.lufy/config/project.yaml`. Usa capacidades de `.agents/skills` cuando sean pertinentes; un registro estructurado no es requisito para responder al usuario.
