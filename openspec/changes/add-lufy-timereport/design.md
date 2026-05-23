## Context

LUFY-3 requiere un reporte `/lufy.timereport` que ayude a entender el tiempo invertido, ROI y actividad técnica del repositorio sin depender de servicios externos. El repositorio usa OpenCode localmente, con patrones de comandos en `.opencode/commands/*` que delegan en skills bajo `.opencode/skills/*`. La fuente local observada de OpenCode es SQLite en `~/.local/share/opencode/opencode.db`, con tablas como `project`, `workspace`, `session`, `message`, `part` y `event`; para este repositorio las sesiones se filtran por `session.directory = /Users/adrianrojas/Desktop/projects/lufy-ai` o el directorio actual equivalente.

Observatory ya modela conceptos útiles como sesiones, mensajes, partes, tool calls, subagents, costs/tokens y duraciones, pero este cambio define una capability de reporte independiente, privada por defecto y orientada a HTML autocontenido. `.opencode/project.yaml` puede no existir todavía, por lo que el stack debe degradar a detección heurística o estado `No configurado`.

## Goals / Non-Goals

**Goals:**
- Crear un contrato para `.opencode/commands/lufy.timereport.md` y `.opencode/skills/lufy.timereport/SKILL.md`.
- Generar un HTML offline/autocontenido en una ruta temporal por defecto o ruta configurable por el usuario.
- Usar OpenCode SQLite local como fuente primaria de sesiones y actividad, Git como fuente secundaria para commits/LOC neto, y `.opencode/project.yaml` como metadata opcional de stack.
- Reportar wall-clock, AI working time, tiempo humano activo, LOC neto, commits, tool calls, top tools, subagents, skills, fases y stack detectado.
- Mantener privacidad por defecto: métricas estructurales/sanitizadas sin prompts, outputs, tool payloads, diffs ni `session_diff`.
- Definir degradaciones explícitas cuando falten DB, Git o `.opencode/project.yaml`.
- Validar con fixtures/datos sanitizados y comandos read-only.

**Non-Goals:**
- No implementar todavía el comando, skill, parser, generador HTML ni tests.
- No leer JSONL ni `session_diff` en el alcance inicial, salvo como extensión futura con opt-in explícito.
- No enviar datos a red, dashboards remotos ni APIs externas.
- No persistir reportes en el repositorio por defecto ni modificar historial Git.
- No calcular productividad individual con contenido semántico de prompts/respuestas.

## Decisions

1. **Fuente primaria: SQLite local de OpenCode.**
   - Decisión: consultar `~/.local/share/opencode/opencode.db` en modo read-only y filtrar sesiones por directorio del proyecto.
   - Razón: es la fuente real observada y contiene sesiones, mensajes, partes, eventos y metadatos suficientes para métricas estructurales.
   - Alternativas: JSONL de sesiones o `session_diff`; se descartan del alcance inicial porque no están confirmados como fuente principal y pueden exponer contenido sensible.

2. **Privacidad por defecto y datos sanitizados.**
   - Decisión: el reporte solo incluirá agregados, conteos, duraciones, nombres de herramientas/subagents/skills/fases y metadata de stack; no incluirá prompts, outputs, argumentos completos de herramientas, diffs ni `session_diff`.
   - Razón: el reporte es fácil de compartir y debe minimizar fuga de IP, credenciales o contenido conversacional.
   - Alternativas: incluir extractos o diffs para explicar cada fase; se deja fuera hasta que exista opt-in explícito y reglas de redacción.

3. **HTML offline/autocontenido.**
   - Decisión: generar un único `.html` con CSS/JS embebido, sin CDNs ni recursos remotos, en `/tmp/lufy-timereport-<timestamp>.html` o ruta indicada por el usuario.
   - Razón: facilita revisión local y sharing manual sin dependencias de red.
   - Alternativas: Markdown/JSON como output principal; pueden ser formatos futuros, pero el backlog exige HTML autocontenido.

4. **Heurísticas de tiempo explícitas y reproducibles.**
   - Decisión: documentar y mostrar en el reporte las reglas usadas:
     - Wall-clock: diferencia entre primer y último timestamp de sesiones incluidas, agrupadas por ventana seleccionada.
     - AI working time: suma de intervalos con actividad de assistant/tool/eventos automáticos, con cap por gaps largos para evitar inflar pausas.
     - Tiempo humano activo: suma de intervalos alrededor de mensajes de usuario y gaps cortos entre interacción humana y respuesta, también con cap configurable.
   - Razón: OpenCode no necesariamente almacena todas las duraciones semánticas necesarias; heurísticas transparentes evitan falsa precisión.
   - Alternativas: usar solo duración de sesión o solo timestamps de mensajes; sería menos útil para distinguir espera de IA vs intervención humana.

5. **Git como fuente secundaria para commits y LOC neto.**
   - Decisión: usar comandos Git read-only para contar commits y calcular LOC neto en el rango temporal/scope del reporte cuando Git esté disponible.
   - Razón: Git es la fuente confiable para cambios versionados y no requiere leer contenido sensible por defecto.
   - Alternativas: derivar LOC desde tool calls o diffs de sesión; se evita por sensibilidad y menor confiabilidad.

6. **Stack opcional con degradación.**
   - Decisión: preferir `.opencode/project.yaml` para stack/tooling y degradar a heurística por archivos conocidos o `No configurado` si falta.
   - Razón: `project.yaml` es parte del roadmap stack-aware pero puede no existir en todos los repositorios.
   - Alternativas: exigir `project.yaml`; bloquearía el uso inicial y contradice degradación deseada.

7. **Comando slash como fachada del skill.**
   - Decisión: agregar `.opencode/commands/lufy.timereport.md` para invocación consistente con comandos locales y delegar la lógica en `.opencode/skills/lufy.timereport/SKILL.md`.
   - Razón: mantiene el patrón local de comandos en español que cargan skills.
   - Alternativas: solo skill manual; reduciría discoverability de `/lufy.timereport`.

## Risks / Trade-offs

- **Schema SQLite de OpenCode cambia** → Mitigar con introspección defensiva de tablas/columnas, errores accionables y fixtures que cubran columnas mínimas.
- **Timestamps incompletos o inconsistentes** → Mitigar mostrando cobertura y supuestos; si faltan timestamps críticos, degradar métricas de tiempo a `No disponible` en vez de inventarlas.
- **Métricas de AI/humano pueden ser aproximadas** → Mitigar declarando heurísticas, caps de gaps y confidence/limitations dentro del reporte.
- **Git no disponible o fuera de repositorio** → Mitigar reportando commits/LOC como `No disponible` y continuar con métricas OpenCode.
- **`.opencode/project.yaml` ausente o inválido** → Mitigar usando stack heurístico o `No configurado`, sin fallar el reporte completo.
- **Riesgo de privacidad por campos sensibles en DB** → Mitigar con allowlist de campos agregados y pruebas que fallen si aparecen prompts, outputs, diffs o payloads completos en el HTML por defecto.
- **HTML autocontenido crece demasiado** → Mitigar con agregados y tablas resumidas por defecto; no embeber datos crudos.

## Migration Plan

1. Implementar el comando y skill como archivos nuevos sin modificar capacidades existentes.
2. Agregar helpers/fixtures sanitizados donde se decida en implementación, manteniendo lectura local read-only.
3. Validar con `openspec validate add-lufy-timereport --strict`, revisión estática, fixtures sanitizados y comandos read-only de Git/SQLite cuando existan.
4. Rollback: eliminar los archivos nuevos del comando/skill/tests/fixtures; no hay migración de datos ni cambios persistentes requeridos.

## Open Questions

- ¿El rango por defecto del reporte debe ser la sesión actual, las últimas 24 horas o el rango completo del repositorio? Propuesta inicial: sesión actual si puede inferirse; si no, últimas 24 horas con opción de override.
- ¿Se requiere output JSON adicional para automatización futura? Fuera de alcance inicial salvo que implementación lo necesite para tests internos.
