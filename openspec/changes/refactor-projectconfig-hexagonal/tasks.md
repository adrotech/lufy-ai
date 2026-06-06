## 1. Preparacion

- [ ] Confirmar baseline con `go test ./internal/projectconfig ./internal/cli`.
- [ ] Registrar snapshot estructural del YAML generado para Go, TS/Next, backend+frontend e unsupported stack.
- [ ] Identificar imports externos actuales que deben quedar en adapters/store.

## 2. Modelos y defaults

- [ ] Extraer modelos YAML a archivo dedicado sin cambiar tags ni nombres.
- [ ] Extraer defaults de TDD, workflow limits y `AgentLens`.
- [ ] Mantener tests existentes verdes.

## 3. Rescan y merge

- [ ] Extraer `RescanMerger`, `RescanPlan`, `DriftItem` y merge helpers.
- [ ] Cubrir preservacion de `workflow_limits`, stacks, `project_profile.surfaces` y extras desconocidos.
- [ ] Mantener mensajes de drift existentes o documentar cualquier ajuste.

## 4. Detectores

- [ ] Extraer detector Go.
- [ ] Extraer detector JavaScript/TypeScript.
- [ ] Extraer detector Python.
- [ ] Extraer detector JVM.
- [ ] Extraer detectores unsupported/infra.
- [ ] Extraer detector de project surfaces.
- [ ] Agregar registry/coordinador liviano para detectores.

## 5. Service y store

- [ ] Reducir `Service.Run` a orquestacion de scan/merge/prompt/write.
- [ ] Aislar `ConfigStore` filesystem y YAML marshal/load.
- [ ] Mantener `ProfilePrompt` como puerto de aplicacion.

## 6. Validacion final

- [ ] Ejecutar `go test ./internal/projectconfig ./internal/cli`.
- [ ] Ejecutar `scripts/validate.sh`.
- [ ] Ejecutar `openspec validate "refactor-projectconfig-hexagonal" --strict`.
- [ ] Revisar diff final y confirmar que no cambia el YAML publico salvo ordenamientos inevitables justificados.
