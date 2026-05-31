# Proposal: CLI flags de seleccion para harness adapters

## Problema

El refactor hexagonal ya introdujo modelos de tool adapter, metodologia por tier, ownership en manifest y renderer compatible con OpenCode/OpenSpec. Sin embargo, la CLI todavia no expone una forma explicita y validada de seleccionar esos adapters al instalar, sincronizar o verificar.

Mientras no exista esa interfaz, `lufy-ai` sigue funcionando como preset implicito aunque el core ya modele el harness como neutral. Esto dificulta preparar compatibilidad futura con Codex, Claude Code y Lufy SDD sin romper defaults.

## Objetivo

Agregar una superficie CLI minima para declarar el contexto de harness efectivo:

- `--tool opencode` como seleccion explicita del unico adapter real soportado.
- overrides por tier de metodologia para `install`, con defaults actuales preservados.
- rechazo claro de tools o metodologias no soportadas para escritura real.
- `verify`/`status --json` como fuente inspeccionable de `tool` y `methodologyByTier`.

## Alcance

Incluye:

- parsing y validacion de `--tool`.
- parsing y validacion de overrides por tier, por ejemplo `--methodology-tier T3:none`.
- bloqueo de `none` inseguro en T1 y T2 sin justificacion en este slice.
- propagacion del contexto seleccionado al manifest v2.
- tests de CLI, dominio, install y status/verify JSON cuando aplique.
- documentacion de uso de flags.

No incluye:

- soporte real de `codex` o `claude-code`.
- escritura experimental de assets Codex/Claude.
- implementacion de `lufy-sdd` full/lite.
- migration wizard interactivo.

## Compatibilidad

`lufy-ai install` sin flags debe conservar el preset actual:

```yaml
tool: opencode
methodology_by_tier:
  T1: openspec/full/required
  T2: openspec/lite/required
  T3: none/none/not-required
```

`--tool opencode` debe ser equivalente al default.
