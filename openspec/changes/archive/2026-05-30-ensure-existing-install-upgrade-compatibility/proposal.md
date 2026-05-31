# Proposal: garantizar compatibilidad de install/update actual

## Contexto

El preset actual de usuarios productivos es `tool=opencode` con metodología `openspec`. El refactor hexagonal ya enruta el catálogo efectivo por adapters, pero antes de seguir abstrayendo más piezas necesitamos fijar evidencia de que usuarios existentes pueden instalar, sincronizar y verificar sin cambios de flags ni migraciones manuales.

## Objetivo

Garantizar que `lufy-ai install`, `lufy-ai sync` y `lufy-ai verify` preserven el comportamiento observable del preset OpenCode/OpenSpec actual después de introducir adapters de tool/metodología.

## No objetivos

- No habilitar escritura real para Codex o Claude Code.
- No cambiar defaults públicos.
- No instalar `lufy-sdd` en usuarios existentes salvo selección explícita.

## Impacto esperado

- Cobertura de regresión para instalación default sin `.lufy/sdd`.
- Cobertura de upgrade/sync sobre targets ya instalados.
- Cobertura de verify posterior a sync para manifest y catálogo adapter-driven.
