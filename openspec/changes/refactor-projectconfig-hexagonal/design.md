## Contexto

`projectconfig` empezo como scanner stack-aware acotado, pero ahora tambien modela `project_profile.surfaces`, selector interactivo, rescan drift y preservacion de overrides. El archivo principal supera un limite razonable para mantenimiento porque junta reglas de dominio, aplicacion e infraestructura.

El refactor debe ser behavior-preserving: primero separar responsabilidades, despues habilitar nuevas capacidades con menor riesgo.

## Arquitectura propuesta

Mantener el paquete publico `internal/projectconfig` para minimizar churn de imports, pero separar internamente por responsabilidades:

- `model.go`: `ProjectConfig`, `Stack`, `ProjectProfile`, `ProjectSurface`, `AgentLens`, `WorkflowLimits`.
- `defaults.go`: defaults de TDD, workflow limits y lenses por superficie.
- `service.go`: caso de uso `Run`, `Ensure`, opciones y wiring.
- `store.go`: `Load`, `Marshal`, `ConfigStore` filesystem.
- `rescan.go`: `RescanPlan`, `DriftItem`, merge y preservacion de overrides.
- `scan.go`: coordinador de detectores.
- `detector_*.go`: strategies por stack/superficie.

Si la separacion por subpaquetes ayuda sin inflar imports, se puede usar:

- `internal/projectconfig/domain`
- `internal/projectconfig/application`
- `internal/projectconfig/adapters`

Pero el primer objetivo es SRP y testabilidad, no mover carpetas por estetica.

## Principios

- SRP: modelos, scan, merge, store y reporting cambian por razones distintas.
- OCP: agregar `astro`, `expo`, `dotnet`, `terraform` o nuevas surfaces no debe tocar un scanner gigante.
- DIP: service depende de interfaces chicas (`Scanner`, `Store`, `Clock`, `ProfilePrompt`).
- Clean Code: nombres concretos, funciones cortas, errores accionables en espanol.
- Tests AAA: nuevos tests deben tener setup/act/assert legible aunque no usen comentarios formales.

## Estrategia de migracion

1. Mover tipos/defaults sin cambiar tests.
2. Extraer rescan y merge, correr tests actuales.
3. Extraer detectores por familia, manteniendo outputs equivalentes.
4. Reducir `Service.Run` a orquestacion.
5. Validar paquete completo y `scripts/validate.sh`.

## Riesgos

- Refactor mecanico grande con bajo valor si se hace en una sola PR.
- Cambios involuntarios en YAML por orden/cero-values.
- Duplicacion temporal entre detectores durante la extraccion.

Mitigacion: slices pequenos, comparaciones de YAML/estructura y validacion agrupada al final de cada slice.
