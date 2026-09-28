# Verificación Slice D — harness, seguridad y documentación

## Estado

`validated` para el alcance del Slice D. La implementación y documentación del programa quedan completas; verify final, sync, delivery y cierre continúan pendientes.

## Resultado funcional

- `.lufy/contracts/adaptive-routing.md` define la frontera neutral: `role_hint` es una capacidad temporal sugerida, nunca identidad, permiso, ownership ni autoridad de gate.
- Orchestrator y router OpenCode/Codex consumen recomendaciones como evidencia de planificación solamente; mantienen roles existentes, `gate_advanced=false` y el routing SDD determinista como fallback.
- `disabled`, `shadow` y `advisory` conservan la misma semántica en contrato, `AGENTS.md.template`, harness, adapters y documentación. No existe modo autónomo.
- Delivery, seguridad, contratos públicos, database schema y destructive migrations preemptan scoring y requieren humano/orchestrator.
- Yield se documenta como liberación segura: checkpoint content-free durable antes de liberar lease/budget; conflictos o storage unavailable mantienen la assignment activa.
- Catálogo raíz y bundle embebido incluyen el nuevo contrato y preservan paridad exacta de contenido administrado.

## Evidencia

| Comando | Resultado | Evidencia |
| --- | --- | --- |
| `go test ./internal/assets ./internal/instructions/... ./internal/adapters/tool/opencode ./internal/adapters/tool/codex` | passed | catálogo raíz/embebido, role registry/render y paridad adapters |
| `./scripts/check-harness-coupling.sh` | passed | superficies neutrales sin acoplamiento indebido y contratos existentes completos |
| `go test ./internal/adaptive/... ./internal/projectconfig ./internal/runledger ./internal/cli ./internal/tui/commandpalette` | passed | regresión focalizada runtime/CLI/config/ledger |
| `go test ./internal/instructions/registry ./internal/instructions/render ./internal/adapters/tool/opencode` | passed | YAML de roles y golden renderer actualizados |
| `rg -n --glob '!**/*_test.go' '(raw_prompt|raw_output|raw_path|prompt_text|response_text|summary_text|secret_value|hypothesis_text|failed_attempt_text|super-secret)' internal/adaptive internal/runledger internal/cli/app_adaptive.go` | passed, sin matches | auditoría recursiva de canaries sensibles fuera de tests |

## Riesgos pendientes

- Falta ejecutar la validación final agrupada completa, build, race, `scripts/validate.sh` y strict SDD validate.
- Sync a specs activas, issue `#222`, commit/push/PR y checks remotos requieren sus gates posteriores; ninguna señal adaptativa los autoriza.
