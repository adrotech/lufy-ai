# Primeros pasos con lufy-ai

## Qué es `lufy-ai`

`lufy-ai` es un kit instalable para sumar un flujo AI-first a un repositorio existente. No crea una aplicación ni instala templates por stack; copia assets operativos para usar OpenCode, OpenSpec, subagentes especializados y reglas de delivery trazable.

Incluye:

- agentes OpenCode con responsabilidades separadas;
- comandos slash `/opsx-*` para el ciclo OpenSpec;
- política de delivery en `.opencode/policies/delivery.md`;
- plugin local Agent Observatory para la TUI;
- CLI Go `lufy-ai` para `install`, `verify`, `backup`, `restore`, `sync` y `version`;
- wrapper estricto `scripts/install.sh` que delega en `lufy-ai install`.

## Requisitos

- Para uso final: una release publicada con tag `v*` y un directorio de instalación escribible que puedas agregar a `PATH`.
- Para desarrollo/contribución: Go y un checkout de este repositorio para compilar desde `tools/lufy-cli-go/`.
- OpenCode en el repositorio destino para consumir agentes, comandos y plugin.
- Engram es opcional; usa `--no-engram` para omitirlo.

## Instalación rápida

### 1. Instalar el binario sin clone desde una release `v*`

El soporte de release/checkout standalone existe en esta rama, pero **no hay release publicada hasta crear y publicar un tag `v*`**. Cuando exista una release, usa pinning explícito:

```bash
curl -fsSL https://raw.githubusercontent.com/adrotech/lufy-ai/vX.Y.Z/scripts/bootstrap.sh -o /tmp/lufy-bootstrap.sh
less /tmp/lufy-bootstrap.sh
bash /tmp/lufy-bootstrap.sh --version vX.Y.Z --install-dir "$HOME/.local/bin"
```

También existe un atajo directo, documentado junto a la alternativa inspeccionable anterior:

```bash
curl -fsSL https://raw.githubusercontent.com/adrotech/lufy-ai/vX.Y.Z/scripts/bootstrap.sh \
  | bash -s -- --version vX.Y.Z --install-dir "$HOME/.local/bin"
```

El bootstrap:

1. detecta OS/arch soportado;
2. resuelve la versión seleccionada (`vX.Y.Z`, o `latest` solo como conveniencia explícita no reproducible);
3. descarga artifact y checksums de la misma release;
4. verifica SHA-256 antes de instalar o ejecutar el binario;
5. copia solo `lufy-ai` al directorio elegido y muestra guía de `PATH`.

No ejecuta `lufy-ai install` contra ningún proyecto por defecto.

### 2. Revisar el plan con `--dry-run`

```bash
lufy-ai version
lufy-ai install --target /ruta/a/tu/proyecto --dry-run --yes --no-engram
```

### 3. Aplicar la instalación

```bash
lufy-ai install --target /ruta/a/tu/proyecto --yes --no-engram
```

### 4. Verificar el target instalado

```bash
lufy-ai verify --target /ruta/a/tu/proyecto --no-engram
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
4. copia assets gestionados del catálogo (`.opencode`, `AGENTS.md`, `tui.json`, `openspec` base);
5. crea o mergea `opencode.json` de forma conservadora: preserva claves desconocidas, agrega solo estructura mínima gestionada y, si Engram está habilitado, conserva otros MCP locales dentro de `mcp`;
6. no trata `opencode.json` como asset completo por hash: queda fuera del manifest de assets completos y se valida por JSON/estructura mínima durante `verify`;
7. registra `.lufy-ai/install-state.json` con hashes SHA-256 para assets completos gestionados;
8. evita sobrescribir archivos con drift local;
9. crea backups antes de actualizaciones gestionadas cuando corresponde;
10. omite Engram con `--no-engram` o lo resuelve desde `PATH` cuando aplique.

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
| `lufy-ai sync` | Reaplica assets gestionados sin tocar drift local ni archivos fuera del catálogo. |
| `lufy-ai version` | Muestra versión, commit, build date, GOOS y GOARCH; si falta metadata de linker reporta development build. |

Detalles técnicos y comandos de validación: [`tools/lufy-cli-go/README.md`](../tools/lufy-cli-go/README.md).

## Uso después de instalar

1. Revisa `AGENTS.md` en el repositorio destino y ajusta convenciones locales.
2. Reinicia OpenCode para cargar agentes, comandos y plugin.
3. Usa `/opsx-explore` para investigar antes de cambios amplios.
4. Usa `/opsx-propose`, `/opsx-apply`, `/opsx-verify` y `/opsx-archive` para cambios OpenSpec.
5. Deja Git/GitHub en manos de `delivery` solo con autorización explícita.

## Flujo de contribución y release del repositorio

- Abre PRs normales desde ramas `feature/*`, `fix/*`, `chore/*` o equivalentes hacia `develop`.
- Reserva `main` para producción/estabilidad: promociones `develop` → `main` o hotfix/release explícitamente autorizados.
- Crea tags estables `v*` solo sobre commits alcanzables desde `origin/main`. El workflow de release bloquea publicación si el tag apunta a un commit que aún vive solo en `develop`.
- Consulta [`docs/github-branch-settings.md`](github-branch-settings.md) para configurar default branch `develop` y protección de `develop`/`main` en GitHub.

## Comandos slash disponibles

| Comando | Descripción |
| --- | --- |
| `/opsx-explore` | Explora requisitos, impacto o código en modo read-only. |
| `/opsx-propose` | Genera artefactos OpenSpec de un cambio. |
| `/opsx-apply` | Implementa tareas de un cambio activo. |
| `/opsx-verify` | Verifica completitud y coherencia contra artefactos. |
| `/opsx-archive` | Archiva un cambio terminado cuando cumple gates. |

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

## Templates y stack detection

Los templates por stack, detección de stack y subagentes adicionales son roadmap, no estado instalable actual. Ver [`docs/roadmap.md`](roadmap.md) para el contexto futuro.

## Solución de problemas

### El wrapper no encuentra `lufy-ai`

Compila el binario local:

```bash
cd /tmp/lufy-ai/tools/lufy-cli-go
mkdir -p bin
go build -o bin/lufy-ai ./cmd/lufy-ai
```

### No existe una release para `vX.Y.Z`

El bootstrap depende de GitHub Releases. Si aún no se creó un tag `v*` y se publicó la release correspondiente, usa el flujo de desarrollo/contribuidor o espera a que exista el artifact versionado con su archivo de checksums.

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
- [OpenSpec Overview](../openspec/README.md)
- [CLI Go README](../tools/lufy-cli-go/README.md)
- [GitHub branch settings](github-branch-settings.md)
- [Roadmap](roadmap.md)
