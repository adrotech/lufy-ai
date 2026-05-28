---
description: Onboarding guiado LUFY para validar instalación y explicar una demo T3 stack-aware.
agent: orchestrator
---

Ejecuta `/lufy.onboard` delegando la guía al skill local `lufy.onboard`.

## Entradas

- `--demo`: preparar una demo T3 dummy adaptada al stack declarado en `.opencode/project.yaml`.
- `--dry-run`: no realizar mutaciones; explicar pasos, archivos candidatos y validación detectada.

## Comportamiento

1. Carga el skill `lufy.onboard`.
2. Valida de forma estática la instalación local mínima.
3. Lee `.opencode/project.yaml` como fuente canónica de stacks y toolchain.
4. Para `--demo --dry-run`, explica un T3 dummy stack-aware sin mutar archivos.
5. Para `--demo` sin dry-run, guía el cambio y pide confirmación/deriva a `implementer` antes de crear artefactos; no hace delivery.

## Degradación esperada

- Si falta `.opencode/project.yaml`, recomendar `lufy-ai init [--target <path>]` antes de continuar.
- Si el stack es unsupported, `supported: false` o contiene comandos `TODO`, explicar el flujo de forma estática y pedir completar `.opencode/project.yaml`.

## Ejemplos

```text
/lufy.onboard --demo --dry-run
/lufy.onboard --demo
```
