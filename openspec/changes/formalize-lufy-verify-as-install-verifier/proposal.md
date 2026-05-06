# Formalizar `lufy-ai verify` como verificador canónico de instalación

## Summary

Formalizar `lufy-ai verify` como la única comprobación canónica de instalaciones `lufy-ai`, reforzando sus checks estructurales y de hashes gestionados sin crear `scripts/verify-install.sh`.

## Motivation

El roadmap histórico mencionaba un posible `scripts/verify-install.sh`, pero la instalación ya migró la lógica crítica a la CLI Go. Mantener un segundo verificador Bash introduciría duplicación, divergencia de criterios y riesgo de reintroducir paths o validaciones legacy. La verificación debe vivir junto al catálogo de assets gestionados, el manifest `.lufy-ai/install-state.json` y el cálculo SHA-256 existente.

## Scope

- Reforzar `lufy-ai verify` para validar explícitamente estructura/categorías críticas instaladas.
- Validar presencia de `.lufy-ai/install-state.json`, archivos críticos y manifest de assets gestionados.
- Mantener la verificación de hashes SHA-256 para todos los assets gestionados registrados.
- Actualizar tests y documentación para que `lufy-ai verify` sea el camino recomendado.
- Quitar `scripts/verify-install.sh` como objetivo futuro.

## Non-Goals

- No crear `scripts/verify-install.sh`.
- No cambiar `scripts/install.sh` ni reintroducir fallback legacy.
- No modificar el esquema de `.lufy-ai/install-state.json`.
- No añadir nuevos comandos de CLI ni cambiar defaults de auth/puertos/configuración global.

## Acceptance Criteria

- `lufy-ai verify --target <dir> --no-engram` falla si faltan categorías críticas: `.opencode/agents`, `.opencode/commands`, `.opencode/skills`, `.opencode/plugins`, `.opencode/policies`.
- `lufy-ai verify` falla si faltan archivos críticos: `.opencode/plugins/agent-observatory.tsx`, `AGENTS.md`, `tui.json`, `openspec/config.yaml` o `.lufy-ai/install-state.json`.
- `lufy-ai verify` falla si un archivo crítico gestionado no está registrado en el manifest o si cualquier asset gestionado tiene drift de SHA-256.
- La documentación apunta a `lufy-ai verify` y no propone crear/integrar `scripts/verify-install.sh`.
- Tests Go y validación OpenSpec/diff solicitada pasan o reportan evidencia real.
