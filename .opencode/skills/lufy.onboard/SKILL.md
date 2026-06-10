---
name: lufy.onboard
description: Guía onboarding de LUFY para usuarios nuevos; valida instalación, lee .lufy/config/project.yaml y demuestra un cambio T3 dummy stack-aware con --demo y --dry-run.
license: MIT
compatibility: OpenCode skill autocontenido; usa solo lectura local de .lufy/config/project.yaml y no ejecuta delivery.
metadata:
  author: lufy-ai
  version: "1.0"
---

# Skill: lufy.onboard

Onboarding guiado para usuarios nuevos de LUFY. Explica el flujo SDD proporcional, valida que la instalación tenga los artefactos mínimos y genera una demo T3 dummy adaptada al stack declarado en `.lufy/config/project.yaml` en menos de 10 minutos.

El comando slash `.opencode/commands/lufy.onboard.md` es una fachada discoverable y debe delegar en este skill.

## Entradas soportadas

- `--demo`: construir una guía de cambio T3 dummy stack-aware para el repositorio actual.
- `--dry-run`: no proponer ni aplicar mutaciones; solo explicar qué se haría, qué archivos tocaría y qué validación usaría.
- Sin flags: mostrar onboarding breve, estado de instalación y próximos pasos seguros.

## Reglas de seguridad

1. No hacer commit, push, PR, delivery ni sincronización de Projects.
2. No modificar archivos durante `--dry-run`.
3. En modo `--demo` sin `--dry-run`, si el usuario quiere crear artefactos reales, pedir confirmación explícita y derivar a `implementer` con alcance T3; no ejecutar delivery.
4. No inventar toolchain. Nombrar comandos únicamente si existen en `.lufy/config/project.yaml` y no son `TODO`.
5. Si falta `.lufy/config/project.yaml`, detener la demo stack-aware y recomendar:

   ```bash
   lufy-ai init [--target <path>]
   ```

## Validación de instalación

Antes de la demo, revisar de forma estática/read-only:

- Existe `AGENTS.md`.
- Existe `.lufy/config/project.yaml` para demo stack-aware.
- Existen `.opencode/agents/`, `.opencode/commands/` y `.opencode/skills/` cuando el usuario quiere validar instalación local.
- Recordar que el usuario debe reiniciar OpenCode si el skill/comando fue instalado durante la sesión actual.

Si alguno falta, reportar estado, impacto y acción recomendada. La ausencia de `.lufy/config/project.yaml` bloquea solo la demo stack-aware confiable.

## Lectura de `.lufy/config/project.yaml`

Usar `stacks[]` como fuente canónica. Para cada stack considerar:

- `id`
- `supported`
- `frameworks`
- `test_runner.command`
- `test_runner.coverage_command`
- `formatter.command`
- `linter.command`
- `static_analysis.command`
- `notes`

Comandos válidos: valores no vacíos y que no empiecen con `TODO`. Si `supported: false`, si el stack no está en la lista soportada o si todos los comandos relevantes son `TODO`/vacíos, degradar a explicación estática y pedir completar `.lufy/config/project.yaml`.

Stacks soportados para demo guiada:

- `go`
- `typescript`
- `javascript`
- `python`
- `java`
- `kotlin`

## Selección de demo T3 dummy

Elegir un único stack soportado. Si hay varios, priorizar el primer stack soportado con mayor cantidad de comandos válidos; si empatan, usar el primero del archivo y mencionar que el usuario puede pedir otro stack.

La demo debe explicar:

1. **Objetivo T3**: cambio local, reversible, de bajo riesgo.
2. **Archivos candidatos**: 1-2 archivos como máximo.
3. **Pasos**: inspección mínima, edición propuesta, validación agrupada.
4. **Conceptos LUFY**: T3 Express, Result Contract, validación proporcional y delivery separado.
5. **Comandos detectados**: test, formatter, linter y static analysis solo si vienen de `project.yaml` y no son `TODO`.

Plantillas de dummy por stack:

- `go`: agregar o explicar una prueba pequeña para una función pura existente, o un comentario/doc mínimo asociado a un paquete. Validar con `test_runner.command`; si existe, sumar `static_analysis.command`; formatter con `formatter.command` solo si se requiere edición `.go`.
- `typescript`/`javascript`: agregar o explicar una prueba mínima o ajuste de texto en una unidad existente. Validar con `test_runner.command`; sumar `static_analysis.command` para TypeScript si existe; usar `linter.command`/`formatter.command` solo si están configurados.
- `python`: agregar o explicar una prueba pequeña de función pura o ejemplo documental cercano al módulo. Validar con `test_runner.command`; sumar `static_analysis.command`, `linter.command` y `formatter.command` solo si existen.
- `java`/`kotlin`: agregar o explicar una prueba mínima en el paquete correspondiente o ajuste documental del módulo. Validar con `test_runner.command`; sumar `static_analysis.command`, `linter.command` y `formatter.command` solo si existen.

Si no hay un archivo objetivo obvio, mantener la demo en modo explicación y proponer que `explorer` identifique un candidato antes de editar.

## Comportamiento por modo

### `/lufy.onboard --demo --dry-run`

Responder sin mutaciones con esta estructura:

1. Estado de instalación.
2. Stack detectado y fuente: `.lufy/config/project.yaml`.
3. Demo T3 dummy propuesta.
4. Comandos de validación detectados, agrupados y marcados como `no ejecutados`.
5. Qué haría `implementer` si el usuario confirma.
6. Result Contract compacto con `status: ready` o `blocked` si falta configuración.

### `/lufy.onboard --demo`

Responder con guía accionable. Si el usuario pide aplicar el dummy:

- Pedir confirmación explícita del archivo/cambio propuesto.
- Derivar a `implementer` con tier T3 y validación agrupada.
- Mantener delivery como paso separado y no autorizado.

### Sin `.lufy/config/project.yaml`

Responder:

- No es posible una demo stack-aware confiable.
- Ejecuta `lufy-ai init [--target <path>]` y revisa/completa `stacks[]`.
- Reintenta `/lufy.onboard --demo --dry-run`.

### Stack unsupported o incompleto

Responder con degradación estática:

- Identificar `id`, `supported` y campos `TODO`/vacíos.
- Explicar conceptos LUFY sin sugerir comandos inventados.
- Pedir completar `.lufy/config/project.yaml` o ejecutar `lufy-ai init --rescan [--target <path>]` cuando aplique.

## Ejemplo de salida esperada

```yaml
estado: ready
instalacion:
  project_yaml: presente
stack:
  id: go
  fuente: .lufy/config/project.yaml
demo_t3:
  objetivo: "Agregar una prueba mínima para una función pura existente"
  mutaciones: "ninguna en --dry-run"
validacion_detectada:
  test: "go test ./..."
  static_analysis: "go vet ./..."
  formatter: "gofmt -w"
siguiente_paso: "Si confirmas, implementer puede aplicar un T3 acotado; delivery queda separado."
```
