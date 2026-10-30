# Configuración del proyecto

La configuración vive en `.lufy/config/project.yaml`, con `schema_version: 2`. El tool soportado es `codex`, con instalación project-local y ejecución `inline`. Los formatos anteriores se rechazan; no hay conversión automática.

## Mínimo y valores predeterminados

```yaml
schema_version: 2
tool: codex
execution:
  mode: inline
  delegation: auto
```

El producto usa `auto` si no se elige una política de delegación. Para trabajar siempre en el principal, elegí `off`; `on_request` requiere un pedido explícito. Las instrucciones del usuario y los permisos del host tienen precedencia. La instalación personal de los ensayos eligió `off`; eso no cambia el default.

`lufy-ai init --target /ruta/al/proyecto` genera configuración a partir del proyecto. `scan` vuelve a detectar stacks y superficies preservando overrides; revisá el diff después. `plan --json` resuelve esa información sin ejecutar los comandos declarados. Necesita superficies configuradas y, si no se pasan `--files`, una base Git válida.

## Campos que orientan el trabajo

| Campo | Uso |
| --- | --- |
| `project_profile.surfaces` | Superficies, raíces, conexiones, arquitectura y expectativas estructurales. Las restricciones explícitas son criterios de aceptación. |
| `stacks` | Toolchains y comandos detectados/declarados por stack. Deben existir en el proyecto real. |
| `validation` | Comprobaciones y comandos permitidos declarados. Su presencia no prueba ejecución ni amplía permisos. |
| `tdd` | Preferencias de pruebas cuando sean pertinentes. |
| `workflow_limits` | Límites de tamaño/revisión, preflight, reglas de parada y estrategia de entrega. Única fuente canónica de límites. |
| `parallel_execution` | Orientación para colaboración independiente y validación después de integrar. |

No consumir `loc_budget` ni `delivery_strategy` top-level como límites. `workflow_limits.routing.strategy: outcome` orienta por resultado; no selecciona una metodología por tier. Los valores generados son configuración visible, no un permiso para ampliar alcance o entregar.

## Capacidades locales

| Campo | Default del generador | Efecto |
| --- | --- | --- |
| `memory` | Obsidian, `.lufy/memory`, Git `ignored` | Memoria portable; requiere inicializar sus archivos. |
| `context_graph` | Habilitado, `.lufy/context`, cache y exclusiones sensibles | Índice reconstruible; se construye explícitamente o con `setup`. |
| `run_ledger` | Habilitado, `.lufy/runtime` | Permite productores explícitos. El hook instalado no registra ejecuciones automáticamente. |
| `adaptive_routing` | `enabled: false`, modo `shadow` | Experimento de asignación; no ejecuta agentes ni cambia la delegación. |
| `parallel_execution` | Habilitado, máximo 3 agentes, archivos independientes y plan de integración | Orientación subordinada a delegación y capacidades del host. |

Habilitar una capacidad en YAML no demuestra que esté operativa. `memory status`, `context status`, `skills status` y `doctor` observan su estado. El registry y el grafo son derivados; preservá sus fuentes al reconstruirlos.

## Configuración del host

`.codex/config.toml` solicita capacidades del host; `.codex/hooks.json` conecta sólo `SessionStart` con `lufy-ai lifecycle codex`. El hook entrega orientación compacta sobre skills, memoria y contexto sin guardar conversaciones ni producir eventos de review. Requiere el ejecutable en el PATH de Codex.

Las reglas distribuidas y la política del proyecto se complementan con las restricciones reales del host. Ninguna concede permisos que el entorno no exponga.

El hook puede refrescar el registry derivado de skills mediante `skills ensure`; no modifica notas de memoria ni reconstruye el grafo.
