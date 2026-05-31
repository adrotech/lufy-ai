# Design: seleccion explicita de tool y metodologia

## Decisiones

1. El unico `tool` escribible en este slice es `opencode`.
2. `codex` y `claude-code` pueden aparecer como valor conocido en documentacion futura, pero no se aceptan en comandos mutantes hasta tener adapters dry-run validados.
3. Los overrides por tier se expresan como `TIER:METHODOLOGY[/MODE]`.
4. Si el modo no se especifica, se infiere de forma conservadora:
   - `openspec` en T1 -> `full`, required `true`.
   - `openspec` en T2 -> `lite`, required `true`.
   - `none` -> `none`, required `false`.
   - `lufy-sdd` queda reservado, pero no instalable mientras no exista adapter real.
5. `none` en T1/T2 queda bloqueado por default en CLI para evitar perder trazabilidad.

## Forma CLI propuesta

```bash
lufy-ai install --tool opencode
lufy-ai install --methodology-tier T3:none
lufy-ai install --methodology-tier T2:openspec/lite
lufy-ai verify --tool opencode --json
lufy-ai status --json
```

Se permite repetir `--methodology-tier` para varios tiers.

## Modelo interno

Se reutiliza `domain.HarnessConfig`:

```go
type HarnessConfig struct {
  Tool              ToolID
  MethodologyByTier MethodologyByTier
}
```

La CLI traduce flags a `HarnessConfig`, valida con reglas de dominio y lo pasa a servicios de install/verify/status cuando aplique.

## Persistencia

El manifest v2 ya contiene:

- `tool`
- `methodologyByTier`
- ownership por asset

Este slice debe asegurar que el manifest refleja el contexto seleccionado, no solo defaults hardcodeados.

## Reglas de error

- `--tool codex` o `--tool claude-code`: error de uso hasta que exista adapter real habilitado.
- `--methodology-tier T4:none`: error de uso por tier invalido.
- `--methodology-tier T1:none`: error de uso por metodologia insegura para T1.
- `--methodology-tier T2:none`: error de uso por metodologia insegura para T2 en este slice.
- `--methodology-tier T3:openspec/full`: permitido si el usuario quiere gobernanza formal en T3.

## Riesgos

- Endurecer demasiado la CLI puede bloquear experimentacion futura. Mitigacion: documentar que adapters dry-run llegaran en cambios separados.
- Mezclar seleccion de metodologia con instalacion de assets puede crear expectativas de Lufy SDD. Mitigacion: `lufy-sdd` sigue reservado pero no escribible.
