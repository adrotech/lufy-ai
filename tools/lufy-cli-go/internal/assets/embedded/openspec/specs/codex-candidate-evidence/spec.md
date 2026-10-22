# Codex candidate evidence

## Purpose
Vincular evidencia declarada a identidades Git y políticas observadas para detectar obsolescencia sin avanzar gates.

## Requirements

### Requirement: Exact candidate evidence binding
Lufy SHALL vincular evidencia canónica validada a repositorio, base, candidato comprometido, diff de paths/modos y políticas efectivas conocidas, sin modificar Result Contract v1 ni avanzar gates.

#### Scenario: Current candidate
- **WHEN** el candidato está limpio y todas las identidades coinciden con el sidecar
- **THEN** verify devuelve current sin mutar Git ni evidencia

#### Scenario: Stale or dirty candidate
- **WHEN** cambia HEAD/base, repositorio, paths/modos, política, config o evidencia, o hay cambios sin commit
- **THEN** el binding no se acepta como vigente y se requiere nueva captura tras revisión

#### Scenario: Strict sidecar
- **WHEN** el sidecar contiene claves duplicadas/unknown, trailing data o identidad incompleta
- **THEN** se rechaza sin fabricar autoridad ni reescribir el resultado original

#### Scenario: Read-only execution
- **WHEN** se captura o verifica un candidato
- **THEN** se usan lecturas Git bounded sin hooks, shell, external diff, textconv, modelos ni transiciones de gates

#### Scenario: Executable repository configuration
- **WHEN** hay filtros Git configurados o gitlinks/submódulos
- **THEN** se rechaza antes de status y sin ejecutar callbacks ni entrar a repositorios hijos
