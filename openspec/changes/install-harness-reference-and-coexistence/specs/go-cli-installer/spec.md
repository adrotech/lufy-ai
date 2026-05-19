## MODIFIED Requirements

### Requirement: Idempotencia y preservación de trabajo del usuario
La instalación SHALL ser idempotente y MUST NOT sobrescribir trabajo del usuario sin estrategia explícita, backup y confirmación cuando corresponda; `AGENTS.md` SHALL be treated as user-owned and integrated only through the minimal `@lufy-ia.harness.md` reference.

#### Scenario: Reinstalación sin cambios
- **WHEN** el usuario ejecuta `install` dos veces sobre el mismo target sin modificaciones intermedias
- **THEN** la segunda ejecución reporta archivos idénticos como `skip` o equivalente y no produce conflictos falsos

#### Scenario: AGENTS.md existente
- **WHEN** el target ya contiene `AGENTS.md` sin referencia al harness
- **THEN** la CLI preserva el archivo, planifica únicamente la inserción de la referencia `@lufy-ia.harness.md` con backup y confirmación/`--yes`, y no inserta el contenido completo de Lufy

#### Scenario: AGENTS.md ausente
- **WHEN** el target no contiene `AGENTS.md`
- **THEN** install crea un archivo mínimo user-owned que referencia `@lufy-ia.harness.md` y no lo registra como asset completo gestionado por hash

#### Scenario: AGENTS.md con referencia existente
- **WHEN** el target contiene `AGENTS.md` con la referencia `@lufy-ia.harness.md`
- **THEN** install no duplica la referencia y no reescribe `AGENTS.md` solo por esa integración

#### Scenario: Archivo desconocido del usuario
- **WHEN** el target contiene archivos no gestionados por `lufy-ai`
- **THEN** la CLI no los borra ni modifica durante `install`

### Requirement: `lufy-ai verify` canónico
La CLI Go SHALL usar `lufy-ai verify` como verificador canónico de instalaciones y MUST NOT requerir ni introducir `scripts/verify-install.sh`; verify SHALL validate `lufy-ia.harness.md` as the managed agent-instructions asset and validate `AGENTS.md` as a user-owned reference integration.

#### Scenario: Verificación estructural de categorías críticas
- **WHEN** el usuario ejecuta `lufy-ai verify --target <dir> --no-engram` sobre un target instalado
- **THEN** la CLI valida que `.opencode/agents`, `.opencode/commands`, `.opencode/skills`, `.opencode/plugins` y `.opencode/policies` existen como directorios seguros no symlink

#### Scenario: Verificación de archivos críticos
- **WHEN** el usuario ejecuta `lufy-ai verify --target <dir> --no-engram` sobre un target instalado
- **THEN** la CLI valida que `.opencode/plugins/agent-observatory.tsx`, `lufy-ia.harness.md`, `tui.json`, `openspec/config.yaml` y `.lufy-ai/install-state.json` existen como archivos seguros no symlink, y valida `AGENTS.md` como archivo user-owned que referencia el harness cuando esté presente

#### Scenario: Archivos críticos presentes en manifest
- **WHEN** un archivo crítico gestionado existe en el target pero no está registrado en `.lufy-ai/install-state.json`
- **THEN** `lufy-ai verify` falla indicando que el asset clave no está en el manifest; esta regla aplica a `lufy-ia.harness.md` y no exige entrada de manifest para `AGENTS.md`

#### Scenario: Hashes de assets gestionados
- **WHEN** un asset listado en `.lufy-ai/install-state.json` existe pero su SHA-256 actual no coincide con `targetSHA256`
- **THEN** `lufy-ai verify` falla reportando drift con hashes abreviados expected/actual

#### Scenario: Verificación de opencode merge-managed
- **WHEN** el usuario ejecuta `lufy-ai verify --target <dir> --no-engram` sobre un target instalado
- **THEN** la CLI valida que `opencode.json` sea JSON parseable y contenga la estructura mínima merge-managed sin requerir entrada de hash completo en el manifest

#### Scenario: Verificación de referencia AGENTS
- **WHEN** el usuario ejecuta `lufy-ai verify --target <dir> --no-engram` sobre un target instalado
- **THEN** la CLI valida que `AGENTS.md` contenga `@lufy-ia.harness.md` como requisito o warning accionable sin comparar hash completo de `AGENTS.md`

#### Scenario: No existe script verificador paralelo
- **WHEN** se documenta o valida una instalación local/CI
- **THEN** la guía usa `lufy-ai verify` y no define `scripts/verify-install.sh` como objetivo ni dependencia

### Requirement: Comando sync de CLI Go
La CLI Go SHALL exponer `lufy-ai sync` como comando para sincronizar assets gestionados de forma segura en un target existente y aplicar merges seguros para assets `merge-json` explícitos, actualizando `lufy-ia.harness.md` como asset gestionado sin mutar `AGENTS.md` automáticamente.

#### Scenario: Help incluye sync
- **WHEN** el usuario solicita ayuda de la CLI o del comando `sync`
- **THEN** la salida describe `sync`, sus flags soportados y que opera sobre assets gestionados con manifest/hash/backup

#### Scenario: Sync delega fuera de main
- **WHEN** `cmd/lufy-ai/main.go` recibe el comando `sync`
- **THEN** delega la lógica de negocio a paquetes internos en vez de implementar planificación o copia completa dentro de `main.go`

#### Scenario: Wrapper Bash no cambia para sync
- **WHEN** se inspecciona `scripts/install.sh` después de añadir `sync`
- **THEN** permanece como wrapper estricto de `lufy-ai install` y no contiene lógica propia ni fallback legacy para sincronizar assets

#### Scenario: Sync aplica merge-json de opencode
- **WHEN** un target instalado tiene `opencode.json` válido que necesita claves merge-managed mínimas
- **THEN** `sync` planifica/aplica `merge-json` para `opencode.json`, preserva claves desconocidas y no usa `copy` ni `update-managed` por hash para ese archivo

#### Scenario: Sync aplica harness gestionado
- **WHEN** un target instalado tiene `lufy-ia.harness.md` registrado sin drift local y el source del harness cambió
- **THEN** `sync` planifica/aplica backup y `update-managed` para `lufy-ia.harness.md` y actualiza su entrada de manifest

#### Scenario: Sync no auto-repara AGENTS
- **WHEN** un target instalado tiene `AGENTS.md` sin la referencia `@lufy-ia.harness.md`
- **THEN** `sync` reporta warning o acción explícita requerida y MUST NOT modificar `AGENTS.md` silenciosamente
