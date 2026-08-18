# Tasks: Completar Lufy SDD Full

## 1. Contrato y bootstrap gestionado

- [x] Definir `config.yaml`, README, actions y templates canónicos bajo `.lufy/sdd/` para su instalación en `.lufy/workflows/sdd/`.
- [x] Marcar correctamente ownership y mode full/lite en catálogo, adapters y copias embebidas.
- [x] Cubrir paridad source/embedded y ausencia de assets OpenSpec en selección exclusivamente Lufy SDD.

## 2. Dominio y parser Lufy SDD

- [x] Crear paquete interno `lufysdd` con modelos de change, diagnostics, report schema y resolución segura de paths.
- [x] Implementar parser de delta requirements/scenarios con líneas y orden determinístico.
- [x] Implementar validación normal/strict, tasks y digest estable.
- [x] Agregar unit tests positivos, negativos, symlink y determinismo.

## 3. Lifecycle de creación e inspección

- [x] Implementar `new`, scaffolding exclusivo y rechazo de IDs/paths inseguros.
- [x] Implementar `status` derivado y salida humana/JSON.
- [x] Implementar `validate` sin mutar fuentes/specs, con refresh del overview derivado y códigos de salida CLI coherentes.
- [x] Integrar `lufy-ai sdd` en help, command palette y tests de CLI.

## 4. Sync seguro de specs

- [x] Implementar plan ADDED/MODIFIED/REMOVED con preflight global y rechazo de ambigüedad.
- [x] Implementar apply con temporales, backup y rollback acotado.
- [x] Persistir digest de sync solo después de éxito completo.
- [x] Cubrir create/modify/remove, delta cambiado, fallo intermedio y symlinks.

## 5. Archive con gates

- [x] Exigir strict validation, tasks completas, digest sincronizado y evidencia regular.
- [x] Implementar move exclusivo hacia archive fechado sin overwrite ni follow de symlinks.
- [x] Cubrir gates incompletos, colisión de destino y archive exitoso.

## 6. Acciones y skills instalables

- [x] Crear acciones tool-neutral explore/propose/apply/verify/sync/archive.
- [x] Crear wrappers OpenCode condicionados a metodología Lufy SDD.
- [x] Crear wrappers Codex condicionados a metodología Lufy SDD.
- [x] Verificar que todos reportan metodología, estado de gates y siguiente acción consistentes.

## 7. Adapter, verify y documentación

- [x] Reemplazar el check informativo de `VerifyWorkflow` por validación estructural/semántica accionable.
- [x] Documentar uso diario, coexistencia y migración manual en README y docs.
- [x] Actualizar status, roadmap y backlog sin declarar disponibilidad antes de validación/delivery.

## 8. Overview HTML y mode Lite

- [x] Implementar renderer HTML autocontenido, determinístico, atómico y compartido por Full/Lite.
- [x] Materializar y refrescar `change-overview.html` automáticamente en new, validate, sync y archive, manteniendo status read-only.
- [x] Hacer scaffolding, validación, estado, sync y archive conscientes de `mode: full|lite`.
- [x] Cubrir Full/Lite, refresh tras editar Markdown, determinismo, ausencia de recursos externos y paths no regulares.
- [x] Actualizar actions, wrappers, bootstrap y documentación sin agregar un comando o skill de render dedicado.

## 9. Integración completa con harness y agentes

- [x] Hacer `sdd-router` methodology-aware y eliminar defaults que fuerzan OpenSpec para T1/T2.
- [x] Integrar lifecycle y overview automático Lufy SDD en orchestrator y Result Contract.
- [x] Alinear implementer, validator, reviewer, delivery, `AGENTS.md` y harness portable con Full/Lite nativo.
- [x] Permitir artifacts user-owned Lufy SDD en PR guard sin habilitar metadata `.lufy/` general.
- [x] Sincronizar copias embebidas y cubrir routing/ownership con checks estáticos y tests.

## 10. Validación agrupada

> **Nota de entrega — 2026-08-18:** la implementación y los checks estáticos disponibles están completos (36/40). Quedan pendientes `gofmt`, tests/coverage/build Go, smokes Full/Lite y validación OpenSpec strict/all porque este entorno no dispone de los binarios `go` ni `openspec`. Ejecutar estos cuatro gates en CI o en un entorno con el toolchain instalado antes de mergear o promocionar la funcionalidad.

- [ ] Ejecutar `gofmt` sobre archivos Go modificados.
- [ ] Ejecutar `go test ./...`, coverage y build desde `tools/lufy-cli-go`.
- [ ] Ejecutar smokes install/sync/verify con `lufy-sdd/full` y `lufy-sdd/lite`.
- [ ] Ejecutar `openspec validate complete-lufy-sdd-full-workflow --strict` y `openspec validate --all`.
- [x] Ejecutar scripts de paridad, shell, release docs y `git diff --check`.
