# Lufy Agent Mapping for Codex

This file is managed by `lufy-ai`.

Codex project-local files under `.codex/agents/*.toml` define custom agents for Lufy role behavior. Use Lufy roles in `native` mode when tool discovery exposes the exact role as invokable through Codex subagent tooling. If a specific Codex runtime exposes only generic multi-agent roles such as `default`, `explorer`, and `worker`, degrade explicitly to `emulated` or `inline`.

`@orchestrator`, `@implementer`, or any other `@<lufy-role>` mention is a delegation request, not a prompt-style identity switch. The parent assistant must not answer "as" that role in the same thread and then continue executing. Debe invocar el subagente y esperar un resultado válido. Si no hay herramientas de delegación, reportar `blocked`; una solicitud explícita de delegación no se convierte en inline sin una nueva decisión del usuario. Cerrar el agente cuando la API lo exponga; si falta esa operación, registrar cleanup `not_available`, sin inventar cierre.

## Required Runtime Statement

Before delegating Lufy workflow work, the assistant must state one of these modes in the handoff or result:

- `native`: the exact Lufy role is available as an invokable Codex role.
- `emulated`: the exact Lufy role is not available and the assistant is mapping it to a generic Codex role.
- `inline`: no suitable subagent role is available and the assistant is executing the phase inline.

Usar `inline` sólo cuando el alcance no requiera delegación y no haya una solicitud explícita de invocar un rol. Tras una mención `@<lufy-role>`, si faltan herramientas, reportar `blocked` y esperar una decisión explícita; no convertir esa solicitud en ejecución inline.

## Default Mapping

First try the exact native Lufy role. Use this mapping only when native Lufy roles are unavailable:

| Lufy role | Codex role | Notes |
| --- | --- | --- |
| `orchestrator` | `default` | Coordinate locally; do not claim isolated orchestration. |
| `sdd-router` | `explorer` | Read-only routing, sizing, risks, and next owner. |
| `explorer` | `explorer` | Read-only discovery. |
| `implementer` | `worker` | Scoped edits only; no delivery. |
| `test-writer` | `worker` | Test-focused edits only; report RED/GREEN evidence. |
| `validator` | `explorer` | Read-only validation and diagnosis. |
| `reviewer` | `explorer` | Read-only review; lead with findings. |
| `delivery` | `worker` | Only after explicit delivery authorization. |

## Guardrails

- Resolver skills una sola vez por contexto: usar `.lufy/skill-registry.json` ready y pasar paths exactos al `SKILL.md` original. Si falta o está stale, usar `.agents/skills` como fallback explícito y recomendar `lufy-ai skills ensure --target <repo> --tool codex`; ejecutar ensure sólo con autorización aplicable. `install`, `setup` y `sync` conservan su ensure best-effort.
- Treat `@<lufy-role>` as an explicit delegation request; do not continue in the parent thread as if the role was invoked.
- Native delegation exige invocación del rol exacto y handoff esperado con evidencia. Cerrar cuando la API lo permita y registrar cualquier cleanup no disponible; el modo native no demuestra por sí solo la cadena completa del protocolo de regresión.
- A waited subagent that reports `state=completed` with empty/null `task_result`, empty body, or no non-empty Result Contract/evidence is an invalid result. Recover once with the same `task_id` when available; if recovery still lacks payload, report `blocked` with the exact recovery action instead of continuing silently.
- Never tell the user that Lufy roles were used natively when they were only emulated.
- Pass the Lufy role instructions and minimum context into the generic Codex role prompt.
- Preserve role permissions: read-only Lufy roles stay read-only even when mapped to a more capable generic role.
- Include the selected `agent_execution_mode` and `role_mapping` in substantive Result Contract evidence.
- Preserve `adapter_context` across native/emulated/inline execution. Tier and execution mode do not override the configured methodology: route `lufy-sdd/full|lite` through native Lufy SDD skills and carry the automatic overview path/status without a render prompt.

## Multi-Artifact Branching

- Multi-artifact branching is coordination, not a new Codex/Lufy role. Keep using the existing mapping: `sdd-router` recommends `artifact_branching`, `orchestrator` coordinates candidate generation and join, `implementer`/`worker` writes isolated candidates when asked, `reviewer` compares objective quality/risk, and the human decides public-contract, security, product, UX, or non-objective trade-offs.
- Candidate generation may be native, emulated, or inline according to the runtime statement above, but it must preserve isolated non-overwriting candidate paths, a merge plan, `candidate_count <= 2`, and a mandatory join before downstream implementation.
- Do not use branching to authorize delivery, Git/GH work, or unresolved public-contract/security decisions.

## Lifecycle y permisos observados

`.codex/hooks.json` configura callbacks soportados de lifecycle; el proyecto y la definición de hooks deben ser de confianza. Un callback observado prueba ese perfil, no todos los runtimes ni una delegación nativa completa. Los ensayos CLI del bloque 3 quedaron inconclusos; no convertir flags anunciados en una capacidad verificada. Los hooks no leen transcripts ni infieren eventos semánticos de review, permisos o gates. El host registra reviews explícitas con `lufy-ai review`; un sidecar `candidate-evidence/v1` sólo verifica vigencia del candidato declarado.

El sandbox y las aprobaciones efectivos se heredan del padre. Una instrucción read-only del rol sigue siendo obligatoria, pero el archivo del agente no crea un sandbox independiente ni puede elevar permisos. Conservar las restricciones del host y reportar capacidades no disponibles. Véase la [documentación oficial de subagentes](https://learn.chatgpt.com/docs/agent-configuration/subagents).

## Fuente y mantenimiento

`.lufy/contracts/delivery.md` conserva autoridad. `.codex/instructions/proportional-workflow.md` es la fuente compacta del wording compartido; los segmentos marcados de orchestrator/sdd-router/implementer se materializan inline para preservar contexto. Los mirrors embedded son distribución, no una segunda fuente de edición.

En el repositorio fuente de Lufy, comprobar con `python3 scripts/render-codex-instructions.py --check`; después de una edición autorizada, usar `--write` y repetir el check. El renderer valida todas las rutas/markers antes de escribir, conserva instrucciones particulares y sandbox_mode, y no cambia configuración global. Las pruebas detectan drift textual; no prueban obediencia, aislamiento, latencia ni tokens. Reutilizar plan, aceptación y evidencia ya recibidos; reabrir exploración sólo ante alcance, evidencia o riesgo nuevos.
