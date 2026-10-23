# Diseño

El texto proporcional existente se conserva byte a byte como fuente compacta de render; delivery.md conserva precedencia. Marcadores Markdown dentro del developer_instructions delimitan únicamente el bloque común. Un renderer Python stdlib con `--check` por defecto y `--write` explícito verifica fuente/mirror y tres roles/mirrors antes de escribir, rechazando symlinks/markers ambiguos. No reescribe el TOML ni toca sandbox_mode, modelos, hooks, roles especializados o reglas de delivery. No reduce contexto inline antes del experimento.

El mapping es la fuente Codex de selección native/emulated/inline y recuperación; README remite a ella en vez de duplicar el procedimiento completo. Hooks configurados requieren confianza del proyecto/definición; callbacks observados prueban sólo ese perfil. El ensayo CLI del bloque 3 inconcluso no invalida el lifecycle implementado ni demuestra native chain. Las instrucciones read-only son restricciones del rol y no una garantía independiente de sandbox contra el host: preservar el padre y no elevar permisos.

Tests offline comprueban drift, render idempotente, preservación de instrucciones particulares, rechazo de markers/symlinks y paridad completa de .codex con assets. No evalúan obediencia del modelo, latencia ni aislamiento efectivo. Las antiguas comprobaciones proporcionales siguen protegiendo el contenido inline.
