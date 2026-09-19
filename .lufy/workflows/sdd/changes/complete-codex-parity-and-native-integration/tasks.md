# Tasks: complete-codex-parity-and-native-integration

## Review Slices

| Slice | Objective | Expected Surface | Validation | Risk | PR Guidance |
| --- | --- | --- | --- | --- | --- |
| A | Eliminar divergencia config/state | projectconfig, installer, CLI, verify | unit + install E2E | critical | PR independiente recomendado |
| B | Activar lifecycle y policy Codex | `.codex`, tool adapter, verify | hook/rules tests + Codex probe | high | despues de A |
| C | Hacer contracts Codex autocontenidos | assets, skills, policies, renderer | catalog + smoke Codex-only | high | puede separarse de B si no comparte renderer |
| D | Cerrar E2E y documentacion | integration tests, docs, embedded | validate agrupado + matrix | medium | join final antes de release |

- [x] 1. Slice A - Canonical harness integrity
  - [x] 1.1 Agregar tests RED donde `install --tool codex --methodology-tier ...` exige la misma seleccion en `project.yaml` e install state.
  - [x] 1.2 Definir `HarnessConfig` efectivo con provenance y precedencia flags > project config > install state > defaults.
  - [x] 1.3 Implementar merge atomico de tool/metodologias en `project.yaml` preservando stacks, surfaces, workflow limits, memoria, contexto y unknown fields.
  - [x] 1.4 Reutilizar la resolucion efectiva en install, sync, skills, setup, info y comandos que hoy dependen de defaults implicitos.
  - [x] 1.5 Agregar reconciliacion config/state a `verify` y `doctor`, con error y recovery no destructivo ante mismatch.
  - [x] 1.6 Cubrir fresh install, config existente, install existente, flag explicito, mismatch y rollback/error parcial.
  - [x] 1.7 Ejecutar tests focalizados y registrar evidencia GREEN del slice.

- [x] 2. Slice B - Native Codex lifecycle and policy
  - [x] 2.1 Sustituir `hooks.json` vacio por lifecycle minimo SessionStart, SubagentStop, Stop y SessionEnd usando scripts/subcomandos portables.
  - [x] 2.2 Asegurar skill registry y orientar context/memory best-effort sin exponer contenido privado.
  - [x] 2.3 Implementar diagnostico de payload de subagentes y status de workflow sin mutar runtime ni marcar gates automaticamente.
  - [x] 2.4 Crear rules Codex conservadoras para Git/GH mutante, force push y comandos destructivos, con `match`/`not_match`.
  - [x] 2.5 Actualizar config Codex a keys soportadas y declarar hooks de forma explicita; eliminar `max_depth` si no existe evidencia runtime.
  - [x] 2.6 Hacer `verify --deep` y `doctor` adapter-aware, sin warnings/recovery OpenCode para Codex.
  - [x] 2.7 Probar hooks/rules estructuralmente y con `codex features list` / `codex execpolicy check` cuando esten disponibles.

- [x] 3. Slice C - Self-contained contracts and skill parity
  - [x] 3.1 Inventariar invariantes de roles, delivery, Result Contract, OpenSpec y Lufy SDD que deben ser tool-neutral.
  - [x] 3.2 Extraer fuente neutral o renderer compartido con overlays Codex/OpenCode, sin duplicar contratos completos manualmente.
  - [x] 3.3 Eliminar referencias obligatorias de skills Codex a `.opencode/policies`, templates o scripts no instalados.
  - [x] 3.4 Incluir templates/references necesarios en `.agents/skills` y su catalogo efectivo.
  - [x] 3.5 Mantener native/emulated/inline, aislamiento de permisos y recuperacion de resultados vacios en custom agents.
  - [x] 3.6 Agregar metadata Codex opcional para skills solo cuando mejore discovery sin volverla requisito runtime.
  - [x] 3.7 Probar una instalacion Codex-only eliminando cualquier dependencia accidental de `.opencode`.

- [x] 4. Slice D - Runtime E2E, documentation and release readiness
  - [x] 4.1 Construir matriz temporal: Codex+OpenSpec, Codex+Lufy SDD Full/Lite y T3 none.
  - [x] 4.2 Verificar install, idempotencia, sync dry-run/real, skills, doctor, verify, conflicts y restore por escenario aplicable.
  - [x] 4.3 Verificar discovery de ocho roles y skills efectivos con runtime Codex cuando este disponible; degradar explicitamente en CI sin Codex.
  - [x] 4.4 Probar Lufy SDD new/validate/sync/archive y overview automatico desde un target Codex limpio.
  - [x] 4.5 Actualizar README, installation, architecture, status, roadmap y troubleshooting con capacidades actuales y limites reales.
  - [x] 4.6 Sincronizar assets fuente/embedded y asegurar tests de catalogo por adapter/metodologia.
  - [x] 4.7 Ejecutar validacion agrupada completa, coverage y builds disponibles.

- [x] 5. Full SDD gates and delivery readiness
  - [x] 5.1 Ejecutar `lufy-ai sdd validate --change complete-codex-parity-and-native-integration --strict` tras cada join relevante y al cierre.
  - [x] 5.2 Registrar evidencia por slice bajo `.lufy/workflows/sdd/verification/complete-codex-parity-and-native-integration/`.
  - [x] 5.3 Verificar scenarios contra codigo/tests y revisar `change-overview.html` derivado actualizado.
  - [x] 5.4 Ejecutar sync de specs Full despues de validacion y resolver ambiguedades antes de archive.
  - [x] 5.5 Mantener delivery como `delivery_pending` hasta autorizacion explicita, checks remotos exitosos y trazabilidad completa.
