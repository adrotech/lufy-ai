# Instalación de lufy-ai

Esta guía cubre la instalación del binario `lufy-ai`, la configuración de `PATH` por sistema/shell y la instalación de assets en un repositorio destino.

Versión estable actual: `v0.3.0`.

## Requisitos

- Un directorio escribible para el binario, por ejemplo `~/.local/bin` en macOS/Linux/WSL.
- Acceso a una release publicada de GitHub con artifacts y checksums.
- Un repositorio destino donde instalar los assets de OpenCode/OpenSpec.

El bootstrap Bash aplica a entornos Unix-like: macOS, Linux y WSL. En Windows nativo usa el binario manual si la release incluye `lufy-ai_<version>_windows_amd64.zip` o `lufy-ai_<version>_windows_arm64.zip`.

## Instalación rápida con bootstrap

Usa una versión explícita; `latest` existe, pero no es reproducible.

```bash
curl -fsSL https://raw.githubusercontent.com/adrotech/lufy-ai/v0.3.0/scripts/bootstrap.sh -o /tmp/lufy-bootstrap.sh
less /tmp/lufy-bootstrap.sh
bash /tmp/lufy-bootstrap.sh --version v0.3.0 --install-dir "$HOME/.local/bin"
```

Atajo directo solo si ya revisaste el script y aceptas ejecutarlo desde la URL fijada:

```bash
curl -fsSL https://raw.githubusercontent.com/adrotech/lufy-ai/v0.3.0/scripts/bootstrap.sh \
  | bash -s -- --version v0.3.0 --install-dir "$HOME/.local/bin"
```

El bootstrap detecta OS/arch, descarga el artifact `lufy-ai_<version>_<os>_<arch>`, verifica SHA-256 contra los checksums de la misma release e instala solo el binario. No ejecuta `lufy-ai install` contra tu proyecto.

## macOS

macOS usa `zsh` por defecto. Apple Silicon normalmente usa `darwin_arm64`; Intel usa `darwin_amd64`. El bootstrap lo detecta automáticamente.

Instala en `~/.local/bin`:

```bash
bash /tmp/lufy-bootstrap.sh --version v0.3.0 --install-dir "$HOME/.local/bin"
```

Si `~/.local/bin` no está en tu `PATH`, agrega una de estas configuraciones y abre una terminal nueva:

### zsh

```zsh
export PATH="$HOME/.local/bin:$PATH"
```

Guárdalo en `~/.zshrc` si quieres hacerlo persistente.

### bash

```bash
export PATH="$HOME/.local/bin:$PATH"
```

Guárdalo en `~/.bashrc` o `~/.bash_profile` según tu entorno.

### fish

```fish
fish_add_path $HOME/.local/bin
```

Alternativa compatible si no quieres usar `fish_add_path`:

```fish
set -gx PATH $HOME/.local/bin $PATH
```

## Linux

En Linux se recomienda `~/.local/bin` para instalaciones de usuario:

```bash
bash /tmp/lufy-bootstrap.sh --version v0.3.0 --install-dir "$HOME/.local/bin"
```

Configura el `PATH` según tu shell:

### bash

```bash
export PATH="$HOME/.local/bin:$PATH"
```

Normalmente se guarda en `~/.bashrc`.

### zsh

```zsh
export PATH="$HOME/.local/bin:$PATH"
```

Normalmente se guarda en `~/.zshrc`.

### fish

```fish
fish_add_path $HOME/.local/bin
```

Alternativa:

```fish
set -gx PATH $HOME/.local/bin $PATH
```

## Windows

### Windows nativo: PowerShell/cmd

El bootstrap Bash no está pensado para PowerShell/cmd nativos. Si la release incluye `lufy-ai_v0.3.0_windows_amd64.zip`:

1. Descarga el zip y el archivo `lufy-ai_v0.3.0_checksums.txt` desde la release.
2. Verifica el checksum antes de usar el binario.
3. Extrae `lufy-ai.exe` en un directorio de usuario, por ejemplo `%USERPROFILE%\\bin`.
4. Agrega ese directorio al `Path` de usuario desde la configuración de Windows.
5. Abre una nueva terminal PowerShell o cmd.

Verificación de hash en PowerShell:

```powershell
Get-FileHash .\lufy-ai_v0.3.0_windows_amd64.zip -Algorithm SHA256
```

Compara el resultado con la entrada del archivo de checksums.

### WSL

En WSL usa el flujo Linux con bootstrap Bash y `~/.local/bin` dentro de la distribución WSL.

## Verificación post-install

Primero confirma que el binario está disponible:

```bash
lufy-ai version
```

Luego revisa el plan de instalación en el repositorio destino:

```bash
lufy-ai install --target /ruta/a/tu/proyecto --scope project --dry-run --yes --no-engram
```

Aplica la instalación:

```bash
lufy-ai install --target /ruta/a/tu/proyecto --scope project --yes --no-engram
```

`--scope=project` preserva el comportamiento actual. `--scope=global` y `--scope=both` resuelven además la raíz global de OpenCode desde `XDG_CONFIG_HOME` o `HOME`, pero siguen siendo opt-in hasta completar validación de release.

Después de `install`, ejecuta el verificador canónico:

```bash
lufy-ai verify --target /ruta/a/tu/proyecto --scope project --no-engram
```

Para automatización o CI puedes usar salida JSON:

```bash
lufy-ai verify --target /ruta/a/tu/proyecto --no-engram --json
lufy-ai status --target /ruta/a/tu/proyecto --json --verbose
```

Si un asset `no-replace` tiene drift local, install/sync preservan el archivo original y escriben `<archivo>.lufy-new`. Revisa el estado y resuelve manualmente o con un merge tool:

```bash
lufy-ai status --target /ruta/a/tu/proyecto --verbose
LUFY_MERGE_TOOL="tu-merge-tool" lufy-ai merge --target /ruta/a/tu/proyecto tui.json
```

Para descubrir y restaurar backups:

```bash
lufy-ai restore --target /ruta/a/tu/proyecto --list
lufy-ai restore --target /ruta/a/tu/proyecto --backup <id-o-ruta> --dry-run
lufy-ai restore --target /ruta/a/tu/proyecto --backup <id-o-ruta> --yes
```

Para validaciones opt-in de referencias de plugins en `tui.json`/`opencode.json`:

```bash
lufy-ai verify --target /ruta/a/tu/proyecto --no-engram --deep
```

## Actualizar el binario

Usa una versión fija; `upgrade` rechaza `latest` para mantener reproducibilidad:

```bash
lufy-ai upgrade --to v0.3.0
```

Para revisar sin reemplazar el binario:

```bash
lufy-ai upgrade --to v0.3.0 --dry-run
```

`upgrade` descarga el artifact de la plataforma actual, verifica SHA-256 contra `checksums.txt`, extrae el binario y reemplaza el ejecutable actual de forma atómica.

## Flujo con clone local para desarrollo

Usa este camino para contribuir o probar cambios locales antes de publicar una release:

```bash
git clone https://github.com/adrotech/lufy-ai.git /tmp/lufy-ai
cd /tmp/lufy-ai/tools/lufy-cli-go
mkdir -p bin
go build -o bin/lufy-ai ./cmd/lufy-ai
```

El wrapper local `scripts/install.sh` busca primero `tools/lufy-cli-go/bin/lufy-ai` dentro del checkout y luego `lufy-ai` en `PATH`. No descarga releases ni reintroduce fallback legacy.

```bash
/tmp/lufy-ai/scripts/install.sh --target /ruta/a/tu/proyecto --dry-run --yes --no-engram
/tmp/lufy-ai/scripts/install.sh --target /ruta/a/tu/proyecto --yes --no-engram
/tmp/lufy-ai/tools/lufy-cli-go/bin/lufy-ai verify --target /ruta/a/tu/proyecto --no-engram
```

## Troubleshooting

### `command not found: lufy-ai`

1. Verifica que el binario exista:

   ```bash
   ls -l "$HOME/.local/bin/lufy-ai"
   ```

2. Revisa tu `PATH` actual:

   ```bash
   printf '%s\n' "$PATH"
   ```

   En fish:

   ```fish
   printf '%s\n' $PATH
   ```

3. Ejecuta por ruta absoluta para confirmar que el binario funciona:

   ```bash
   "$HOME/.local/bin/lufy-ai" version
   ```

4. Agrega `~/.local/bin` a tu shell. Para bash/zsh usa `export PATH="$HOME/.local/bin:$PATH"`; para fish usa `fish_add_path $HOME/.local/bin` o `set -gx PATH $HOME/.local/bin $PATH`.

### `lufy-ai verify` falla después de instalar

Revisa el error exacto. `verify` valida estructura, estado `.lufy-ai/install-state.json`, hashes SHA-256 y configuración gestionada. Si editaste assets gestionados localmente, puede reportar drift o conflictos que requieren revisión manual.

### No existe artifact para mi plataforma

Los artifacts actuales cubren `darwin/amd64`, `darwin/arm64`, `linux/amd64`, `linux/arm64`, `windows/amd64` y `windows/arm64`. Si tu plataforma no está publicada, usa un entorno soportado o compila desde el clone local con Go.
