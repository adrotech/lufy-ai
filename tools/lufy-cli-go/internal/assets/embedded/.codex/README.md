# Lufy Codex Surface

This directory contains project-local Codex configuration managed by `lufy-ai`.

- `agents/`: custom Codex agents for Lufy roles.
- `lufy-agent-mapping.md`: required native/emulated/inline role mapping guidance for Codex runtimes, including `@<lufy-role>` delegation handling.
- `hooks.json`: lifecycle Lufy para orientación, diagnóstico de handoffs, estado SDD y validación best-effort de memoria. Los hooks no leen transcripts ni avanzan gates.
- `rules/`: policy conservadora para delivery y comandos destructivos; una decisión técnica `prompt` nunca reemplaza autorización del usuario.
- `config.toml`: minimal project Codex configuration with multi-agent enabled.

Codex loads project-local `.codex` layers only when the project is trusted.
Review and trust changed hooks with `/hooks`. If `lufy-ai` is unavailable, SessionStart emits recovery and continues without advancing workflow state.
Use Lufy custom agents natively when tool discovery exposes them. Treat `@orchestrator` and other `@<lufy-role>` mentions as delegation requests: spawn the requested subagent, wait for its result, close it, then synthesize. If a Codex runtime exposes only generic invokable roles (`default`, `explorer`, `worker`) or no subagent tooling, degrade explicitly through the mapping guide instead of reporting native Lufy subagents or continuing inline silently.
