# Design: Installable Lufy SDD methodology

## Decision

El catalogo se mantiene como fuente de assets disponibles, pero el instalador y sync deben derivar un catalogo efectivo desde `HarnessConfig`. El filtro conserva assets core/tool y deja pasar solo las metodologias presentes en `methodologyByTier`.

## Asset model

Assets nuevos:

```text
.lufy/sdd/README.md
.lufy/sdd/changes/.gitkeep
.lufy/sdd/decisions/.gitkeep
.lufy/sdd/verification/.gitkeep
.lufy/sdd/specs/.gitkeep
```

Ownership:

- `methodology: lufy-sdd`
- `component: methodology-surface`
- `tool: opencode` mientras OpenCode sea la unica tool instalable

`lite` no requiere `.lufy/sdd/specs`; el filtro omite esa rama salvo que exista al menos un tier `lufy-sdd/full`.

## CLI

`parseMethodologyTier` debe aceptar:

- `T1:lufy-sdd` como `full`
- `T2:lufy-sdd` como `lite`
- `T3:lufy-sdd` como `lite`
- modos explicitos `full` y `lite`

`none` sigue gobernado por policy: bloqueado en T1 y requiere justificacion futura en T2.

## Verify and sync

`verify` debe derivar requirements desde el manifest instalado. OpenSpec deja de ser un fallback incondicional; si un manifest usa solo `lufy-sdd`, verify no debe exigir `openspec/config.yaml`.

`sync` debe usar la metodologia del manifest previo para calcular assets nuevos y retirados, evitando mezclar assets de metodologias no seleccionadas.
