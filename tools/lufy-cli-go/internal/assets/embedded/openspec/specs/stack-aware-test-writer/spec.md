# stack-aware-test-writer Specification

## Purpose
TBD - created by archiving change add-stack-aware-test-writer. Update Purpose after archive.
## Requirements
### Requirement: Stack-aware test writer agent
The harness SHALL provide a `test-writer` OpenCode subagent that writes and revises tests using the target repository's `.lufy/config/project.yaml` stack configuration when available.

#### Scenario: Agent reads stack-specific test configuration
- **GIVEN** a target repository contains `.lufy/config/project.yaml` with a supported stack entry and test command metadata
- **WHEN** `test-writer` is assigned a T1 or T2 change requiring substantive tests
- **THEN** it uses the configured stack test command, coverage threshold and anti-pattern guidance for that stack instead of assuming a fixed language toolchain

#### Scenario: Agent reports missing stack config
- **GIVEN** a target repository does not contain `.lufy/config/project.yaml` or the relevant stack lacks test command metadata
- **WHEN** `test-writer` is assigned a change requiring tests
- **THEN** it reports the missing configuration as `not_available` or `blocked` evidence and recommends `lufy-ai init` or a manual project config update without inventing test commands

### Requirement: Observable TDD phase evidence
`test-writer` SHALL report RED, GREEN, TRIANGULATE and REFACTOR phase evidence when a TDD cycle applies to the assigned change.

#### Scenario: Full TDD cycle is recorded
- **WHEN** `test-writer` completes a TDD-applicable T1 or T2 task
- **THEN** its result contract includes the test files changed, phase-by-phase evidence, commands run and whether each phase passed, failed, was blocked or was not applicable

#### Scenario: TDD phase is not applicable
- **WHEN** a requested test change does not require one of the TDD phases
- **THEN** `test-writer` marks that phase as `not_applicable` with a concise reason instead of omitting the phase silently

### Requirement: Stack anti-pattern enforcement
`test-writer` SHALL use stack-specific anti-pattern guidance from `.lufy/config/project.yaml` when proposing or modifying tests.

#### Scenario: Anti-pattern guidance exists
- **GIVEN** `.lufy/config/project.yaml` defines anti-patterns for the relevant stack
- **WHEN** `test-writer` creates or edits tests
- **THEN** it checks the proposed test approach against those anti-patterns and reports any avoided or unresolved anti-patterns in its result contract

#### Scenario: Anti-pattern guidance is absent
- **GIVEN** `.lufy/config/project.yaml` does not define anti-patterns for the relevant stack
- **WHEN** `test-writer` creates or edits tests
- **THEN** it reports anti-pattern guidance as `not_available` and does not invent project-specific rules

### Requirement: Bounded permissions and delivery separation
`test-writer` SHALL be limited to test-focused implementation and validation evidence, and SHALL NOT perform Git/GH delivery operations.

#### Scenario: Delivery is requested from test writer
- **WHEN** `test-writer` completes test work and delivery is still required
- **THEN** it reports the next owner as `delivery` or `implementer` as appropriate and MUST NOT commit, push, create PRs or update GitHub Projects

### Requirement: AAA structure for substantive tests
Los tests nuevos o modificados para cambios T1/T2 sustantivos SHALL mantener estructura Arrange/Act/Assert observable, explicita o idiomatica, para que reviewer y validator puedan auditar intencion, accion y resultado sin reconstruir el flujo completo.

#### Scenario: New behavior test is structured
- **WHEN** un cambio agrega un test para comportamiento nuevo o modificado
- **THEN** el test SHALL separar claramente setup, accion y assertions mediante orden, helpers, nombres o comentarios
- **AND** SHALL evitar multiples actos no relacionados dentro del mismo test salvo que sea una prueba de flujo end-to-end deliberada

#### Scenario: Integration test documents multiple phases
- **WHEN** un test de integracion necesita varias acciones para cubrir install, sync, verify, backup, restore o rollback
- **THEN** el test SHALL hacer observable la razon de cada fase mediante nombres, helpers o bloques
- **AND** SHALL preservar assertions de no-mutacion cuando el comportamiento prometa dry-run, preservacion de usuario o rollback

### Requirement: TDD evidence for architecture refactor slices
Los slices de refactor que cambien comportamiento observable SHALL reportar evidencia TDD proporcional, o marcar fases como `not_applicable` con razon concreta.

#### Scenario: Behavior-changing slice
- **WHEN** un slice cambia planificacion, ejecucion, verificacion, backup, sync, status o salida CLI
- **THEN** el resultado SHALL reportar RED, GREEN, TRIANGULATE y REFACTOR como `passed`, `failed`, `blocked` o `not_applicable`
- **AND** SHALL incluir comandos reales ejecutados

#### Scenario: Pure structural slice
- **WHEN** un slice solo mueve codigo sin cambiar comportamiento observable
- **THEN** puede marcar RED o TRIANGULATE como `not_applicable`
- **AND** SHALL compensar con tests existentes, tests de caracterizacion o validacion agrupada que demuestre no-regresion
