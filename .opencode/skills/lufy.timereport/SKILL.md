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

Genera un reporte local de tiempo/ROI para el repositorio actual. El reporte es un único HTML autocontenido, sin recursos remotos, diseñado para revisión offline y para compartir solo métricas estructurales sanitizadas.

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

El comando slash `.opencode/commands/lufy.timereport.md` es solo una fachada discoverable y debe delegar en este skill.

## Fuentes de datos y degradación

### OpenCode SQLite primario

- Ruta por defecto: `~/.local/share/opencode/opencode.db`.
- Apertura read-only: URI SQLite `file:<path>?mode=ro`.
- Tablas esperadas e introspeccionadas defensivamente: `project`, `workspace`, `session`, `message`, `part` y `event`.
- Filtrado por repositorio: incluir sesiones cuyo `directory`, `path`, `cwd`, `root`, `project_path`, `workspace_path` o relación con `project`/`workspace` coincida con `--target-dir` cuando esas columnas existan. Si el schema no expone directorio, degradar con limitación visible en vez de mezclar datos silenciosamente.
- Schema mismatch: si falta una tabla/columna requerida para una métrica, marcar solo esa sección como `No disponible` o `Estimación parcial` e incluir la fuente faltante en Limitaciones.

### Git secundario

- Operaciones permitidas: `git rev-parse --is-inside-work-tree` y `git log --numstat` con rango temporal cuando aplique.
- Métricas: conteo de commits y LOC neto (`líneas agregadas - líneas eliminadas`).
- Si el target no es repositorio Git o Git no está disponible, continuar con secciones no-Git y marcar commits/LOC como `No disponible`.

### Stack opcional

- Preferir `.opencode/project.yaml` si existe y contiene metadata simple de stack/tooling.
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

Toda métrica de tiempo con timestamps insuficientes debe mostrarse como `No disponible` o `Estimación parcial`; no se debe inventar precisión.

## Output contract

El HTML debe incluir secciones visibles para:

- wall-clock
- AI working time
- tiempo humano activo
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
