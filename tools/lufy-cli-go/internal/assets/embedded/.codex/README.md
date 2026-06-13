# Lufy Codex Surface

This directory contains project-local Codex configuration managed by `lufy-ai`.

- `agents/`: custom Codex agents for Lufy roles.
- `lufy-agent-mapping.md`: required native/emulated/inline role mapping guidance for Codex runtimes.
- `hooks.json`: lifecycle hook placeholders for future local validation/memory workflows.
- `rules/`: project-local execution policy rules.
- `config.toml`: minimal project Codex configuration.

Codex loads project-local `.codex` layers only when the project is trusted.
Use Lufy custom agents natively when tool discovery exposes them. If a Codex runtime exposes only generic invokable roles (`default`, `explorer`, `worker`), degrade explicitly through the mapping guide instead of reporting native Lufy subagents.
