## MODIFIED Requirements

### Requirement: Merge conservador de opencode.json
La CLI SHALL crear o mergear `opencode.json` mediante JSON válido, preservando claves desconocidas del usuario, y SHALL tratarlo como configuración `merge-json` especial en vez de asset completo gestionado por hash.

#### Scenario: Crear opencode.json faltante
- **WHEN** el target no contiene `opencode.json` y la instalación requiere configuración OpenCode
- **THEN** la CLI crea un JSON válido con las claves gestionadas mínimas por `lufy-ai`

#### Scenario: Preservar claves existentes
- **WHEN** el target contiene `opencode.json` válido con claves no gestionadas
- **THEN** la CLI preserva esas claves y modifica solo secciones gestionadas por `lufy-ai`

#### Scenario: JSON inválido
- **WHEN** el target contiene `opencode.json` inválido
- **THEN** la CLI falla sin sobrescribirlo y reporta una instrucción accionable para corregir o respaldar el archivo

#### Scenario: Opencode no se registra con hash completo
- **WHEN** `install` o `sync` escriben `opencode.json` mediante `merge-json`
- **THEN** `.lufy-ai/install-state.json` no contiene una entrada de asset completo para `opencode.json` ni requiere comparar su SHA-256 como asset gestionado

### Requirement: Comando sync de CLI Go
La CLI Go SHALL exponer `lufy-ai sync` como comando para sincronizar assets gestionados de forma segura en un target existente y aplicar merges seguros para assets `merge-json` explícitos.

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

### Requirement: `lufy-ai verify` canónico
La CLI Go SHALL usar `lufy-ai verify` como verificador canónico de instalaciones y MUST NOT requerir ni introducir `scripts/verify-install.sh`.

#### Scenario: Verificación estructural de categorías críticas
- **WHEN** el usuario ejecuta `lufy-ai verify --target <dir> --no-engram` sobre un target instalado
- **THEN** la CLI valida que `.opencode/agents`, `.opencode/commands`, `.opencode/skills`, `.opencode/plugins` y `.opencode/policies` existen como directorios seguros no symlink

#### Scenario: Verificación de archivos críticos
- **WHEN** el usuario ejecuta `lufy-ai verify --target <dir> --no-engram` sobre un target instalado
- **THEN** la CLI valida que `.opencode/plugins/agent-observatory.tsx`, `AGENTS.md`, `tui.json`, `openspec/config.yaml` y `.lufy-ai/install-state.json` existen como archivos seguros no symlink

#### Scenario: Archivos críticos presentes en manifest
- **WHEN** un archivo crítico gestionado existe en el target pero no está registrado en `.lufy-ai/install-state.json`
- **THEN** `lufy-ai verify` falla indicando que el asset clave no está en el manifest

#### Scenario: Hashes de assets gestionados
- **WHEN** un asset listado en `.lufy-ai/install-state.json` existe pero su SHA-256 actual no coincide con `targetSHA256`
- **THEN** `lufy-ai verify` falla reportando drift con hashes abreviados expected/actual

#### Scenario: Verificación de opencode merge-managed
- **WHEN** el usuario ejecuta `lufy-ai verify --target <dir> --no-engram` sobre un target instalado
- **THEN** la CLI valida que `opencode.json` sea JSON parseable y contenga la estructura mínima merge-managed sin requerir entrada de hash completo en el manifest

#### Scenario: No existe script verificador paralelo
- **WHEN** se documenta o valida una instalación local/CI
- **THEN** la guía usa `lufy-ai verify` y no define `scripts/verify-install.sh` como objetivo ni dependencia
