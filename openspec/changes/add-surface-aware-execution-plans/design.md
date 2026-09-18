## Contexto

`ProjectSurface` contiene roots, stacks, conexiones, arquitectura y `AgentLens`, pero no existe un caso de uso que resuelva cuál de esas superficies aplica a un cambio. Los agentes reciben instrucciones para leer el YAML y toman la decisión en texto libre.

El nuevo motor debe mantenerse read-only, determinístico y separado de Git/YAML para que otras herramientas puedan aportar las mismas entradas en el futuro.

## Arquitectura

El paquete `internal/surfaceplan` se divide por responsabilidad sin crear subpaquetes ceremoniales:

- `model.go`: valores de dominio del plan, decisiones, contratos y reglas de validación.
- `ports.go`: `ProjectSource` y `ChangeSource`.
- `resolver.go`: `SurfaceResolverStrategy`, pura y determinística.
- `validation_factory.go`: `ValidationPlanFactory`, pura y data-driven.
- `adapters.go`: adapters para `projectconfig` y `git diff --name-only`.
- `service.go`: caso de uso y composición por constructor.

La CLI sólo parsea flags y representa el resultado. No resuelve roots, contratos ni validaciones.

## Resolución

1. `--surface` distinto de `auto` gana y acepta ID exacto o tipo único.
2. `--files` gana sobre Git como fuente de archivos.
3. Sin `--files`, `git diff --name-only <base>` entrega el alcance.
4. Para cada archivo se elige la root no compuesta más específica.
5. Un empate entre superficies conectadas se eleva a su composición fullstack.
6. Si no hay archivos, una única superficie hoja se elige; varias hojas conectadas forman composición; lo demás devuelve ambigüedad accionable.

Las decisiones quedan en `decisions[]` con código, razón y evidencia.

## Validación

La Factory combina:

- `agent_lens.validation_expectations` de las superficies activas;
- comandos conocidos de los stacks vinculados;
- señales del diff: UI, contrato y persistencia;
- capacidades declaradas o pasadas por CLI.

Cada `ValidationRule` declara ID, categoría, trigger, obligatoriedad, superficies, comandos sugeridos, evidencia y razón. Los comandos son evidencia de configuración, no autorización ni ejecución.

## Compatibilidad

`ProjectSurface.capabilities` es opcional y mantiene `schema_version: 1`: YAML anterior continúa válido y los campos extra siguen preservándose. `rescan` conserva capacidades del usuario. La salida JSON del comando comienza en `surface-execution-plan/v1`.

## Seguridad

- El motor nunca ejecuta comandos de validación.
- El adapter Git usa argumentos directos, no shell.
- Paths absolutos o que escapen del target se rechazan como entrada de archivos.
- La salida explica cuándo usa una composición conservadora.

