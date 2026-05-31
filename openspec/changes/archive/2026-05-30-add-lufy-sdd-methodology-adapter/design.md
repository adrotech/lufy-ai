# Design: Lufy SDD methodology foundation

## Adapter

El adapter `lufy-sdd` implementa `ports.MethodologyAdapter`:

- `ID() = lufy-sdd`
- `SupportedModes() = full, lite`
- `RenderWorkflow()` devuelve asset specs conceptuales bajo `.lufy/sdd/`
- `VerifyWorkflow()` devuelve checks informativos hasta que haya persistencia real

## Modos

- `lite`: pensado para T2, con `changes/`, `decisions/` y `verification/`.
- `full`: pensado para T1, con `changes/`, `specs/`, `decisions/` y `verification/`.

## CLI

Aunque el adapter queda registrado, la CLI mutante mantiene `lufy-sdd` bloqueado en `--methodology-tier`. El motivo es que `install` todavia consume el catalogo actual OpenCode/OpenSpec y no debe persistir una metodologia seleccionada que no coincide con los assets instalados.

La habilitacion de `lufy-sdd` en install/sync requiere un spec posterior que conecte:

- methodology adapter;
- instruction renderer;
- catalog filtering;
- manifest ownership;
- verify/status.

## Riesgos

- Confundir adapter foundation con metodologia usable por usuarios finales.
- Duplicar OpenSpec bajo otro nombre.

Mitigacion: mantener alcance de foundation, tests de paths `.lufy/sdd/` y docs claras.
