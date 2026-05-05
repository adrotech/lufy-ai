# Roadmap de hardening y evolución de lufy-ai

Este roadmap captura trabajo futuro taggeable derivado del análisis comparativo con `Gentleman-Programming/gentle-ai`. El alcance real de `lufy-ai` sigue siendo un kit OpenCode/OpenSpec instalable en repositorios destino. La CLI Go actual existe para instalar, verificar y sincronizar assets gestionados del kit; no convierte el repo en un framework de aplicación ni en un producto multiagente separado.

## Visión

Convertir `lufy-ai` en una capa operativa instalable, segura e idempotente para proyectos existentes, con validación básica automatizada y capacidad de reaplicar assets gestionados sin destruir configuración local.

La evolución debe priorizar primero la seguridad del instalador y la confianza del usuario antes de expandir templates, automatizaciones o empaquetado como producto.

## Principios de diseño

- **Instalación segura por defecto**: ningún archivo crítico debe sobrescribirse sin backup, estrategia explícita o confirmación.
- **Idempotencia verificable**: ejecutar el instalador o sync más de una vez debe producir un estado estable y auditable.
- **Compatibilidad con repositorios existentes**: respetar `.opencode/`, `AGENTS.md`, `tui.json`, `opencode.json` y `openspec/` cuando ya existan.
- **Flags predecibles para automatización**: `--dry-run`, `--yes`, `--no-engram` y `--target` deben funcionar sin prompts inesperados.
- **Detección portable**: usar `command -v engram` para resolver Engram, evitando paths hardcodeados como `/opt/homebrew/bin/engram`.
- **Validación simple antes de crecer**: añadir checks básicos de shell, JSON e instalación en temp dir antes de introducir features mayores.
- **Documentación honesta**: documentar solo templates y assets que realmente se instalan.
- **Evolución incremental**: productizar como CLI queda para una fase futura, no como prioridad inmediata.
- **Migración compatible de Bash a Go**: Bash queda como wrapper estricto de compatibilidad que delega en `lufy-ai install`, sin fallback legacy.

## Estado instalable vs roadmap

Esta página contiene ideas futuras y decisiones estratégicas. Salvo que una sección indique explícitamente que algo ya existe en la rama actual, no debe leerse como contrato instalable.

Estado actual documentable:

- CLI Go en `tools/lufy-cli-go/` con `install`, `verify`, `backup`, `restore` y `sync`.
- Assets gestionados con estado `.lufy-ai/install-state.json`, hashes SHA-256, idempotencia y backups antes de actualizaciones gestionadas.
- Wrapper `scripts/install.sh` estricto, sin fallback legacy ni detección de stack en Bash.
- Workflow mínimo `.github/workflows/go-cli-install.yml` presente en esta rama para tests/build/smokes de la CLI Go y `git diff --check`; su existencia no implica archive automático de proposals OpenSpec.

No son capacidades instalables actuales:

- templates por stack como `frontend-react`, `frontend-nextjs`, `frontend-astro`, `mobile-expo` o `backend-spring`;
- detección automática de stack;
- subagentes especializados adicionales como `infra-cloud-sre`, `react-ui`, `nextjs-app-router` o `astro-islands-content`.

Esos elementos se conservan abajo como roadmap para futuras iteraciones y solo deberían moverse al README cuando existan como assets reales, estén instalados por la CLI y tengan validación local/CI coherente.

## Prioridades por fases

### Fase 1 — Hardening del installer y seguridad de instalación

Objetivo: hacer que `scripts/install.sh` sea seguro, auditable e idempotente para uso local y CI.

- `RM-001`: Añadir flags `--dry-run`, `--yes`, `--no-engram` y `--target` robusto.
- `RM-002`: Implementar backup/rollback mínimo antes de tocar `.opencode/`, `AGENTS.md`, `tui.json`, `opencode.json` y `openspec/`.
- `RM-003`: Resolver Engram con `command -v engram` y escribir ese path en `opencode.json` solo cuando aplique.
- `RM-004`: Crear `scripts/verify-install.sh` para validar estructura instalada y JSON parseable.
- `RM-005`: Mejorar merge/idempotencia para no sobrescribir configs existentes; detectar bloques gestionados o pedir estrategia.

### Fase 2 — Validación continua mínima

Objetivo: asegurar que el kit instalable no se rompa por cambios documentales, scripts o JSON.

- `RM-006`: Añadir GitHub Actions básica para `shellcheck scripts/install.sh`, validación de JSON y test de instalación en temp dir.
- `RM-007`: Integrar `scripts/verify-install.sh` en el workflow y documentar cómo ejecutarlo localmente.

### Fase 3 — Sync seguro de assets gestionados

Objetivo: permitir reaplicar assets de `lufy-ai` sin convertir la instalación en una sobrescritura destructiva.

- `RM-008`: Añadir sync simple para reaplicar assets gestionados de forma segura.
- `RM-009`: Definir manifest o metadata de assets gestionados para distinguir archivos propios de personalizaciones locales.

### Fase 4 — Templates reales y documentación sincronizada

Objetivo: introducir templates por stack solo cuando existan como archivos instalables y estén validados.

- `RM-010`: Añadir templates reales por stack únicamente si se instalan realmente.
- `RM-011`: Separar roadmap del README moviendo futuro/plantillas a `docs/roadmap.md` y dejando README como estado real + quickstart.

### Fase 5 — Productización futura

Objetivo: evaluar empaquetado y DX avanzada cuando el instalador, sync y CI ya sean confiables.

- `RM-012`: Evaluar productización como CLI como fase futura, no prioridad inmediata.
- `RM-013`: Migrar installer a CLI Go y mantener Bash solo como wrapper estricto de compatibilidad, sin lógica legacy.

## Propuesta/design por iniciativa

### `RM-001` — Flags robustos del installer

**Propuesta**: reemplazar el manejo posicional actual por parsing explícito de argumentos.

**Design**:

- Soportar `--target <dir>` y mantener compatibilidad con el argumento posicional actual cuando no haya flag.
- `--dry-run` debe listar acciones planeadas sin escribir archivos ni clonar/cambiar estado persistente salvo temporales limpiables.
- `--yes` debe aceptar prompts seguros, pero no debe ocultar conflictos destructivos sin backup.
- `--no-engram` debe saltar detección, prompts y cambios MCP de Engram.
- Flags desconocidos deben fallar con ayuda breve y código distinto de cero.

### `RM-002` — Backup/rollback mínimo

**Propuesta**: crear un backup por ejecución antes de cualquier escritura sobre assets sensibles.

**Design**:

- Directorio sugerido: `.lufy-ai/backups/<timestamp>/` dentro del target o temp externo configurable.
- Respaldar solo si existe el path destino y será modificado.
- Registrar un manifest con paths respaldados, acciones realizadas y timestamp.
- Si una acción falla, intentar rollback de paths tocados en esa ejecución y reportar el estado final.
- No borrar backups automáticamente en la primera iteración.

### `RM-003` — Engram portable

**Propuesta**: resolver Engram con `command -v engram` al momento de instalar.

**Design**:

- Si `--no-engram`, no consultar ni escribir integración Engram.
- Si Engram existe y el usuario acepta integración, escribir el path resuelto en `opencode.json`.
- Si no existe, dejar MCP deshabilitado sin path hardcodeado o con una recomendación explícita de instalación.
- Evitar depender de `/opt/homebrew/bin/engram`, porque rompe Linux, macOS no-Homebrew y entornos CI.

### `RM-004` — `scripts/verify-install.sh`

**Propuesta**: añadir un verificador local para confirmar que un target quedó instalable.

**Design**:

- Aceptar `--target <dir>` y validar desde fuera del proyecto destino.
- Confirmar presencia de `.opencode/agents`, `.opencode/commands`, `.opencode/skills`, `.opencode/plugins`, `.opencode/policies`, `AGENTS.md`, `tui.json`, `opencode.json` y `openspec/` cuando corresponda.
- Validar JSON con `python3 -m json.tool` o herramienta disponible documentada.
- Fallar con mensajes accionables y código distinto de cero.

### `RM-005` — Merge/idempotencia de configs

**Propuesta**: distinguir archivos gestionados de archivos del usuario y evitar sobrescrituras silenciosas.

**Design**:

- Para archivos completos gestionados, comparar checksums o manifest antes de reemplazar.
- Para `AGENTS.md`, usar bloque gestionado con marcadores si se decide mezclar contenido.
- Para `opencode.json` y `tui.json`, preferir merge JSON conservador o pedir estrategia: `skip`, `backup-and-replace`, `merge`.
- El instalador repetido debe reportar `unchanged`, `updated`, `skipped` o `conflict` por asset.

### `RM-006` — GitHub Actions básica

**Propuesta**: crear un workflow CI mínimo enfocado en assets reales del repo.

**Design**:

- `shellcheck scripts/install.sh` y, cuando exista, `scripts/verify-install.sh`.
- Validar JSON de `tui.json`, `.opencode/package.json`, `.opencode/package-lock.json` y archivos JSON relevantes.
- Ejecutar instalación en un temp dir y luego `scripts/verify-install.sh --target <temp>`.
- No agregar pipelines de build/test de producto inexistente.

### `RM-007` — Verificación local documentada e integrada

**Propuesta**: convertir `scripts/verify-install.sh` en la comprobación estándar de instalación local y CI.

**Design**:

- Documentar en README o `docs/getting-started.md` el comando local recomendado.
- Mantener el script sin dependencias pesadas para que funcione en macOS/Linux con herramientas comunes.
- Reutilizar los mismos checks en CI para evitar divergencia entre validación local y remota.
- Reportar claramente qué asset falta o qué JSON no parsea.

### `RM-008` — Sync seguro de assets gestionados

**Propuesta**: agregar un modo simple para reaplicar assets de `lufy-ai` luego de una instalación inicial.

**Design**:

- Puede iniciar como `scripts/install.sh --sync --target <dir>` o script separado si eso reduce complejidad.
- Debe reutilizar backup, manifest e idempotencia de Fase 1.
- Debe limitarse a assets gestionados: `.opencode/agents`, `.opencode/commands`, `.opencode/skills`, `.opencode/policies`, `.opencode/plugins` y templates reales.
- No debe reescribir convenciones locales sin estrategia explícita.

### `RM-009` — Manifest de assets gestionados

**Propuesta**: registrar qué archivos pertenecen a `lufy-ai` para habilitar sync e idempotencia sin adivinar.

**Design**:

- Crear un manifest versionado con path, tipo de asset, estrategia de actualización y checksum opcional.
- Distinguir assets reemplazables completos de archivos que requieren merge o confirmación.
- Usar el manifest tanto en instalación como en sync y verificación.
- Evitar incluir archivos generados o personalizaciones del proyecto destino.

### `RM-010` — Templates reales por stack

**Propuesta**: no prometer templates de stack hasta que existan y se instalen.

**Design**:

- Cada template debe tener archivos concretos, criterio de instalación y verificación.
- Actualizar README solo cuando el template exista como asset real.
- Mantener templates recomendados como roadmap mientras no existan: `frontend-react`, `frontend-nextjs`, `frontend-astro`, `mobile-expo`, `backend-spring`.
- Evitar `backend-node` genérico salvo que se defina un caso de uso backend concreto.

### `RM-011` — README como estado real y roadmap separado

**Propuesta**: mantener README enfocado en quickstart, estado actual y enlaces; mover contenido futuro a `docs/roadmap.md` cuando se haga la reestructuración documental.

**Design**:

- Por ahora, enlazar este roadmap desde README sin reescritura amplia.
- En una iteración futura, migrar secciones especulativas de templates/subagentes a este documento o a docs específicos.
- Evitar que README venda templates o features no instalables.
- Mantener una navegación corta hacia `docs/getting-started.md`, `docs/roadmap.md` y `openspec/README.md`.

### `RM-012` — CLI futura

**Propuesta**: evaluar CLI después de estabilizar instalador, CI, sync y templates reales.

**Design**:

- No bloquear Fase 1 por selección de lenguaje o empaquetado.
- Considerar CLI solo si reduce complejidad del shell o mejora instalación multiplataforma.
- Requisitos mínimos futuros: comandos `install`, `verify`, `sync`, `doctor`, versionado y distribución clara.

### `RM-013` — Migrar installer a CLI Go

**Propuesta**: productizar el installer como CLI Go (`lufy-ai`) cuando el hardening inicial esté validado, moviendo progresivamente la lógica crítica fuera de Bash.

**Rationale Go vs Rust**:

- Go es suficiente para CLI, filesystem, JSON, OS detection y distribución como single-binary.
- Rust no es necesario para esta etapa porque el problema principal es idempotencia, backup/rollback, portabilidad y UX, no performance ni memory-safety.
- Mantener Bash solo como wrapper estricto de compatibilidad reduce fricción para usuarios actuales sin seguir aumentando la complejidad del shell ni duplicar rutas legacy.

**Design por fases**:

1. Mantener Bash como wrapper estricto que delega en la CLI Go; si no hay binario local o en `PATH`, falla con instrucciones de build local.
2. Crear una CLI Go mínima con comandos base: `lufy-ai install --target . --dry-run --yes --no-engram`, `lufy-ai verify --target .`, `lufy-ai backup` y `lufy-ai restore`.
3. Mover al binario la detección de entorno, Engram portable, backup/rollback, merge/idempotencia y `verify`.
4. Agregar `sync`/`update` después de estabilizar instalación, verificación y rollback.
5. Evaluar una TUI Go como opción futura; no es prioridad inicial.

**Estado actual (2026-05-05):**

- ✅ Scaffolding Go en carpeta dedicada `tools/lufy-cli-go/` con `go.mod` propio.
- ✅ Comandos base cableados (`install`, `verify`, `backup`, `restore`) en slice mínimo funcional.
- ✅ Wrapper `scripts/install.sh` delega exclusivamente a `lufy-ai install`, usando `tools/lufy-cli-go/bin/lufy-ai` o `lufy-ai` en `PATH`, sin fallback legacy.
- ✅ Resolución Engram portable por `PATH` sin hardcode nuevo en la ruta de migración.
- ✅ Smoke E2E reproducible validado en temp dir para install real + verify + idempotencia básica (2da ejecución con `skip`) + backup/restore (dry-run y real) sobre conflicto controlado de `AGENTS.md`.
- ✅ Install real copia assets gestionados del catálogo, escribe `.lufy-ai/install-state.json` con SHA-256 y evita sobrescribir drift local.
- ✅ `backup`/`restore` usan `manifest.json`, hashes, `targetRoot` y backup de recovery antes de restauraciones reales.
- ✅ `sync` reaplica assets gestionados con hash/idempotencia y backup previo, bloqueando drift local, estado ausente/corrupto y escapes por symlink/path inseguro.
- 🔄 Pendiente inmediato: decidir si `opencode.json` entra como asset gestionado futuro con merge conservador.

## Criterios de aceptación

### Criterios generales

- Cada iniciativa conserva un ID estable `RM-###` para issues, ramas o tags.
- Los cambios siguen siendo compatibles con el rol real del repo: kit OpenCode/OpenSpec instalable.
- La documentación humana permanece en español y preserva identificadores técnicos.
- No se documenta como existente algo que no esté implementado o instalado realmente.

### Criterios por iniciativa

| ID | Criterios de aceptación |
| --- | --- |
| `RM-001` | `scripts/install.sh --dry-run --target <dir>` no escribe en el target; `--yes` no requiere prompts seguros; `--no-engram` omite Engram; flags inválidos fallan con ayuda. |
| `RM-002` | Antes de modificar paths sensibles se crea backup; ante fallo se intenta rollback; queda manifest de la ejecución. |
| `RM-003` | `opencode.json` no contiene `/opt/homebrew/bin/engram` hardcodeado cuando la integración se genera; usa el resultado de `command -v engram` o queda deshabilitada. |
| `RM-004` | `scripts/verify-install.sh --target <dir>` valida estructura, JSON parseable y presencia de commands/skills/plugin. |
| `RM-005` | Reinstalar sobre un target existente no sobrescribe configs sin estrategia; los conflictos se reportan de forma accionable. |
| `RM-006` | Existe workflow en `.github/workflows/` con tests/build Go, smoke de instalación en temp dir, wrapper smoke y checks estáticos disponibles; `shellcheck` queda como mejora opcional si se incorpora al runner. |
| `RM-007` | La verificación local está documentada y CI ejecuta el mismo script o checks equivalentes. |
| `RM-008` | Sync reaplica assets gestionados con backup y reporte por archivo, sin modificar personalizaciones fuera de scope. |
| `RM-009` | Existe manifest de assets gestionados y lo consumen instalación/sync/verificación cuando aplique. |
| `RM-010` | Cualquier template documentado como disponible existe como asset instalable y es verificado por CI o script local. |
| `RM-011` | README queda centrado en estado real + quickstart; contenido futuro vive en docs de roadmap o diseño. |
| `RM-012` | Hay decisión documentada de CLI con alcance, tradeoffs y prerequisitos cumplidos antes de implementarla. |
| `RM-013` | Existe plan de migración a CLI Go; Bash queda limitado a wrapper estricto sin fallback legacy; la primera CLI cubre `install`, `verify`, `backup` y `restore` antes de `sync`/`update`. |

## Riesgos/dependencias

- `shellcheck` puede no estar disponible localmente; CI debe instalarlo o usar una acción mantenida.
- Merge JSON sin tooling dedicado puede volverse frágil en shell; si crece, conviene evaluar helper en Python o CLI futura.
- Rollback parcial puede ser complejo si se mezclan copias recursivas y merges; la primera versión debe limitar acciones y registrar manifest.
- `--yes` puede ser peligroso si se interpreta como permiso destructivo; debe combinarse con backup y políticas conservadoras.
- Sync requiere una fuente de verdad de assets gestionados; sin manifest puede pisar personalizaciones.
- Templates por stack pueden reintroducir drift documental si README se actualiza antes de tener archivos reales.
- Migrar lógica desde Bash a Go puede dejar brechas mientras la CLI completa assets gestionados; el wrapper evita duplicación al no mantener fallback legacy.

## Etiquetas sugeridas

- `roadmap`
- `installer-hardening`
- `install-safety`
- `idempotency`
- `backup-rollback`
- `engram`
- `ci`
- `verify-install`
- `sync`
- `templates`
- `documentation`
- `future-cli`
- `go-cli`
