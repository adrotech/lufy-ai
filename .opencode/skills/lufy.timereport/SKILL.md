---
name: lufy.timereport
description: Genera un reporte HTML offline/autocontenido de tiempo, actividad y ROI desde OpenCode SQLite, Git y metadata local del proyecto, con privacidad por defecto.
license: MIT
compatibility: OpenCode skill autocontenido; usa Python 3 estándar y comandos Git read-only cuando estén disponibles.
metadata:
  author: lufy-ai
  version: "1.0"
---

# Skill: lufy.timereport

Genera un Developer Impact Report local para la tarea actual del usuario dentro del repositorio. El reporte es un único HTML autocontenido, sin recursos remotos, diseñado para revisión offline y para explicar cómo la IA ayudó en el trabajo diario con métricas estructurales sanitizadas.

## Uso recomendado

Desde la raíz del repositorio:

```bash
python3 .opencode/skills/lufy.timereport/scripts/generate_timereport.py
```

Opciones soportadas:

- `--output <ruta>`: ruta del HTML. Default: `/tmp/lufy-timereport-<timestamp>.html`.
- `--target-dir <ruta>`: repositorio/directorio objetivo. Default: directorio actual.
- `--db <ruta>`: SQLite OpenCode. Default: `~/.local/share/opencode/opencode.db`.
- `--from <fecha>` / `--to <fecha>`: rango temporal opcional para Git y actividad con timestamps.
- `--scope task|repo`: alcance. Default `task`, que toma la sesión raíz de la tarea solicitada y sus subagentes. `repo` genera el reporte global del repositorio.
- `--session-id <id>`: ancla explícita. Si el id pertenece a un subagente, el reporte sube a la sesión raíz e incluye todo el árbol.
- `--tier T1|T2|T3`: tier original de la tarea cuando el agente lo conoce.
- `--change <id>`: spec/change OpenSpec o LUFY SDD asociado cuando existe.

El comando slash `.opencode/commands/lufy.timereport.md` es solo una fachada discoverable y debe delegar en este skill.

## Fuentes de datos y degradación

### OpenCode SQLite primario

- Ruta por defecto: `~/.local/share/opencode/opencode.db`.
- Apertura read-only: URI SQLite `file:<path>?mode=ro`.
- Tablas esperadas e introspeccionadas defensivamente: `project`, `workspace`, `session`, `message`, `part` y `event`.
- Filtrado por repositorio: incluir sesiones cuyo `directory`, `path`, `cwd`, `root`, `project_path`, `workspace_path` o relación con `project`/`workspace` coincida con `--target-dir` cuando esas columnas existan. Si el schema no expone directorio, degradar con limitación visible en vez de mezclar datos silenciosamente.
- Schema mismatch: si falta una tabla/columna requerida para una métrica, marcar solo esa sección como `No disponible` o `Estimación parcial` e incluir la fuente faltante en Limitaciones.

### Alcance de tarea

- El default es `--scope task`, no el histórico completo.
- La tarea se define como sesión raíz OpenCode más reciente para `--target-dir` y todos sus descendientes por `parent_id`.
- Si se pasa `--session-id`, usar esa sesión como ancla: subir por `parent_id` hasta la raíz disponible y luego incluir descendientes.
- El reporte debe mostrar el contrato de alcance: scope, tier, sesión raíz, sesión ancla, cantidad de sesiones incluidas, ventana temporal, metodología y spec/change cuando exista.
- Si se requiere el comportamiento anterior de todo el repositorio, usar `--scope repo` explícitamente.
- Cuando `--scope task` no recibe `--from/--to`, Git se limita a la ventana temporal inferida de la tarea para no mezclar commits ajenos.

### Git secundario

- Operaciones permitidas: `git rev-parse --is-inside-work-tree` y `git log --numstat` con rango temporal cuando aplique.
- Métricas: conteo de commits y LOC neto (`líneas agregadas - líneas eliminadas`).
- Si el target no es repositorio Git o Git no está disponible, continuar con secciones no-Git y marcar commits/LOC como `No disponible`.

### Stack opcional

- Preferir `.lufy/config/project.yaml` si existe y contiene metadata simple de stack/tooling.
- Si falta o no puede interpretarse, usar heurística por archivos conocidos (`go.mod`, `package.json`, `pyproject.toml`, `Cargo.toml`, etc.).
- Si tampoco hay señales, reportar `No configurado`.

### Exclusiones explícitas

- No leer JSONL por defecto.
- No leer ni emitir `session_diff`.
- No emitir prompts, assistant outputs, argumentos completos, outputs de tools, contenidos de archivos ni diffs.

## Métricas y heurísticas determinísticas

- **Wall-clock**: diferencia entre el primer y último timestamp incluido para el rango seleccionado.
- **AI working time**: suma de intervalos entre actividad `assistant`, tools y eventos automáticos, con cap de gap de 5 minutos para evitar contar pausas largas.
- **Tiempo humano activo**: suma de ventanas alrededor de mensajes `user` y gaps cortos adyacentes, con cap de gap de 10 minutos.
- **Tool calls / top tools**: conteos por nombre de tool únicamente.
- **Subagents / skills**: conteos por campos estructurales (`agent`, `command`, nombres de skill detectados) sin inputs ni outputs.
- **Fases**: timeline inferido por actividad estructural: exploración (lecturas/búsquedas), implementación (ediciones/parches), validación (tests/checks/OpenSpec), delivery-readiness (Git/GH/status sin ejecutar delivery).
- **Paso a paso**: timeline estructural de la tarea con tramos sanitizados, qué tipo de trabajo ocurrió, por qué ocurrió, actores/tools/skills involucrados, duración total, tiempo IA y tiempo humano por tramo.
- **Impacto diario**: resumen ejecutivo del valor para el desarrollador: pedido original, aporte de la IA, resultado revisable y evidencia local.
- **Aprendizajes y pivots**: aprendizajes inferidos desde stack, metodología, skills y limitaciones. Los pivots explícitos se reportan solo cuando existan señales estructuradas; no se inventan motivos conversacionales.
- **Tier / spec**: `--tier` y `--change` ganan si se pasan. Sin flags, inferir solo de forma conservadora: OpenSpec/LUFY SDD por skills/señales locales o `none`; no inventar un change id.

## Estilo visual

- El HTML sigue una estética tipo Notion: fondo cálido claro, tipografía de sistema/Inter, propiedades de página, callouts suaves y tablas tipo database.
- Mantener el diseño content-first: métricas como propiedades y evidencia revisable, no dashboard pesado.
- No usar recursos remotos, CDNs, imágenes ni fonts externas.

Toda métrica de tiempo con timestamps insuficientes debe mostrarse como `No disponible` o `Estimación parcial`; no se debe inventar precisión.

## Output contract

El HTML debe incluir secciones visibles para:

- propiedades de la tarea
- resumen ejecutivo
- impacto diario
- wall-clock
- AI working time
- tiempo humano activo
- paso a paso con qué se hizo, por qué, duración, tiempo IA y tiempo humano
- aprendizajes y pivots
- LOC neto
- commits
- tool calls
- top tools
- subagents
- skills
- fases/timeline
- stack detectado
- metodología y limitaciones

El archivo debe tener CSS embebido y no debe contener referencias `http://`, `https://`, CDNs ni assets externos.

## Validación local

Smoke read-only con fixtures sanitizados:

```bash
python3 .opencode/skills/lufy.timereport/tests/smoke_timereport.py
```

Validación OpenSpec:

```bash
openspec validate add-lufy-timereport --strict
```

El smoke crea datos sintéticos temporales y verifica que el HTML contenga secciones requeridas y no contenga prompts, outputs, diffs ni payloads completos de fixture.
