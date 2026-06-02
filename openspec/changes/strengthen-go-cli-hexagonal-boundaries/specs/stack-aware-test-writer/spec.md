## ADDED Requirements

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
