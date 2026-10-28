# AGENTS.md

Guía de trabajo para `lufy-ai`.

- Respuestas, documentación humana y comentarios en español; conserva identificadores técnicos.
- CLI Go: `tools/lufy-cli-go`. `scripts/install.sh` es su wrapper estricto.
- No hay package.json ni toolchain Node raíz. No inventes comandos de validación.
- Usa `scripts/validate.sh` para la validación agrupada pertinente; reporta comandos y resultados reales.
- Analiza dependencias e interconexiones antes de modificar. Preserva trabajo local ajeno.
- Mantén paridad entre assets fuente seleccionados y sus mirrors en `tools/lufy-cli-go/internal/assets/embedded`.
- `workflow_limits` es la única fuente de límites del proyecto; no consumas aliases top-level.
- El catálogo activo es Codex inline. No reintroduzcas selección de metodología por tier ni lectores de formatos anteriores.
- `develop` es integración, `main` estable; entrega sólo con autorización explícita.

<!-- LUFY:BEGIN codex-harness -->
# Lufy AI Harness

- Responde en español y trabaja como agente principal inline.
- Lee `execution` en `.lufy/config/project.yaml` y aplica `.lufy/contracts/work.md`.
- Entiende el resultado pedido y el contexto existente; actúa dentro del alcance autorizado hasta comprobarlo.
- Adapta la evidencia a análisis, propuestas, escritura, código, datos u operaciones. No exijas tiers, metodología ni handoffs para avanzar.
- Delega según `execution.delegation`: `off`, `on_request` o `auto`. Una instrucción inline del usuario prevalece; delegar nunca amplía permisos.
- Usa capacidades y skills locales pertinentes, con contexto mínimo, sin imponer una cadena de roles.
- Consulta sólo decisiones materiales que no puedas resolver; continúa el trabajo independiente mientras esperas.
- Preserva trabajo ajeno, comprueba el resultado y reporta evidencia real y límites pendientes.
- Commit, push, PR, publicación y operaciones externas requieren autorización explícita; aplica `.lufy/contracts/delivery.md` cuando corresponda.
<!-- LUFY:END codex-harness -->
