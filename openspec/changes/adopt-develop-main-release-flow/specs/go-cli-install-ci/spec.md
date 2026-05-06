## MODIFIED Requirements

### Requirement: CI mínima del instalador Go
El sistema SHALL ejecutar una validación continua mínima para la CLI Go y el wrapper de instalación en GitHub Actions sobre PRs y pushes dirigidos a `develop` y `main`.

#### Scenario: Tests y build Go en CI
- **WHEN** se abre o actualiza un pull request hacia `develop` o `main` que afecta el repositorio
- **THEN** el workflow ejecuta `go test ./...` y `go build ./cmd/lufy-ai` desde `tools/lufy-cli-go/` sin depender de toolchains Node/TS en la raíz

#### Scenario: Pushes protegidos en develop y main
- **WHEN** hay un push en `develop` o `main` que afecta rutas cubiertas por el workflow
- **THEN** el workflow ejecuta el gate mínimo de tests, build, smokes, sanity OpenSpec condicional y `git diff --check`

#### Scenario: Ramas legacy no son base del gate normal
- **WHEN** se configura el trigger del workflow de instalación Go
- **THEN** no usa `development` ni `master` como ramas normales de PR/push para este flujo

#### Scenario: Smoke de instalación en target temporal
- **WHEN** el workflow compila el binario `lufy-ai`
- **THEN** ejecuta un smoke en un directorio temporal que cubre dry-run sin mutaciones, install real, `verify`, idempotencia básica, `backup` y `restore`

#### Scenario: Wrapper Bash validado por CI
- **WHEN** existe `tools/lufy-cli-go/bin/lufy-ai` construido durante el job
- **THEN** el workflow ejecuta `scripts/install.sh` contra un target temporal y confirma que delega en la CLI Go sin usar fallback legacy

#### Scenario: CI portable sin Engram obligatorio
- **WHEN** el workflow ejecuta smokes del instalador
- **THEN** usa `--no-engram` para no depender de que `engram` exista en el runner
