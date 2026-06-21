# Instalación de lufy-ai

Esta guía cubre:

- instalación del binario `lufy-ai`;
- configuración de `PATH`;
- instalación de assets en un repositorio destino;
- verificación, sync, uninstall y reinstall;
- troubleshooting básico.

Versión estable objetivo: `v0.6.11`.

## Requisitos

- Un directorio escribible para el binario, por ejemplo `~/.local/bin`.
- Acceso a una GitHub Release publicada con artifacts y checksums.
- Un repositorio destino donde instalar el harness.
- OpenCode si vas a usar el adapter escribible actual.

El bootstrap Bash aplica a macOS, Linux y WSL. En Windows nativo usa el binario publicado para Windows si la release lo incluye.

## Instalar el binario

Usa una versión explícita. `latest` existe como conveniencia, pero no es reproducible.

```bash
curl -fsSL https://raw.githubusercontent.com/adrotech/lufy-ai/v0.6.11/scripts/bootstrap.sh -o /tmp/lufy-bootstrap.sh
less /tmp/lufy-bootstrap.sh
bash /tmp/lufy-bootstrap.sh --version v0.6.11 --install-dir "$HOME/.local/bin"
```

El bootstrap:

1. detecta OS/arch;
2. descarga el artifact `lufy-ai_<version>_<os>_<arch>`;
3. verifica SHA-256 contra checksums de la release;
4. instala solo el binario;
5. no ejecuta `lufy-ai install` contra ningún proyecto.

Atajo directo, solo si ya revisaste el script:

```bash
curl -fsSL https://raw.githubusercontent.com/adrotech/lufy-ai/v0.6.11/scripts/bootstrap.sh \
  | bash -s -- --version v0.6.11 --install-dir "$HOME/.local/bin"
```

## PATH por shell

### bash/zsh

```bash
export PATH="$HOME/.local/bin:$PATH"
```

Guárdalo en `~/.bashrc`, `~/.bash_profile` o `~/.zshrc` según tu shell.

### fish

```fish
fish_add_path $HOME/.local/bin
```

Alternativa:

```fish
set -gx PATH $HOME/.local/bin $PATH
```

### Windows nativo

1. Descarga `lufy-ai_v0.6.11_windows_amd64.zip` o `lufy-ai_v0.6.11_windows_arm64.zip`.
2. Descarga `lufy-ai_v0.6.11_checksums.txt`.
3. Verifica el hash:

   ```powershell
   Get-FileHash .\lufy-ai_v0.6.11_windows_amd64.zip -Algorithm SHA256
   ```

4. Extrae `lufy-ai.exe` en un directorio de usuario.
5. Agrega ese directorio al `Path` de usuario.
6. Abre una terminal nueva.

## Instalar en un repositorio

Primero revisa la versión:

```bash
lufy-ai version
```

Luego revisa el plan sin mutar:

```bash
lufy-ai install --target /ruta/a/tu/proyecto --scope project --tool opencode --dry-run --yes
```

Aplica:

```bash
lufy-ai install --target /ruta/a/tu/proyecto --scope project --tool opencode --yes
```

Verifica:

```bash
lufy-ai verify --target /ruta/a/tu/proyecto --scope project --tool opencode
lufy-ai status --target /ruta/a/tu/proyecto --verbose
lufy-ai info --target /ruta/a/tu/proyecto
lufy-ai doctor --target /ruta/a/tu/proyecto
```

Inicializa memoria Obsidian portable cuando quieras que los agentes conserven decisiones, reglas, flows y lessons durables en el repo destino:

```bash
lufy-ai memory init --target /ruta/a/tu/proyecto
lufy-ai memory status --target /ruta/a/tu/proyecto
lufy-ai memory validate --target /ruta/a/tu/proyecto
lufy-ai memory search --target /ruta/a/tu/proyecto "routing"
lufy-ai memory capture --target /ruta/a/tu/proyecto --title "Regla de memoria" --type rule "Las correcciones del usuario a la IA se guardan como memoria durable."
lufy-ai memory connect --target /ruta/a/tu/proyecto regla-de-memoria app-profile
lufy-ai memory index --target /ruta/a/tu/proyecto
```

`memory init` crea `.lufy/memory` con `MEMORY.md`, `inbox/`, `knowledge/`, `maps/_app-profile.md`, `index/backlinks.json` y `.gitignore`. `memory capture` crea o actualiza notas durables, `memory connect` enlaza notas existentes sin backlinks rotos y `memory index` reconstruye el índice local. Por default las notas privadas quedan ignoradas por Git; `install` y `sync` solo gestionan comandos, skills, hooks y templates de memoria, no el contenido de `inbox/` ni `knowledge/`.

## Selección de tool y metodología

Los tool adapters escribibles actuales son `opencode` y `codex`. `opencode` sigue siendo el default efectivo.

```bash
lufy-ai install --target <repo> --tool opencode --yes
lufy-ai install --target <repo> --tool codex --yes
```

Sin `--tool`, el default efectivo sigue siendo `opencode`.

Las metodologías soportadas por configuración son:

- `openspec`;
- `lufy-sdd`;
- `none`.

Se seleccionan por tier:

```bash
lufy-ai install --target <repo> --methodology-tier T3:none --yes
lufy-ai install --target <repo> --methodology-tier T2:openspec/lite --methodology-tier T3:none --yes
lufy-ai install --target <repo> --methodology-tier T2:lufy-sdd/lite --yes
```

Restricciones actuales:

- `T1:none` está bloqueado para comandos mutantes;
- `T2:none` está bloqueado para comandos mutantes;
- `--tool claude-code` está bloqueado para escritura;
- `claude-code` existe solo como dry-run/preview.

## Qué queda instalado

En scope `project`, la CLI gestiona:

- `.opencode/agents`;
- `.opencode/commands`;
- `.opencode/skills`;
- `.agents/skills` cuando `--tool codex`;
- `.codex` cuando `--tool codex`;
- `.opencode/templates`;
- `.opencode/policies`;
- `.opencode/plugins`;
- `.opencode/agent-observatory`;
- `lufy-ia.harness.md`;
- `tui.json`;
- `openspec/` cuando la metodología lo requiere;
- `.lufy/workflows/sdd/` cuando se selecciona `lufy-sdd`;
- `.lufy/managed-state/install-state.json`.

`AGENTS.md` es user-owned. `install` solo agrega la integración LUFY gestionada:

```text
<!-- LUFY:BEGIN codex-harness -->
```

La CLI sigue reconociendo la referencia legacy `@lufy-ia.harness.md` para compatibilidad. `opencode.json` es user-owned/merge-managed. La CLI preserva claves desconocidas y no lo registra como asset completo por hash.

En Codex, `lufy-ai install --tool codex` instala archivos `.codex/agents/*.toml` como custom agents para roles Lufy y habilita `features.multi_agent` en `.codex/config.toml`. La superficie instala `.codex/lufy-agent-mapping.md` y el bloque gestionado de `AGENTS.md` exige que el assistant declare si ejecuta roles en modo `native`, `emulated` o `inline`; cuando tool discovery expone el rol Lufy exacto debe usar `native`, y cuando solo existan roles genéricos usa el mapeo documentado en vez de afirmar que usó subagentes Lufy nativos. En Codex, `@orchestrator` o `@<rol-lufy>` se trata como solicitud de delegación: el assistant padre debe hacer spawn/wait/close del subagente cuando esté disponible, o reportar que la delegación no está disponible antes de continuar; no debe responder como ese rol inline y seguir ejecutando.




## Defaults de `project.yaml`

`lufy-ai init` y `lufy-ai scan --rescan` escriben o preservan estos bloques:

```yaml
memory:
  provider: obsidian
  root: .lufy/memory
  git_policy: ignored
  schema_version: 1
  search: rg
  backlinks_index: .lufy/memory/index/backlinks.json
parallel_execution:
  enabled: true
  strategy: independent_review_slices
  max_parallel_agents: 3
  requires_independent_files: true
  requires_merge_plan: true
  validation_mode: grouped_after_join
```

El paralelismo no es un pipeline fijo: `sdd-router` lo recomienda solo para `review_slices` independientes con archivos separados y plan de merge. Delivery, migraciones, contratos públicos no cerrados y archivos compartidos quedan en ejecución secuencial.

## Manifest y backups

`.lufy/managed-state/install-state.json` usa schema v2 y registra:

- `tool`;
- `methodologyByTier`;
- ownership por asset;
- policy;
- scope;
- hashes SHA-256;
- ancestors cuando corresponde.

Backups se escriben en:

```text
.lufy/managed-state/backups/<timestamp>/
```

Antes de mutaciones reales, `install`, `sync`, `restore` y `uninstall` crean backups cuando corresponde.

## Sync

`sync` reaplica assets gestionados desde el catálogo actual hacia un target instalado. Solo actualiza archivos sin drift local y preserva assets frozen con `pin`.

```bash
lufy-ai sync --target /ruta/a/tu/proyecto --dry-run --yes
lufy-ai sync --target /ruta/a/tu/proyecto --yes
lufy-ai verify --target /ruta/a/tu/proyecto
```

Si necesitas conservar un override local sobre un asset gestionado mientras actualizas el resto del kit:

```bash
lufy-ai pin --target /ruta/a/tu/proyecto --reason "override local" lufy-ia.harness.md
lufy-ai sync --target /ruta/a/tu/proyecto --dry-run --yes
lufy-ai status --target /ruta/a/tu/proyecto --verbose
lufy-ai unpin --target /ruta/a/tu/proyecto lufy-ia.harness.md
```

Mientras el asset esté pinned/frozen, `sync` lo reporta como `pinned-skip` y no avanza sus hashes registrados.

Si un asset no reemplazable tiene drift local, la CLI preserva el archivo y puede generar `<archivo>.lufy-new`.

Para resolver:

```bash
lufy-ai status --target /ruta/a/tu/proyecto --verbose
LUFY_MERGE_TOOL="tu-merge-tool" lufy-ai merge --target /ruta/a/tu/proyecto <path>
lufy-ai merge --target /ruta/a/tu/proyecto --accept-theirs <path>
lufy-ai merge --target /ruta/a/tu/proyecto --accept-ours <path>
```

Después de resolver, `merge` actualiza el manifest, refresca el ancestor seguro y remueve `<archivo>.lufy-new`. `doctor` falla si todavía quedan conflictos pendientes para evitar cerrar un estado parcialmente reconciliado.

## Uninstall y reinstall

`uninstall` remueve solo assets gestionados por Lufy cuando el hash actual coincide con el manifest. Si detecta drift, bloquea y no muta.

Dry-run:

```bash
lufy-ai uninstall --target /ruta/a/tu/proyecto --dry-run
```

Aplicar:

```bash
lufy-ai uninstall --target /ruta/a/tu/proyecto --yes
```

Comportamiento:

- crea backup previo;
- borra assets gestionados sin drift;
- borra ancestors gestionados sin drift;
- remueve `.lufy/managed-state/install-state.json`;
- preserva `.lufy/managed-state/backups`;
- preserva `opencode.json`;
- preserva `AGENTS.md` y remueve solo la integración LUFY gestionada, ya sea el bloque nuevo o la línea legacy `@lufy-ia.harness.md`;
- limpia directorios gestionados que queden vacíos.

Reinstalar:

```bash
lufy-ai install --target /ruta/a/tu/proyecto --tool opencode --yes
lufy-ai verify --target /ruta/a/tu/proyecto --tool opencode --quiet
```

`--keep-state` existe para diagnóstico: conserva `.lufy/managed-state/install-state.json` aunque remueva assets. No es el flujo normal.

## Restore

Listar backups:

```bash
lufy-ai restore --target /ruta/a/tu/proyecto --list
```

Revisar:

```bash
lufy-ai restore --target /ruta/a/tu/proyecto --backup <id-o-ruta> --dry-run
```

Aplicar:

```bash
lufy-ai restore --target /ruta/a/tu/proyecto --backup <id-o-ruta> --yes
```

`restore` valida `targetRoot`, paths seguros y hashes antes de escribir.

## Upgrade del binario

`upgrade` requiere versión fija.

```bash
lufy-ai upgrade --to v0.6.11 --dry-run
lufy-ai upgrade --to v0.6.11
```

Descarga el artifact de la plataforma actual, verifica SHA-256 y reemplaza el ejecutable de forma atómica.

## Desarrollo con clone local

```bash
git clone https://github.com/adrotech/lufy-ai.git /tmp/lufy-ai
cd /tmp/lufy-ai/tools/lufy-cli-go
mkdir -p bin
go build -o bin/lufy-ai ./cmd/lufy-ai
```

Usar el binario local:

```bash
/tmp/lufy-ai/tools/lufy-cli-go/bin/lufy-ai install --target /ruta/a/tu/proyecto --dry-run --yes
/tmp/lufy-ai/tools/lufy-cli-go/bin/lufy-ai install --target /ruta/a/tu/proyecto --yes
/tmp/lufy-ai/tools/lufy-cli-go/bin/lufy-ai verify --target /ruta/a/tu/proyecto
```

El wrapper:

```bash
/tmp/lufy-ai/scripts/install.sh --target /ruta/a/tu/proyecto --yes
```

resuelve primero `tools/lufy-cli-go/bin/lufy-ai` y luego `lufy-ai` en `PATH`. No descarga releases ni usa fallback legacy.

## Troubleshooting

### `command not found: lufy-ai`

```bash
ls -l "$HOME/.local/bin/lufy-ai"
printf '%s\n' "$PATH"
"$HOME/.local/bin/lufy-ai" version
```

Agrega `~/.local/bin` a tu `PATH`.

### `verify` falla después de instalar

Revisa el error exacto. `verify` valida:

- `.lufy/managed-state/install-state.json`;
- estructura crítica;
- hashes SHA-256;
- JSON gestionado;
- bloque LUFY gestionado o referencia legacy `@lufy-ia.harness.md` en `AGENTS.md`;
- adapter esperado si usas `--tool`.

Si falta la referencia en `AGENTS.md`, ejecuta:

```bash
lufy-ai install --target <dir> --yes
```

o agrégala manualmente.

### `uninstall` bloquea por drift

Eso es intencional. Significa que un asset gestionado fue modificado localmente. Revisa:

```bash
lufy-ai status --target <dir> --verbose
lufy-ai doctor --target <dir>
```

Luego decide si conservar el cambio, restaurar desde backup o reinstalar sobre un estado limpio.

### No existe artifact para mi plataforma

Los artifacts objetivo son:

- `darwin/amd64`;
- `darwin/arm64`;
- `linux/amd64`;
- `linux/arm64`;
- `windows/amd64`;
- `windows/arm64`.

Si tu plataforma no está publicada, usa un entorno soportado o compila desde clone local con Go.
