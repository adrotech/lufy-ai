# Proposal: Lufy SDD methodology adapter foundation

## Why

El core ya reconoce `lufy-sdd` como metodologia futura, pero no existe adapter que modele sus modos ni su superficie esperada. Para avanzar sin romper el preset OpenCode/OpenSpec, necesitamos una primera fundacion verificable del adapter Lufy SDD.

## What Changes

- Registrar `lufy-sdd` como methodology adapter.
- Soportar modos `lite` y `full`.
- Renderizar specs conceptuales para `.lufy/sdd/` sin integrarlas todavia al catalogo instalable default.
- Mantener `install --methodology-tier ...:lufy-sdd/...` bloqueado hasta un spec posterior que conecte renderer/catalog/sync.
- Documentar que Lufy SDD existe como foundation, no como workflow default.

## Non-goals

- No migrar cambios OpenSpec existentes a Lufy SDD.
- No reemplazar `/opsx-*`.
- No escribir `.lufy/sdd/` desde `install`.
- No cambiar defaults T1/T2/T3.
