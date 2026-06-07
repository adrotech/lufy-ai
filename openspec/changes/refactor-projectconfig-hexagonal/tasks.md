## 1. Preparacion

- [x] Confirmar baseline con `go test ./internal/projectconfig ./internal/cli`.
- [x] Registrar snapshot estructural del YAML generado para Go, TS/Next, backend+frontend e unsupported stack.
- [x] Identificar imports externos actuales que deben quedar en adapters/store.

## 2. Modelos y defaults

- [x] Extraer modelos YAML a archivo dedicado sin cambiar tags ni nombres.
- [x] Extraer defaults de TDD, workflow limits y `AgentLens`.
- [x] Mantener tests existentes verdes.

## 3. Rescan y merge

- [x] Extraer `RescanMerger`, `RescanPlan`, `DriftItem` y merge helpers.
- [x] Cubrir preservacion de `workflow_limits`, stacks, `project_profile.surfaces` y extras desconocidos.
- [x] Mantener mensajes de drift existentes o documentar cualquier ajuste.

## 4. Detectores

- [x] Extraer detector Go.
- [x] Extraer detector JavaScript/TypeScript.
- [x] Extraer detector Python.
- [x] Extraer detector JVM.
- [x] Extraer detectores unsupported/infra.
- [x] Extraer detector de project surfaces.
- [x] Agregar registry/coordinador liviano para detectores.

## 5. Service y store

- [x] Reducir `Service.Run` a orquestacion de scan/merge/prompt/write.
- [x] Aislar `ConfigStore` filesystem y YAML marshal/load.
- [x] Mantener `ProfilePrompt` como puerto de aplicacion.

## 6. Validacion final

- [x] Ejecutar `go test ./internal/projectconfig ./internal/cli`.
- [x] Ejecutar `scripts/validate.sh`.
- [x] Ejecutar `openspec validate "refactor-projectconfig-hexagonal" --strict`.
- [x] Revisar diff final y confirmar que no cambia el YAML publico salvo ordenamientos inevitables justificados.
