# Primeros pasos con lufy-ai

## Qué es `lufy-ai`

`lufy-ai` es un kit instalable para sumar un flujo AI-first a un repositorio existente. No crea una aplicación ni instala templates por stack; copia assets operativos para usar el preset actual OpenCode/OpenSpec, harness SDD proporcional, subagentes especializados y reglas de delivery trazable.

El diseño del harness ya separa core, tool adapter y methodology adapter. Para usuarios finales no cambia el quickstart: `lufy-ai install` sigue instalando OpenCode/OpenSpec por defecto. La separación permite que futuros adapters como Codex o Claude Code no obliguen a reescribir tiers, roles, Result Contract ni policies.

Incluye:

- agentes OpenCode con responsabilidades separadas;
- `sdd-router` para clasificar T1 Full SDD, T2 SDD Lite o T3 Express antes de usar flujos pesados;
- comandos slash `/opsx-*` para el ciclo OpenSpec core v2;
- templates operativos `.opencode/templates/sdd-lite.md` y `.opencode/templates/result-contract.md`;
- política de delivery en `.opencode/policies/delivery.md`;
- plugin local Agent Observatory para la TUI;
- CLI Go `lufy-ai` para `install`, `verify`, `backup`, `restore`, `sync`, `status`, `upgrade` y `version`;
- wrapper estricto `scripts/install.sh` que delega en `lufy-ai install`.

## Requisitos

- Para uso final: una release publicada con tag `v*` y un directorio de instalación escribible que puedas agregar a `PATH`.
- Para desarrollo/contribución: Go y un checkout de este repositorio para compilar desde `tools/lufy-cli-go/`.
- OpenCode en el repositorio destino para consumir agentes, comandos y plugin.
- Engram es opcional; usa `--no-engram` para omitirlo.

## Instalación rápida

Versión estable objetivo: `v0.4.0`. El paso a paso completo por OS/shell, incluyendo `PATH` para bash, zsh y fish, está en [`docs/installation.md`](installation.md).

### 1. Instalar el binario sin clone desde una release estable

```bash
curl -fsSL https://raw.githubusercontent.com/adrotech/lufy-ai/v0.4.0/scripts/bootstrap.sh -o /tmp/lufy-bootstrap.sh
less /tmp/lufy-bootstrap.sh
bash /tmp/lufy-bootstrap.sh --version v0.4.0 --install-dir "$HOME/.local/bin"
```

Si `~/.local/bin` no está en `PATH`, configura tu shell antes de continuar. Ejemplos rápidos:

```bash
# bash/zsh
export PATH="$HOME/.local/bin:$PATH"
```

```fish
# fish
fish_add_path $HOME/.local/bin
```

El bootstrap detecta OS/arch, verifica SHA-256 e instala solo el binario. No ejecuta `lufy-ai install` contra ningún proyecto por defecto.

### 2. Revisar el plan con `--dry-run`

```bash
lufy-ai version
lufy-ai install --target /ruta/a/tu/proyecto --tool opencode --dry-run --yes --no-engram
```

### 3. Aplicar la instalación

```bash
lufy-ai install --target /ruta/a/tu/proyecto --tool opencode --yes --no-engram
```

### 4. Verificar el target instalado

```bash
lufy-ai verify --target /ruta/a/tu/proyecto --tool opencode --no-engram
lufy-ai status --target /ruta/a/tu/proyecto
```

## Flujo de desarrollo/contribuidor con clone local

Usa este camino para trabajar en este repositorio o validar cambios antes de que exista una release publicada:

```bash
git clone https://github.com/adrotech/lufy-ai.git /tmp/lufy-ai
cd /tmp/lufy-ai/tools/lufy-cli-go
mkdir -p bin
go build -o bin/lufy-ai ./cmd/lufy-ai
```

`scripts/install.sh` busca primero ese binario local (`tools/lufy-cli-go/bin/lufy-ai`) y luego `lufy-ai` en `PATH`. Si no existe ninguno, falla sin fallback legacy y muestra la instrucción de build. Este wrapper local no descarga releases.

Revisar el plan con `--dry-run`:

```bash
/tmp/lufy-ai/scripts/install.sh --target /ruta/a/tu/proyecto --dry-run --yes --no-engram
```

Forma equivalente con la CLI:

```bash
/tmp/lufy-ai/tools/lufy-cli-go/bin/lufy-ai install --target /ruta/a/tu/proyecto --dry-run --yes --no-engram
```

Aplicar la instalación:

```bash
/tmp/lufy-ai/scripts/install.sh --target /ruta/a/tu/proyecto --yes --no-engram
```

El argumento posicional histórico se conserva como alias de target:

```bash
/tmp/lufy-ai/scripts/install.sh /ruta/a/tu/proyecto --yes --no-engram
```

Verificar el target instalado:

```bash
/tmp/lufy-ai/tools/lufy-cli-go/bin/lufy-ai verify --target /ruta/a/tu/proyecto --no-engram
```

## Qué hace el instalador actual

`lufy-ai install`:

1. resuelve `--target` a una ruta segura;
2. construye un plan de instalación;
3. respeta `--dry-run` sin mutaciones;
4. copia assets gestionados del catálogo (`.opencode`, `.opencode/templates`, `lufy-ia.harness.md`, `tui.json`, `openspec` base);
5. crea o mergea `opencode.json` de forma conservadora: preserva claves desconocidas, agrega solo estructura mínima gestionada y, si Engram está habilitado, conserva otros MCP locales dentro de `mcp`;
6. no trata `opencode.json` como asset completo por hash: queda fuera del manifest de assets completos y se valida por JSON/estructura mínima durante `verify`;
7. trata `AGENTS.md` como user-owned: si falta lo crea mínimo con `@lufy-ia.harness.md`; si existe agrega solo esa referencia con backup/`--yes`; si ya está presente no lo reescribe ni duplica;
8. registra `.lufy-ai/install-state.json` con hashes SHA-256 para assets completos gestionados, incluyendo `lufy-ia.harness.md` y excluyendo `AGENTS.md`;
9. registra ownership por asset (`tool`, `methodology`, `component`, `scope`) y la metodología efectiva por tier;
10. evita sobrescribir archivos con drift local;
11. crea backups antes de actualizaciones gestionadas cuando corresponde;
12. omite Engram con `--no-engram` o lo resuelve desde `PATH` cuando aplique.

Si `opencode.json` existente no es JSON válido, o si `mcp` existe con un tipo incompatible cuando debe agregarse Engram, `install`/`sync` fallan sin sobrescribirlo y piden corregir o respaldar el archivo.

Flags frecuentes:

| Flag | Uso |
| --- | --- |
| `--target <dir>` | Repositorio destino. |
| `--dry-run` | Imprime el plan sin escribir archivos. |
| `--yes` | Autoriza mutaciones reales cuando el plan es seguro. |
| `--no-engram` | Omite integración Engram. |
| `--backup <path>` | Ruta de backup usada por `restore`. |

## Comandos de la CLI Go

| Comando | Estado actual |
| --- | --- |
| `lufy-ai install` | Instala assets gestionados con estado SHA-256 e idempotencia. |
| `lufy-ai verify` | Valida estructura, estado y hashes del target. |
| `lufy-ai backup` | Crea backup multiasset con `manifest.json`. |
| `lufy-ai restore` | Restaura desde backup y valida seguridad del manifest. |
| `lufy-ai sync` | Reaplica assets gestionados, actualiza `lufy-ia.harness.md` y preserva `AGENTS.md`; si falta `@lufy-ia.harness.md`, reporta acción explícita sin auto-reparar. |
| `lufy-ai status` | Resume estado instalado, drift local, faltantes y errores; soporta `--json` y `--verbose`. |
| `lufy-ai upgrade` | Actualiza el binario a una versión fija verificando checksum antes de reemplazarlo. |
| `lufy-ai version` | Muestra versión, commit, build date, GOOS y GOARCH; si falta metadata de linker reporta development build. |

Flags útiles de verificación:

| Flag | Uso |
| --- | --- |
| `verify --json` | Emite reporte estructurado para CI/automatización. |
| `install --tool opencode` | Selecciona explícitamente el adapter actual; sin flag conserva el default. |
| `install --methodology-tier T3:none` | Override repetible de metodología por tier; acepta `openspec`, `lufy-sdd` o `none` donde la policy lo permite. |
| `verify --tool opencode` | Valida que el manifest instalado use el adapter esperado. |
| `verify --quiet` | Suprime salida humana por stdout. |
| `verify --verbose` | Agrega diagnóstico adicional. |
| `verify --deep` | Valida referencias de plugins en `tui.json` y `opencode.json`. |

Detalles técnicos y comandos de validación: [`tools/lufy-cli-go/README.md`](../tools/lufy-cli-go/README.md).

## Uso después de instalar

1. Revisa `AGENTS.md` en el repositorio destino y ajusta convenciones locales; conserva una referencia `@lufy-ia.harness.md` para cargar el harness gestionado.
2. Reinicia OpenCode para cargar agentes, comandos, templates y plugin.
3. Deja que `sdd-router` clasifique cambios no triviales en modo read-only/no-shell: T1 Full SDD, T2 SDD Lite o T3 Express.
4. Usa `/opsx-explore` y `/opsx-propose` para T1 o cambios con alta incertidumbre.
5. Usa `.opencode/templates/sdd-lite.md` para T2 cuando baste un mini-spec profesional con criterios `WHEN`/`THEN`.
6. Usa `/opsx-apply`, `/opsx-verify`, `/opsx-sync` y `/opsx-archive` según corresponda.
7. Usa `opsx-version` para reportar la fuente OpenSpec efectiva: `PATH`, cache local o baseline embebida offline.
8. Deja Git/GitHub en manos de `delivery` solo con autorización explícita.

Para features grandes, piensa en el reviewer humano desde el diseño. El harness puede proponer `review_slices`: subproblemas pequeños con objetivo, archivos esperados, criterios `WHEN`/`THEN`, validación y guía de PR. Úsalos en T1 y T2 con varios riesgos; evita fragmentar T3 o cambios pequeños sin necesidad.

### Migración mínima desde assets `v0.2.0`

Al sincronizar desde una fuente con OpenSpec core v2, `lufy-ai sync` agrega `openspec/UPSTREAM.json`, `/opsx-sync`, `opsx-version` y la skill `openspec-sync` como assets gestionados. `UPSTREAM.json` ahora también declara la versión mínima compatible y el orden de resolución stay-updated: `openspec` en `PATH`, cache `.lufy-ai/openspec-cache/<version>/manifest.json` y baseline embebida offline. `install` y `sync` no descargan ni ejecutan OpenSpec remoto por defecto. Si un target tenía cambios locales en assets gestionados, aplican las mismas reglas de drift de `v0.2.0`: se preserva el cambio local, se reporta conflicto o `.lufy-new` según la policy y no se pisa trabajo del usuario.

Los cambios OpenSpec nuevos deben escribir specs delta bajo `openspec/changes/<change>/specs/` con markers `ADDED`, `MODIFIED` o `REMOVED`. Antes de archivar, corre `/opsx-sync <change>` para aplicar esos deltas a `openspec/specs/` sin mover el change.

## Flujo de contribución y release del repositorio

- Abre PRs normales desde ramas `feature/*`, `fix/*`, `chore/*` o equivalentes hacia `develop`; `delivery` debe consultar/esperar checks remotos del PR con evidencia antes de reportar `delivered` o `closed`.
- Reserva `main` para producción/estabilidad: promociones `develop` → `main` o hotfix/release explícitamente autorizados.
- Crea tags estables `v*` solo sobre commits alcanzables desde `origin/main`. El workflow de release bloquea publicación si el tag apunta a un commit que aún vive solo en `develop`.
- Consulta [`docs/github-branch-settings.md`](github-branch-settings.md) para configurar default branch `develop` y protección de `develop`/`main` en GitHub.

## Comandos slash disponibles

Los comandos `/opsx-*` son el namespace estable del workflow OpenSpec y se mantienen con ese nombre para no romper guías, skills ni memoria operativa existente. Los comandos `/lufy.*` son extras propios del kit Lufy y pueden complementar el flujo sin reemplazar ni renombrar ningún `/opsx-*`.

| Namespace | Comando | Descripción |
| --- | --- | --- |
| OpenSpec | `/opsx-explore` | Explora requisitos, impacto o código en modo read-only. |
| OpenSpec | `/opsx-propose` | Genera artefactos OpenSpec de un cambio. |
| OpenSpec | `/opsx-apply` | Implementa tareas de un cambio activo. |
| OpenSpec | `/opsx-verify` | Verifica completitud y coherencia contra artefactos. |
| OpenSpec | `/opsx-sync` | Aplica deltas validados a specs principales sin archivar. |
| OpenSpec | `/opsx-archive` | Archiva un cambio terminado cuando cumple gates. |
| OpenSpec | `/opsx-version` | Reporta fuente OpenSpec efectiva y diagnósticos de fallback desde `openspec/UPSTREAM.json`. |
| Lufy | `/lufy.timereport` | Genera un reporte local de tiempo/ROI como extra del kit. |

## Validación local disponible

Desde `tools/lufy-cli-go/`:

```bash
go test ./...
go build ./cmd/lufy-ai
scripts/smoke-install.sh
```

Desde la raíz del repo:

```bash
tools/lufy-cli-go/scripts/smoke-wrapper.sh
tools/lufy-cli-go/scripts/smoke-release-artifacts.sh
tools/lufy-cli-go/scripts/smoke-bootstrap.sh
git diff --check
```

El workflow `.github/workflows/go-cli-install.yml` existe en esta rama y cubre un gate mínimo para la CLI Go y el wrapper en PRs/pushes a `develop` y `main`. Que exista el workflow no reemplaza la validación local ni implica que otras proposals ya estén archivadas.

El workflow `.github/workflows/release.yml` construye artifacts versionados, checksums y smokes de release/bootstrap. Solo corre para tags `v*` y publica GitHub Releases si el commit taggeado es alcanzable desde `origin/main`; no hay release estable desde `develop` sin promoción.

No hay toolchain Node/TypeScript de producto en la raíz; no asumas `npm test`, `npm run typecheck` ni `tsc` global.

## Harness routing, templates y stack detection

Los templates operativos de proceso (`sdd-lite` y `result-contract`) sí son assets instalables. Los templates por stack, detección de stack integrada en CLI y subagentes de dominio adicionales siguen siendo roadmap. AutoSkills puede sugerirse como bootstrap opcional mediante `npx autoskills --dry-run`, pero no reemplaza skills locales ni se ejecuta sin autorización explícita. Ver [`docs/roadmap.md`](roadmap.md) para el contexto futuro.

## Solución de problemas

Para problemas de instalación del binario, `PATH`, fish o ejecución por ruta absoluta, consulta [`docs/installation.md#troubleshooting`](installation.md#troubleshooting).

### El wrapper no encuentra `lufy-ai`

Compila el binario local:

```bash
cd /tmp/lufy-ai/tools/lufy-cli-go
mkdir -p bin
go build -o bin/lufy-ai ./cmd/lufy-ai
```

### No existe una release para la versión seleccionada

El bootstrap depende de GitHub Releases. Si la versión seleccionada no tiene artifact para tu plataforma o checksums publicados, usa el flujo de desarrollo/contribuidor o espera a que exista la release correspondiente.

### Los agentes no cargan

1. Reinicia OpenCode.
2. Verifica que exista `.opencode/agents/` en el target.
3. Revisa que `AGENTS.md` contenga las convenciones del proyecto.

### El plugin TUI no aparece

1. Verifica `tui.json` en la raíz del target.
2. Confirma la ruta local del plugin en `tui.json`.
3. Usa `/observatory` para abrir/toggle del panel.

## Más documentación

- [README raíz](../README.md)
- [Instalación completa](installation.md)
- [OpenSpec Overview](../openspec/README.md)
- [CLI Go README](../tools/lufy-cli-go/README.md)
- [GitHub branch settings](github-branch-settings.md)
- [Roadmap](roadmap.md)
