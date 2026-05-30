# Proposal: Enable Lufy SDD as installable methodology

## Problem

`lufy-sdd` ya existe como ID y adapter foundation, pero la CLI mutante sigue bloqueando su seleccion. Esto deja la etapa 12 del plan hexagonal incompleta: el usuario no puede elegir `lufy-sdd/lite` o `lufy-sdd/full` por tier.

El catalogo actual tambien instala OpenSpec de forma global, sin filtrar por metodologia efectiva. Para que Lufy sea un harness abstracto, los assets metodologicos deben depender de `methodologyByTier`.

## Goals

- Permitir `--methodology-tier T2:lufy-sdd/lite` y `T1:lufy-sdd/full`.
- Agregar assets gestionados minimos bajo `.lufy/sdd/`.
- Filtrar assets metodologicos del catalogo segun las metodologias seleccionadas por tier.
- Mantener OpenCode como unica tool instalable y OpenSpec como default compatible.
- Mantener `none` disponible para T3 sin instalar assets metodologicos formales.

## Non-goals

- No reemplazar el workflow OpenSpec default.
- No crear todavia comandos `/lufy.sdd-*`.
- No migrar cambios existentes de OpenSpec a Lufy SDD.
- No habilitar Codex o Claude Code para escritura real.
